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
	"time"

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	"go.uber.org/zap"
)

// RateLimitMiddleware provides rate limiting middleware
type RateLimitMiddleware struct {
	cache   *cache.Service
	config  *config.RateLimitConfig
	logger  *zap.Logger
	enabled bool
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(cache *cache.Service, cfg *config.RateLimitConfig, logger *zap.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cache:   cache,
		config:  cfg,
		logger:  logger,
		enabled: cfg.Enabled,
	}
}

// Limit applies rate limiting based on IP address or API key
func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return m.LimitWithConfig(next, m.config.Requests, m.config.Window)
}

// LimitWithConfig applies rate limiting with custom requests and window
func (m *RateLimitMiddleware) LimitWithConfig(next http.Handler, requests int, window time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get identifier (API key or IP address)
		identifier := m.getIdentifier(r)
		key := fmt.Sprintf("rate_limit:%s", identifier)

		// Check rate limit
		allowed, err := m.cache.CheckRateLimit(
			r.Context(),
			key,
			requests,
			window,
		)

		if err != nil {
			m.logger.Error("Failed to check rate limit", zap.Error(err))
			// Allow request on error to avoid blocking legitimate traffic
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			m.logger.Warn("Rate limit exceeded",
				zap.String("identifier", identifier),
				zap.String("path", r.URL.Path),
			)

			respondJSON(w, http.StatusTooManyRequests, models.NewErrorResponse(
				models.ErrCodeRateLimitExceeded,
				fmt.Sprintf("Rate limit exceeded. Max %d requests per %s", requests, window),
				nil,
			))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LimitLogin applies stricter rate limiting for login endpoints (anti-brute force)
// Default: 5 attempts per 15 minutes per IP
func (m *RateLimitMiddleware) LimitLogin(next http.Handler) http.Handler {
	return m.LimitWithConfig(next, 5, 15*time.Minute)
}

// getIdentifier extracts identifier from request (API key or IP)
func (m *RateLimitMiddleware) getIdentifier(r *http.Request) string {
	// Try API key first
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("apikey:%s", apiKey)
	}

	// Try user ID from context (from JWT)
	if userID, ok := r.Context().Value("userID").(string); ok && userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}

	// Fall back to IP address
	return fmt.Sprintf("ip:%s", getClientIP(r))
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Use RemoteAddr
	return r.RemoteAddr
}

