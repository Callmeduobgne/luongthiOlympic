// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chaincode

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ibn-network/backend/internal/infrastructure/cache"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// VerifyService handles product verification by hash/blockhash
type VerifyService struct {
	db          *pgxpool.Pool
	cache       *cache.MultiLayerCache
	teaTraceSvc TeaTraceService
	logger      *zap.Logger
}

// NewVerifyService creates a new verify service
func NewVerifyService(
	db *pgxpool.Pool,
	cache *cache.MultiLayerCache,
	teaTraceSvc TeaTraceService,
	logger *zap.Logger,
) *VerifyService {
	return &VerifyService{
		db:          db,
		cache:       cache,
		teaTraceSvc: teaTraceSvc,
		logger:      logger,
	}
}

// VerifyByHash verifies a product by hash/blockhash/transaction ID
// Uses cache (24h TTL) and database queries with indexes
func (s *VerifyService) VerifyByHash(ctx context.Context, hash string) (*VerifyByHashResponse, error) {
	hash = trimHash(hash)
	if hash == "" {
		return &VerifyByHashResponse{
			IsValid: false,
			Message: "Hash không hợp lệ",
		}, nil
	}

	// Check cache first (24h TTL)
	cacheKey := fmt.Sprintf("verify:hash:%s", hash)
	var cachedResult VerifyByHashResponse
	
	// Try to get from cache
	err := s.cache.Get(ctx, cacheKey, &cachedResult, func(ctx context.Context) (interface{}, error) {
		// Cache miss - will query database below
		return nil, fmt.Errorf("cache miss")
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,  // L1: 5 minutes
		L2TTL: 24 * time.Hour,  // L2: 24 hours
	})
	
	if err == nil {
		s.logger.Debug("Hash verification result retrieved from cache",
			zap.String("hash", hash[:16]+"..."),
			zap.Bool("is_valid", cachedResult.IsValid),
		)
		return &cachedResult, nil
	}

	// Query database with indexes
	result, err := s.queryDatabase(ctx, hash)
	if err != nil {
		s.logger.Error("Failed to query database for hash verification",
			zap.String("hash", hash[:16]+"..."),
			zap.Error(err),
		)
		return &VerifyByHashResponse{
			IsValid: false,
			Message: "Không thể xác thực sản phẩm. Vui lòng thử lại sau.",
		}, nil
	}

	// Cache result for 24 hours (will be cached automatically by Get method above)
	// But we also set it explicitly here to ensure it's cached
	if cacheErr := s.cache.Set(ctx, cacheKey, result, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,  // L1: 5 minutes
		L2TTL: 24 * time.Hour,   // L2: 24 hours
	}); cacheErr != nil {
		s.logger.Warn("Failed to cache verification result", zap.Error(cacheErr))
	}

	return result, nil
}

