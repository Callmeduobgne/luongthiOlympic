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
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ibn-network/backend/internal/services/analytics/metrics"
	blockchainInfo "github.com/ibn-network/backend/internal/services/blockchain/info"
	"github.com/ibn-network/backend/internal/services/auth"
	"go.uber.org/zap"
)

const (
	// Production best practice: WebSocket timeouts
	writeWait      = 10 * time.Second // Time allowed to write a message
	pongWait       = 60 * time.Second // Time allowed to read pong
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024      // 512KB max message size
	updateInterval = 10 * time.Second // Update interval (production: 10s, not 2s)
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development, restrict in production
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// BlockchainInfoService interface for blockchain info operations
type BlockchainInfoService interface {
	GetChannelInfo(ctx context.Context) (*blockchainInfo.ChannelInfo, error)
}

// Handler handles dashboard WebSocket connections
type Handler struct {
	metricsService  *metrics.Service
	blockchainInfo  BlockchainInfoService
	authService     *auth.Service
	logger          *zap.Logger
}

// NewHandler creates a new dashboard handler
func NewHandler(
	metricsService *metrics.Service,
	blockchainInfo BlockchainInfoService,
	authService *auth.Service,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		metricsService: metricsService,
		blockchainInfo: blockchainInfo,
		authService:    authService,
		logger:         logger,
	}
}

// HandleWebSocket handles WebSocket connections for dashboard real-time updates
// GET /api/v1/dashboard/ws/{channel}?token=<jwt_token>
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get channel from URL
	channel := chi.URLParam(r, "channel")
	if channel == "" {
		channel = "ibnchannel" // Default channel
	}

	// Get token from query parameter (WebSocket doesn't support headers easily)
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header as fallback
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Verify token
	claims, err := h.authService.VerifyAccessToken(token)
	if err != nil {
		h.logger.Warn("WebSocket authentication failed", zap.Error(err))
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established",
		zap.String("channel", channel),
		zap.String("user_id", claims.UserID.String()),
	)

	// Create context for this connection with timeout
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Set read deadline and pong handler
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Set read limit for message size (production best practice)
	conn.SetReadLimit(maxMessageSize)

	// Set close handler
	conn.SetCloseHandler(func(code int, text string) error {
		h.logger.Info("WebSocket connection closed",
			zap.String("channel", channel),
			zap.Int("code", code),
			zap.String("reason", text),
			zap.String("user_id", claims.UserID.String()),
		)
		return nil
	})

	// Send initial data
	if err := h.sendInitialData(ctx, conn, channel); err != nil {
		h.logger.Error("Failed to send initial data", zap.Error(err))
		return
	}

	// Start ping ticker
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Start update ticker (production: 10s interval)
	updateTicker := time.NewTicker(updateInterval)
	defer updateTicker.Stop()

	// Message channel with buffer
	messageChan := make(chan []byte, 256)
	errorChan := make(chan error, 1)

	// Read messages in goroutine
	go func() {
		defer close(errorChan)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Error("WebSocket read error", zap.Error(err))
				}
				errorChan <- err
				return
			}
			// Check message size (production best practice)
			if len(message) > maxMessageSize {
				h.logger.Warn("WebSocket message too large", zap.Int("size", len(message)))
				errorChan <- websocket.ErrReadLimit
				return
			}
			select {
			case messageChan <- message:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket connection closed by context")
			return

		case <-pingTicker.C:
			// Send ping with write timeout
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("Failed to send ping", zap.Error(err))
				return
			}

		case <-updateTicker.C:
			// Send periodic updates with timeout
			if err := h.sendUpdate(ctx, conn, channel); err != nil {
				h.logger.Error("Failed to send update", zap.Error(err))
				return
			}

		case message := <-messageChan:
			// Handle incoming messages
			h.handleMessage(ctx, conn, message, channel)

		case err := <-errorChan:
			if err != nil {
				h.logger.Info("WebSocket connection closed", zap.Error(err))
				return
			}
		}
	}
}

