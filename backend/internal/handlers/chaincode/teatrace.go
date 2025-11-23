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

package chaincode

import (
	"encoding/json"
	"net/http"

	"github.com/ibn-network/backend/internal/services/blockchain/chaincode"
	"github.com/ibn-network/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// TeaTraceHandler handles HTTP requests for TeaTrace chaincode
type TeaTraceHandler struct {
	service chaincode.TeaTraceService // Use interface to support both implementations
	logger  *zap.Logger
}

// NewTeaTraceHandler creates a new TeaTrace handler
func NewTeaTraceHandler(service chaincode.TeaTraceService, logger *zap.Logger) *TeaTraceHandler {
	return &TeaTraceHandler{
		service: service,
		logger:  logger,
	}
}

// CreateBatch handles creating a new tea batch
func (h *TeaTraceHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req chaincode.CreateBatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(&req); validationErrors != nil {
		h.logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	txID, err := h.service.CreateBatch(r.Context(), req.BatchID, req.FarmName, req.HarvestDate, req.Certification, req.CertificateID)
	if err != nil {
		h.logger.Error("Failed to create batch", zap.Error(err))
		http.Error(w, "Failed to create batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"tx_id":    txID,
		"batch_id": req.BatchID,
		"message":  "Batch created successfully",
	})
}

// GetBatch handles retrieving a tea batch by ID
func (h *TeaTraceHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		http.Error(w, "Batch ID is required", http.StatusBadRequest)
		return
	}

	batch, err := h.service.GetBatch(r.Context(), batchID)
	if err != nil {
		h.logger.Error("Failed to get batch", zap.String("batch_id", batchID), zap.Error(err))
		http.Error(w, "Batch not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batch)
}

// GetAllBatches handles retrieving all tea batches
func (h *TeaTraceHandler) GetAllBatches(w http.ResponseWriter, r *http.Request) {
	batches, err := h.service.GetAllBatches(r.Context())
	if err != nil {
		h.logger.Error("Failed to get all batches", zap.Error(err))
		http.Error(w, "Failed to get batches", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"batches": batches,
		"count":   len(batches),
	})
}

// VerifyBatch handles verifying a tea batch
func (h *TeaTraceHandler) VerifyBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		http.Error(w, "Batch ID is required", http.StatusBadRequest)
		return
	}

	var req chaincode.VerifyBatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.BatchID = batchID // Set from URL param

	// Validate request
	if validationErrors := utils.ValidateStruct(&req); validationErrors != nil {
		h.logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	txID, err := h.service.VerifyBatch(r.Context(), req.BatchID, req.VerificationHash)
	if err != nil {
		h.logger.Error("Failed to verify batch", zap.Error(err))
		http.Error(w, "Failed to verify batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"tx_id":    txID,
		"batch_id": req.BatchID,
		"message":  "Batch verified successfully",
	})
}

// UpdateBatchStatus handles updating the status of a tea batch
func (h *TeaTraceHandler) UpdateBatchStatus(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		http.Error(w, "Batch ID is required", http.StatusBadRequest)
		return
	}

	var req chaincode.UpdateBatchStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.BatchID = batchID // Set from URL param

	// Validate request
	if validationErrors := utils.ValidateStruct(&req); validationErrors != nil {
		h.logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	txID, err := h.service.UpdateBatchStatus(r.Context(), req.BatchID, req.Status)
	if err != nil {
		h.logger.Error("Failed to update batch status", zap.Error(err))
		http.Error(w, "Failed to update batch status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"tx_id":    txID,
		"batch_id": batchID,
		"status":   req.Status,
		"message":  "Batch status updated successfully",
	})
}

// HealthCheck handles chaincode health check
func (h *TeaTraceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		h.logger.Error("Chaincode health check failed", zap.Error(err))
		http.Error(w, "Chaincode unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"message": "Chaincode is operational",
	})
}

