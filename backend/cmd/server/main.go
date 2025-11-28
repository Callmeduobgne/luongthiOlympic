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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ibn-network/backend/internal/config"
	authHandler "github.com/ibn-network/backend/internal/handlers/auth"
	teatraceHandler "github.com/ibn-network/backend/internal/handlers/teatrace"
	"github.com/ibn-network/backend/internal/infrastructure/cache"
	"github.com/ibn-network/backend/internal/infrastructure/database"
	"github.com/ibn-network/backend/internal/infrastructure/gateway"
	authMiddleware "github.com/ibn-network/backend/internal/middleware"
	"github.com/ibn-network/backend/internal/services/auth"
	teatraceService "github.com/ibn-network/backend/internal/services/teatrace"
	"github.com/ibn-network/backend/internal/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// @title IBN Backend API
// @version 1.0
// @description Backend API for IBN Network - Blockchain traceability system
// @termsOfService http://swagger.io/terms/
// @contact.name IBN Network Team
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9090
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := utils.NewLogger(&cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting IBN Backend",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
		zap.String("env", cfg.Server.Env),
	)

	// Initialize database pool
	dbPool, err := database.NewPool(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database pool", zap.Error(err))
	}
	defer dbPool.Close()

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(&cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis cache", zap.Error(err))
	}
	defer redisCache.Close()

	// Initialize memory cache (L1)
	memoryCache := cache.NewMemoryCache(&cache.MemoryCacheConfig{
		DefaultTTL:      5 * time.Minute,
		CleanupInterval: 10 * time.Minute,
		MaxSize:         100 * 1024 * 1024, // 100MB
	}, logger)

	// Initialize multi-layer cache
	multiCache := cache.NewMultiLayerCache(memoryCache, redisCache, dbPool.Primary(), logger)

	// Initialize repositories
	authRepo := auth.NewRepository(dbPool.Primary())

	// Initialize Gateway client
	gatewayClient := gateway.NewClient(&gateway.Config{
		BaseURL: cfg.Gateway.BaseURL,
		APIKey:  cfg.Gateway.APIKey,
		Timeout: cfg.Gateway.Timeout,
		Logger:  logger,
	})

	// Initialize services
	authService := auth.NewService(authRepo, multiCache, &cfg.JWT, logger)
	teatraceService := teatraceService.NewService(gatewayClient, logger)

	// Initialize handlers
	authHandlerInstance := authHandler.NewHandler(authService, logger)
	teatraceHandlerInstance := teatraceHandler.NewHandler(teatraceService, logger)

	// Setup routes
	router := setupRoutes(cfg, authHandlerInstance, teatraceHandlerInstance, authService, logger)

	// Create HTTP server
	addr := cfg.Server.Address()
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("IBN Backend listening", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down IBN Backend...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server exited gracefully")
	}
}

// setupRoutes configures all HTTP routes
func setupRoutes(
	cfg *config.Config,
	authHandler *authHandler.Handler,
	teatraceHandler *teatraceHandler.Handler,
	authService *auth.Service,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS - Simple implementation
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-API-Key, Idempotency-Key")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoints
	healthHandler := func(status, message string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if r.Method == http.MethodHead {
				return
			}
			fmt.Fprintf(w, `{"status":"%s","service":"%s"}`, status, message)
		}
	}

	r.Method(http.MethodGet, "/health", healthHandler("healthy", "ibn-backend"))
	r.Method(http.MethodHead, "/health", healthHandler("healthy", "ibn-backend"))

	r.Method(http.MethodGet, "/ready", healthHandler("ready", "ibn-backend"))
	r.Method(http.MethodHead, "/ready", healthHandler("ready", "ibn-backend"))

	// Metrics endpoint for Prometheus scraping
	r.Handle("/metrics", promhttp.Handler())

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			// Public auth routes
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				authMW := authMiddleware.NewAuthMiddleware(authService, logger)
				r.Use(authMW.Authenticate)

				// Profile
				r.Get("/profile", authHandler.GetProfile)
				r.Post("/profile/avatar", authHandler.UploadAvatar)

				// API Keys
				r.Post("/api-keys", authHandler.CreateAPIKey)
			})
		})

		// Tea Traceability routes (public)
		r.Route("/teatrace", func(r chi.Router) {
			r.Post("/verify-by-hash", teatraceHandler.VerifyByHash)
		})

		// TODO: Add other routes (blockchain, chaincode, metrics, etc.)
	})

	return r
}
