package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// AdminOnly middleware ensures only admin users can access the route
func AdminOnly(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get role from context (set by auth middleware)
			role, ok := r.Context().Value("role").(string)
			if !ok || role != "admin" {
				logger.Warn("Admin access denied",
					zap.String("path", r.URL.Path),
					zap.String("role", role),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(models.NewErrorResponse(
					models.ErrCodeForbidden,
					"Admin privileges required",
					nil,
				))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

