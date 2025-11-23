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
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/rollback"
	"go.uber.org/zap"
)

// RollbackHandler handles rollback operations
type RollbackHandler struct {
	rollbackService *rollback.Service
	logger          *zap.Logger
}

// NewRollbackHandler creates a new rollback handler
func NewRollbackHandler(rollbackService *rollback.Service, logger *zap.Logger) *RollbackHandler {
	return &RollbackHandler{
		rollbackService: rollbackService,
		logger:          logger,
	}
}

// CreateRollbackRequest represents the request body for creating a rollback
type CreateRollbackRequest struct {
	ChaincodeName string                 `json:"chaincode_name"`
	ChannelName   string                 `json:"channel_name"`
	ToVersionID   *uuid.UUID             `json:"to_version_id,omitempty"` // Optional: if nil, rollback to previous
	Reason        *string                `json:"reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreateRollback creates a new rollback operation
func (h *RollbackHandler) CreateRollback(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		h.respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	// Parse request body
	var req CreateRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.ChaincodeName == "" {
		h.respondError(w, http.StatusBadRequest, "chaincode_name is required")
		return
	}
	if req.ChannelName == "" {
		h.respondError(w, http.StatusBadRequest, "channel_name is required")
		return
	}

	// Create rollback operation
	rollbackReq := &rollback.CreateRollbackRequest{
		ChaincodeName: req.ChaincodeName,
		ChannelName:   req.ChannelName,
		ToVersionID:   req.ToVersionID,
		Reason:        req.Reason,
		RequestedBy:   userID,
		Metadata:      req.Metadata,
	}

	op, err := h.rollbackService.CreateRollbackOperation(r.Context(), rollbackReq)
	if err != nil {
		h.logger.Error("Failed to create rollback operation", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to create rollback operation: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, op)
}

// ExecuteRollback executes a rollback operation
func (h *RollbackHandler) ExecuteRollback(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		h.respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	// Get operation ID from URL
	operationIDStr := chi.URLParam(r, "id")
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid operation ID")
		return
	}

	// Execute rollback
	if err := h.rollbackService.ExecuteRollback(r.Context(), operationID, userID); err != nil {
		h.logger.Error("Failed to execute rollback", zap.Error(err), zap.String("operation_id", operationID.String()))
		h.respondError(w, http.StatusInternalServerError, "failed to execute rollback: "+err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, "rollback executed successfully")
}

// GetRollback retrieves a rollback operation by ID
func (h *RollbackHandler) GetRollback(w http.ResponseWriter, r *http.Request) {
	// Get operation ID from URL
	operationIDStr := chi.URLParam(r, "id")
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid operation ID")
		return
	}

	// Get rollback operation
	op, err := h.rollbackService.GetRollbackOperation(r.Context(), operationID)
	if err != nil {
		h.logger.Error("Failed to get rollback operation", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "rollback operation not found")
		return
	}

	h.respondJSON(w, http.StatusOK, op)
}

// ListRollbacks lists rollback operations with filters
func (h *RollbackHandler) ListRollbacks(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	filters := &rollback.RollbackFilters{}

	if chaincodeName := r.URL.Query().Get("chaincode_name"); chaincodeName != "" {
		filters.ChaincodeName = &chaincodeName
	}

	if channelName := r.URL.Query().Get("channel_name"); channelName != "" {
		filters.ChannelName = &channelName
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 50 // Default limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// List rollback operations
	operations, err := h.rollbackService.ListRollbackOperations(r.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list rollback operations", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to list rollback operations: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"operations": operations,
		"count":      len(operations),
	})
}

// GetRollbackHistory retrieves rollback history for an operation
func (h *RollbackHandler) GetRollbackHistory(w http.ResponseWriter, r *http.Request) {
	// Get operation ID from URL
	operationIDStr := chi.URLParam(r, "id")
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid operation ID")
		return
	}

	// Get rollback history
	history, err := h.rollbackService.GetRollbackHistory(r.Context(), operationID)
	if err != nil {
		h.logger.Error("Failed to get rollback history", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get rollback history: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// CancelRollback cancels a pending rollback operation
func (h *RollbackHandler) CancelRollback(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		h.respondError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "invalid user ID")
		return
	}

	// Get operation ID from URL
	operationIDStr := chi.URLParam(r, "id")
	operationID, err := uuid.Parse(operationIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid operation ID")
		return
	}

	// Cancel rollback
	if err := h.rollbackService.CancelRollback(r.Context(), operationID, userID); err != nil {
		h.logger.Error("Failed to cancel rollback", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to cancel rollback: "+err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, "rollback cancelled successfully")
}

// Helper methods
func (h *RollbackHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RollbackHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

func (h *RollbackHandler) respondSuccess(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"success": true,
		"message": message,
	})
}

