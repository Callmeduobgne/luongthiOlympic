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
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ibn-network/backend/internal/services/blockchain/info"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// InfoHandler handles blockchain info HTTP requests
type InfoHandler struct {
	infoService info.InfoService // Use interface to support both Service and ServiceViaGateway
	logger      *zap.Logger
}

// NewInfoHandler creates a new blockchain info handler
func NewInfoHandler(infoService info.InfoService, logger *zap.Logger) *InfoHandler {
	return &InfoHandler{
		infoService: infoService,
		logger:      logger,
	}
}

// GetChannelInfo handles getting channel information
func (h *InfoHandler) GetChannelInfo(w http.ResponseWriter, r *http.Request) {
	channelInfo, err := h.infoService.GetChannelInfo(r.Context())
	if err != nil {
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

