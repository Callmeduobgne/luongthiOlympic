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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/cache"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"go.uber.org/zap"
)

// BatchHandler handles tea batch operations
type BatchHandler struct {
	contract *fabric.ContractService
	cache    *cache.Service
	logger   *zap.Logger
}

// NewBatchHandler creates a new batch handler
func NewBatchHandler(contract *fabric.ContractService, cache *cache.Service, logger *zap.Logger) *BatchHandler {
	return &BatchHandler{
		contract: contract,
		cache:    cache,
		logger:   logger,
	}
}

// CreateBatch godoc
// @Summary Create new tea batch
// @Description Create a new tea batch on blockchain
// @Tags batches
// @Accept json
// @Produce json
// @Param batch body models.CreateBatchRequest true "Batch data"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Security BearerAuth
// @Router /batches [post]
func (h *BatchHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Create batch on blockchain
	batch, err := h.contract.CreateBatch(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create batch", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to create batch",
			err.Error(),
		))
		return
	}

	// Cache the batch
	cacheKey := fmt.Sprintf("batch:%s", batch.BatchID)
	_ = h.cache.SetJSON(r.Context(), cacheKey, batch, 5*time.Minute)

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(batch))
}

// GetBatchInfo godoc
// @Summary Get batch information
// @Description Get tea batch information by ID
// @Tags batches
// @Accept json
// @Produce json
// @Param id path string true "Batch ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /batches/{id} [get]
func (h *BatchHandler) GetBatchInfo(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "id")
	if batchID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Batch ID is required",
			nil,
		))
		return
	}

	// Try cache first
	cacheKey := fmt.Sprintf("batch:%s", batchID)
	var batch models.TeaBatch
	err := h.cache.GetJSON(r.Context(), cacheKey, &batch)
	if err == nil && batch.BatchID != "" {
		h.logger.Debug("Batch retrieved from cache", zap.String("batchId", batchID))
		respondJSON(w, http.StatusOK, models.NewSuccessResponse(batch))
		return
	}

	// Query from blockchain
	batchPtr, err := h.contract.GetBatchInfo(r.Context(), batchID)
	if err != nil {
		// Check if it's "not found" error
		if strings.Contains(err.Error(), "does not exist") {
			h.logger.Debug("Batch not found", zap.String("batchId", batchID))
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeBatchNotFound,
				fmt.Sprintf("Batch with ID '%s' not found", batchID),
				nil,
			))
			return
		}
		// Other errors
		h.logger.Error("Failed to get batch info", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to get batch info",
			err.Error(),
		))
		return
	}

	// Check if batch is nil (shouldn't happen with new chaincode, but safety check)
	if batchPtr == nil {
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeBatchNotFound,
			fmt.Sprintf("Batch with ID '%s' not found", batchID),
			nil,
		))
		return
	}

	// Cache the result
	_ = h.cache.SetJSON(r.Context(), cacheKey, batchPtr, 5*time.Minute)

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(batchPtr))
}

// VerifyBatch godoc
// @Summary Verify tea batch
// @Description Verify tea batch hash
// @Tags batches
// @Accept json
// @Produce json
// @Param id path string true "Batch ID"
// @Param request body models.VerifyBatchRequest true "Hash input"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /batches/{id}/verify [post]
func (h *BatchHandler) VerifyBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "id")
	if batchID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Batch ID is required",
			nil,
		))
		return
	}

	var req models.VerifyBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Verify batch on blockchain
	response, err := h.contract.VerifyBatch(r.Context(), batchID, req.HashInput)
	if err != nil {
		h.logger.Error("Failed to verify batch", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeVerificationFailed,
			"Failed to verify batch",
			err.Error(),
		))
		return
	}

	// Invalidate cache if batch was updated
	if response.IsValid && response.Batch.Status == models.StatusVerified {
		cacheKey := fmt.Sprintf("batch:%s", batchID)
		_ = h.cache.Delete(r.Context(), cacheKey)
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// UpdateBatchStatus godoc
// @Summary Update batch status
// @Description Update tea batch status
// @Tags batches
// @Accept json
// @Produce json
// @Param id path string true "Batch ID"
// @Param request body models.UpdateBatchStatusRequest true "Status update"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /batches/{id}/status [patch]
func (h *BatchHandler) UpdateBatchStatus(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "id")
	if batchID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Batch ID is required",
			nil,
		))
		return
	}

	var req models.UpdateBatchStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Update status on blockchain
	batch, err := h.contract.UpdateBatchStatus(r.Context(), batchID, req.Status)
	if err != nil {
		h.logger.Error("Failed to update batch status", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to update batch status",
			err.Error(),
		))
		return
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("batch:%s", batchID)
	_ = h.cache.Delete(r.Context(), cacheKey)

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(batch))
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

