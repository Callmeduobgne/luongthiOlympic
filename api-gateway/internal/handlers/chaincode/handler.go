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

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/chaincode"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	transactionservice "github.com/ibn-network/api-gateway/internal/services/transaction"
	"go.uber.org/zap"
)

// ChaincodeHandler handles chaincode lifecycle operations
type ChaincodeHandler struct {
	chaincodeService  *chaincode.Service
	genericService    *fabric.ChaincodeService
	transactionService *transactionservice.Service
	logger            *zap.Logger
}

// NewChaincodeHandler creates a new chaincode handler
func NewChaincodeHandler(
	chaincodeService *chaincode.Service,
	genericService *fabric.ChaincodeService,
	transactionService *transactionservice.Service,
	logger *zap.Logger,
) *ChaincodeHandler {
	return &ChaincodeHandler{
		chaincodeService:  chaincodeService,
		genericService:     genericService,
		transactionService: transactionService,
		logger:            logger,
	}
}

// ListInstalled - DEPRECATED: Moved to Admin Service
// This method is kept for backward compatibility but routes have been removed
// Use Backend API: GET /api/v1/chaincode/installed (which calls Admin Service)
func (h *ChaincodeHandler) ListInstalled(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusGone, models.NewErrorResponse(
		"DEPRECATED",
		"Chaincode lifecycle operations have been moved to Admin Service. Please use Backend API endpoints.",
		"Use: GET /api/v1/chaincode/installed via Backend",
	))
}

// ListCommitted - DEPRECATED: Moved to Admin Service
// This method is kept for backward compatibility but routes have been removed
// Use Backend API: GET /api/v1/chaincode/committed (which calls Admin Service)
func (h *ChaincodeHandler) ListCommitted(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusGone, models.NewErrorResponse(
		"DEPRECATED",
		"Chaincode lifecycle operations have been moved to Admin Service. Please use Backend API endpoints.",
		"Use: GET /api/v1/chaincode/committed via Backend",
	))
}

// GetCommittedInfo - DEPRECATED: Moved to Admin Service
// This method is kept for backward compatibility but routes have been removed
// Use Backend API: GET /api/v1/chaincode/committed/{name} (which calls Admin Service)
func (h *ChaincodeHandler) GetCommittedInfo(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusGone, models.NewErrorResponse(
		"DEPRECATED",
		"Chaincode lifecycle operations have been moved to Admin Service. Please use Backend API endpoints.",
		"Use: GET /api/v1/chaincode/committed/{name} via Backend",
	))
}

// Install - DEPRECATED: Moved to Admin Service
// This method is kept for backward compatibility but routes have been removed
// Use Backend API: POST /api/v1/chaincode/install (which calls Admin Service)
func (h *ChaincodeHandler) Install(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusGone, models.NewErrorResponse(
		"DEPRECATED",
		"Chaincode lifecycle operations have been moved to Admin Service. Please use Backend API endpoints.",
		"Use: POST /api/v1/chaincode/install via Backend",
	))
}

// Approve - DEPRECATED: Moved to Admin Service
// This method is kept for backward compatibility but routes have been removed
// Use Backend API: POST /api/v1/chaincode/approve (which calls Admin Service)
func (h *ChaincodeHandler) Approve(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusGone, models.NewErrorResponse(
		"DEPRECATED",
		"Chaincode lifecycle operations have been moved to Admin Service. Please use Backend API endpoints.",
		"Use: POST /api/v1/chaincode/approve via Backend",
	))
}

// Commit - DEPRECATED: Moved to Admin Service
// This method is kept for backward compatibility but routes have been removed
// Use Backend API: POST /api/v1/chaincode/commit (which calls Admin Service)
func (h *ChaincodeHandler) Commit(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusGone, models.NewErrorResponse(
		"DEPRECATED",
		"Chaincode lifecycle operations have been moved to Admin Service. Please use Backend API endpoints.",
		"Use: POST /api/v1/chaincode/commit via Backend",
	))
}

