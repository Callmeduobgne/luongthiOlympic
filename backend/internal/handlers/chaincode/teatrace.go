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
	service            chaincode.TeaTraceService       // Use interface to support both implementations
	verifyService      *chaincode.VerifyService        // Service for hash verification (legacy)
	merkleVerifyService *chaincode.MerkleVerifyService // Service for Merkle proof verification (new)
	logger             *zap.Logger
}

// NewTeaTraceHandler creates a new TeaTrace handler
func NewTeaTraceHandler(service chaincode.TeaTraceService, logger *zap.Logger) *TeaTraceHandler {
	return &TeaTraceHandler{
		service: service,
		logger:  logger,
	}
}

// NewTeaTraceHandlerWithVerify creates a new TeaTrace handler with verify service
func NewTeaTraceHandlerWithVerify(service chaincode.TeaTraceService, verifyService *chaincode.VerifyService, logger *zap.Logger) *TeaTraceHandler {
	return &TeaTraceHandler{
		service:       service,
		verifyService: verifyService,
		logger:        logger,
	}
}

// NewTeaTraceHandlerWithMerkleVerify creates a new TeaTrace handler with Merkle verify service
// This is the recommended constructor for new deployments
func NewTeaTraceHandlerWithMerkleVerify(
	service chaincode.TeaTraceService,
	verifyService *chaincode.VerifyService,
	merkleVerifyService *chaincode.MerkleVerifyService,
	logger *zap.Logger,
) *TeaTraceHandler {
	return &TeaTraceHandler{
		service:             service,
		verifyService:       verifyService,
		merkleVerifyService: merkleVerifyService,
		logger:              logger,
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

	// Check if JWT token is in context
	ctx := r.Context()
	if token, ok := ctx.Value("jwt_token").(string); ok {
		h.logger.Info("JWT token found in context", zap.String("token_preview", token[:20]+"..."))
	} else {
		h.logger.Warn("No JWT token in context for CreateBatch")
	}
	
	txID, err := h.service.CreateBatch(ctx, req.BatchID, req.FarmName, req.HarvestDate, req.Certification, req.CertificateID)
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

// CreatePackage handles creating a new tea package
// POST /api/v1/teatrace/packages
func (h *TeaTraceHandler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PackageID     string  `json:"package_id" validate:"required"`
		BatchID       string  `json:"batch_id" validate:"required"`
		Weight        float64 `json:"weight" validate:"required,gt=0"`
		ProductionDate string `json:"production_date" validate:"required"`
		ExpiryDate    string  `json:"expiry_date,omitempty"`
	}

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

	txID, err := h.service.CreatePackage(r.Context(), req.PackageID, req.BatchID, req.Weight, req.ProductionDate, req.ExpiryDate)
	if err != nil {
		h.logger.Error("Failed to create package", zap.Error(err))
		http.Error(w, "Failed to create package: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"tx_id":     txID,
		"package_id": req.PackageID,
		"message":   "Package created successfully",
	})
}

// GetPackage handles retrieving a tea package by ID
// GET /api/v1/teatrace/packages/:packageId
func (h *TeaTraceHandler) GetPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "packageId")
	if packageID == "" {
		http.Error(w, "Package ID is required", http.StatusBadRequest)
		return
	}

	pkg, err := h.service.GetPackage(r.Context(), packageID)
	if err != nil {
		h.logger.Error("Failed to get package", 
			zap.String("package_id", packageID), 
			zap.Error(err))
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pkg)
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

// VerifyByHash handles product verification by hash/blockhash
// Now supports Merkle proof for ultra-fast verification (< 10ms)
// POST /api/v1/teatrace/verify-by-hash
func (h *TeaTraceHandler) VerifyByHash(w http.ResponseWriter, r *http.Request) {
	var req chaincode.VerifyByHashRequest

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

	var result *chaincode.VerifyByHashResponse
	var err error

	// Priority 1: Try Merkle proof verification if available
	useMerkleProof := h.merkleVerifyService != nil && len(req.MerkleProof) > 0 && req.MerkleRoot != ""
	
	if useMerkleProof {
		h.logger.Info("Using Merkle proof verification",
			zap.String("hash", req.Hash[:min(16, len(req.Hash))]),
			zap.Int("proof_steps", len(req.MerkleProof)),
		)
		
		result, err = h.merkleVerifyService.VerifyWithMerkleProof(
			r.Context(),
			req.Hash,
			req.MerkleProof,
			req.MerkleRoot,
			req.BlockNumber,
		)
		
		if err == nil {
			// Success with Merkle proof
			h.logger.Info("Verification completed via Merkle proof",
				zap.String("hash", req.Hash[:min(16, len(req.Hash))]),
				zap.Bool("is_valid", result.IsValid),
			)
		} else {
			// Merkle proof failed, will fallback to legacy
			h.logger.Warn("Merkle proof verification failed, falling back to legacy",
				zap.Error(err),
			)
			useMerkleProof = false // Force fallback
		}
	}
	
	// Priority 2: Legacy hash verification (if Merkle proof not used or failed)
	if !useMerkleProof || err != nil {
		if h.verifyService == nil {
			h.logger.Error("No verification service available")
			http.Error(w, "Service unavailable", http.StatusInternalServerError)
			return
		}

		h.logger.Info("Using legacy hash verification",
			zap.String("hash", req.Hash[:min(16, len(req.Hash))]),
		)

		result, err = h.verifyService.VerifyByHash(r.Context(), req.Hash)
		if err != nil {
			h.logger.Error("Failed to verify hash", zap.Error(err))
			http.Error(w, "Failed to verify hash", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

