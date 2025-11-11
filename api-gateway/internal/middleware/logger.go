package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoggerMiddleware provides request logging middleware
type LoggerMiddleware struct {
	logger *zap.Logger
}

// NewLoggerMiddleware creates a new logger middleware
func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

// Log logs HTTP requests with correlation ID
func (m *LoggerMiddleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate correlation ID
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add correlation ID to response headers
		w.Header().Set("X-Correlation-ID", correlationID)

		// Create wrapped response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Add correlation ID to context
		ctx := r.Context()
		ctx = withCorrelationID(ctx, correlationID)

		// Process request
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Log request
		duration := time.Since(start)
		m.logger.Info("HTTP request",
			zap.String("correlation_id", correlationID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// withCorrelationID adds correlation ID to context
func withCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, "correlationID", correlationID)
}

// GetCorrelationID retrieves correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value("correlationID").(string); ok {
		return id
	}
	return ""
}

