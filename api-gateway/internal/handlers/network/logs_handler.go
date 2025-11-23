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
	"net/http"
	"strconv"

	"github.com/ibn-network/api-gateway/internal/models"
	networkservice "github.com/ibn-network/api-gateway/internal/services/network"
	"go.uber.org/zap"
)

// LogsHandler handles network logs queries
type LogsHandler struct {
	logsService *networkservice.LogsService
	logger      *zap.Logger
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(logsService *networkservice.LogsService, logger *zap.Logger) *LogsHandler {
	return &LogsHandler{
		logsService: logsService,
		logger:      logger,
	}
}

// GetLogs godoc
// @Summary Get network logs
// @Description Get real-time logs from network nodes (peers, orderers) via Loki
// @Tags network
// @Accept json
// @Produce json
// @Param container query string false "Filter by container name (peer, orderer)"
// @Param level query string false "Filter by log level (error, warn, info, debug)"
// @Param limit query int false "Max number of logs (default: 500, max: 1000)"
// @Param since query string false "Time range (e.g., '5m', '1h', '30m')"
// @Success 200 {object} models.APIResponse{data=[]network.LogEntry}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/logs [get]
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := networkservice.LogQueryParams{
		Container: r.URL.Query().Get("container"),
		Level:     r.URL.Query().Get("level"),
		Limit:     500, // Default limit
		Since:     r.URL.Query().Get("since"),
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			if limit > 0 && limit <= 1000 {
				params.Limit = limit
			}
		}
	}

	// Default since to 1 hour if not specified
	if params.Since == "" {
		params.Since = "1h"
	}

	// Query logs
	logs, err := h.logsService.QueryLogs(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to query logs", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to query logs",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(logs))
}