// queryDatabase queries database to find transaction/batch/package by hash
func (s *VerifyService) queryDatabase(ctx context.Context, hash string) (*VerifyByHashResponse, error) {
	// Option 1: Query transaction by tx_id (with index)
	// Note: transactions table is in auth schema (shared with api-gateway)
	var txID, chaincodeName string
	var argsJSON []byte
	queryTx := `
		SELECT tx_id, chaincode_name, args
		FROM auth.transactions
		WHERE tx_id = $1
		LIMIT 1
	`
	err := s.db.QueryRow(ctx, queryTx, hash).Scan(&txID, &chaincodeName, &argsJSON)
	if err == nil && chaincodeName == "teaTraceCC" {
		// Found transaction related to teaTraceCC
		response := &VerifyByHashResponse{
			IsValid:       true,
			Message:       "Sản phẩm thuộc thương hiệu chúng tôi (Verified)",
			TransactionID: txID,
			EntityType:    "transaction",
			VerifiedAt:    time.Now(),
			VerificationMethod: "blockchain_query",
		}
		s.logger.Info("VERIFY DEBUG: Transaction found", zap.String("txID", txID))

		// Try to parse args to get more info
		var args []string
		if len(argsJSON) > 0 {
			_ = json.Unmarshal(argsJSON, &args)
		}

		// If args available, try to fetch details
		if len(args) > 0 {
			// Check function name (we need to fetch function name too)
			var functionName string
			_ = s.db.QueryRow(ctx, "SELECT function_name FROM auth.transactions WHERE tx_id = $1", txID).Scan(&functionName)

			if functionName == "createPackage" && len(args) >= 2 {
				// args: [packageId, batchId, weight, productionDate, expiryDate?]
				packageID := args[0]
				batchID := args[1]
				response.PackageID = packageID
				response.BatchID = batchID
				response.EntityType = "package"

				// Initialize ProductDetails from args (fallback)
				response.ProductDetails = &ProductDetails{
					ProductionDate: args[3],
				}
				if len(args) >= 3 {
					fmt.Sscanf(args[2], "%f", &response.ProductDetails.Weight)
				}
				if len(args) >= 5 {
					response.ProductDetails.ExpiryDate = args[4]
				}

				// Fetch package details (live data)
				pkg, err := s.teaTraceSvc.GetPackage(ctx, packageID)
				if err == nil {
					response.ProductDetails.Status = string(pkg.Status)
					// Update with live data if available
					if pkg.ExpiryDate != "" {
						response.ProductDetails.ExpiryDate = pkg.ExpiryDate
					}
				}

				// Fetch batch details (for farm info)
				batch, err := s.teaTraceSvc.GetBatch(ctx, batchID)
				if err == nil {
					response.ProductDetails.FarmLocation = batch.FarmLocation
					response.ProductDetails.HarvestDate = batch.HarvestDate
					response.ProductDetails.ProcessingInfo = batch.ProcessingInfo
					response.ProductDetails.QualityCert = batch.QualityCert
				}
			} else if functionName == "createBatch" && len(args) >= 1 {
				// args: [batchId, farmName, harvestDate, certification, certificateID]
				batchID := args[0]
				response.BatchID = batchID
				response.EntityType = "batch"
				
				// Initialize ProductDetails from args
				response.ProductDetails = &ProductDetails{}
				if len(args) >= 3 {
					response.ProductDetails.FarmLocation = args[1]
					response.ProductDetails.HarvestDate = args[2]
				}
				if len(args) >= 4 {
					response.ProductDetails.QualityCert = args[3]
				}

				// Fetch batch details (live data)
				batch, err := s.teaTraceSvc.GetBatch(ctx, batchID)
				if err == nil {
					response.ProductDetails.FarmLocation = batch.FarmLocation
					response.ProductDetails.HarvestDate = batch.HarvestDate
					response.ProductDetails.ProcessingInfo = batch.ProcessingInfo
					response.ProductDetails.QualityCert = batch.QualityCert
					response.ProductDetails.Status = string(batch.Status)
				}
			}
		}

		return response, nil
	}

	// Option 2: Query all batches and check verificationHash
	// Since we don't have direct batch table, query via chaincode
	batches, err := s.teaTraceSvc.GetAllBatches(ctx)
	if err == nil {
		for _, batch := range batches {
			if batch.VerificationHash == hash {
				response := &VerifyByHashResponse{
					IsValid:       true,
					Message:       "Sản phẩm thuộc thương hiệu chúng tôi",
					BatchID:       batch.BatchID,
					EntityType:    "batch",
					VerifiedAt:    time.Now(),
					VerificationMethod: "blockchain_query",
				}
				
				response.ProductDetails = &ProductDetails{
					FarmLocation:   batch.FarmLocation,
					HarvestDate:    batch.HarvestDate,
					ProcessingInfo: batch.ProcessingInfo,
					QualityCert:    batch.QualityCert,
					Status:         string(batch.Status),
				}
				
				return response, nil
			}
		}
	}

	// Option 3: Try to get package by querying chaincode (if hash is package blockHash)
	// This is expensive, so we do it last
	// We would need to query all packages, but that's not efficient
	// Instead, we'll check if hash matches any known pattern

	// Not found - product doesn't belong to our brand
	return &VerifyByHashResponse{
		IsValid: false,
		Message: "Sản phẩm không thuộc thương hiệu chúng tôi",
	}, nil
}

// trimHash trims and normalizes hash input
func trimHash(hash string) string {
	// Remove whitespace
	hash = strings.TrimSpace(hash)
	// Convert to lowercase for consistency
	return strings.ToLower(hash)
}

