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
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	networkservice "github.com/ibn-network/api-gateway/internal/services/network"
	"go.uber.org/zap"
)

// DiscoveryHandler handles network discovery operations
type DiscoveryHandler struct {
	discoveryService *networkservice.DiscoveryService
	logger           *zap.Logger
}

// NewDiscoveryHandler creates a new discovery handler
func NewDiscoveryHandler(discoveryService *networkservice.DiscoveryService, logger *zap.Logger) *DiscoveryHandler {
	return &DiscoveryHandler{
		discoveryService: discoveryService,
		logger:           logger,
	}
}


// ListPeers godoc
// @Summary List peers
// @Description List all peers in the network
// @Tags network
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.PeerInfo}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/peers [get]
func (h *DiscoveryHandler) ListPeers(w http.ResponseWriter, r *http.Request) {
	peers, err := h.discoveryService.ListPeers(r.Context())
	if err != nil {
		h.logger.Error("Failed to list peers", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list peers",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(peers))
}

// GetPeer godoc
// @Summary Get peer details
// @Description Get peer information by ID
// @Tags network
// @Produce json
// @Param id path string true "Peer ID or name"
// @Success 200 {object} models.APIResponse{data=models.PeerInfo}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/peers/{id} [get]
func (h *DiscoveryHandler) GetPeer(w http.ResponseWriter, r *http.Request) {
	peerID := chi.URLParam(r, "id")
	if peerID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Peer ID is required",
			nil,
		))
		return
	}

	peer, err := h.discoveryService.GetPeer(r.Context(), peerID)
	if err != nil {
		h.logger.Error("Failed to get peer", zap.String("peerId", peerID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Peer '%s' not found", peerID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get peer",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(peer))
}

// ListOrderers godoc
// @Summary List orderers
// @Description List all orderers in the network
// @Tags network
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.OrdererInfo}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/orderers [get]
func (h *DiscoveryHandler) ListOrderers(w http.ResponseWriter, r *http.Request) {
	orderers, err := h.discoveryService.ListOrderers(r.Context())
	if err != nil {
		h.logger.Error("Failed to list orderers", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list orderers",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(orderers))
}

// GetOrderer godoc
// @Summary Get orderer details
// @Description Get orderer information by ID
// @Tags network
// @Produce json
// @Param id path string true "Orderer ID or name"
// @Success 200 {object} models.APIResponse{data=models.OrdererInfo}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/orderers/{id} [get]
func (h *DiscoveryHandler) GetOrderer(w http.ResponseWriter, r *http.Request) {
	ordererID := chi.URLParam(r, "id")
	if ordererID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Orderer ID is required",
			nil,
		))
		return
	}

	orderer, err := h.discoveryService.GetOrderer(r.Context(), ordererID)
	if err != nil {
		h.logger.Error("Failed to get orderer", zap.String("ordererId", ordererID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Orderer '%s' not found", ordererID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get orderer",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(orderer))
}

// ListCAs godoc
// @Summary List CAs
// @Description List all Fabric CAs in the network
// @Tags network
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.CAInfo}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/cas [get]
func (h *DiscoveryHandler) ListCAs(w http.ResponseWriter, r *http.Request) {
	cas, err := h.discoveryService.ListCAs(r.Context())
	if err != nil {
		h.logger.Error("Failed to list CAs", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list CAs",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(cas))
}

// GetTopology godoc
// @Summary Get network topology
// @Description Get network topology (peers, orderers, CAs, channels)
// @Tags network
// @Produce json
// @Success 200 {object} models.APIResponse{data=models.NetworkTopology}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/topology [get]
func (h *DiscoveryHandler) GetTopology(w http.ResponseWriter, r *http.Request) {
	topology, err := h.discoveryService.GetTopology(r.Context())
	if err != nil {
		h.logger.Error("Failed to get topology", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get topology",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(topology))
}

// GetPeersInChannel godoc
// @Summary Get peers in channel
// @Description Get all peers in a specific channel
// @Tags network
// @Produce json
// @Param channel path string true "Channel name"
// @Success 200 {object} models.APIResponse{data=[]models.PeerInfo}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels/{channel}/peers [get]
func (h *DiscoveryHandler) GetPeersInChannel(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	peers, err := h.discoveryService.GetPeersInChannel(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to get peers in channel", zap.String("channel", channelName), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Channel '%s' not found", channelName),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get peers in channel",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(peers))
}

// CheckPeerHealth godoc
// @Summary Check peer health
// @Description Check health status of a peer
// @Tags network
// @Produce json
// @Param id path string true "Peer ID or name"
// @Success 200 {object} models.APIResponse{data=models.HealthStatus}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/health/peers/{id} [get]
func (h *DiscoveryHandler) CheckPeerHealth(w http.ResponseWriter, r *http.Request) {
	peerID := chi.URLParam(r, "id")
	if peerID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Peer ID is required",
			nil,
		))
		return
	}

	health, err := h.discoveryService.CheckPeerHealth(r.Context(), peerID)
	if err != nil {
		h.logger.Error("Failed to check peer health", zap.String("peerId", peerID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Peer '%s' not found", peerID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to check peer health",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(health))
}

// CheckAllPeersHealth godoc
// @Summary Check all peers health
// @Description Check health status of all peers
// @Tags network
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.HealthStatus}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/health/peers [get]
func (h *DiscoveryHandler) CheckAllPeersHealth(w http.ResponseWriter, r *http.Request) {
	healthStatuses, err := h.discoveryService.CheckAllPeersHealth(r.Context())
	if err != nil {
		h.logger.Error("Failed to check all peers health", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to check peers health",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(healthStatuses))
}

// CheckOrdererHealth godoc
// @Summary Check orderer health
// @Description Check health status of an orderer
// @Tags network
// @Produce json
// @Param id path string true "Orderer ID or name"
// @Success 200 {object} models.APIResponse{data=models.HealthStatus}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/health/orderers/{id} [get]
func (h *DiscoveryHandler) CheckOrdererHealth(w http.ResponseWriter, r *http.Request) {
	ordererID := chi.URLParam(r, "id")
	if ordererID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Orderer ID is required",
			nil,
		))
		return
	}

	health, err := h.discoveryService.CheckOrdererHealth(r.Context(), ordererID)
	if err != nil {
		h.logger.Error("Failed to check orderer health", zap.String("ordererId", ordererID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Orderer '%s' not found", ordererID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to check orderer health",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(health))
}

// CheckAllOrderersHealth godoc
// @Summary Check all orderers health
// @Description Check health status of all orderers
// @Tags network
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]models.HealthStatus}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/health/orderers [get]
func (h *DiscoveryHandler) CheckAllOrderersHealth(w http.ResponseWriter, r *http.Request) {
	healthStatuses, err := h.discoveryService.CheckAllOrderersHealth(r.Context())
	if err != nil {
		h.logger.Error("Failed to check all orderers health", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to check orderers health",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(healthStatuses))
}

