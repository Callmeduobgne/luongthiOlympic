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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/network"
	"go.uber.org/zap"
)

// NetworkHandler handles network information queries
type NetworkHandler struct {
	networkService *network.Service
	logger         *zap.Logger
}

// NewNetworkHandler creates a new network handler
func NewNetworkHandler(networkService *network.Service, logger *zap.Logger) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
		logger:         logger,
	}
}

// ListChannels godoc
// @Summary List channels
// @Description List all channels accessible by the gateway
// @Tags network
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels [get]
func (h *NetworkHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := h.networkService.ListChannels(r.Context())
	if err != nil {
		h.logger.Error("Failed to list channels", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to list channels",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(channels))
}

// GetChannelInfo godoc
// @Summary Get channel information
// @Description Get detailed information about a channel
// @Tags network
// @Accept json
// @Produce json
// @Param name path string true "Channel name"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels/{name} [get]
func (h *NetworkHandler) GetChannelInfo(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	info, err := h.networkService.GetChannelInfo(r.Context(), channelName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Channel '%s' not found", channelName),
				err.Error(),
			))
			return
		}
		h.logger.Error("Failed to get channel info", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to get channel info",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(info))
}

// GetChannelConfig godoc
// @Summary Get channel configuration
// @Description Get channel configuration details
// @Tags network
// @Accept json
// @Produce json
// @Param name path string true "Channel name"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels/{name}/config [get]
func (h *NetworkHandler) GetChannelConfig(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "name")
	if channelName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name is required",
			nil,
		))
		return
	}

	config, err := h.networkService.GetChannelConfig(r.Context(), channelName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Channel '%s' not found", channelName),
				err.Error(),
			))
			return
		}
		h.logger.Error("Failed to get channel config", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to get channel config",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(config))
}

// GetNetworkInfo godoc
// @Summary Get network information
// @Description Get overall network information (channels, peers, orderers)
// @Tags network
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/info [get]
func (h *NetworkHandler) GetNetworkInfo(w http.ResponseWriter, r *http.Request) {
	info, err := h.networkService.GetNetworkInfo(r.Context())
	if err != nil {
		h.logger.Error("Failed to get network info", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to get network info",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(info))
}

// GetBlockInfo godoc
// @Summary Get block information
// @Description Get block information by block number
// @Tags network
// @Accept json
// @Produce json
// @Param channel path string true "Channel name"
// @Param number path string true "Block number"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels/{channel}/blocks/{number} [get]
func (h *NetworkHandler) GetBlockInfo(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	blockNumberStr := chi.URLParam(r, "number")

	if channelName == "" || blockNumberStr == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name and block number are required",
			nil,
		))
		return
	}

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid block number",
			err.Error(),
		))
		return
	}

	blockInfo, err := h.networkService.GetBlockInfo(r.Context(), channelName, blockNumber)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Block %d not found on channel '%s'", blockNumber, channelName),
				err.Error(),
			))
			return
		}
		h.logger.Error("Failed to get block info", zap.Error(err))
		respondJSON(w, http.StatusNotImplemented, models.NewErrorResponse(
			"NOT_IMPLEMENTED",
			"Getting block by number requires peer query or admin API",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(blockInfo))
}

// GetTransactionInfo godoc
// @Summary Get transaction information
// @Description Get transaction information by transaction ID
// @Tags network
// @Accept json
// @Produce json
// @Param channel path string true "Channel name"
// @Param txid path string true "Transaction ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels/{channel}/transactions/{txid} [get]
func (h *NetworkHandler) GetTransactionInfo(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	txID := chi.URLParam(r, "txid")

	if channelName == "" || txID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name and transaction ID are required",
			nil,
		))
		return
	}

	txInfo, err := h.networkService.GetTransactionInfo(r.Context(), channelName, txID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Transaction '%s' not found on channel '%s'", txID, channelName),
				err.Error(),
			))
			return
		}
		h.logger.Error("Failed to get transaction info", zap.Error(err))
		respondJSON(w, http.StatusNotImplemented, models.NewErrorResponse(
			"NOT_IMPLEMENTED",
			"Getting transaction by ID requires peer query or admin API",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(txInfo))
}

// GetChaincodeInfoOnChannel godoc
// @Summary Get chaincode information on channel
// @Description Get chaincode information on a specific channel
// @Tags network
// @Accept json
// @Produce json
// @Param channel path string true "Channel name"
// @Param chaincode path string true "Chaincode name"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /network/channels/{channel}/chaincode/{chaincode} [get]
func (h *NetworkHandler) GetChaincodeInfoOnChannel(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	chaincodeName := chi.URLParam(r, "chaincode")

	if channelName == "" || chaincodeName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name and chaincode name are required",
			nil,
		))
		return
	}

	info, err := h.networkService.GetChaincodeInfoOnChannel(r.Context(), channelName, chaincodeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Chaincode '%s' not found on channel '%s'", chaincodeName, channelName),
				err.Error(),
			))
			return
		}
		h.logger.Error("Failed to get chaincode info", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to get chaincode info",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(info))
}

// respondJSON is a helper function to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

