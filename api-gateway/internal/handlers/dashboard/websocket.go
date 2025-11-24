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

package dashboard

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/middleware"
	"github.com/ibn-network/api-gateway/internal/services/explorer"
	"github.com/ibn-network/api-gateway/internal/services/metrics"
	"github.com/ibn-network/api-gateway/internal/services/network"
	"go.uber.org/zap"
)

// WebSocket connection metrics
var (
	activeConnections   int64
	connectionMutex     sync.RWMutex
	totalConnections    int64
	totalMessages       int64
	connectionErrors    int64
)

// getUpgrader creates a WebSocket upgrader with config
func getUpgrader(cfg *config.WebSocketConfig) websocket.Upgrader {
	allowedOrigins := make(map[string]bool)
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[strings.ToLower(origin)] = true
	}

	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Production best practice: Validate origin to prevent CSRF attacks
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// If no allowed origins configured, allow all (development mode)
			if len(allowedOrigins) == 0 {
				return true
			}
			// Check if origin is in allowed list
			return allowedOrigins[strings.ToLower(origin)]
		},
		// Production best practice: Enable compression for better performance
		EnableCompression: cfg.EnableCompression,
	}
}

// DashboardWebSocketHandler handles WebSocket connections for dashboard real-time updates
type DashboardWebSocketHandler struct {
	metricsService     *metrics.Service
	explorerService    *explorer.Service
	networkService     *network.Service
	wsConfig           *config.WebSocketConfig
	rateLimitMW        *middleware.WebSocketRateLimitMiddleware
	logger             *zap.Logger
}

// NewDashboardWebSocketHandler creates a new dashboard WebSocket handler
func NewDashboardWebSocketHandler(
	metricsService *metrics.Service,
	explorerService *explorer.Service,
	networkService *network.Service,
	wsConfig *config.WebSocketConfig,
	rateLimitMW *middleware.WebSocketRateLimitMiddleware,
	logger *zap.Logger,
) *DashboardWebSocketHandler {
	return &DashboardWebSocketHandler{
		metricsService:  metricsService,
		explorerService: explorerService,
		networkService:  networkService,
		wsConfig:        wsConfig,
		rateLimitMW:     rateLimitMW,
		logger:          logger,
	}
}

// Handle handles WebSocket connections for dashboard
// Production best practice: Proper error handling and connection management
func (h *DashboardWebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	if channelName == "" {
		channelName = "ibnchannel" // Default channel
	}

	// Production best practice: Validate channel name
	if !isValidChannelName(channelName) {
		h.logger.Warn("Invalid channel name", zap.String("channel", channelName))
		http.Error(w, "Invalid channel name", http.StatusBadRequest)
		connectionMutex.Lock()
		connectionErrors++
		connectionMutex.Unlock()
		return
	}

	// Production best practice: Check connection limits
	userID := ""
	if userIDVal := r.Context().Value("userID"); userIDVal != nil {
		if id, ok := userIDVal.(string); ok {
			userID = id
		}
	}

	// Get client IP
	ip := getClientIP(r)

	// Check connection limits
	if h.rateLimitMW != nil {
		if err := h.rateLimitMW.LimitConnection(
			h.wsConfig.MaxConnectionsPerIP,
			h.wsConfig.MaxConnectionsPerUser,
			userID,
			r,
		); err != nil {
			h.logger.Warn("WebSocket connection limit exceeded",
				zap.String("ip", ip),
				zap.String("user_id", userID),
			)
			http.Error(w, "Connection limit exceeded", http.StatusTooManyRequests)
			connectionMutex.Lock()
			connectionErrors++
			connectionMutex.Unlock()
			return
		}
	}

	// Get upgrader with config
	upgrader := getUpgrader(h.wsConfig)

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket",
			zap.Error(err),
			zap.String("channel", channelName),
			zap.String("remote_addr", r.RemoteAddr),
		)
		// Production best practice: Return proper HTTP error
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		connectionMutex.Lock()
		connectionErrors++
		connectionMutex.Unlock()
		return
	}
	defer conn.Close()

	// Increment connection count
	connectionMutex.Lock()
	activeConnections++
	totalConnections++
	connectionMutex.Unlock()

	// Decrement on close
	defer func() {
		connectionMutex.Lock()
		activeConnections--
		connectionMutex.Unlock()
		if h.rateLimitMW != nil {
			h.rateLimitMW.DecrementConnection(ip, userID)
		}
	}()

	// Increment in rate limiter
	if h.rateLimitMW != nil {
		h.rateLimitMW.IncrementConnection(ip, userID)
	}
	
	// Production best practice: Set connection close handler
	conn.SetCloseHandler(func(code int, text string) error {
		h.logger.Info("WebSocket connection closed",
			zap.String("channel", channelName),
			zap.Int("code", code),
			zap.String("reason", text),
			zap.String("user_id", userID),
		)
		return nil
	})

	h.logger.Info("Dashboard WebSocket connection established",
		zap.String("channel", channelName),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Send initial data
	if err := h.sendInitialData(conn, channelName); err != nil {
		h.logger.Error("Failed to send initial data", zap.Error(err))
		return
	}

	// Production best practice: Use config values for ping/pong
	pingInterval := h.wsConfig.PingInterval
	if pingInterval == 0 {
		pingInterval = 30 * time.Second
	}
	pongTimeout := h.wsConfig.PongTimeout
	if pongTimeout == 0 {
		pongTimeout = 60 * time.Second
	}

	// Set up ping/pong
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	// Set up read deadline
	conn.SetReadDeadline(time.Now().Add(pongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	// Start periodic updates
	updateTicker := time.NewTicker(10 * time.Second) // Update every 10 seconds
	defer updateTicker.Stop()

	// Handle close messages in a separate goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Error("WebSocket error", zap.Error(err))
				}
				return
			}
			// Production best practice: Track messages and check rate limit
			connectionMutex.Lock()
			totalMessages++
			connectionMutex.Unlock()

			// Check message rate limit
			if h.rateLimitMW != nil && h.wsConfig.RateLimitEnabled {
				if err := h.rateLimitMW.CheckMessageRate(ip, userID); err != nil {
					h.logger.Warn("WebSocket message rate limit exceeded",
						zap.String("ip", ip),
						zap.String("user_id", userID),
					)
					_ = conn.WriteMessage(websocket.CloseMessage, []byte("Rate limit exceeded"))
					return
				}
			}

			// Log message (optional - can be disabled in production)
			if len(message) > 0 {
				h.logger.Debug("WebSocket message received",
					zap.String("channel", channelName),
					zap.Int("size", len(message)),
				)
			}
		}
	}()

	// Handle messages and updates
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Send ping
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("Failed to send ping", zap.Error(err))
				return
			}

		case <-updateTicker.C:
			// Send periodic updates
			if err := h.sendUpdates(conn, channelName); err != nil {
				h.logger.Error("Failed to send updates", zap.Error(err))
				return
			}
		}
	}
}

