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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/utils"
	"go.uber.org/zap"
)

// AuthService interface for validating API keys
type AuthService interface {
	ValidateAPIKey(ctx context.Context, apiKey string) (*models.UserInfo, error)
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	jwtSecret  []byte
	issuer     string
	logger     *zap.Logger
	authService AuthService // Optional: for API key validation
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(cfg *config.JWTConfig, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(cfg.Secret),
		issuer:    cfg.Issuer,
		logger:    logger,
	}
}

// SetAuthService sets the auth service for API key validation
func (m *AuthMiddleware) SetAuthService(authService AuthService) {
	m.authService = authService
}

// Authenticate validates JWT token or API key
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userInfo *models.UserInfo
		var err error

		// Try API key first (X-API-Key header)
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			m.logger.Info("API key found in request",
				zap.String("path", r.URL.Path),
				zap.String("key_preview", apiKey[:20]+"..."),
				zap.Bool("auth_service_set", m.authService != nil),
			)
			if m.authService != nil {
				userInfo, err = m.authService.ValidateAPIKey(r.Context(), apiKey)
				if err == nil && userInfo != nil {
					m.logger.Info("API key validated successfully",
						zap.String("user_id", userInfo.ID),
						zap.String("email", userInfo.Email),
					)
					// API key is valid, add user info to context
					ctx := context.WithValue(r.Context(), "userID", userInfo.ID)
					ctx = context.WithValue(ctx, "email", userInfo.Email)
					ctx = context.WithValue(ctx, "mspId", userInfo.MSPId)
					ctx = context.WithValue(ctx, "role", userInfo.Role)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				m.logger.Warn("API key validation failed",
					zap.String("path", r.URL.Path),
					zap.Error(err),
				)
			} else {
				m.logger.Warn("API key found but authService is nil",
					zap.String("path", r.URL.Path),
				)
			}
			// If API key validation failed, continue to try JWT
		} else {
			// Debug: Log all headers to see what we receive
			headerList := []string{}
			for k, v := range r.Header {
				headerList = append(headerList, fmt.Sprintf("%s=%v", k, v))
			}
			m.logger.Info("No API key in request - checking headers",
				zap.String("path", r.URL.Path),
				zap.String("headers", strings.Join(headerList, "; ")),
			)
		}

		// Try JWT token from multiple sources:
		// 1. Authorization header (Bearer <token>)
		// 2. Query parameter (?token=...)
		var tokenString string
		
		// First, try Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Check if it starts with "Bearer "
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
		
		// If no token from header, try query parameter (for WebSocket)
		if tokenString == "" {
			tokenString = r.URL.Query().Get("token")
			if tokenString != "" {
				m.logger.Debug("Token found in query parameter",
					zap.String("path", r.URL.Path),
					zap.String("query", r.URL.RawQuery),
				)
			}
		}

		// If still no token, return error
		if tokenString == "" {
			m.logger.Warn("No token found in header or query parameter",
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.String("auth_header", authHeader),
			)
			// For WebSocket requests, return plain HTTP error (don't write JSON)
			if r.Header.Get("Upgrade") == "websocket" {
				http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
				return
			}
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Missing authorization header, API key, or token query parameter",
				nil,
			))
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid or expired token",
				nil,
			))
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid token claims",
				nil,
			))
			return
		}

		// Validate issuer (accept both backend and gateway issuers for service-to-service communication)
		if iss, ok := claims["iss"].(string); ok {
			// Accept both "ibn-network" (backend) and "ibn-api-gateway" (gateway) issuers
			if iss != m.issuer && iss != "ibn-network" {
				respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
					models.ErrCodeUnauthorized,
					"Invalid token issuer",
					nil,
				))
				return
			}
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "userID", claims["userId"])
		ctx = context.WithValue(ctx, "email", claims["email"])
		ctx = context.WithValue(ctx, "mspId", claims["mspId"])
		ctx = context.WithValue(ctx, "role", claims["role"])

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth validates token if present, but allows request to continue if not
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// If header is present, validate it
		m.Authenticate(next).ServeHTTP(w, r)
	})
}

// GenerateToken generates a new JWT token
func (m *AuthMiddleware) GenerateToken(userID, email, mspID, role string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"mspId":  mspID,
		"role":   role,
		"iss":    m.issuer,
		"iat":    now.Unix(),
		"exp":    now.Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token
func (m *AuthMiddleware) GenerateRefreshToken() (string, error) {
	return utils.GenerateRandomString(32)
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		// Use json encoder
		json.NewEncoder(w).Encode(payload)
	}
}

