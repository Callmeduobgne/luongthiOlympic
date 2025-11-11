package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ibn-network/api-gateway/internal/config"
)

// CORSMiddleware provides CORS middleware
type CORSMiddleware struct {
	config *config.CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(cfg *config.CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: cfg,
	}
}

// Handler applies CORS headers
func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if m.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Set other CORS headers
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(m.config.ExposedHeaders, ", "))
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", m.config.MaxAge))

		if m.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if an origin is in the allowed list
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range m.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

