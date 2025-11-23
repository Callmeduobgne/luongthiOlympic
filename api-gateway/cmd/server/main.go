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

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/handlers"
	aclhandler "github.com/ibn-network/api-gateway/internal/handlers/acl"
	audithandler "github.com/ibn-network/api-gateway/internal/handlers/audit"
	authhandler "github.com/ibn-network/api-gateway/internal/handlers/auth"
	chaincodehandler "github.com/ibn-network/api-gateway/internal/handlers/chaincode"
	channelhandler "github.com/ibn-network/api-gateway/internal/handlers/channel"
	eventhandler "github.com/ibn-network/api-gateway/internal/handlers/event"
	explorerhandler "github.com/ibn-network/api-gateway/internal/handlers/explorer"
	metricshandler "github.com/ibn-network/api-gateway/internal/handlers/metrics"
	networkhandler "github.com/ibn-network/api-gateway/internal/handlers/network"
	transactionhandler "github.com/ibn-network/api-gateway/internal/handlers/transaction"
	"github.com/ibn-network/api-gateway/internal/handlers/users"
	"github.com/ibn-network/api-gateway/internal/middleware"
	"github.com/ibn-network/api-gateway/internal/routes"
	aclservice "github.com/ibn-network/api-gateway/internal/services/acl"
	auditservice "github.com/ibn-network/api-gateway/internal/services/audit"
	authservice "github.com/ibn-network/api-gateway/internal/services/auth"
	caservice "github.com/ibn-network/api-gateway/internal/services/ca"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	chaincodeservice "github.com/ibn-network/api-gateway/internal/services/chaincode"
	channelservice "github.com/ibn-network/api-gateway/internal/services/channel"
	eventservice "github.com/ibn-network/api-gateway/internal/services/event"
	explorerservice "github.com/ibn-network/api-gateway/internal/services/explorer"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"github.com/ibn-network/api-gateway/internal/services/identity"
	indexerservice "github.com/ibn-network/api-gateway/internal/services/indexer"
	metricsservice "github.com/ibn-network/api-gateway/internal/services/metrics"
	networkservice "github.com/ibn-network/api-gateway/internal/services/network"
	transactionservice "github.com/ibn-network/api-gateway/internal/services/transaction"
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

	// Initialize generic chaincode service
	genericChaincodeService := fabric.NewChaincodeService(fabricGateway)

	// Initialize middleware
	authMW := middleware.NewAuthMiddleware(&cfg.JWT, logger)
	rateLimitMW := middleware.NewRateLimitMiddleware(redisService, &cfg.RateLimit, logger)
	loggerMW := middleware.NewLoggerMiddleware(logger)
	corsMW := middleware.NewCORSMiddleware(&cfg.CORS)
	recoveryMW := middleware.NewRecoveryMiddleware(logger)
	tracingMW := middleware.NewTracingMiddleware()

	// Initialize WebSocket rate limit middleware
	var wsRateLimitMW *middleware.WebSocketRateLimitMiddleware
	if cfg.WebSocket.RateLimitEnabled {
		wsRateLimitMW = middleware.NewWebSocketRateLimitMiddleware(
			redisService,
			cfg.WebSocket.RateLimitEnabled,
			cfg.WebSocket.RateLimitMessages,
			cfg.WebSocket.RateLimitWindow,
			logger,
		)
	}

	// Initialize identity service
	identityService, err := identity.NewService(&cfg.CA, logger)
	if err != nil {
		logger.Warn("Failed to initialize identity service", zap.Error(err))
		// Continue without identity service (optional feature)
		identityService = nil
	}

	// Initialize CA service (optional - requires CA server)
	var caService *caservice.Service
	if cfg.CA.URL != "" {
		caService, err = caservice.NewService(&cfg.CA, logger)
		if err != nil {
			logger.Warn("Failed to initialize CA service", zap.Error(err))
			caService = nil
		} else {
			logger.Info("CA service initialized", zap.String("url", cfg.CA.URL))
		}
	}

	// Chaincode lifecycle service removed - lifecycle operations now handled by Admin Service
	// API Gateway only handles business transactions (invoke/query) via Fabric Gateway SDK
	var chaincodeService *chaincodeservice.Service = nil

	// Initialize auth service (with Fabric identity verification and cache for lockout)
	authService := authservice.NewService(db, authMW, &cfg.JWT, identityService, redisService, logger)

	// Set auth service in middleware for API key validation
	authMW.SetAuthService(authService)

	// Initialize handlers
	batchHandler := handlers.NewBatchHandler(contractService, redisService, logger)
	healthHandler := handlers.NewHealthHandler(db, redisService, fabricGateway, logger, version)
	authHandler := authhandler.NewAuthHandler(authService, cfg.Upstream.BackendBaseURL, logger)

	var userHandler *users.UserHandler
	if identityService != nil {
		// Pass CA service if available
		var caServiceInterface users.CAServiceInterface
		if caService != nil {
			caServiceInterface = caService
		}
		userHandler = users.NewUserHandler(identityService, caServiceInterface, logger)
	}

	// Initialize transaction service (needed for chaincode handler)
	transactionService := transactionservice.NewService(db, genericChaincodeService, logger)

	// Create chaincode handler - generic service is always available
	// chaincodeService is optional (for lifecycle management)
	var chaincodeHandler *chaincodehandler.ChaincodeHandler
	if genericChaincodeService != nil {
		chaincodeHandler = chaincodehandler.NewChaincodeHandler(chaincodeService, genericChaincodeService, transactionService, logger)
	}

	// Initialize network service
	networkService := networkservice.NewService(fabricGateway.GetGatewayClient(), &cfg.Fabric, logger)
	networkHandler := networkhandler.NewNetworkHandler(networkService, logger)

	// Initialize discovery service
	discoveryService := networkservice.NewDiscoveryService(fabricGateway.GetGatewayClient(), &cfg.Fabric, logger)
	discoveryHandler := networkhandler.NewDiscoveryHandler(discoveryService, logger)

	// Initialize logs service (for querying Loki)
	lokiURL := os.Getenv("LOKI_URL")
	if lokiURL == "" {
		lokiURL = "http://loki:3100" // Default Loki URL
	}
	logsService := networkservice.NewLogsService(lokiURL, logger)
	logsHandler := networkhandler.NewLogsHandler(logsService, logger)

	// Initialize channel service
	channelService := channelservice.NewService(fabricGateway.GetGatewayClient(), &cfg.Fabric, logger)
	channelHandler := channelhandler.NewChannelHandler(channelService, logger)

	// Transaction service already initialized above for chaincode handler
	transactionHandler := transactionhandler.NewTransactionHandler(transactionService, logger)

	// Initialize event system services
	webhookClient := eventservice.NewWebhookClient(logger)
	wsManager := eventservice.NewWebSocketManager(logger)
	eventDispatcher := eventservice.NewEventDispatcher(webhookClient, wsManager, logger)
	eventListenerService := eventservice.NewListenerService(fabricGateway, eventDispatcher, logger)
	subscriptionService := eventservice.NewSubscriptionService(db, logger)
	eventHandler := eventhandler.NewEventHandler(
		subscriptionService,
		eventListenerService,
		eventDispatcher,
		wsManager,
		logger,
	)

	// Initialize block indexer service
	indexerService := indexerservice.NewService(
		fabricGateway.GetGatewayClient(),
		&cfg.Fabric,
		db,
		logger,
	)

	// Start block indexer in background
	indexerCtx := context.Background()
	if err := indexerService.Start(indexerCtx); err != nil {
		logger.Error("Failed to start block indexer", zap.Error(err))
	} else {
		logger.Info("Block indexer started successfully")
	}
	defer indexerService.Stop()

	// Initialize block explorer service
	explorerService := explorerservice.NewService(
		fabricGateway.GetGatewayClient(),
		&cfg.Fabric,
		transactionService,
		db,
		logger,
	)
	explorerHandler := explorerhandler.NewHandler(explorerService, logger)

	// Initialize audit service
	auditService := auditservice.NewService(db, logger)
	auditHandler := audithandler.NewHandler(auditService, logger)
	auditMW := middleware.NewAuditMiddleware(auditService, logger)

	// Initialize metrics service
	metricsService := metricsservice.NewService(
		db,
		transactionService,
		explorerService,
		auditService,
		logger,
	)
	metricsHandler := metricshandler.NewHandler(metricsService, logger)

	// Initialize ACL service
	aclService := aclservice.NewService(db, logger)
	aclHandler := aclhandler.NewACLHandler(aclService, logger)

	// Setup routes
	router := routes.SetupRoutes(
		batchHandler,
		healthHandler,
		authHandler,
		userHandler,
		aclHandler,
		chaincodeHandler,
		channelHandler,
		networkHandler,
		discoveryHandler,
		logsHandler,
		transactionHandler,
		eventHandler,
		explorerHandler,
		auditHandler,
		metricsHandler,
		metricsService,
		explorerService,
		networkService,
		authMW,
		rateLimitMW,
		loggerMW,
		corsMW,
		recoveryMW,
		tracingMW,
		auditMW,
		&cfg.WebSocket,
		wsRateLimitMW,
		logger,
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
