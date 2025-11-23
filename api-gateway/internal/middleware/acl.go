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
	"fmt"
	"net/http"
	"strings"

	"github.com/ibn-network/api-gateway/internal/models"
	aclservice "github.com/ibn-network/api-gateway/internal/services/acl"
	"go.uber.org/zap"
)

// ACLMiddleware provides permission checking middleware
type ACLMiddleware struct {
	aclService *aclservice.Service
	logger     *zap.Logger
}

// NewACLMiddleware creates a new ACL middleware
func NewACLMiddleware(aclService *aclservice.Service, logger *zap.Logger) *ACLMiddleware {
	return &ACLMiddleware{
		aclService: aclService,
		logger:     logger,
	}
}

// RequirePermission creates a middleware that checks if the user has permission
// to perform an action on a resource
func (m *ACLMiddleware) RequirePermission(resourceType, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by auth middleware)
			userID, ok := r.Context().Value("userID").(string)
			if !ok || userID == "" {
				m.logger.Warn("ACL check failed: user ID not found in context")
				respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
					models.ErrCodeUnauthorized,
					"Authentication required",
					nil,
				))
				return
			}

			// Determine resource from request
			resource := m.getResourceFromRequest(r, resourceType)

			// Check permission
			req := &models.CheckPermissionRequest{
				UserID:       userID,
				ResourceType: resourceType,
				Resource:     resource,
				Action:       action,
			}

			result, err := m.aclService.CheckPermission(r.Context(), req)
			if err != nil {
				m.logger.Error("Failed to check permission", zap.Error(err))
				respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
					models.ErrCodeInternalError,
					"Failed to check permission",
					err.Error(),
				))
				return
			}

			if !result.Allowed {
				m.logger.Warn("Permission denied",
					zap.String("userId", userID),
					zap.String("resourceType", resourceType),
					zap.String("resource", resource),
					zap.String("action", action),
					zap.String("reason", result.Reason),
				)
				respondJSON(w, http.StatusForbidden, models.NewErrorResponse(
					models.ErrCodeForbidden,
					fmt.Sprintf("Permission denied: %s", result.Reason),
					result,
				))
				return
			}

			// Permission granted, continue
			next.ServeHTTP(w, r)
		})
	}
}

// getResourceFromRequest extracts the resource identifier from the request
func (m *ACLMiddleware) getResourceFromRequest(r *http.Request, resourceType string) string {
	switch resourceType {
	case "channel":
		// Extract channel name from path (e.g., /api/v1/channels/{name})
		if strings.Contains(r.URL.Path, "/channels/") {
			parts := strings.Split(r.URL.Path, "/channels/")
			if len(parts) > 1 {
				channelName := strings.Split(parts[1], "/")[0]
				if channelName != "" {
					return channelName
				}
			}
		}
		// Fallback: use path
		return r.URL.Path

	case "chaincode":
		// Extract chaincode name from path (e.g., /api/v1/channels/{channel}/chaincodes/{name})
		if strings.Contains(r.URL.Path, "/chaincodes/") {
			parts := strings.Split(r.URL.Path, "/chaincodes/")
			if len(parts) > 1 {
				chaincodeName := strings.Split(parts[1], "/")[0]
				if chaincodeName != "" {
					return chaincodeName
				}
			}
		}
		// Fallback: use path
		return r.URL.Path

	case "endpoint":
		// Use the full path as resource
		return r.URL.Path

	default:
		// Default: use path
		return r.URL.Path
	}
}

// RequireAction creates a middleware that checks permission based on HTTP method
// Maps HTTP methods to actions: GET/HEAD -> read, POST/PUT/PATCH -> write, DELETE -> admin
func (m *ACLMiddleware) RequireAction(resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Map HTTP method to action
			action := m.getActionFromMethod(r.Method)

			// Use RequirePermission with the determined action
			m.RequirePermission(resourceType, action)(next).ServeHTTP(w, r)
		})
	}
}

// getActionFromMethod maps HTTP methods to ACL actions
func (m *ACLMiddleware) getActionFromMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET", "HEAD":
		return "read"
	case "POST", "PUT", "PATCH":
		return "write"
	case "DELETE":
		return "admin"
	default:
		return "read"
	}
}

