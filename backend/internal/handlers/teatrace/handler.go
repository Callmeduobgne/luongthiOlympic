package teatrace

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/backend/internal/services/teatrace"
	"go.uber.org/zap"
)

// Handler handles Tea Traceability HTTP requests
type Handler struct {
	service *teatrace.Service
	logger  *zap.Logger
}

// NewHandler creates a new Tea Traceability handler
func NewHandler(service *teatrace.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// VerifyRequest represents the request body for verification
type VerifyRequest struct {
	Hash string `json:"hash"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// VerifyByHash verifies an entity by its hash
func (h *Handler) VerifyByHash(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if req.Hash == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Hash is required",
			},
		})
		return
	}

	result, err := h.service.VerifyByHash(r.Context(), req.Hash)
	if err != nil {
		h.logger.Error("Failed to verify by hash",
			zap.String("hash", req.Hash),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// CreateBatchRequest represents the request body for creating a batch
type CreateBatchRequest struct {
	BatchID       string `json:"batch_id"`
	FarmName      string `json:"farm_name"`
	HarvestDate   string `json:"harvest_date"`
	Certification string `json:"certification"`
	CertificateID string `json:"certificate_id"`
}

// CreateBatch creates a new tea batch
func (h *Handler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req CreateBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Basic validation
	if req.BatchID == "" || req.FarmName == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Batch ID and Farm Name are required",
			},
		})
		return
	}

	result, err := h.service.CreateBatch(r.Context(), req.BatchID, req.FarmName, req.HarvestDate, req.Certification, req.CertificateID)
	if err != nil {
		h.logger.Error("Failed to create batch",
			zap.String("batch_id", req.BatchID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetBatch retrieves a tea batch by ID
func (h *Handler) GetBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "id")
	if batchID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Batch ID is required",
			},
		})
		return
	}

	result, err := h.service.GetBatch(r.Context(), batchID)
	if err != nil {
		h.logger.Error("Failed to get batch",
			zap.String("batch_id", batchID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	if result == nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "Batch not found",
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// CreatePackageRequest represents the request body for creating a package
type CreatePackageRequest struct {
	PackageID      string  `json:"package_id"`
	BatchID        string  `json:"batch_id"`
	Weight         float64 `json:"weight"`
	ProductionDate string  `json:"production_date"`
	ExpiryDate     string  `json:"expiry_date"`
}

// CreatePackage creates a new tea package
func (h *Handler) CreatePackage(w http.ResponseWriter, r *http.Request) {
	var req CreatePackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	// Basic validation
	if req.PackageID == "" || req.BatchID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Package ID and Batch ID are required",
			},
		})
		return
	}

	result, err := h.service.CreatePackage(r.Context(), req.PackageID, req.BatchID, req.Weight, req.ProductionDate, req.ExpiryDate)
	if err != nil {
		h.logger.Error("Failed to create package",
			zap.String("package_id", req.PackageID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetPackage retrieves a tea package by ID
func (h *Handler) GetPackage(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "packageId")
	if packageID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Package ID is required",
			},
		})
		return
	}

	result, err := h.service.GetPackage(r.Context(), packageID)
	if err != nil {
		h.logger.Error("Failed to get package",
			zap.String("package_id", packageID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	if result == nil {
		respondJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "Package not found",
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// GetPackagesByBatch retrieves all packages for a specific batch
func (h *Handler) GetPackagesByBatch(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Batch ID is required",
			},
		})
		return
	}

	// Parse pagination parameters
	limit := 100
	offset := 0
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	packages, total, err := h.service.GetPackagesByBatch(r.Context(), batchID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get packages by batch",
			zap.String("batch_id", batchID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"packages": packages,
			"total":    total,
		},
	})
}
