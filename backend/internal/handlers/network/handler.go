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

package network

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ibn-network/backend/internal/services/network"
	"go.uber.org/zap"
)

// Handler handles network-related HTTP requests
type Handler struct {
	service *network.Service
	logger  *zap.Logger
}

// NewHandler creates a new network handler
func NewHandler(service *network.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetLogs handles GET /api/v1/network/logs
// @Summary Get network logs from Loki
// @Description Query logs from Loki with optional container and search filters
// @Tags network
// @Accept json
// @Produce json
// @Param container query string false "Container name filter (can be specified multiple times)"
// @Param since query string false "Time range (e.g., '1h', '30m')" default("1h")
// @Param limit query int false "Maximum number of logs" default(500)
// @Param search query string false "Search query"
// @Success 200 {object} map[string]interface{} "Logs response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/logs [get]
// @Security BearerAuth
func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	
	req := &network.QueryLogsRequest{
		Since:  query.Get("since"),
		Search: query.Get("search"),
		Limit:  500, // Default limit
	}
	
	// Parse containers (can be multiple)
	containers := query["container"]
	if len(containers) > 0 {
		req.Containers = containers
	}
	
	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}
	
	// Default since if not provided
	if req.Since == "" {
		req.Since = "1h"
	}
	
	// Query logs
	logs, err := h.service.QueryLogs(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to query logs", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "QUERY_LOGS_FAILED",
				"message": "Failed to query logs from Loki",
				"details": err.Error(),
			},
		})
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

