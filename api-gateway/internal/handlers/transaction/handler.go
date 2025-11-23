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

package transaction

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/transaction"
	"go.uber.org/zap"
)

// TransactionHandler handles transaction operations
type TransactionHandler struct {
	transactionService *transaction.Service
	logger             *zap.Logger
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transactionService *transaction.Service, logger *zap.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

// SubmitTransaction godoc
// @Summary Submit a transaction
// @Description Submit a new transaction to the blockchain and track it
// @Tags transactions
// @Accept json
// @Produce json
// @Param request body models.TransactionRequest true "Transaction request"
// @Success 201 {object} models.APIResponse{data=models.TransactionResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /transactions [post]
func (h *TransactionHandler) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Get user ID and API key ID from context (set by auth middleware)
	userID := ""
	if uid, ok := r.Context().Value("userID").(string); ok {
		userID = uid
	}

	apiKeyID := ""
	if akid, ok := r.Context().Value("apiKeyID").(string); ok {
		apiKeyID = akid
	}

	// Extract user certificate from headers (sent by Backend)
	userCert := r.Header.Get("X-User-Cert")
	userKey := r.Header.Get("X-User-Key")
	userMSPID := r.Header.Get("X-User-MSPID")

	// Add cert to context for Gateway service
	ctx := r.Context()
	if userCert != "" {
		ctx = context.WithValue(ctx, "user_cert", userCert)
	}
	if userKey != "" {
		ctx = context.WithValue(ctx, "user_key", userKey)
	}
	if userMSPID != "" {
		ctx = context.WithValue(ctx, "user_msp_id", userMSPID)
	}

	// Submit transaction
	response, err := h.transactionService.SubmitTransaction(ctx, &req, userID, apiKeyID)
	if err != nil {
		h.logger.Error("Failed to submit transaction",
			zap.String("channel", req.ChannelName),
			zap.String("chaincode", req.ChaincodeName),
			zap.String("function", req.FunctionName),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to submit transaction",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(response))
}

// GetTransaction godoc
// @Summary Get transaction details
// @Description Get transaction details by ID or TxID
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID or TxID"
// @Success 200 {object} models.APIResponse{data=models.Transaction}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	idOrTxID := chi.URLParam(r, "id")
	if idOrTxID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Transaction ID is required",
			nil,
		))
		return
	}

	tx, err := h.transactionService.GetTransaction(r.Context(), idOrTxID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeTransactionNotFound,
				"Transaction not found",
				err.Error(),
			))
		} else {
			h.logger.Error("Failed to get transaction", zap.String("id", idOrTxID), zap.Error(err))
			respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeInternalError,
				"Failed to get transaction",
				err.Error(),
			))
		}
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(tx))
}

// GetTransactionStatus godoc
// @Summary Get transaction status
// @Description Get current transaction status
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID or TxID"
// @Success 200 {object} models.APIResponse{data=object}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /transactions/{id}/status [get]
func (h *TransactionHandler) GetTransactionStatus(w http.ResponseWriter, r *http.Request) {
	idOrTxID := chi.URLParam(r, "id")
	if idOrTxID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Transaction ID is required",
			nil,
		))
		return
	}

	status, err := h.transactionService.GetTransactionStatus(r.Context(), idOrTxID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeTransactionNotFound,
				"Transaction not found",
				err.Error(),
			))
		} else {
			h.logger.Error("Failed to get transaction status", zap.String("id", idOrTxID), zap.Error(err))
			respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeInternalError,
				"Failed to get transaction status",
				err.Error(),
			))
		}
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"status": status,
	}))
}

// GetTransactionReceipt godoc
// @Summary Get transaction receipt
// @Description Get transaction receipt with block information
// @Tags transactions
// @Produce json
// @Param id path string true "Transaction ID or TxID"
// @Success 200 {object} models.APIResponse{data=models.TransactionReceipt}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /transactions/{id}/receipt [get]
func (h *TransactionHandler) GetTransactionReceipt(w http.ResponseWriter, r *http.Request) {
	idOrTxID := chi.URLParam(r, "id")
	if idOrTxID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Transaction ID is required",
			nil,
		))
		return
	}

	receipt, err := h.transactionService.GetTransactionReceipt(r.Context(), idOrTxID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeTransactionNotFound,
				"Transaction not found",
				err.Error(),
			))
		} else {
			h.logger.Error("Failed to get transaction receipt", zap.String("id", idOrTxID), zap.Error(err))
			respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeInternalError,
				"Failed to get transaction receipt",
				err.Error(),
			))
		}
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(receipt))
}

// ListTransactions godoc
// @Summary List transactions
// @Description List transactions with optional filters
// @Tags transactions
// @Produce json
// @Param channel query string false "Channel name"
// @Param chaincode query string false "Chaincode name"
// @Param status query string false "Transaction status"
// @Param userId query string false "User ID"
// @Param limit query int false "Limit (default: 50)"
// @Param offset query int false "Offset (default: 0)"
// @Param startTime query string false "Start time (RFC3339)"
// @Param endTime query string false "End time (RFC3339)"
// @Success 200 {object} models.APIResponse{data=object}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /transactions [get]
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	query := &models.TransactionListQuery{}

	// Parse query parameters
	if channelName := r.URL.Query().Get("channel"); channelName != "" {
		query.ChannelName = channelName
	}
	if chaincodeName := r.URL.Query().Get("chaincode"); chaincodeName != "" {
		query.ChaincodeName = chaincodeName
	}
	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = models.TransactionStatus(status)
	}
	if userID := r.URL.Query().Get("userId"); userID != "" {
		query.UserID = userID
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			query.Offset = offset
		}
	}
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}
	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	// List transactions
	transactions, total, err := h.transactionService.ListTransactions(r.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list transactions", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list transactions",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"transactions": transactions,
		"total":        total,
		"limit":        query.Limit,
		"offset":       query.Offset,
	}))
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

