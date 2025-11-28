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

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/backend/internal/services/network"
	"go.uber.org/zap"
)

// Handler handles network-related HTTP requests
type Handler struct {
	logsService      *network.Service
	discoveryService *network.DiscoveryService
	logger           *zap.Logger
}

// NewHandler creates a new network handler
func NewHandler(logsService *network.Service, discoveryService *network.DiscoveryService, logger *zap.Logger) *Handler {
	return &Handler{
		logsService:      logsService,
		discoveryService: discoveryService,
		logger:           logger,
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
	logs, err := h.logsService.QueryLogs(r.Context(), req)
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

// GetNetworkInfo handles GET /api/v1/network/info
// @Summary Get network information
// @Description Get overall network information (channels, peers, orderers, MSPs)
// @Tags network
// @Produce json
// @Success 200 {object} map[string]interface{} "Network info response"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/info [get]
// @Security BearerAuth
func (h *Handler) GetNetworkInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.discoveryService.GetNetworkInfo(r.Context())
	if err != nil {
		h.logger.Error("Failed to get network info", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "GET_NETWORK_INFO_FAILED",
				"message": "Failed to get network info from Gateway",
				"details": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    info,
	})
}

// ListPeers handles GET /api/v1/network/peers
// @Summary List all peers
// @Description List all peers in the network
// @Tags network
// @Produce json
// @Success 200 {object} map[string]interface{} "Peers response"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/peers [get]
// @Security BearerAuth
func (h *Handler) ListPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := h.discoveryService.ListPeers(r.Context())
	if err != nil {
		h.logger.Error("Failed to list peers", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "LIST_PEERS_FAILED",
				"message": "Failed to list peers from Gateway",
				"details": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    peers,
	})
}

// ListOrderers handles GET /api/v1/network/orderers
// @Summary List all orderers
// @Description List all orderers in the network
// @Tags network
// @Produce json
// @Success 200 {object} map[string]interface{} "Orderers response"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/orderers [get]
// @Security BearerAuth
func (h *Handler) ListOrderers(w http.ResponseWriter, r *http.Request) {
	orderers, err := h.discoveryService.ListOrderers(r.Context())
	if err != nil {
		h.logger.Error("Failed to list orderers", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "LIST_ORDERERS_FAILED",
				"message": "Failed to list orderers from Gateway",
				"details": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    orderers,
	})
}

// ListChannels handles GET /api/v1/network/channels
// @Summary List all channels
// @Description List all channels in the network
// @Tags network
// @Produce json
// @Success 200 {object} map[string]interface{} "Channels response"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/channels [get]
// @Security BearerAuth
func (h *Handler) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.discoveryService.ListChannels(r.Context())
	if err != nil {
		h.logger.Error("Failed to list channels", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "LIST_CHANNELS_FAILED",
				"message": "Failed to list channels from Gateway",
				"details": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    channels,
	})
}

// GetChannelInfo handles GET /api/v1/network/channels/{name}
// @Summary Get channel information
// @Description Get detailed information about a specific channel
// @Tags network
// @Produce json
// @Param name path string true "Channel name"
// @Success 200 {object} map[string]interface{} "Channel info response"
// @Failure 404 {object} map[string]interface{} "Channel not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/channels/{name} [get]
// @Security BearerAuth
func (h *Handler) GetChannelInfo(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "INVALID_CHANNEL_NAME",
				"message": "Channel name is required",
			},
		})
		return
	}

	channelInfo, err := h.discoveryService.GetChannelInfo(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to get channel info", zap.Error(err), zap.String("channel", channelName))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "GET_CHANNEL_INFO_FAILED",
				"message": "Failed to get channel info from Gateway",
				"details": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    channelInfo,
	})
}

// GetTopology handles GET /api/v1/network/topology
// @Summary Get network topology
// @Description Get complete network topology (peers, orderers, CAs, channels, MSPs)
// @Tags network
// @Produce json
// @Success 200 {object} map[string]interface{} "Topology response"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/network/topology [get]
// @Security BearerAuth
func (h *Handler) GetTopology(w http.ResponseWriter, r *http.Request) {
	topology, err := h.discoveryService.GetTopology(r.Context())
	if err != nil {
		h.logger.Error("Failed to get topology", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "GET_TOPOLOGY_FAILED",
				"message": "Failed to get topology from Gateway",
				"details": err.Error(),
			},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    topology,
	})
}