// sendInitialData sends initial dashboard data
func (h *DashboardWebSocketHandler) sendInitialData(conn *websocket.Conn, channelName string) error {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get metrics
	metricsData, err := h.metricsService.GetMetricsSummary(ctx, channelName)
	if err != nil {
		h.logger.Warn("Failed to get metrics", zap.Error(err))
		metricsData = nil
	}

	// Get blocks
	blocks, _, err := h.explorerService.ListBlocks(ctx, channelName, 10, 0)
	if err != nil {
		h.logger.Warn("Failed to get blocks", zap.Error(err))
		blocks = nil
	}

	// Get network info
	networkInfo, err := h.networkService.GetNetworkInfo(ctx)
	if err != nil {
		h.logger.Warn("Failed to get network info", zap.Error(err))
		networkInfo = nil
	}

	// Send initial data
	initialData := map[string]interface{}{
		"type":        "initial",
		"metrics":     metricsData,
		"blocks":      blocks,
		"networkInfo": networkInfo,
	}

	return conn.WriteJSON(initialData)
}

// sendUpdates sends periodic updates
func (h *DashboardWebSocketHandler) sendUpdates(conn *websocket.Conn, channelName string) error {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get latest metrics
	metricsData, err := h.metricsService.GetMetricsSummary(ctx, channelName)
	if err == nil && metricsData != nil {
		update := map[string]interface{}{
			"type":    "metrics:update",
			"metrics": metricsData,
		}
		if err := conn.WriteJSON(update); err != nil {
			return err
		}
	}

	// Get latest blocks
	blocks, _, err := h.explorerService.ListBlocks(ctx, channelName, 10, 0)
	if err == nil && blocks != nil {
		update := map[string]interface{}{
			"type":   "blocks:update",
			"blocks": blocks,
		}
		if err := conn.WriteJSON(update); err != nil {
			return err
		}
	}

	return nil
}


// isValidChannelName validates channel name format
// Production best practice: Validate input to prevent injection attacks
func isValidChannelName(channelName string) bool {
	// Channel name should be alphanumeric with hyphens/underscores, 1-64 chars
	if len(channelName) == 0 || len(channelName) > 64 {
		return false
	}
	// Simple validation - in production, use regex or more strict validation
	for _, char := range channelName {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}
	return true
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take first IP if multiple
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	// Fallback to RemoteAddr
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

// GetWebSocketMetrics returns WebSocket connection metrics
// Production best practice: Expose metrics for monitoring
func GetWebSocketMetrics() map[string]interface{} {
	connectionMutex.RLock()
	defer connectionMutex.RUnlock()
	return map[string]interface{}{
		"active_connections": activeConnections,
		"total_connections":  totalConnections,
		"total_messages":     totalMessages,
		"connection_errors":  connectionErrors,
	}
}

		"total_connections":  totalConnections,
		"total_messages":     totalMessages,
		"connection_errors":  connectionErrors,
	}
}
