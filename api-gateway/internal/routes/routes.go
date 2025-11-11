package routes

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ibn-network/api-gateway/internal/handlers"
	"github.com/ibn-network/api-gateway/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	batchHandler *handlers.BatchHandler,
	healthHandler *handlers.HealthHandler,
	authMW *middleware.AuthMiddleware,
	rateLimitMW *middleware.RateLimitMiddleware,
	loggerMW *middleware.LoggerMiddleware,
	corsMW *middleware.CORSMiddleware,
	recoveryMW *middleware.RecoveryMiddleware,
	tracingMW *middleware.TracingMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(recoveryMW.Recover)
	r.Use(corsMW.Handler)
	r.Use(loggerMW.Log)
	r.Use(tracingMW.Trace)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Compress(5))

	// Health endpoints (no auth required)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)
	r.Get("/live", healthHandler.Live)

	// Metrics endpoint (no auth required)
	r.Handle("/metrics", handlers.MetricsHandler())

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Batch routes
		r.Route("/batches", func(r chi.Router) {
			// Public route - no auth required
			r.Get("/{id}", batchHandler.GetBatchInfo)

			// Protected routes - require authentication and rate limiting
			r.Group(func(r chi.Router) {
				r.Use(authMW.Authenticate)
				r.Use(rateLimitMW.Limit)

				r.Post("/", batchHandler.CreateBatch)
				r.Post("/{id}/verify", batchHandler.VerifyBatch)
				r.Patch("/{id}/status", batchHandler.UpdateBatchStatus)
			})
		})
	})

	return r
}

