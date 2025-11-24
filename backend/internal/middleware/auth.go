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

	"github.com/ibn-network/backend/internal/services/auth"
	"go.uber.org/zap"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	authService *auth.Service
	logger      *zap.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *auth.Service, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// Authenticate middleware checks JWT token or API key
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try API key first
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			user, err := m.authService.VerifyAPIKey(r.Context(), apiKey)
			if err != nil {
				m.respondError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), "user_id", user.ID)
			ctx = context.WithValue(ctx, "user_email", user.Email)
			ctx = context.WithValue(ctx, "user_role", user.Role)
			if user.MSPID != nil {
				ctx = context.WithValue(ctx, "user_msp_id", *user.MSPID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Try JWT token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Verify token
		claims, err := m.authService.VerifyAccessToken(tokenString)
		if err != nil {
			m.logger.Warn("Token verification failed", zap.Error(err))
			m.respondError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "jwt_token", tokenString) // Add JWT token for Gateway
		if claims.MSPID != nil {
			ctx = context.WithValue(ctx, "user_msp_id", *claims.MSPID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware checks if user has required role
func (m *AuthMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("user_role").(string)
			if !ok {
				m.respondError(w, http.StatusForbidden, "Access denied")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.logger.Warn("Access denied due to insufficient role",
					zap.String("required_roles", strings.Join(roles, ",")),
					zap.String("user_role", userRole),
				)
				m.respondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireMSP middleware checks if user belongs to required MSP
func (m *AuthMiddleware) RequireMSP(mspIDs ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userMSPID, ok := r.Context().Value("user_msp_id").(string)
			if !ok {
				m.respondError(w, http.StatusForbidden, "MSP ID required")
				return
			}

			// Check if user has required MSP
			hasMSP := false
			for _, mspID := range mspIDs {
				if userMSPID == mspID {
					hasMSP = true
					break
				}
			}

			if !hasMSP {
				m.logger.Warn("Access denied due to wrong MSP",
					zap.String("required_msp", strings.Join(mspIDs, ",")),
					zap.String("user_msp", userMSPID),
				)
				m.respondError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Optional middleware makes authentication optional
func (m *AuthMiddleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to authenticate but don't fail if no credentials
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			user, err := m.authService.VerifyAPIKey(r.Context(), apiKey)
			if err == nil {
				ctx := context.WithValue(r.Context(), "user_id", user.ID)
				ctx = context.WithValue(ctx, "user_email", user.Email)
				ctx = context.WithValue(ctx, "user_role", user.Role)
				if user.MSPID != nil {
					ctx = context.WithValue(ctx, "user_msp_id", *user.MSPID)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				claims, err := m.authService.VerifyAccessToken(parts[1])
				if err == nil {
					ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
					ctx = context.WithValue(ctx, "user_email", claims.Email)
					ctx = context.WithValue(ctx, "user_role", claims.Role)
					if claims.MSPID != nil {
						ctx = context.WithValue(ctx, "user_msp_id", *claims.MSPID)
					}
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// No valid credentials, continue without authentication
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}


