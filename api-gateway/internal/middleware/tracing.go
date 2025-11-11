package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware provides OpenTelemetry tracing middleware
type TracingMiddleware struct {
	tracer trace.Tracer
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{
		tracer: otel.Tracer("http-server"),
	}
}

// Trace traces HTTP requests
func (m *TracingMiddleware) Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := m.tracer.Start(r.Context(), r.URL.Path)
		defer span.End()

		// Set attributes
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.String("http.remote_addr", r.RemoteAddr),
		)

		// Wrap response writer to capture status code
		wrapped := &tracingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Set status code attribute
		span.SetAttributes(attribute.Int("http.status_code", wrapped.statusCode))

		// Set span status based on HTTP status
		if wrapped.statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	})
}

// tracingResponseWriter wraps http.ResponseWriter to capture status code
type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *tracingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

