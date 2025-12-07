package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// APIKeyAuth middleware validates API key for internal service-to-service authentication
func APIKeyAuth(expectedAPIKey string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}

			// Get API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Try Authorization header
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					apiKey = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if apiKey == "" || apiKey != expectedAPIKey {
				logger.Warn("Unauthorized API key attempt",
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"unauthorized","message":"Invalid or missing API key"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

