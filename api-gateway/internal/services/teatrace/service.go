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

package teatrace

import (
	"context"
	"fmt"
	"time"

	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/transaction"
	"go.uber.org/zap"
)

// Service handles Tea Traceability operations
type Service struct {
	transactionService *transaction.Service
	logger             *zap.Logger
}

// NewService creates a new Tea Traceability service
func NewService(
	transactionService *transaction.Service,
	logger *zap.Logger,
) *Service {
	return &Service{
		transactionService: transactionService,
		logger:             logger,
	}
}

// VerifyResult represents the result of a verification
type VerifyResult struct {
	IsValid            bool              `json:"is_valid"`
	Message            string            `json:"message"`
	TransactionID      string            `json:"transaction_id,omitempty"`
	BatchID            string            `json:"batch_id,omitempty"`
	PackageID          string            `json:"package_id,omitempty"`
	BlockNumber        uint64            `json:"block_number,omitempty"`
	EntityType         string            `json:"entity_type,omitempty"`
	VerifiedAt         string            `json:"verified_at"`
	VerificationMethod string            `json:"verification_method"`
	ProductDetails     map[string]interface{} `json:"product_details,omitempty"`
}

// VerifyByHash verifies an entity by its hash (Transaction ID)
func (s *Service) VerifyByHash(ctx context.Context, hash string) (*VerifyResult, error) {
	s.logger.Info("Verifying by hash", zap.String("hash", hash))

	// 1. Try to find transaction by ID
	tx, err := s.transactionService.GetTransaction(ctx, hash)
	if err != nil {
		s.logger.Warn("Transaction not found", zap.String("hash", hash), zap.Error(err))
		return &VerifyResult{
			IsValid: false,
			Message: "Không tìm thấy thông tin giao dịch với mã này.",
		}, nil
	}

	// 2. Analyze transaction to extract details
	result := &VerifyResult{
		IsValid:            true,
		Message:            "Giao dịch hợp lệ và đã được ghi nhận trên Blockchain.",
		TransactionID:      tx.TxID,
		BlockNumber:        tx.BlockNumber,
		VerifiedAt:         time.Now().Format(time.RFC3339),
		VerificationMethod: "blockchain_query",
		ProductDetails:     make(map[string]interface{}),
	}

	// Extract details from arguments
	// Assuming args structure based on chaincode function
	// createBatch: [batchId, farmLocation, harvestDate, processingInfo, qualityCert]
	// createPackage: [packageId, batchId, weight, productionDate, expiryDate?, qrCode?]
	
	if len(tx.Args) > 0 {
		switch tx.FunctionName {
		case "createBatch":
			result.EntityType = "batch"
			if len(tx.Args) >= 1 {
				result.BatchID = tx.Args[0]
			}
			if len(tx.Args) >= 5 {
				result.ProductDetails["farm_location"] = tx.Args[1]
				result.ProductDetails["harvest_date"] = tx.Args[2]
				result.ProductDetails["processing_info"] = tx.Args[3]
				result.ProductDetails["quality_cert"] = tx.Args[4]
			}
		case "createPackage":
			result.EntityType = "package"
			if len(tx.Args) >= 1 {
				result.PackageID = tx.Args[0]
			}
			if len(tx.Args) >= 2 {
				result.BatchID = tx.Args[1]
			}
			if len(tx.Args) >= 4 {
				result.ProductDetails["weight"] = tx.Args[2]
				result.ProductDetails["production_date"] = tx.Args[3]
			}
			if len(tx.Args) >= 5 {
				result.ProductDetails["expiry_date"] = tx.Args[4]
			}
		default:
			result.EntityType = "transaction"
			result.Message = fmt.Sprintf("Giao dịch hợp lệ (Loại: %s)", tx.FunctionName)
		}
	}

	// Check transaction status
	if tx.Status != models.TransactionStatusValid {
		result.IsValid = false
		result.Message = fmt.Sprintf("Giao dịch tồn tại nhưng không hợp lệ (Trạng thái: %s)", tx.Status)
	}

	return result, nil
}
