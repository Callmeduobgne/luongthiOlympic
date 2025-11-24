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

package event

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024
)

// WebSocketManager manages WebSocket connections and SSE connections
type WebSocketManager struct {
	connections map[string]*websocket.Conn // subscription ID -> connection
	sseClients  map[string]*SSEClient     // subscription ID -> SSE client
	logger      *zap.Logger
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
	shutdown    chan struct{} // Channel to signal shutdown
	shutdownMu  sync.Mutex
	isShutdown  bool
}

// SSEClient represents a Server-Sent Events client
type SSEClient struct {
	ID       string
	Channel  chan *models.ChaincodeEvent
	Done     chan bool
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(logger *zap.Logger) *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*websocket.Conn),
		sseClients:  make(map[string]*SSEClient),
		logger:      logger,
		shutdown:    make(chan struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now (should be configured in production)
				return true
			},
		},
	}
}

// RegisterWebSocketConnection registers a new WebSocket connection
func (m *WebSocketManager) RegisterWebSocketConnection(subscriptionID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[subscriptionID] = conn

	// Start ping/pong handler
	go m.handlePingPong(subscriptionID, conn)

	m.logger.Info("WebSocket connection registered",
		zap.String("subscription_id", subscriptionID),
	)
}

// UnregisterWebSocketConnection unregisters a WebSocket connection
func (m *WebSocketManager) UnregisterWebSocketConnection(subscriptionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.connections[subscriptionID]; ok {
		conn.Close()
		delete(m.connections, subscriptionID)
		m.logger.Info("WebSocket connection unregistered",
			zap.String("subscription_id", subscriptionID),
		)
	}
}

// RegisterSSEClient registers a new SSE client
func (m *WebSocketManager) RegisterSSEClient(subscriptionID string) *SSEClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := &SSEClient{
		ID:      subscriptionID,
		Channel: make(chan *models.ChaincodeEvent, 10),
		Done:    make(chan bool),
	}

	m.sseClients[subscriptionID] = client

	m.logger.Info("SSE client registered",
		zap.String("subscription_id", subscriptionID),
	)

	return client
}

// UnregisterSSEClient unregisters an SSE client
func (m *WebSocketManager) UnregisterSSEClient(subscriptionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.sseClients[subscriptionID]; ok {
		close(client.Channel)
		close(client.Done)
		delete(m.sseClients, subscriptionID)
		m.logger.Info("SSE client unregistered",
			zap.String("subscription_id", subscriptionID),
		)
	}
}

// SendEvent sends event to WebSocket client
func (m *WebSocketManager) SendEvent(subscriptionID string, event *models.ChaincodeEvent) error {
	m.mu.RLock()
	conn, ok := m.connections[subscriptionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("WebSocket connection not found: %s", subscriptionID)
	}

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))

	// Send JSON message
	if err := conn.WriteJSON(event); err != nil {
		m.logger.Error("Failed to send WebSocket event",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// SendSSEEvent sends event to SSE client
func (m *WebSocketManager) SendSSEEvent(subscriptionID string, event *models.ChaincodeEvent) error {
	m.mu.RLock()
	client, ok := m.sseClients[subscriptionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("SSE client not found: %s", subscriptionID)
	}

	// Non-blocking send
	select {
	case client.Channel <- event:
		return nil
	case <-time.After(1 * time.Second):
		return fmt.Errorf("SSE client channel full: %s", subscriptionID)
	}
}

// handlePingPong handles ping/pong messages for WebSocket connection
func (m *WebSocketManager) handlePingPong(subscriptionID string, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-m.shutdown:
			// Graceful shutdown: close connection
			m.logger.Info("Closing WebSocket connection due to shutdown",
				zap.String("subscription_id", subscriptionID),
			)
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"))
			m.UnregisterWebSocketConnection(subscriptionID)
			return

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				m.logger.Warn("Failed to send ping",
					zap.String("subscription_id", subscriptionID),
					zap.Error(err),
				)
				m.UnregisterWebSocketConnection(subscriptionID)
				return
			}
		}
	}
}

// GetUpgrader returns the WebSocket upgrader
func (m *WebSocketManager) GetUpgrader() websocket.Upgrader {
	return m.upgrader
}

// Shutdown gracefully shuts down all WebSocket and SSE connections
// Production best practice: Clean shutdown to prevent connection leaks
func (m *WebSocketManager) Shutdown() error {
	m.shutdownMu.Lock()
	if m.isShutdown {
		m.shutdownMu.Unlock()
		return nil
	}
	m.isShutdown = true
	close(m.shutdown)
	m.shutdownMu.Unlock()

	m.logger.Info("Shutting down WebSocket manager")

	// Close all WebSocket connections
	m.mu.Lock()
	connCount := len(m.connections)
	for subscriptionID, conn := range m.connections {
		m.logger.Info("Closing WebSocket connection",
			zap.String("subscription_id", subscriptionID),
		)
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server shutting down"))
		conn.Close()
	}
	m.connections = make(map[string]*websocket.Conn)

	// Close all SSE clients
	sseCount := len(m.sseClients)
	for subscriptionID, client := range m.sseClients {
		m.logger.Info("Closing SSE client",
			zap.String("subscription_id", subscriptionID),
		)
		close(client.Channel)
		close(client.Done)
	}
	m.sseClients = make(map[string]*SSEClient)
	m.mu.Unlock()

	m.logger.Info("WebSocket manager shutdown complete",
		zap.Int("websocket_connections", connCount),
		zap.Int("sse_clients", sseCount),
	)

	return nil
}

