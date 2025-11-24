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

package fabric

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// ConnectionManager manages gRPC connection lifecycle with reconnection logic
type ConnectionManager struct {
	cfg           *Config
	credentials   *GRPCCredentials
	conn          *grpc.ClientConn
	logger        *zap.Logger
	mu            sync.RWMutex
	
	// Reconnection settings
	maxRetries    int
	retryDelay    time.Duration
	currentRetry  int
	isConnected   bool
	
	// Health check
	healthCheckInterval time.Duration
	stopHealthCheck     chan struct{}
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(cfg *Config, logger *zap.Logger) (*ConnectionManager, error) {
	credentials, err := NewGRPCCredentials(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC credentials: %w", err)
	}

	cm := &ConnectionManager{
		cfg:                 cfg,
		credentials:         credentials,
		logger:              logger,
		maxRetries:          5,
		retryDelay:          5 * time.Second,
		healthCheckInterval: 30 * time.Second,
		stopHealthCheck:     make(chan struct{}),
	}

	return cm, nil
}

// Connect establishes connection with retry logic
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isConnected && cm.conn != nil {
		state := cm.conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			cm.logger.Info("Already connected to peer", zap.String("state", state.String()))
			return nil
		}
	}

	var lastErr error
	for attempt := 1; attempt <= cm.maxRetries; attempt++ {
		cm.logger.Info("Attempting to connect to peer",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", cm.maxRetries),
			zap.String("endpoint", cm.cfg.PeerEndpoint),
		)

		conn, err := cm.dialWithContext(ctx)
		if err != nil {
			lastErr = err
			cm.logger.Warn("Connection attempt failed",
				zap.Int("attempt", attempt),
				zap.Error(err),
			)

			if attempt < cm.maxRetries {
				// Exponential backoff
				backoff := cm.retryDelay * time.Duration(attempt)
				cm.logger.Info("Retrying connection", zap.Duration("backoff", backoff))
				
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			continue
		}

		// Connection successful
		cm.conn = conn
		cm.isConnected = true
		cm.currentRetry = 0
		
		cm.logger.Info("Successfully connected to peer",
			zap.String("endpoint", cm.cfg.PeerEndpoint),
			zap.String("msp_id", cm.cfg.MSPID),
		)

		// Start health check
		go cm.startHealthCheck()

		return nil
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", cm.maxRetries, lastErr)
}

// dialWithContext creates a gRPC connection with context
func (cm *ConnectionManager) dialWithContext(ctx context.Context) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := cm.credentials.GetDialOptions()
	
	conn, err := grpc.DialContext(dialCtx, cm.cfg.PeerEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	// Wait for connection to be ready
	readyCtx, readyCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readyCancel()

	if !conn.WaitForStateChange(readyCtx, connectivity.Connecting) {
		conn.Close()
		return nil, fmt.Errorf("connection timeout")
	}

	state := conn.GetState()
	if state != connectivity.Ready && state != connectivity.Idle {
		conn.Close()
		return nil, fmt.Errorf("connection not ready, state: %s", state)
	}

	return conn, nil
}

// GetConnection returns the current gRPC connection
func (cm *ConnectionManager) GetConnection() (*grpc.ClientConn, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.isConnected || cm.conn == nil {
		return nil, fmt.Errorf("not connected to peer")
	}

	state := cm.conn.GetState()
	if state == connectivity.Shutdown || state == connectivity.TransientFailure {
		return nil, fmt.Errorf("connection is in %s state", state)
	}

	return cm.conn, nil
}

// Reconnect attempts to reconnect if connection is lost
func (cm *ConnectionManager) Reconnect(ctx context.Context) error {
	cm.mu.Lock()
	cm.isConnected = false
	if cm.conn != nil {
		cm.conn.Close()
		cm.conn = nil
	}
	cm.mu.Unlock()

	cm.logger.Warn("Reconnecting to peer", zap.String("endpoint", cm.cfg.PeerEndpoint))
	return cm.Connect(ctx)
}

// startHealthCheck monitors connection health
func (cm *ConnectionManager) startHealthCheck() {
	ticker := time.NewTicker(cm.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.checkHealth()
		case <-cm.stopHealthCheck:
			cm.logger.Info("Stopping health check")
			return
		}
	}
}

// checkHealth checks connection health and reconnects if needed
func (cm *ConnectionManager) checkHealth() {
	cm.mu.RLock()
	conn := cm.conn
	cm.mu.RUnlock()

	if conn == nil {
		cm.logger.Warn("Connection is nil, attempting reconnect")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := cm.Reconnect(ctx); err != nil {
			cm.logger.Error("Health check reconnection failed", zap.Error(err))
		}
		return
	}

	state := conn.GetState()
	cm.logger.Debug("Connection health check", zap.String("state", state.String()))

	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		cm.logger.Warn("Unhealthy connection detected, reconnecting",
			zap.String("state", state.String()),
		)
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := cm.Reconnect(ctx); err != nil {
			cm.logger.Error("Health check reconnection failed", zap.Error(err))
		}
		
	case connectivity.Ready, connectivity.Idle:
		cm.logger.Debug("Connection healthy", zap.String("state", state.String()))
		
	case connectivity.Connecting:
		cm.logger.Info("Connection in progress", zap.String("state", state.String()))
	}
}

// IsConnected returns connection status
func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.isConnected
}

// Close closes the connection
func (cm *ConnectionManager) Close() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	close(cm.stopHealthCheck)

	if cm.conn != nil {
		err := cm.conn.Close()
		cm.conn = nil
		cm.isConnected = false
		cm.logger.Info("Connection closed")
		return err
	}

	return nil
}

// GetConnectionState returns the current connection state
func (cm *ConnectionManager) GetConnectionState() connectivity.State {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.conn == nil {
		return connectivity.Shutdown
	}

	return cm.conn.GetState()
}