// sendInitialData sends initial dashboard data
func (h *Handler) sendInitialData(ctx context.Context, conn *websocket.Conn, channel string) error {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	snapshot := h.metricsService.GetSnapshot()

	// Get channel info (for blocks info)
	// Note: GetChannelInfo doesn't take channel parameter, uses default from service config
	channelInfo, err := h.blockchainInfo.GetChannelInfo(ctx)
	var blocksInfo interface{}
	if err != nil {
		h.logger.Warn("Failed to get channel info", zap.Error(err), zap.String("channel", channel))
		blocksInfo = nil
	} else {
		blocksInfo = channelInfo
	}

	message := map[string]interface{}{
		"type":        "initial",
		"metrics":     snapshot,
		"blocks":      blocksInfo,
		"networkInfo":  nil, // Can be added later
		"timestamp":   time.Now().Unix(),
	}

	return conn.WriteJSON(message)
}

// sendUpdate sends periodic dashboard updates
func (h *Handler) sendUpdate(ctx context.Context, conn *websocket.Conn, channel string) error {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	snapshot := h.metricsService.GetSnapshot()

	message := map[string]interface{}{
		"type":      "metrics:update",
		"metrics":   snapshot,
		"timestamp": time.Now().Unix(),
	}

	return conn.WriteJSON(message)
}

// handleMessage handles incoming WebSocket messages
func (h *Handler) handleMessage(ctx context.Context, conn *websocket.Conn, message []byte, channel string) {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	// Check context cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		h.logger.Warn("Failed to parse WebSocket message", zap.Error(err))
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// Respond to ping
		_ = conn.WriteJSON(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Unix(),
		})

	case "request:blocks":
		// Send blocks update
		// Note: GetChannelInfo doesn't take channel parameter, uses default from service config
		channelInfo, err := h.blockchainInfo.GetChannelInfo(ctx)
		if err != nil {
			h.logger.Warn("Failed to get channel info for blocks request", zap.Error(err))
			return
		}
		_ = conn.WriteJSON(map[string]interface{}{
			"type":      "blocks:update",
			"blocks":    channelInfo,
			"timestamp": time.Now().Unix(),
		})

	case "request:metrics":
		// Send metrics update
		snapshot := h.metricsService.GetSnapshot()
		_ = conn.WriteJSON(map[string]interface{}{
			"type":      "metrics:update",
			"metrics":   snapshot,
			"timestamp": time.Now().Unix(),
		})
	}
}


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
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ibn-network/backend/internal/services/analytics/metrics"
	blockchainInfo "github.com/ibn-network/backend/internal/services/blockchain/info"
	"github.com/ibn-network/backend/internal/services/auth"
	"go.uber.org/zap"
)

const (
	// Production best practice: WebSocket timeouts
	writeWait      = 10 * time.Second // Time allowed to write a message
	pongWait       = 60 * time.Second // Time allowed to read pong
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024      // 512KB max message size
	updateInterval = 10 * time.Second // Update interval (production: 10s, not 2s)
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development, restrict in production
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// BlockchainInfoService interface for blockchain info operations
type BlockchainInfoService interface {
	GetChannelInfo(ctx context.Context) (*blockchainInfo.ChannelInfo, error)
}

// Handler handles dashboard WebSocket connections
type Handler struct {
	metricsService  *metrics.Service
	blockchainInfo  BlockchainInfoService
	authService     *auth.Service
	logger          *zap.Logger
}

// NewHandler creates a new dashboard handler
func NewHandler(
	metricsService *metrics.Service,
	blockchainInfo BlockchainInfoService,
	authService *auth.Service,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		metricsService: metricsService,
		blockchainInfo: blockchainInfo,
		authService:    authService,
		logger:         logger,
	}
}

