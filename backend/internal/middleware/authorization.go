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
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/authorization"
	"go.uber.org/zap"
)

// AuthorizationMiddleware handles authorization checks
type AuthorizationMiddleware struct {
	authzService *authorization.Service
	logger       *zap.Logger
}

// NewAuthorizationMiddleware creates a new authorization middleware
func NewAuthorizationMiddleware(authzService *authorization.Service, logger *zap.Logger) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		authzService: authzService,
		logger:       logger,
	}
}

// RequirePermission creates a middleware that requires a specific permission
func (m *AuthorizationMiddleware) RequirePermission(resource, action, scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by auth middleware)
			userIDVal := r.Context().Value("user_id")
			if userIDVal == nil {
				m.respondError(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			userIDStr, ok := userIDVal.(string)
			if !ok {
				m.respondError(w, http.StatusUnauthorized, "Invalid user ID")
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				m.respondError(w, http.StatusUnauthorized, "Invalid user ID format")
				return
			}

			// Extract resource ID from URL if needed
			resourceID := m.extractResourceID(r, resource)

			// Get IP address
			ipAddress := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ipAddress = strings.Split(forwarded, ",")[0]
			}

			// Build authorization request
			authzReq := &authorization.AuthorizeRequest{
				UserID:      userID,
				Resource:    resource,
				Action:      action,
				Scope:       scope,
				ResourceID:  resourceID,
				IPAddress:   ipAddress,
				ResourceAttrs: make(map[string]interface{}),
			}

			// Check authorization
			allowed, err := m.authzService.Authorize(r.Context(), authzReq)
			if err != nil {
				m.logger.Error("Authorization check failed", zap.Error(err))
				m.respondError(w, http.StatusInternalServerError, "Authorization check failed")
				return
			}

			if !allowed {
				m.logger.Warn("Access denied",
					zap.String("user_id", userIDStr),
					zap.String("resource", resource),
					zap.String("action", action),
					zap.String("scope", scope),
				)
				m.respondError(w, http.StatusForbidden, "Access denied")
				return
			}

			// Add authorization info to context
			ctx := context.WithValue(r.Context(), "authorized_resource", resource)
			ctx = context.WithValue(ctx, "authorized_action", action)
			ctx = context.WithValue(ctx, "authorized_scope", scope)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAnyPermission creates a middleware that requires any of the specified permissions
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions []Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context
			userIDVal := r.Context().Value("user_id")
			if userIDVal == nil {
				m.respondError(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			userIDStr, ok := userIDVal.(string)
			if !ok {
				m.respondError(w, http.StatusUnauthorized, "Invalid user ID")
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				m.respondError(w, http.StatusUnauthorized, "Invalid user ID format")
				return
			}

			// Try each permission
			for _, perm := range permissions {
				resourceID := m.extractResourceID(r, perm.Resource)

				authzReq := &authorization.AuthorizeRequest{
					UserID:      userID,
					Resource:    perm.Resource,
					Action:      perm.Action,
					Scope:       perm.Scope,
					ResourceID:  resourceID,
					IPAddress:   r.RemoteAddr,
					ResourceAttrs: make(map[string]interface{}),
				}

				allowed, err := m.authzService.Authorize(r.Context(), authzReq)
				if err != nil {
					m.logger.Warn("Authorization check failed", zap.Error(err))
					continue
				}

				if allowed {
					// Permission granted, continue
					ctx := context.WithValue(r.Context(), "authorized_resource", perm.Resource)
					ctx = context.WithValue(ctx, "authorized_action", perm.Action)
					ctx = context.WithValue(ctx, "authorized_scope", perm.Scope)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// No permission granted
			m.logger.Warn("Access denied - no matching permission",
				zap.String("user_id", userIDStr),
				zap.Int("permissions_checked", len(permissions)),
			)
			m.respondError(w, http.StatusForbidden, "Access denied")
		})
	}
}

// RequireRole creates a middleware that requires one of the specified roles
func (m *AuthorizationMiddleware) RequireRole(roleName string, additionalRoles ...string) func(http.Handler) http.Handler {
	allowedRoles := append([]string{roleName}, additionalRoles...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user roles from context (set by auth middleware)
			rolesVal := r.Context().Value("user_roles")
			if rolesVal == nil {
				// Fallback to single role
				roleVal := r.Context().Value("user_role")
				if roleVal == nil {
					m.respondError(w, http.StatusUnauthorized, "User role not found")
					return
				}
				role, ok := roleVal.(string)
				if !ok || !containsRole(role, allowedRoles) {
					m.respondError(w, http.StatusForbidden, "Insufficient permissions")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			roles, ok := rolesVal.([]string)
			if !ok {
				m.respondError(w, http.StatusUnauthorized, "Invalid user roles")
				return
			}

			// Check if user has required role
			for _, role := range roles {
				if containsRole(role, allowedRoles) {
					next.ServeHTTP(w, r)
					return
				}
			}

			m.respondError(w, http.StatusForbidden, "Insufficient permissions")
		})
	}
}

func containsRole(role string, allowed []string) bool {
	for _, r := range allowed {
		if role == r {
			return true
		}
	}
	return false
}

// Permission represents a permission requirement
type Permission struct {
	Resource string
	Action   string
	Scope    string
}

// extractResourceID extracts resource ID from URL path
func (m *AuthorizationMiddleware) extractResourceID(r *http.Request, resource string) *string {
	// Try to extract ID from URL path
	// Pattern: /api/v1/{resource}/{id}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	
	// Look for resource in path
	for i, part := range pathParts {
		if part == resource && i+1 < len(pathParts) {
			id := pathParts[i+1]
			// Validate UUID format
			if _, err := uuid.Parse(id); err == nil {
				return &id
			}
		}
	}

	// Try chi URL param
	if id := r.URL.Query().Get("id"); id != "" {
		if _, err := uuid.Parse(id); err == nil {
			return &id
		}
	}

	return nil
}

// respondError sends an error response
func (m *AuthorizationMiddleware) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

