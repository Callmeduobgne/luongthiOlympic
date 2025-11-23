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

package channel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	channelservice "github.com/ibn-network/api-gateway/internal/services/channel"
	"go.uber.org/zap"
)

// ChannelHandler handles channel management operations
type ChannelHandler struct {
	channelService *channelservice.Service
	logger         *zap.Logger
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(channelService *channelservice.Service, logger *zap.Logger) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
		logger:         logger,
	}
}

// CreateChannel godoc
// @Summary Create new channel
// @Description Create a new channel in the network
// @Tags channel
// @Accept json
// @Produce json
// @Param request body models.CreateChannelRequest true "Channel creation request"
// @Success 201 {object} models.APIResponse{data=models.CreateChannelResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels [post]
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var req models.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	if req.Consortium == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Consortium is required",
			nil,
		))
		return
	}

	if len(req.Organizations) == 0 {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"At least one organization is required",
			nil,
		))
		return
	}

	response, err := h.channelService.CreateChannel(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create channel", zap.String("channel", req.Name), zap.Error(err))
		if strings.Contains(err.Error(), "already exists") {
			respondJSON(w, http.StatusConflict, models.NewErrorResponse(
				models.ErrCodeConflict,
				fmt.Sprintf("Channel '%s' already exists", req.Name),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create channel",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(response))
}

// UpdateChannelConfig godoc
// @Summary Update channel configuration
// @Description Update channel configuration
// @Tags channel
// @Accept json
// @Produce json
// @Param name path string true "Channel name"
// @Param request body models.UpdateChannelConfigRequest true "Config update request"
// @Success 200 {object} models.APIResponse{data=models.UpdateChannelConfigResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels/{name}/config [patch]
func (h *ChannelHandler) UpdateChannelConfig(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	var req models.UpdateChannelConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if req.ConfigUpdate == nil || len(req.ConfigUpdate) == 0 {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Config update is required",
			nil,
		))
		return
	}

	response, err := h.channelService.UpdateChannelConfig(r.Context(), channelName, &req)
	if err != nil {
		h.logger.Error("Failed to update channel config", zap.String("channel", channelName), zap.Error(err))
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
			"Failed to update channel config",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// JoinPeer godoc
// @Summary Join peer to channel
// @Description Join a peer to a channel
// @Tags channel
// @Accept json
// @Produce json
// @Param name path string true "Channel name"
// @Param request body models.JoinChannelRequest true "Join request"
// @Success 200 {object} models.APIResponse{data=models.JoinChannelResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels/{name}/join [post]
func (h *ChannelHandler) JoinPeer(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	var req models.JoinChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	if req.PeerAddress == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Peer address is required",
			nil,
		))
		return
	}

	response, err := h.channelService.JoinPeer(r.Context(), channelName, &req)
	if err != nil {
		h.logger.Error("Failed to join peer to channel", zap.String("channel", channelName), zap.String("peer", req.PeerAddress), zap.Error(err))
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
			"Failed to join peer to channel",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// ListChannelMembers godoc
// @Summary List channel members
// @Description List all members (organizations) in a channel
// @Tags channel
// @Accept json
// @Produce json
// @Param name path string true "Channel name"
// @Success 200 {object} models.APIResponse{data=models.ListChannelMembersResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels/{name}/members [get]
func (h *ChannelHandler) ListChannelMembers(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	response, err := h.channelService.ListChannelMembers(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to list channel members", zap.String("channel", channelName), zap.Error(err))
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
			"Failed to list channel members",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// ListChannelPeers godoc
// @Summary List peers in channel
// @Description List all peers in a channel
// @Tags channel
// @Accept json
// @Produce json
// @Param name path string true "Channel name"
// @Success 200 {object} models.APIResponse{data=models.ListChannelPeersResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels/{name}/peers [get]
func (h *ChannelHandler) ListChannelPeers(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	response, err := h.channelService.ListChannelPeers(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to list channel peers", zap.String("channel", channelName), zap.Error(err))
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
			"Failed to list channel peers",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// respondJSON is a helper function to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

