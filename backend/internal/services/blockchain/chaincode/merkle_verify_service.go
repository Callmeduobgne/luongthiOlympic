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
	"context"
	"fmt"
	"time"

	"github.com/ibn-network/backend/internal/infrastructure/cache"
	"github.com/ibn-network/backend/internal/utils/merkle"
	"go.uber.org/zap"
)

// MerkleVerifyService handles product verification using Merkle proofs
// This service provides ultra-fast verification (< 10ms) without blockchain queries
type MerkleVerifyService struct {
	cache            *cache.MultiLayerCache
	teaTraceSvc      TeaTraceService // For blockchain fallback
	logger           *zap.Logger
}

// NewMerkleVerifyService creates a new Merkle verification service
func NewMerkleVerifyService(
	cache *cache.MultiLayerCache,
	teaTraceSvc TeaTraceService,
	logger *zap.Logger,
) *MerkleVerifyService {
	return &MerkleVerifyService{
		cache:       cache,
		teaTraceSvc: teaTraceSvc,
		logger:      logger,
	}
}

// VerifyWithMerkleProof verifies a transaction using Merkle proof
// This is the fast path - no blockchain query needed
// Returns verification result with metadata
func (s *MerkleVerifyService) VerifyWithMerkleProof(
	ctx context.Context,
	txID string,
	proof []merkle.ProofStep,
	merkleRoot string,
	blockNumber uint64,
) (*VerifyByHashResponse, error) {
	startTime := time.Now()

	// Validate inputs
	if txID == "" {
		return &VerifyByHashResponse{
			IsValid: false,
			Message: "Transaction ID không hợp lệ",
		}, nil
	}

	if len(proof) == 0 {
		s.logger.Warn("Empty Merkle proof provided, falling back to blockchain",
			zap.String("tx_id", txID),
		)
		return s.FallbackToBlockchainVerify(ctx, txID)
	}

	if merkleRoot == "" {
		s.logger.Warn("Empty Merkle root provided, falling back to blockchain",
			zap.String("tx_id", txID),
		)
		return s.FallbackToBlockchainVerify(ctx, txID)
	}

	// Check cache first (to avoid re-verification)
	cacheKey := fmt.Sprintf("merkle:verify:%s:%s", txID, merkleRoot)
	var cachedResult VerifyByHashResponse
	
	err := s.cache.Get(ctx, cacheKey, &cachedResult, func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Errorf("cache miss")
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,  // L1: 5 minutes
		L2TTL: 24 * time.Hour,   // L2: 24 hours
	})
	
	if err == nil {
		s.logger.Debug("Merkle verification result from cache",
			zap.String("tx_id", txID[:min(16, len(txID))]),
			zap.Bool("is_valid", cachedResult.IsValid),
			zap.Duration("duration", time.Since(startTime)),
		)
		return &cachedResult, nil
	}

	// Verify Merkle proof cryptographically
	isValid := merkle.VerifyMerkleProof(txID, proof, merkleRoot)

	duration := time.Since(startTime)

	if !isValid {
		s.logger.Warn("Merkle proof verification failed",
			zap.String("tx_id", txID),
			zap.Uint64("block_number", blockNumber),
			zap.Int("proof_steps", len(proof)),
			zap.Duration("duration", duration),
		)
		
		// Proof verification failed - this could mean:
		// 1. Tampered proof
		// 2. Wrong root hash
		// 3. Transaction not in block
		// Fallback to blockchain for authoritative answer
		return s.FallbackToBlockchainVerify(ctx, txID)
	}

	// Proof verified successfully!
	result := &VerifyByHashResponse{
		IsValid:            true,
		Message:            "Sản phẩm chính hãng - Xác thực bằng Merkle proof (cryptographic)",
		TransactionID:      txID,
		BlockNumber:        blockNumber,
		VerifiedAt:         time.Now(),
		VerificationMethod: "merkle_proof",
		EntityType:         "transaction",
	}

	// Cache the result
	if cacheErr := s.cache.Set(ctx, cacheKey, result, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 24 * time.Hour,
	}); cacheErr != nil {
		s.logger.Warn("Failed to cache Merkle verification result", zap.Error(cacheErr))
	}

	s.logger.Info("Merkle proof verified successfully",
		zap.String("tx_id", txID),
		zap.Uint64("block_number", blockNumber),
		zap.Int("proof_steps", len(proof)),
		zap.Duration("duration", duration),
	)

	return result, nil
}

// FallbackToBlockchainVerify queries blockchain when Merkle proof fails or is unavailable
// This is the slow path but provides authoritative verification
func (s *MerkleVerifyService) FallbackToBlockchainVerify(
	ctx context.Context,
	txID string,
) (*VerifyByHashResponse, error) {
	startTime := time.Now()
	
	s.logger.Info("Falling back to blockchain verification",
		zap.String("tx_id", txID),
	)

	// Check cache for blockchain verification result
	cacheKey := fmt.Sprintf("blockchain:verify:%s", txID)
	var cachedResult VerifyByHashResponse
	
	err := s.cache.Get(ctx, cacheKey, &cachedResult, func(ctx context.Context) (interface{}, error) {
		return nil, fmt.Errorf("cache miss")
	}, &cache.CacheTTLs{
		L1TTL: 5 * time.Minute,
		L2TTL: 1 * time.Hour, // Shorter TTL for blockchain results
	})
	
	if err == nil {
		s.logger.Debug("Blockchain verification result from cache",
			zap.String("tx_id", txID[:min(16, len(txID))]),
			zap.Duration("duration", time.Since(startTime)),
		)
		return &cachedResult, nil
	}

	// Query blockchain - try to get batch by verification hash
	// Note: This is a simplified implementation
	// In production, you would query the actual blockchain/chaincode
	batches, err := s.teaTraceSvc.GetAllBatches(ctx)
	if err != nil {
		s.logger.Error("Failed to query blockchain",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		return &VerifyByHashResponse{
			IsValid: false,
			Message: "Không thể xác thực sản phẩm. Vui lòng thử lại sau.",
		}, nil
	}

	// Search for transaction in batches
	for _, batch := range batches {
		if batch.VerificationHash == txID {
			result := &VerifyByHashResponse{
				IsValid:            true,
				Message:            "Sản phẩm chính hãng - Xác thực từ blockchain",
				TransactionID:      txID,
				BatchID:            batch.BatchID,
				VerifiedAt:         time.Now(),
				VerificationMethod: "blockchain_query",
				EntityType:         "batch",
			}

			// Cache the result
			if cacheErr := s.cache.Set(ctx, cacheKey, result, &cache.CacheTTLs{
				L1TTL: 5 * time.Minute,
				L2TTL: 1 * time.Hour,
			}); cacheErr != nil {
				s.logger.Warn("Failed to cache blockchain verification result", zap.Error(cacheErr))
			}

			duration := time.Since(startTime)
			s.logger.Info("Blockchain verification successful",
				zap.String("tx_id", txID),
				zap.String("batch_id", batch.BatchID),
				zap.Duration("duration", duration),
			)

			return result, nil
		}
	}

	// Not found on blockchain
	duration := time.Since(startTime)
	s.logger.Warn("Transaction not found on blockchain",
		zap.String("tx_id", txID),
		zap.Duration("duration", duration),
	)

	result := &VerifyByHashResponse{
		IsValid:            false,
		Message:            "Sản phẩm không thuộc thương hiệu chúng tôi",
		VerifiedAt:         time.Now(),
		VerificationMethod: "blockchain_query",
	}

	return result, nil
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
