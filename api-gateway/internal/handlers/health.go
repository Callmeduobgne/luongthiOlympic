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

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ibn-network/api-gateway/internal/handlers/dashboard"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// HealthHandler handles health check operations
type HealthHandler struct {
	db        *pgxpool.Pool
	redis     *cache.Service
	fabric    *fabric.GatewayService
	logger    *zap.Logger
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	db *pgxpool.Pool,
	redis *cache.Service,
	fabric *fabric.GatewayService,
	logger *zap.Logger,
	version string,
) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redis:     redis,
		fabric:    fabric,
		logger:    logger,
		startTime: time.Now(),
		version:   version,
	}
}

// Health godoc
// @Summary Health check
// @Description Check health status of all services
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Failure 503 {object} models.HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services := make(map[string]string)

	// Check database
	if err := h.db.Ping(ctx); err != nil {
		services["database"] = "unhealthy"
		h.logger.Error("Database health check failed", zap.Error(err))
	} else {
		services["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		services["redis"] = "unhealthy"
		h.logger.Error("Redis health check failed", zap.Error(err))
	} else {
		services["redis"] = "healthy"
	}

	// Check Fabric
	if err := h.fabric.Health(ctx); err != nil {
		services["fabric"] = "unhealthy"
		h.logger.Error("Fabric health check failed", zap.Error(err))
	} else {
		services["fabric"] = "healthy"
	}

	// Determine overall status
	status := "healthy"
	statusCode := http.StatusOK
	for _, svcStatus := range services {
		if svcStatus == "unhealthy" {
			status = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	// Get WebSocket metrics
	wsMetrics := dashboard.GetWebSocketMetrics()

	response := models.HealthResponse{
		Status:   status,
		Version:  h.version,
		Uptime:   int64(time.Since(h.startTime).Seconds()),
		Services: services,
	}

	// Add WebSocket metrics to response (if available)
	if wsMetrics != nil {
		responseMap := map[string]interface{}{
			"status":   response.Status,
			"version":  response.Version,
			"uptime":   response.Uptime,
			"services": response.Services,
			"websocket": wsMetrics,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(responseMap)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Ready godoc
// @Summary Readiness check
// @Description Check if service is ready to accept traffic
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /ready [get]
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if all dependencies are ready
	if err := h.db.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "reason": "database unavailable"})
		return
	}

	if err := h.redis.Health(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "reason": "redis unavailable"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

// Live godoc
// @Summary Liveness check
// @Description Check if service is alive
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /live [get]
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

