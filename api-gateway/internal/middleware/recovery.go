package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// RecoveryMiddleware provides panic recovery middleware
type RecoveryMiddleware struct {
	logger *zap.Logger
}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// Recover recovers from panics and returns 500 error
func (m *RecoveryMiddleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				m.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("stack", string(debug.Stack())),
				)

				// Return 500 error
				respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
					models.ErrCodeInternalServer,
					"Internal server error",
					fmt.Sprintf("%v", err),
				))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

