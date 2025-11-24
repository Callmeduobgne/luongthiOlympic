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

package qrcode

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ibn-network/backend/internal/services/qrcode"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode"
	"github.com/ibn-network/backend/internal/services/blockchain/transaction"
)

// Handler handles QR code and NFC generation requests
type Handler struct {
	logger            *zap.Logger
	qrCodeService     *qrcode.Service
	packageService    chaincode.TeaTraceService
	transactionService *transaction.Service
}

// NewHandler creates a new QR code handler
func NewHandler(
	logger *zap.Logger,
	qrCodeService *qrcode.Service,
	packageService chaincode.TeaTraceService,
	transactionService *transaction.Service,
) *Handler {
	return &Handler{
		logger:            logger,
		qrCodeService:     qrCodeService,
		packageService:    packageService,
		transactionService: transactionService,
	}
}

// GeneratePackageQRCode generates QR code image for a package
// GET /api/v1/qrcode/packages/{packageId}
func (h *Handler) GeneratePackageQRCode(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "packageId")
	if packageID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Package ID is required")
		return
	}

	// Get package info from blockchain
	ctx := r.Context()
	pkg, err := h.packageService.GetPackage(ctx, packageID)
	if err != nil {
		h.logger.Error("Failed to get package", zap.Error(err), zap.String("package_id", packageID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	// Generate QR code PNG
	qrBytes, err := h.qrCodeService.GeneratePackageQRCode(pkg.PackageID, pkg.BlockHash, pkg.TxID)
	if err != nil {
		h.logger.Error("Failed to generate QR code", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate QR code")
		return
	}

	// Return PNG image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"qr-%s.png\"", packageID))
	w.WriteHeader(http.StatusOK)
	w.Write(qrBytes)
}

// GetPackageQRCodeData returns QR code data structure (JSON)
// GET /api/v1/qrcode/packages/{packageId}/data
func (h *Handler) GetPackageQRCodeData(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "packageId")
	if packageID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Package ID is required")
		return
	}

	// Get package info from blockchain
	ctx := r.Context()
	pkg, err := h.packageService.GetPackage(ctx, packageID)
	if err != nil {
		h.logger.Error("Failed to get package", zap.Error(err), zap.String("package_id", packageID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	// Get QR code data
	qrData := h.qrCodeService.GetPackageQRData(pkg.PackageID, pkg.BlockHash, pkg.TxID)

	h.respondSuccess(w, http.StatusOK, qrData)
}

// GetPackageQRCodeBase64 returns QR code as base64 data URI
// GET /api/v1/qrcode/packages/{packageId}/base64
func (h *Handler) GetPackageQRCodeBase64(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "packageId")
	if packageID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Package ID is required")
		return
	}

	// Get package info from blockchain
	ctx := r.Context()
	pkg, err := h.packageService.GetPackage(ctx, packageID)
	if err != nil {
		h.logger.Error("Failed to get package", zap.Error(err), zap.String("package_id", packageID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	// Generate QR code base64 string
	dataURI, err := h.qrCodeService.GeneratePackageQRCodeString(pkg.PackageID, pkg.BlockHash, pkg.TxID)
	if err != nil {
		h.logger.Error("Failed to generate QR code", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate QR code")
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]string{
		"dataUri": dataURI,
		"packageId": packageID,
	})
}

// GenerateNFCPayload generates NFC payload for a package
// GET /api/v1/nfc/packages/{packageId}
func (h *Handler) GenerateNFCPayload(w http.ResponseWriter, r *http.Request) {
	packageID := chi.URLParam(r, "packageId")
	if packageID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Package ID is required")
		return
	}

	// Get package info from blockchain
	ctx := r.Context()
	pkg, err := h.packageService.GetPackage(ctx, packageID)
	if err != nil {
		h.logger.Error("Failed to get package", zap.Error(err), zap.String("package_id", packageID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
		return
	}

	// Generate NFC payload
	nfcPayload, err := h.qrCodeService.GenerateNFCPayload(pkg.PackageID, pkg.BlockHash, pkg.TxID)
	if err != nil {
		h.logger.Error("Failed to generate NFC payload", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate NFC payload")
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]interface{}{
		"packageId": packageID,
		"payload":   nfcPayload,
		"format":    "NDEF",
	})
}

// GenerateBatchQRCode generates QR code image for a batch
// GET /api/v1/qrcode/batches/{batchId}
func (h *Handler) GenerateBatchQRCode(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Batch ID is required")
		return
	}

	// Get batch info from blockchain
	ctx := r.Context()
	batch, err := h.packageService.GetBatch(ctx, batchID)
	if err != nil {
		h.logger.Error("Failed to get batch", zap.Error(err), zap.String("batch_id", batchID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Batch not found")
		return
	}

	// Generate QR code PNG
	qrBytes, err := h.qrCodeService.GenerateBatchQRCode(batch.BatchID, batch.VerificationHash, "")
	if err != nil {
		h.logger.Error("Failed to generate QR code", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate QR code")
		return
	}

	// Return PNG image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"qr-batch-%s.png\"", batchID))
	w.WriteHeader(http.StatusOK)
	w.Write(qrBytes)
}

// GetBatchQRCodeData returns QR code data structure for batch (JSON)
// GET /api/v1/qrcode/batches/{batchId}/data
func (h *Handler) GetBatchQRCodeData(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Batch ID is required")
		return
	}

	// Get batch info from blockchain
	ctx := r.Context()
	batch, err := h.packageService.GetBatch(ctx, batchID)
	if err != nil {
		h.logger.Error("Failed to get batch", zap.Error(err), zap.String("batch_id", batchID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Batch not found")
		return
	}

	// Get QR code data
	qrData := h.qrCodeService.GetBatchQRData(batch.BatchID, batch.VerificationHash, "")

	h.respondSuccess(w, http.StatusOK, qrData)
}

// GetBatchQRCodeBase64 returns QR code as base64 data URI for batch
// GET /api/v1/qrcode/batches/{batchId}/base64
func (h *Handler) GetBatchQRCodeBase64(w http.ResponseWriter, r *http.Request) {
	batchID := chi.URLParam(r, "batchId")
	if batchID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Batch ID is required")
		return
	}

	// Get batch info from blockchain
	ctx := r.Context()
	batch, err := h.packageService.GetBatch(ctx, batchID)
	if err != nil {
		h.logger.Error("Failed to get batch", zap.Error(err), zap.String("batch_id", batchID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Batch not found")
		return
	}

	// Generate QR code base64 string
	dataURI, err := h.qrCodeService.GenerateBatchQRCodeString(batch.BatchID, batch.VerificationHash, "")
	if err != nil {
		h.logger.Error("Failed to generate QR code", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate QR code")
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]string{
		"dataUri": dataURI,
		"batchId": batchID,
	})
}

// GenerateQRCodeFromTransaction generates QR code from transaction ID
// GET /api/v1/qrcode/transactions/{txId}
// This endpoint automatically detects if transaction created a batch or package
func (h *Handler) GenerateQRCodeFromTransaction(w http.ResponseWriter, r *http.Request) {
	txID := chi.URLParam(r, "txId")
	if txID == "" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Transaction ID is required")
		return
	}

	ctx := r.Context()

	// Get transaction from database
	tx, err := h.transactionService.GetTransactionByTxID(ctx, txID)
	if err != nil {
		h.logger.Error("Failed to get transaction", zap.Error(err), zap.String("tx_id", txID))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}

	// Check if transaction is for teaTraceCC
	if tx.ChaincodeID != "teaTraceCC" {
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Transaction is not from teaTraceCC chaincode")
		return
	}

	// Determine QR code type based on function name
	switch tx.FunctionName {
	case "createBatch":
		// Extract batch ID from args
		if len(tx.Args) == 0 {
			h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Transaction args missing batch ID")
			return
		}
		batchID := tx.Args[0]

		// Get batch info
		batch, err := h.packageService.GetBatch(ctx, batchID)
		if err != nil {
			h.logger.Error("Failed to get batch", zap.Error(err), zap.String("batch_id", batchID))
			h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Batch not found")
			return
		}

		// Generate QR code PNG
		qrBytes, err := h.qrCodeService.GenerateBatchQRCode(batch.BatchID, batch.VerificationHash, txID)
		if err != nil {
			h.logger.Error("Failed to generate QR code", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate QR code")
			return
		}

		// Return PNG image
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"qr-batch-%s.png\"", batchID))
		w.WriteHeader(http.StatusOK)
		w.Write(qrBytes)

	case "createPackage":
		// Extract package ID from args
		if len(tx.Args) == 0 {
			h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Transaction args missing package ID")
			return
		}
		packageID := tx.Args[0]

		// Get package info
		pkg, err := h.packageService.GetPackage(ctx, packageID)
		if err != nil {
			h.logger.Error("Failed to get package", zap.Error(err), zap.String("package_id", packageID))
			h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Package not found")
			return
		}

		// Generate QR code PNG
		qrBytes, err := h.qrCodeService.GeneratePackageQRCode(pkg.PackageID, pkg.BlockHash, txID)
		if err != nil {
			h.logger.Error("Failed to generate QR code", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate QR code")
			return
		}

		// Return PNG image
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"qr-package-%s.png\"", packageID))
		w.WriteHeader(http.StatusOK)
		w.Write(qrBytes)

	default:
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", 
			fmt.Sprintf("Transaction function '%s' does not support QR code generation", tx.FunctionName))
	}
}

// Helper methods

func (h *Handler) respondSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func (h *Handler) respondError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

