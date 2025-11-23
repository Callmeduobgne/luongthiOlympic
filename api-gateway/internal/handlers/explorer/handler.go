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

package explorer

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/explorer"
	"go.uber.org/zap"
)

// Handler handles block explorer operations
type Handler struct {
	explorerService *explorer.Service
	logger          *zap.Logger
}

// NewHandler creates a new block explorer handler
func NewHandler(explorerService *explorer.Service, logger *zap.Logger) *Handler {
	return &Handler{
		explorerService: explorerService,
		logger:         logger,
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// GetBlock godoc
// @Summary Get block by number
// @Description Get block information by block number
// @Tags explorer
// @Produce json
// @Param channel path string true "Channel name"
// @Param number path int true "Block number"
// @Success 200 {object} models.APIResponse{data=models.BlockInfo}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /blocks/{channel}/{number} [get]
func (h *Handler) GetBlock(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	blockNumberStr := chi.URLParam(r, "number")

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid block number",
			err.Error(),
		))
		return
	}

	block, err := h.explorerService.GetBlock(r.Context(), channelName, blockNumber)
	if err != nil {
		h.logger.Error("Failed to get block",
			zap.String("channel", channelName),
			zap.Uint64("blockNumber", blockNumber),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get block",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(block))
}

// GetLatestBlock godoc
// @Summary Get latest block
// @Description Get the latest block information
// @Tags explorer
// @Produce json
// @Param channel path string true "Channel name"
// @Success 200 {object} models.APIResponse{data=models.BlockInfo}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /blocks/{channel}/latest [get]
func (h *Handler) GetLatestBlock(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")

	block, err := h.explorerService.GetLatestBlock(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to get latest block",
			zap.String("channel", channelName),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get latest block",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(block))
}

// ListBlocks godoc
// @Summary List blocks
// @Description List blocks with pagination
// @Tags explorer
// @Produce json
// @Param channel path string true "Channel name"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} models.APIResponse{data=object}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /blocks/{channel} [get]
func (h *Handler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	blocks, total, err := h.explorerService.ListBlocks(r.Context(), channelName, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list blocks",
			zap.String("channel", channelName),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list blocks",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"blocks": blocks,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}))
}

// GetTransactionByBlock godoc
// @Summary Get transactions by block
// @Description Get all transactions in a block
// @Tags explorer
// @Produce json
// @Param channel path string true "Channel name"
// @Param number path int true "Block number"
// @Success 200 {object} models.APIResponse{data=[]models.Transaction}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /blocks/{channel}/{number}/transactions [get]
func (h *Handler) GetTransactionByBlock(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	blockNumberStr := chi.URLParam(r, "number")

	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid block number",
			err.Error(),
		))
		return
	}

	transactions, err := h.explorerService.GetTransactionByBlock(r.Context(), channelName, blockNumber)
	if err != nil {
		h.logger.Error("Failed to get transactions by block",
			zap.String("channel", channelName),
			zap.Uint64("blockNumber", blockNumber),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get transactions",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(transactions))
}