// HandleWebSocket handles WebSocket connections for dashboard real-time updates
// GET /api/v1/dashboard/ws/{channel}?token=<jwt_token>
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get channel from URL
	channel := chi.URLParam(r, "channel")
	if channel == "" {
		channel = "ibnchannel" // Default channel
	}

	// Get token from query parameter (WebSocket doesn't support headers easily)
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header as fallback
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Verify token
	claims, err := h.authService.VerifyAccessToken(token)
	if err != nil {
		h.logger.Warn("WebSocket authentication failed", zap.Error(err))
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established",
		zap.String("channel", channel),
		zap.String("user_id", claims.UserID.String()),
	)

	// Create context for this connection with timeout
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Set read deadline and pong handler
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Set read limit for message size (production best practice)
	conn.SetReadLimit(maxMessageSize)

	// Set close handler
	conn.SetCloseHandler(func(code int, text string) error {
		h.logger.Info("WebSocket connection closed",
			zap.String("channel", channel),
			zap.Int("code", code),
			zap.String("reason", text),
			zap.String("user_id", claims.UserID.String()),
		)
		return nil
	})

	// Send initial data
	if err := h.sendInitialData(ctx, conn, channel); err != nil {
		h.logger.Error("Failed to send initial data", zap.Error(err))
		return
	}

	// Start ping ticker
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Start update ticker (production: 10s interval)
	updateTicker := time.NewTicker(updateInterval)
	defer updateTicker.Stop()

	// Message channel with buffer
	messageChan := make(chan []byte, 256)
	errorChan := make(chan error, 1)

	// Read messages in goroutine
	go func() {
		defer close(errorChan)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Error("WebSocket read error", zap.Error(err))
				}
				errorChan <- err
				return
			}
			// Check message size (production best practice)
			if len(message) > maxMessageSize {
				h.logger.Warn("WebSocket message too large", zap.Int("size", len(message)))
				errorChan <- websocket.ErrReadLimit
				return
			}
			select {
			case messageChan <- message:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("WebSocket connection closed by context")
			return

		case <-pingTicker.C:
			// Send ping with write timeout
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("Failed to send ping", zap.Error(err))
				return
			}

		case <-updateTicker.C:
			// Send periodic updates with timeout
			if err := h.sendUpdate(ctx, conn, channel); err != nil {
				h.logger.Error("Failed to send update", zap.Error(err))
				return
			}

		case message := <-messageChan:
			// Handle incoming messages
			h.handleMessage(ctx, conn, message, channel)

		case err := <-errorChan:
			if err != nil {
				h.logger.Info("WebSocket connection closed", zap.Error(err))
				return
			}
		}
	}
}

// sendInitialData sends initial dashboard data
func (h *Handler) sendInitialData(ctx context.Context, conn *websocket.Conn, channel string) error {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	snapshot := h.metricsService.GetSnapshot()

	// Get channel info (for blocks info)
	// Note: GetChannelInfo doesn't take channel parameter, uses default from service config
	channelInfo, err := h.blockchainInfo.GetChannelInfo(ctx)
	var blocksInfo interface{}
	if err != nil {
		h.logger.Warn("Failed to get channel info", zap.Error(err), zap.String("channel", channel))
		blocksInfo = nil
	} else {
		blocksInfo = channelInfo
	}

	message := map[string]interface{}{
		"type":        "initial",
		"metrics":     snapshot,
		"blocks":      blocksInfo,
		"networkInfo":  nil, // Can be added later
		"timestamp":   time.Now().Unix(),
	}

	return conn.WriteJSON(message)
}

// sendUpdate sends periodic dashboard updates
func (h *Handler) sendUpdate(ctx context.Context, conn *websocket.Conn, channel string) error {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	snapshot := h.metricsService.GetSnapshot()

	message := map[string]interface{}{
		"type":      "metrics:update",
		"metrics":   snapshot,
		"timestamp": time.Now().Unix(),
	}

	return conn.WriteJSON(message)
}

// handleMessage handles incoming WebSocket messages
func (h *Handler) handleMessage(ctx context.Context, conn *websocket.Conn, message []byte, channel string) {
	// Production best practice: Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	// Check context cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		h.logger.Warn("Failed to parse WebSocket message", zap.Error(err))
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// Respond to ping
		_ = conn.WriteJSON(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Unix(),
		})

	case "request:blocks":
		// Send blocks update
		// Note: GetChannelInfo doesn't take channel parameter, uses default from service config
		channelInfo, err := h.blockchainInfo.GetChannelInfo(ctx)
		if err != nil {
			h.logger.Warn("Failed to get channel info for blocks request", zap.Error(err))
			return
		}
		_ = conn.WriteJSON(map[string]interface{}{
			"type":      "blocks:update",
			"blocks":    channelInfo,
			"timestamp": time.Now().Unix(),
		})

	case "request:metrics":
		// Send metrics update
		snapshot := h.metricsService.GetSnapshot()
		_ = conn.WriteJSON(map[string]interface{}{
			"type":      "metrics:update",
			"metrics":   snapshot,
			"timestamp": time.Now().Unix(),
		})
	}
}

