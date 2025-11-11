package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/handlers"
	"github.com/ibn-network/api-gateway/internal/middleware"
	"github.com/ibn-network/api-gateway/internal/routes"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"github.com/ibn-network/api-gateway/internal/utils"
	"go.uber.org/zap"
)

// @title IBN API Gateway
// @version 1.0
// @description Production-ready API Gateway for Hyperledger Fabric IBN Network - Tea Traceability
// @termsOfService http://swagger.io/terms/

// @contact.name IBN Network Team
// @contact.email support@ibn.vn

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

const version = "1.0.0"

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

	logger.Info("Starting IBN API Gateway",
		zap.String("version", version),
		zap.String("environment", cfg.Server.Env),
	)

	// Initialize database
	db, err := config.NewPostgresPool(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL")

	// Initialize Redis
	redisService, err := cache.NewService(&cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisService.Close()

	// Initialize Fabric Gateway
	fabricGateway, err := fabric.NewGatewayService(&cfg.Fabric, &cfg.CircuitBreaker, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Fabric Gateway", zap.Error(err))
	}
	defer fabricGateway.Close()

	// Initialize contract service
	contractService := fabric.NewContractService(fabricGateway)

	// Initialize middleware
	authMW := middleware.NewAuthMiddleware(&cfg.JWT, logger)
	rateLimitMW := middleware.NewRateLimitMiddleware(redisService, &cfg.RateLimit, logger)
	loggerMW := middleware.NewLoggerMiddleware(logger)
	corsMW := middleware.NewCORSMiddleware(&cfg.CORS)
	recoveryMW := middleware.NewRecoveryMiddleware(logger)
	tracingMW := middleware.NewTracingMiddleware()

	// Initialize handlers
	batchHandler := handlers.NewBatchHandler(contractService, redisService, logger)
	healthHandler := handlers.NewHealthHandler(db, redisService, fabricGateway, logger, version)

	// Setup routes
	router := routes.SetupRoutes(
		batchHandler,
		healthHandler,
		authMW,
		rateLimitMW,
		loggerMW,
		corsMW,
		recoveryMW,
		tracingMW,
	)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("API Gateway started successfully",
		zap.String("address", addr),
		zap.String("swagger", fmt.Sprintf("http://%s/swagger/index.html", addr)),
	)

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Close services in reverse order
	// Close Fabric gateway
	if err := fabricGateway.Close(); err != nil {
		logger.Error("Failed to close Fabric gateway", zap.Error(err))
	}

	// Close Redis
	if err := redisService.Close(); err != nil {
		logger.Error("Failed to close Redis", zap.Error(err))
	}

	// Close database
	db.Close()

	logger.Info("All services closed")

	logger.Info("Server exited gracefully")
}

