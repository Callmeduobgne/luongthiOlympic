// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ibn-network/api-gateway/internal/services/audit"
	"go.uber.org/zap"
)

// AuditMiddleware logs all API requests to audit log
type AuditMiddleware struct {
	auditService *audit.Service
	logger       *zap.Logger
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditService *audit.Service, logger *zap.Logger) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
		logger:       logger,
	}
}

// Audit logs the request to audit log
func (m *AuditMiddleware) Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip audit logging for health checks, metrics, and WebSocket connections
		// WebSocket connections are long-lived and generate too many audit logs
		if r.URL.Path == "/health" || r.URL.Path == "/ready" || r.URL.Path == "/live" || r.URL.Path == "/metrics" || SkipAuditForWebSocket(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Capture request body if needed (for POST/PUT/PATCH)
		var requestBody []byte
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if r.Body != nil {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
		}

		// Create response writer wrapper to capture status code
		rw := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Execute next handler
		start := time.Now()
		next.ServeHTTP(rw, r)
		duration := time.Since(start)

		// Extract user info from context
		userID := ""
		apiKeyID := ""
		if userIDVal := r.Context().Value("userID"); userIDVal != nil {
			if id, ok := userIDVal.(string); ok {
				userID = id
			}
		}
		if apiKeyIDVal := r.Context().Value("apiKeyID"); apiKeyIDVal != nil {
			if id, ok := apiKeyIDVal.(string); ok {
				apiKeyID = id
			}
		}

		// Extract transaction ID from context if available
		txID := ""
		if txIDVal := r.Context().Value("txID"); txIDVal != nil {
			if id, ok := txIDVal.(string); ok {
				txID = id
			}
		}

		// Determine action from method and path
		action := m.determineAction(r.Method, r.URL.Path)

		// Determine resource type and ID from path
		resourceType, resourceID := m.extractResourceInfo(r.URL.Path)

		// Build details
		details := map[string]interface{}{
			"method":   r.Method,
			"path":     r.URL.Path,
			"query":    r.URL.RawQuery,
			"duration": duration.Milliseconds(),
		}

		// Add request body for non-GET requests (truncated if too long)
		if len(requestBody) > 0 && len(requestBody) < 1000 {
			details["requestBody"] = string(requestBody)
		}

		// Log to audit service (async to avoid blocking)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			logReq := &audit.LogRequest{
				UserID:       userID,
				ApiKeyID:     apiKeyID,
				Action:       action,
				ResourceType: resourceType,
				ResourceID:   resourceID,
				TxID:         txID,
				Status:       http.StatusText(rw.statusCode),
				Details:      details,
				IpAddress:    r.RemoteAddr,
				UserAgent:    r.UserAgent(),
			}

			if err := m.auditService.CreateLog(ctx, logReq); err != nil {
				m.logger.Warn("Failed to create audit log",
					zap.String("action", action),
					zap.Error(err),
				)
			}
		}()
	})
}

// determineAction determines the action from HTTP method and path
func (m *AuditMiddleware) determineAction(method, path string) string {
	// Map HTTP methods to actions
	actionMap := map[string]string{
		http.MethodGet:    "READ",
		http.MethodPost:   "CREATE",
		http.MethodPut:    "UPDATE",
		http.MethodPatch:  "UPDATE",
		http.MethodDelete: "DELETE",
	}

	baseAction := actionMap[method]
	if baseAction == "" {
		baseAction = method
	}

	// Add resource type to action
	if path != "" {
		// Extract resource from path (e.g., /api/v1/transactions -> transactions)
		parts := []string{}
		for _, part := range strings.Split(path, "/") {
			if part != "" && part != "api" && part != "v1" {
				parts = append(parts, part)
			}
		}
		if len(parts) > 0 {
			baseAction = fmt.Sprintf("%s_%s", baseAction, strings.ToUpper(parts[0]))
		}
	}

	return baseAction
}

// extractResourceInfo extracts resource type and ID from path
func (m *AuditMiddleware) extractResourceInfo(path string) (resourceType, resourceID string) {
	parts := []string{}
	for _, part := range strings.Split(path, "/") {
		if part != "" && part != "api" && part != "v1" {
			parts = append(parts, part)
		}
	}

	if len(parts) == 0 {
		return "", ""
	}

	resourceType = parts[0]

	// Extract resource ID (usually the last part if it's a UUID or number)
	if len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		// Check if it looks like an ID (UUID or number)
		if len(lastPart) > 10 || isNumeric(lastPart) {
			resourceID = lastPart
		}
	}

	return resourceType, resourceID
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// auditResponseWriter wraps http.ResponseWriter to capture status code
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *auditResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker for WebSocket support
func (rw *auditResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

