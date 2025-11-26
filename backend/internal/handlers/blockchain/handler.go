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

package blockchain

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ibn-network/backend/internal/services/blockchain/transaction"
	"github.com/ibn-network/backend/internal/services/certificate"
	"github.com/ibn-network/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles blockchain-related HTTP requests
type Handler struct {
	txService      *transaction.Service
	certService   *certificate.Service
	logger         *zap.Logger
}

// NewHandler creates a new blockchain handler
func NewHandler(txService *transaction.Service, certService *certificate.Service, logger *zap.Logger) *Handler {
	return &Handler{
		txService:    txService,
		certService:  certService,
		logger:       logger,
	}
}

// SubmitTransaction handles submitting a transaction to the blockchain
func (h *Handler) SubmitTransaction(w http.ResponseWriter, r *http.Request) {
	var req transaction.SubmitTransactionRequest
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

	// Get user ID from context (set by auth middleware)
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert user_id to UUID (handle both string and UUID types)
	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		var err error
		userID, err = uuid.Parse(v)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}
	default:
		http.Error(w, "Invalid user ID type", http.StatusUnauthorized)
		return
	}

	// Get user certificate with private key (for Gateway submission)
	ctx := r.Context()
	if h.certService != nil {
		certWithKey, err := h.certService.GetActiveCertificateWithKey(ctx, userID)
		if err != nil {
			h.logger.Warn("Failed to get user certificate",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
			// Continue without cert - transaction service will handle error
		} else {
			// Set certificate and private key in context for transaction service
			ctx = context.WithValue(ctx, "user_cert", certWithKey.Certificate)
			ctx = context.WithValue(ctx, "user_key", certWithKey.PrivateKey)
			ctx = context.WithValue(ctx, "user_msp_id", certWithKey.MSPID)
			
			h.logger.Debug("User certificate loaded",
				zap.String("user_id", userID.String()),
				zap.String("cert_id", certWithKey.ID.String()),
				zap.String("msp_id", certWithKey.MSPID),
			)
		}
	}

	tx, err := h.txService.SubmitTransaction(ctx, userID, &req)
	if err != nil {
		h.logger.Error("Failed to submit transaction", zap.Error(err))
		http.Error(w, "Failed to submit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tx)
}

// QueryChaincode handles querying chaincode (read-only)
func (h *Handler) QueryChaincode(w http.ResponseWriter, r *http.Request) {
	var req transaction.QueryTransactionRequest
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

	result, err := h.txService.QueryTransaction(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to query chaincode", zap.Error(err))
		http.Error(w, "Failed to query chaincode", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

// GetTransaction handles retrieving a transaction by ID
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	tx, err := h.txService.GetTransaction(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get transaction", zap.Error(err))
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

// GetTransactionByTxID handles retrieving a transaction by blockchain tx_id
func (h *Handler) GetTransactionByTxID(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txid")
	if txID == "" {
		http.Error(w, "Transaction ID is required", http.StatusBadRequest)
		return
	}

	tx, err := h.txService.GetTransactionByTxID(r.Context(), txID)
	if err != nil {
		h.logger.Error("Failed to get transaction by txid", zap.Error(err))
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

// QueryTransactions handles querying multiple transactions
func (h *Handler) QueryTransactions(w http.ResponseWriter, r *http.Request) {
	var req transaction.QueryTransactionsRequest

	// Parse query parameters
	query := r.URL.Query()
	
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			req.UserID = &userID
		}
	}

	if channelID := query.Get("channel_id"); channelID != "" {
		req.ChannelID = &channelID
	}

	// Support both "chaincode" and "chaincode_id" query parameters
	chaincodeID := query.Get("chaincode")
	if chaincodeID == "" {
		chaincodeID = query.Get("chaincode_id")
	}
	if chaincodeID != "" {
		req.ChaincodeID = &chaincodeID
	}

	if status := query.Get("status"); status != "" {
		req.Status = &status
	}

	// Parse limit and offset from query parameters
	req.Limit = 100 // Default limit
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	req.Offset = 0 // Default offset
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	transactions, err := h.txService.QueryTransactions(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to query transactions", zap.Error(err))
		http.Error(w, "Failed to query transactions", http.StatusInternalServerError)
		return
	}

	// Return response in standard format matching frontend expectation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"transactions": transactions,
			"total":        len(transactions),
			"limit":        req.Limit,
			"offset":       req.Offset,
		},
	})
}

// GetTransactionHistory handles retrieving transaction status history
func (h *Handler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	history, err := h.txService.GetStatusHistory(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get transaction history", zap.Error(err))
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

