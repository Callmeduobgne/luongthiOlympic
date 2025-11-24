// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ibn-network/backend/internal/infrastructure/cache"
	"go.uber.org/zap"
)

// RateLimitMiddleware provides rate limiting middleware
type RateLimitMiddleware struct {
	cache  *cache.MultiLayerCache
	logger *zap.Logger
	enabled bool
}

// NewRateLimitMiddleware creates a new rate limit middleware
func NewRateLimitMiddleware(cache *cache.MultiLayerCache, enabled bool, logger *zap.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		cache:   cache,
		logger:  logger,
		enabled: enabled,
	}
}

// LimitWithConfig applies rate limiting with custom requests and window
func (m *RateLimitMiddleware) LimitWithConfig(next http.Handler, requests int, window time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Get identifier (IP address)
		identifier := m.getIdentifier(r)
		key := fmt.Sprintf("rate_limit:%s", identifier)

		// Check rate limit using MultiLayerCache
		allowed, err := m.cache.CheckRateLimit(r.Context(), key, requests, window)
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
				zap.Int("requests", requests),
				zap.Duration("window", window),
			)

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", requests))
			w.Header().Set("X-RateLimit-Window", window.String())
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(fmt.Sprintf(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Rate limit exceeded. Max %d requests per %s"}}`, requests, window)))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Limit applies rate limiting (default: 10 requests per minute)
func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return m.LimitWithConfig(next, 10, 1*time.Minute)
}

// getIdentifier extracts identifier from request (IP address)
func (m *RateLimitMiddleware) getIdentifier(r *http.Request) string {
	// Check X-Forwarded-For header (from proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take first IP if multiple
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

