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

package routes

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/handlers"
	aclhandler "github.com/ibn-network/api-gateway/internal/handlers/acl"
	audithandler "github.com/ibn-network/api-gateway/internal/handlers/audit"
	authhandler "github.com/ibn-network/api-gateway/internal/handlers/auth"
	chaincodehandler "github.com/ibn-network/api-gateway/internal/handlers/chaincode"
	channelhandler "github.com/ibn-network/api-gateway/internal/handlers/channel"
	dashboardhandler "github.com/ibn-network/api-gateway/internal/handlers/dashboard"
	eventhandler "github.com/ibn-network/api-gateway/internal/handlers/event"
	explorerhandler "github.com/ibn-network/api-gateway/internal/handlers/explorer"
	metricshandler "github.com/ibn-network/api-gateway/internal/handlers/metrics"
	networkhandler "github.com/ibn-network/api-gateway/internal/handlers/network"
	transactionhandler "github.com/ibn-network/api-gateway/internal/handlers/transaction"
	"github.com/ibn-network/api-gateway/internal/handlers/users"
	"github.com/ibn-network/api-gateway/internal/middleware"
	explorerservice "github.com/ibn-network/api-gateway/internal/services/explorer"
	metricsservice "github.com/ibn-network/api-gateway/internal/services/metrics"
	networkservice "github.com/ibn-network/api-gateway/internal/services/network"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	batchHandler *handlers.BatchHandler,
	healthHandler *handlers.HealthHandler,
	authHandler *authhandler.AuthHandler,
	userHandler *users.UserHandler,
	aclHandler *aclhandler.ACLHandler,
	chaincodeHandler *chaincodehandler.ChaincodeHandler,
	channelHandler *channelhandler.ChannelHandler,
	networkHandler *networkhandler.NetworkHandler,
	discoveryHandler *networkhandler.DiscoveryHandler,
	logsHandler *networkhandler.LogsHandler,
	transactionHandler *transactionhandler.TransactionHandler,
	eventHandler *eventhandler.EventHandler,
	explorerHandler *explorerhandler.Handler,
	auditHandler *audithandler.Handler,
	metricsHandler *metricshandler.Handler,
	metricsService *metricsservice.Service,
	explorerService *explorerservice.Service,
	networkService *networkservice.Service,
	authMW *middleware.AuthMiddleware,
	rateLimitMW *middleware.RateLimitMiddleware,
	loggerMW *middleware.LoggerMiddleware,
	corsMW *middleware.CORSMiddleware,
	recoveryMW *middleware.RecoveryMiddleware,
	tracingMW *middleware.TracingMiddleware,
	auditMW *middleware.AuditMiddleware,
	wsConfig *config.WebSocketConfig,
	wsRateLimitMW *middleware.WebSocketRateLimitMiddleware,
	logger *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(recoveryMW.Recover)
	// CORS is handled by Nginx proxy, so we disable it here to avoid duplicate headers
	// r.Use(corsMW.Handler)
	r.Use(loggerMW.Log)
	r.Use(tracingMW.Trace)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)

	// Compression middleware - skip for WebSocket requests
	// Production best practice: WebSocket requires http.Hijacker which compression breaks
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip compression for WebSocket upgrade requests
			upgrade := r.Header.Get("Upgrade")
			connection := r.Header.Get("Connection")
			if strings.ToLower(upgrade) == "websocket" ||
				(connection != "" && strings.Contains(strings.ToLower(connection), "upgrade")) {
				// Bypass compression for WebSocket to preserve http.Hijacker
				next.ServeHTTP(w, r)
				return
			}
			// Apply compression for other requests
			chimiddleware.Compress(5)(next).ServeHTTP(w, r)
		})
	})

	// Audit middleware (log all requests, but skip WebSocket)
	// Production best practice: Skip audit for WebSocket to reduce log volume
	if auditMW != nil {
		r.Use(auditMW.Audit)
	}

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
		// Auth routes (no auth required for login/register)
		if authHandler != nil {
			r.Route("/auth", func(r chi.Router) {
				// Login endpoint with stricter rate limiting (anti-brute force: 5 attempts per 15 minutes)
				r.With(rateLimitMW.LimitLogin).Post("/login", authHandler.Login)
				r.Post("/register", authHandler.Register)
				r.Post("/refresh", authHandler.RefreshToken)

				// Protected routes
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					r.Get("/profile", authHandler.GetProfile)
					r.Post("/avatar", authHandler.UploadAvatar)
					r.Post("/api-keys", authHandler.GenerateAPIKey)
					r.Get("/api-keys", authHandler.ListAPIKeys)
					r.Delete("/api-keys/{id}", authHandler.RevokeAPIKey)
				})
			})
		}

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

		// User/Identity management routes
		if userHandler != nil {
			r.Route("/users", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Read-only routes (any authenticated user)
					r.Get("/", userHandler.ListUsers)
					r.Get("/{id}", userHandler.GetUser)

					// Admin-only routes
					r.Group(func(r chi.Router) {
						r.Use(middleware.AdminOnly(logger))

						r.Post("/enroll", userHandler.Enroll)
						r.Post("/register", userHandler.Register)
						r.Post("/{id}/reenroll", userHandler.Reenroll)
						r.Delete("/{id}/revoke", userHandler.Revoke)
						r.Get("/{id}/certificate", userHandler.GetCertificate)
					})
				})
			})
		}

		// Chaincode management routes
		// NOTE: Lifecycle operations (install/approve/commit/list) have been moved to Admin Service
		// API Gateway only handles business transactions (invoke/query)
		if chaincodeHandler != nil {
			// Generic chaincode invocation routes (business transactions only)
			r.Route("/channels/{channel}/chaincodes/{name}", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Generic invoke and query (business transactions)
					r.Post("/invoke", chaincodeHandler.Invoke)
					r.Post("/query", chaincodeHandler.Query)
				})
			})
		}

		// Channel management routes
		if channelHandler != nil {
			r.Route("/channels", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Admin-only routes (channel operations require admin privileges)
					r.Group(func(r chi.Router) {
						r.Use(middleware.AdminOnly(logger))

						// Create channel
						r.Post("/", channelHandler.CreateChannel)

						// Update channel config
						r.Patch("/{name}/config", channelHandler.UpdateChannelConfig)

						// Join peer to channel
						r.Post("/{name}/join", channelHandler.JoinPeer)
					})

					// Read-only routes (any authenticated user)
					// Get channel config (also available at /network/channels/{name}/config)
					if networkHandler != nil {
						r.Get("/{name}/config", networkHandler.GetChannelConfig)
					}

					// List channel members
					r.Get("/{name}/members", channelHandler.ListChannelMembers)

					// List peers in channel
					r.Get("/{name}/peers", channelHandler.ListChannelPeers)
				})
			})
		}

		// Network routes (information + discovery)
		if networkHandler != nil || discoveryHandler != nil {
			r.Route("/network", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Network information routes
					if networkHandler != nil {
						// Network info
						r.Get("/info", networkHandler.GetNetworkInfo)

						// Channels
						r.Get("/channels", networkHandler.ListChannels)
						r.Get("/channels/{name}", networkHandler.GetChannelInfo)
						r.Get("/channels/{name}/config", networkHandler.GetChannelConfig)

						// Blocks and transactions
						r.Get("/channels/{channel}/blocks/{number}", networkHandler.GetBlockInfo)
						r.Get("/channels/{channel}/transactions/{txid}", networkHandler.GetTransactionInfo)

						// Chaincode on channel
						r.Get("/channels/{channel}/chaincode/{chaincode}", networkHandler.GetChaincodeInfoOnChannel)
					}

					// Network discovery routes
					if discoveryHandler != nil {
						// Peers
						r.Get("/peers", discoveryHandler.ListPeers)
						r.Get("/peers/{id}", discoveryHandler.GetPeer)

						// Orderers
						r.Get("/orderers", discoveryHandler.ListOrderers)
						r.Get("/orderers/{id}", discoveryHandler.GetOrderer)

						// CAs
						r.Get("/cas", discoveryHandler.ListCAs)

						// Topology
						r.Get("/topology", discoveryHandler.GetTopology)

						// Peers in channel
						r.Get("/channels/{channel}/peers", discoveryHandler.GetPeersInChannel)

						// Health checks
						r.Get("/health/peers", discoveryHandler.CheckAllPeersHealth)
						r.Get("/health/peers/{id}", discoveryHandler.CheckPeerHealth)
						r.Get("/health/orderers", discoveryHandler.CheckAllOrderersHealth)
						r.Get("/health/orderers/{id}", discoveryHandler.CheckOrdererHealth)
					}

					// Network logs route
					if logsHandler != nil {
						r.Get("/logs", logsHandler.GetLogs)
					}
				})
			})
		}

		// Transaction management routes
		if transactionHandler != nil {
			r.Route("/transactions", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Submit transaction
					r.Post("/", transactionHandler.SubmitTransaction)

					// List transactions
					r.Get("/", transactionHandler.ListTransactions)

					// Get transaction details
					r.Get("/{id}", transactionHandler.GetTransaction)
					r.Get("/{id}/status", transactionHandler.GetTransactionStatus)
					r.Get("/{id}/receipt", transactionHandler.GetTransactionReceipt)
				})
			})
		}

		// Block explorer routes
		if explorerHandler != nil {
			r.Route("/blocks", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Block operations
					r.Get("/{channel}", explorerHandler.ListBlocks)
					r.Get("/{channel}/latest", explorerHandler.GetLatestBlock)
					r.Get("/{channel}/{number}", explorerHandler.GetBlock)
					r.Get("/{channel}/{number}/transactions", explorerHandler.GetTransactionByBlock)
				})
			})
		}

		// Audit logging routes
		if auditHandler != nil {
			r.Route("/audit", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Audit log endpoints
					r.Get("/logs", auditHandler.ListLogs)
					r.Get("/logs/{id}", auditHandler.GetLog)
				})
			})
		}

		// Metrics routes
		if metricsHandler != nil {
			r.Route("/metrics", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Metrics endpoints
					r.Get("/transactions", metricsHandler.GetTransactionMetrics)
					r.Get("/blocks", metricsHandler.GetBlockMetrics)
					r.Get("/performance", metricsHandler.GetPerformanceMetrics)
					r.Get("/peers", metricsHandler.GetPeerMetrics)
					r.Get("/summary", metricsHandler.GetMetricsSummary)
				})
			})
		}

		// Dashboard WebSocket (simple, no subscription required)
		if metricsService != nil && explorerService != nil && networkService != nil && wsConfig != nil {
			dashboardWSHandler := dashboardhandler.NewDashboardWebSocketHandler(
				metricsService,
				explorerService,
				networkService,
				wsConfig,
				wsRateLimitMW,
				logger,
			)
			r.Route("/dashboard", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Get("/ws/{channel}", dashboardWSHandler.Handle)
				})
			})
		}

		// Event system routes
		if eventHandler != nil {
			r.Route("/events", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Subscription management
					r.Post("/subscriptions", eventHandler.CreateSubscription)
					r.Get("/subscriptions", eventHandler.ListSubscriptions)
					r.Get("/subscriptions/{id}", eventHandler.GetSubscription)
					r.Patch("/subscriptions/{id}", eventHandler.UpdateSubscription)
					r.Delete("/subscriptions/{id}", eventHandler.DeleteSubscription)

					// Real-time event streams
					r.Get("/ws/{subscriptionId}", eventHandler.WebSocketHandler)
					r.Get("/sse/{subscriptionId}", eventHandler.SSEHandler)
				})
			})
		}

		// ACL routes
		if aclHandler != nil {
			r.Route("/acl", func(r chi.Router) {
				// Require authentication
				r.Group(func(r chi.Router) {
					r.Use(authMW.Authenticate)
					r.Use(rateLimitMW.Limit)

					// Policy management (Admin only for write operations)
					r.Group(func(r chi.Router) {
						r.Use(middleware.AdminOnly(logger))

						// Create policy
						r.Post("/policies", aclHandler.CreatePolicy)

						// Update policy
						r.Patch("/policies/{id}", aclHandler.UpdatePolicy)

						// Delete policy
						r.Delete("/policies/{id}", aclHandler.DeletePolicy)
					})

					// Read-only policy operations
					r.Get("/policies", aclHandler.ListPolicies)
					r.Get("/policies/{id}", aclHandler.GetPolicy)

					// Permissions
					r.Get("/permissions", aclHandler.ListPermissions)

					// Permission check
					r.Post("/check", aclHandler.CheckPermission)
				})
			})
		}
	})

	return r
}
