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
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthChecker performs health checks on Fabric peer connections
type HealthChecker struct {
	connManager *ConnectionManager
	logger      *zap.Logger
}

// HealthStatus represents the health status of a connection
type HealthStatus struct {
	Healthy        bool                  `json:"healthy"`
	PeerEndpoint   string                `json:"peer_endpoint"`
	State          connectivity.State    `json:"state"`
	MSPID          string                `json:"msp_id"`
	LastCheckTime  time.Time             `json:"last_check_time"`
	Error          string                `json:"error,omitempty"`
	ResponseTimeMs int64                 `json:"response_time_ms"`
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(connManager *ConnectionManager, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		connManager: connManager,
		logger:      logger,
	}
}

// CheckHealth performs a comprehensive health check
func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
	startTime := time.Now()
	
	status := &HealthStatus{
		PeerEndpoint:  hc.connManager.cfg.PeerEndpoint,
		MSPID:         hc.connManager.cfg.MSPID,
		LastCheckTime: startTime,
	}

	// Check connection state
	conn, err := hc.connManager.GetConnection()
	if err != nil {
		status.Healthy = false
		status.Error = fmt.Sprintf("connection error: %v", err)
		status.State = connectivity.Shutdown
		status.ResponseTimeMs = time.Since(startTime).Milliseconds()
		
		hc.logger.Warn("Health check failed - no connection", zap.Error(err))
		return status
	}

	// Get connection state
	state := conn.GetState()
	status.State = state

	// Check if state is healthy
	switch state {
	case connectivity.Ready, connectivity.Idle:
		status.Healthy = true
		
		// Optional: Perform gRPC health check if available
		if err := hc.performGRPCHealthCheck(ctx, conn); err != nil {
			hc.logger.Debug("gRPC health check not available or failed", zap.Error(err))
			// Don't fail the overall health check if gRPC health service is not available
		}
		
	case connectivity.Connecting:
		status.Healthy = false
		status.Error = "connection in progress"
		
	case connectivity.TransientFailure:
		status.Healthy = false
		status.Error = "transient connection failure"
		
	case connectivity.Shutdown:
		status.Healthy = false
		status.Error = "connection shutdown"
	}

	status.ResponseTimeMs = time.Since(startTime).Milliseconds()
	
	hc.logger.Debug("Health check completed",
		zap.Bool("healthy", status.Healthy),
		zap.String("state", state.String()),
		zap.Int64("response_time_ms", status.ResponseTimeMs),
	)

	return status
}

// performGRPCHealthCheck performs gRPC health check if the service is available
func (hc *HealthChecker) performGRPCHealthCheck(ctx context.Context, conn grpc.ClientConnInterface) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthClient := grpc_health_v1.NewHealthClient(conn)
	
	resp, err := healthClient.Check(checkCtx, &grpc_health_v1.HealthCheckRequest{
		Service: "", // Empty string checks overall server health
	})
	
	if err != nil {
		return fmt.Errorf("gRPC health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving, status: %s", resp.Status.String())
	}

	hc.logger.Debug("gRPC health check passed")
	return nil
}

// WaitForHealthy waits for the connection to become healthy with timeout
func (hc *HealthChecker) WaitForHealthy(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
			
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for healthy connection after %v", timeout)
			}

			status := hc.CheckHealth(ctx)
			if status.Healthy {
				hc.logger.Info("Connection is healthy",
					zap.String("peer", status.PeerEndpoint),
					zap.Int64("response_time_ms", status.ResponseTimeMs),
				)
				return nil
			}

			hc.logger.Debug("Waiting for healthy connection",
				zap.String("state", status.State.String()),
				zap.String("error", status.Error),
			)
		}
	}
}

// StartContinuousHealthCheck starts continuous health monitoring
func (hc *HealthChecker) StartContinuousHealthCheck(ctx context.Context, interval time.Duration, callback func(*HealthStatus)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	hc.logger.Info("Starting continuous health check", zap.Duration("interval", interval))

	for {
		select {
		case <-ctx.Done():
			hc.logger.Info("Stopping continuous health check")
			return
			
		case <-ticker.C:
			status := hc.CheckHealth(ctx)
			
			if callback != nil {
				callback(status)
			}

			if !status.Healthy {
				hc.logger.Warn("Unhealthy connection detected",
					zap.String("peer", status.PeerEndpoint),
					zap.String("error", status.Error),
				)
			}
		}
	}
}