// Invoke godoc
// @Summary Invoke chaincode function
// @Description Submit a transaction to invoke a chaincode function
// @Tags chaincode
// @Accept json
// @Produce json
// @Param channel path string true "Channel name"
// @Param name path string true "Chaincode name"
// @Param request body models.InvokeRequest true "Invoke request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels/{channel}/chaincodes/{name}/invoke [post]
func (h *ChaincodeHandler) Invoke(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	chaincodeName := chi.URLParam(r, "name")

	if channelName == "" || chaincodeName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name and chaincode name are required",
			nil,
		))
		return
	}

	var req models.InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.Function == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Function name is required",
			nil,
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

	// If transaction service is available, use it to track the transaction
	if h.transactionService != nil {
		h.logger.Info("Using transaction service to track transaction",
			zap.String("channel", channelName),
			zap.String("chaincode", chaincodeName),
			zap.String("function", req.Function),
		)

		txReq := &models.TransactionRequest{
			ChannelName:   channelName,
			ChaincodeName: chaincodeName,
			FunctionName:  req.Function,
			Args:          req.Args,
			TransientData: nil, // Convert if needed
		}

		// Submit transaction via transaction service (tracks in database)
		txResponse, err := h.transactionService.SubmitTransaction(r.Context(), txReq, userID, apiKeyID)
		if err != nil {
			h.logger.Error("Failed to submit transaction",
				zap.String("channel", channelName),
				zap.String("chaincode", chaincodeName),
				zap.String("function", req.Function),
				zap.Error(err),
			)
			respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
				models.ErrCodeTransactionFailed,
				"Failed to invoke chaincode",
				err.Error(),
			))
			return
		}

		// Get result from transaction service (it already has the result)
		// We need to query the chaincode to get the result, or use the result from the invoke
		// For now, create a simple response
		invokeResponse := &models.InvokeResponse{
			TxID:        txResponse.TxID,
			Result:      nil, // Result is not stored in transaction, would need to query separately
			Status:      string(txResponse.Status),
			BlockNumber: txResponse.BlockNumber,
			Timestamp:   txResponse.Timestamp.Format("2006-01-02T15:04:05Z"),
		}

		// Try to get result by querying the chaincode (if it's a queryable function)
		// For now, just return the transaction info
		h.logger.Info("Transaction submitted successfully via transaction service",
			zap.String("txId", txResponse.TxID),
			zap.Uint64("blockNumber", txResponse.BlockNumber),
		)

		respondJSON(w, http.StatusOK, models.NewSuccessResponse(invokeResponse))
		return
	}

	// Fallback: Invoke chaincode directly (without tracking)
	h.logger.Warn("Transaction service is nil, using generic service (transaction will not be tracked in database)")
	response, err := h.genericService.Invoke(r.Context(), channelName, chaincodeName, &req)
	if err != nil {
		h.logger.Error("Failed to invoke chaincode",
			zap.String("channel", channelName),
			zap.String("chaincode", chaincodeName),
			zap.String("function", req.Function),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to invoke chaincode",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// Query godoc
// @Summary Query chaincode function
// @Description Evaluate a chaincode function (read-only query)
// @Tags chaincode
// @Accept json
// @Produce json
// @Param channel path string true "Channel name"
// @Param name path string true "Chaincode name"
// @Param request body models.QueryRequest true "Query request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /channels/{channel}/chaincodes/{name}/query [post]
func (h *ChaincodeHandler) Query(w http.ResponseWriter, r *http.Request) {
	channelName := chi.URLParam(r, "channel")
	chaincodeName := chi.URLParam(r, "name")

	if channelName == "" || chaincodeName == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Channel name and chaincode name are required",
			nil,
		))
		return
	}

	var req models.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.Function == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Function name is required",
			nil,
		))
		return
	}

	// Query chaincode
	response, err := h.genericService.Query(r.Context(), channelName, chaincodeName, &req)
	if err != nil {
		h.logger.Error("Failed to query chaincode",
			zap.String("channel", channelName),
			zap.String("chaincode", chaincodeName),
			zap.String("function", req.Function),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to query chaincode",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// respondJSON is a helper function to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

