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
// Production-ready WebSocket with post-connection authentication
// GET /api/v1/dashboard/ws/{channel}
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get channel from URL
	channel := chi.URLParam(r, "channel")
	if channel == "" {
		channel = "ibnchannel" // Default channel
	}

	// Upgrade connection to WebSocket BEFORE authentication
	// This allows us to send proper close frames on auth failure
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", 
			zap.Error(err),
			zap.String("channel", channel),
		)
		return
	}
	defer conn.Close()

	// Track connection start time for metrics
	connectionStart := time.Now()
	defer func() {
		duration := time.Since(connectionStart).Seconds()
		wsConnectionDuration.Observe(duration)
	}()

	h.logger.Debug("WebSocket connection upgraded, waiting for authentication",
		zap.String("channel", channel),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Set read deadline for auth message (5 seconds)
	authTimeout := 5 * time.Second
	conn.SetReadDeadline(time.Now().Add(authTimeout))

	// Wait for authentication message
	// Expected format: {"type": "auth", "token": "<jwt_token>"}
	_, message, err := conn.ReadMessage()
	if err != nil {
		h.logger.Warn("Failed to read auth message",
			zap.Error(err),
			zap.String("channel", channel),
		)
		wsAuthFailuresTotal.WithLabelValues("read_timeout").Inc()
		
		// Send close frame with policy violation
		closeMsg := websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			"Authentication timeout - no auth message received within 5 seconds",
		)
		conn.WriteMessage(websocket.CloseMessage, closeMsg)
		return
	}

	// Parse auth message
	var authMsg struct {
		Type  string `json:"type"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(message, &authMsg); err != nil {
		h.logger.Warn("Failed to parse auth message",
			zap.Error(err),
			zap.String("channel", channel),
		)
		wsAuthFailuresTotal.WithLabelValues("invalid_format").Inc()
		
		closeMsg := websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			"Invalid auth message format - expected JSON with type and token fields",
		)
		conn.WriteMessage(websocket.CloseMessage, closeMsg)
		return
	}

	// Validate auth message type
	if authMsg.Type != "auth" {
		h.logger.Warn("Invalid auth message type",
			zap.String("type", authMsg.Type),
			zap.String("channel", channel),
		)
		wsAuthFailuresTotal.WithLabelValues("invalid_type").Inc()
		
		closeMsg := websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			"First message must be authentication message with type='auth'",
		)
		conn.WriteMessage(websocket.CloseMessage, closeMsg)
		return
	}

	// Validate token is not empty
	if authMsg.Token == "" {
		h.logger.Warn("Empty token in auth message",
			zap.String("channel", channel),
		)
		wsAuthFailuresTotal.WithLabelValues("empty_token").Inc()
		
		closeMsg := websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			"Authentication token is required",
		)
		conn.WriteMessage(websocket.CloseMessage, closeMsg)
		return
	}

	// Verify JWT token
	claims, err := h.authService.VerifyAccessToken(authMsg.Token)
	if err != nil {
		h.logger.Warn("WebSocket authentication failed - invalid token",
			zap.Error(err),
			zap.String("channel", channel),
			zap.Int("token_length", len(authMsg.Token)),
		)
		wsAuthFailuresTotal.WithLabelValues("invalid_token").Inc()
		
		closeMsg := websocket.FormatCloseMessage(
			websocket.ClosePolicyViolation,
			"Invalid or expired authentication token",
		)
		conn.WriteMessage(websocket.CloseMessage, closeMsg)
		return
	}

	// Authentication successful!
	h.logger.Info("WebSocket authentication successful",
		zap.String("channel", channel),
		zap.String("user_id", claims.UserID.String()),
		zap.String("user_email", claims.Email),
		zap.String("user_role", claims.Role),
	)

	// Track active connection
	globalConnectionTracker.increment(channel)
	defer globalConnectionTracker.decrement(channel)

	// Send auth success response
	authResponse := map[string]interface{}{
		"type":    "auth_success",
		"message": "Authentication successful",
		"user": map[string]string{
			"id":    claims.UserID.String(),
			"email": claims.Email,
			"role":  claims.Role,
		},
		"timestamp": time.Now().Unix(),
	}
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteJSON(authResponse); err != nil {
		h.logger.Error("Failed to send auth success response", zap.Error(err))
		return
	}
	wsMessagesSentTotal.Inc()


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
			// Track received message
			wsMessagesReceivedTotal.Inc()
			
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

	// Skip GetChannelInfo call for production optimization
	// Reason: qscc.GetChainInfo is not available in Fabric network, and frontend doesn't use this data
	// Frontend's getLatestBlocks() always returns empty array regardless of channelInfo
	// This reduces Gateway load and eliminates unnecessary error logs
	var blocksInfo interface{} = nil

	message := map[string]interface{}{
		"type":        "initial",
		"metrics":     snapshot,
		"blocks":      blocksInfo,
		"networkInfo":  nil, // Can be added later
		"timestamp":   time.Now().Unix(),
	}

	if err := conn.WriteJSON(message); err != nil {
		return err
	}
	wsMessagesSentTotal.Inc()
	return nil
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

	if err := conn.WriteJSON(message); err != nil {
		return err
	}
	wsMessagesSentTotal.Inc()
	return nil
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
		if err := conn.WriteJSON(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().Unix(),
		}); err == nil {
			wsMessagesSentTotal.Inc()
		}

	case "request:blocks":
		// Send blocks update
		// Skip GetChannelInfo call - qscc.GetChainInfo is not available and frontend doesn't use this data
		if err := conn.WriteJSON(map[string]interface{}{
			"type":      "blocks:update",
			"blocks":    nil, // Frontend doesn't parse this data anyway
			"timestamp": time.Now().Unix(),
		}); err == nil {
			wsMessagesSentTotal.Inc()
		}

	case "request:metrics":
		// Send metrics update
		snapshot := h.metricsService.GetSnapshot()
		if err := conn.WriteJSON(map[string]interface{}{
			"type":      "metrics:update",
			"metrics":   snapshot,
			"timestamp": time.Now().Unix(),
		}); err == nil {
			wsMessagesSentTotal.Inc()
		}
	}
}

