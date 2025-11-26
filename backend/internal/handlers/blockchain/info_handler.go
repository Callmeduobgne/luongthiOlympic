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
	"strings"

	"github.com/ibn-network/backend/internal/services/blockchain/info"
	"github.com/ibn-network/backend/internal/services/certificate"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// InfoHandler handles blockchain info HTTP requests
type InfoHandler struct {
	infoService info.InfoService // Use interface to support both Service and ServiceViaGateway
	certService *certificate.Service // Certificate service for user certificates
	logger      *zap.Logger
}

// NewInfoHandler creates a new blockchain info handler
func NewInfoHandler(infoService info.InfoService, certService *certificate.Service, logger *zap.Logger) *InfoHandler {
	return &InfoHandler{
		infoService: infoService,
		certService: certService,
		logger:      logger,
	}
}

// GetChannelInfo handles getting channel information
func (h *InfoHandler) GetChannelInfo(w http.ResponseWriter, r *http.Request) {
	channelInfo, err := h.infoService.GetChannelInfo(r.Context())
	if err != nil {
		// Check if error is due to Gateway/qscc issues (expected fallback scenario)
		errStr := err.Error()
		if contains(errStr, "gateway returned status") ||
			contains(errStr, "502") ||
			contains(errStr, "function that does not exist") ||
			contains(errStr, "qscc") {
			// Return fallback empty data instead of 500 error
			// This allows frontend to continue working even if GetChainInfo is not available
			h.logger.Debug("GetChannelInfo fallback: Gateway/qscc issue, returning empty data",
				zap.String("error_type", "fallback"),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"channel_id": "ibnchannel",
				"raw_info":   "",
				"size":       0,
			})
			return
		}

		// For unexpected errors, return 500
		h.logger.Error("Failed to get channel info", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get channel info",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channelInfo)
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && strings.Contains(s, substr)))
}

// GetBlockByNumber handles getting block by number
func (h *InfoHandler) GetBlockByNumber(w http.ResponseWriter, r *http.Request) {
	blockNumStr := chi.URLParam(r, "number")
	blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid block number",
		})
		return
	}

	block, err := h.infoService.GetBlockByNumber(r.Context(), blockNum)
	if err != nil {
		h.logger.Error("Failed to get block",
			zap.Uint64("block_number", blockNum),
			zap.Error(err),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get block",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

// GetBlockByTxID handles getting block by transaction ID
func (h *InfoHandler) GetBlockByTxID(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txid")
	if txID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Transaction ID is required",
		})
		return
	}

	block, err := h.infoService.GetBlockByTxID(r.Context(), txID)
	if err != nil {
		h.logger.Error("Failed to get block by txID",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get block",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(block)
}

// GetTransactionByID handles getting transaction by ID
func (h *InfoHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txid")
	if txID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Transaction ID is required",
		})
		return
	}

	txHex, err := h.infoService.GetTransactionByID(r.Context(), txID)
	if err != nil {
		h.logger.Error("Failed to get transaction",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to get transaction",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tx_id":          txID,
		"transaction_hex": txHex,
		"size_bytes":     len(txHex) / 2,
	})
}

// VerifyTransactionByHash verifies a transaction hash by querying directly from blockchain network
// This endpoint is PUBLIC (no authentication required) - anyone can verify transactions
// 
// Behavior:
// - If user is logged in and has certificate: uses user's certificate to query blockchain
// - If user is not logged in or has no certificate: uses Gateway API key (service account) to query
// - Always queries from blockchain network via Gateway, NOT from database
func (h *InfoHandler) VerifyTransactionByHash(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txid")
	if txID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Transaction ID (hash) is required",
			},
		})
		return
	}

	// Get user ID from context (set by auth middleware if user is logged in)
	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		// No user logged in - will use Gateway API key (service account) for query
		h.logger.Debug("No user logged in, querying with Gateway API key (service account)",
			zap.String("tx_id", txID),
		)
	} else {
		// User is logged in - try to use user's certificate if available
		if h.certService != nil {
			certWithKey, err := h.certService.GetActiveCertificateWithKey(ctx, userID)
			if err != nil {
				// User has no certificate - fallback to Gateway API key
				h.logger.Debug("User has no certificate, using Gateway API key for query",
					zap.String("user_id", userID.String()),
					zap.String("tx_id", txID),
				)
			} else {
				// Use user's certificate to query blockchain with user's identity
				ctx = context.WithValue(ctx, "user_cert", certWithKey.Certificate)
				ctx = context.WithValue(ctx, "user_key", certWithKey.PrivateKey)
				ctx = context.WithValue(ctx, "user_msp_id", certWithKey.MSPID)
				
				h.logger.Debug("Using user certificate for verification",
					zap.String("user_id", userID.String()),
					zap.String("cert_id", certWithKey.ID.String()),
					zap.String("msp_id", certWithKey.MSPID),
					zap.String("tx_id", txID),
				)
			}
		}
	}

	// Query transaction directly from blockchain network via Gateway
	// Will use user certificate if available, otherwise Gateway API key
	txHex, err := h.infoService.GetTransactionByID(ctx, txID)
	if err != nil {
		// Transaction not found in blockchain network
		h.logger.Warn("Transaction not found in blockchain network",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"is_valid": false,
				"message":  "Transaction không tồn tại trong blockchain network",
				"tx_id":    txID,
			},
		})
		return
	}

	// Transaction found in blockchain network - verification successful
	h.logger.Info("Transaction verified from blockchain network",
		zap.String("tx_id", txID),
		zap.Int("size_bytes", len(txHex)/2),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"is_valid":      true,
			"message":       "Transaction hợp lệ và tồn tại trong blockchain network",
			"tx_id":         txID,
			"size_bytes":   len(txHex) / 2,
			"verified_from": "blockchain_network",
		},
	})
}

