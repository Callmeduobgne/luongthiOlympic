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

package explorer

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/transaction"
	"go.uber.org/zap"
)

// Service provides block explorer capabilities
type Service struct {
	gateway           *client.Gateway
	config            *config.FabricConfig
	transactionService *transaction.Service
	db                *pgxpool.Pool
	logger            *zap.Logger
}

// NewService creates a new block explorer service
func NewService(
	gateway *client.Gateway,
	cfg *config.FabricConfig,
	transactionService *transaction.Service,
	db *pgxpool.Pool,
	logger *zap.Logger,
) *Service {
	return &Service{
		gateway:           gateway,
		config:            cfg,
		transactionService: transactionService,
		db:                db,
		logger:            logger,
	}
}

// GetBlock gets block information by block number
// First tries to get from blocks table, then falls back to building from transactions
func (s *Service) GetBlock(ctx context.Context, channelName string, blockNumber uint64) (*models.BlockInfo, error) {
	s.logger.Info("Getting block",
		zap.String("channel", channelName),
		zap.Uint64("blockNumber", blockNumber),
	)

	// Try to get block from database first (faster and more complete)
	if s.db != nil {
		block, err := GetBlockFromDB(ctx, s.db, channelName, blockNumber)
		if err == nil && block != nil {
			// Get transactions for this block
			query := &models.TransactionListQuery{
				ChannelName: channelName,
				Limit:        1000,
				Offset:       0,
			}
			transactions, _, err := s.transactionService.ListTransactions(ctx, query)
			if err == nil {
				var blockTxs []string
				for _, tx := range transactions {
					if tx.BlockNumber == blockNumber {
						blockTxs = append(blockTxs, tx.TxID)
					}
				}
				block.Transactions = blockTxs
				block.TransactionCount = len(blockTxs)
			}
			return block, nil
		}
		// If not found in DB, fall back to building from transactions
		s.logger.Debug("Block not found in database, building from transactions",
			zap.String("channel", channelName),
			zap.Uint64("blockNumber", blockNumber),
		)
	}

	// Fallback: Query transactions in this block from database
	query := &models.TransactionListQuery{
		ChannelName: channelName,
		Limit:        1000, // Max transactions per block
		Offset:       0,
	}

	// Get all transactions
	transactions, _, err := s.transactionService.ListTransactions(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Filter transactions by block number
	var blockTxs []string
	var blockTimestamp time.Time
	var blockHash string

	for _, tx := range transactions {
		if tx.BlockNumber == blockNumber {
			blockTxs = append(blockTxs, tx.TxID)
			if blockTimestamp.IsZero() || tx.Timestamp.Before(blockTimestamp) {
				blockTimestamp = tx.Timestamp
			}
			if blockHash == "" && tx.BlockHash != "" {
				blockHash = tx.BlockHash
			}
		}
	}

	// If no transactions found, return empty block
	if len(blockTxs) == 0 {
		return &models.BlockInfo{
			Number:           blockNumber,
			Hash:             "",
			PreviousHash:     "",
			DataHash:         "",
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
			Transactions:     []string{},
			TransactionCount: 0,
		}, nil
	}

	// Get previous block hash (from previous block's transactions)
	var previousHash string
	if blockNumber > 0 {
		prevBlock, err := s.GetBlock(ctx, channelName, blockNumber-1)
		if err == nil && prevBlock != nil {
			previousHash = prevBlock.Hash
		}
	}

	return &models.BlockInfo{
		Number:           blockNumber,
		Hash:             blockHash,
		PreviousHash:     previousHash,
		DataHash:         "",
		Timestamp:        blockTimestamp.UTC().Format(time.RFC3339),
		Transactions:     blockTxs,
		TransactionCount: len(blockTxs),
	}, nil
}

// GetLatestBlock gets the latest block information
func (s *Service) GetLatestBlock(ctx context.Context, channelName string) (*models.BlockInfo, error) {
	s.logger.Info("Getting latest block", zap.String("channel", channelName))

	// Query transactions to find latest block number
	query := &models.TransactionListQuery{
		ChannelName: channelName,
		Limit:        1,
		Offset:       0,
	}

	transactions, _, err := s.transactionService.ListTransactions(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	if len(transactions) == 0 {
		// No transactions, return block 0
		return &models.BlockInfo{
			Number:           0,
			Hash:             "",
			PreviousHash:     "",
			DataHash:         "",
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
			Transactions:     []string{},
			TransactionCount: 0,
		}, nil
	}

	// Find max block number
	maxBlockNumber := uint64(0)
	for _, tx := range transactions {
		if tx.BlockNumber > maxBlockNumber {
			maxBlockNumber = tx.BlockNumber
		}
	}

	// Get all transactions to find latest block
	query.Limit = 10000
	allTxs, _, err := s.transactionService.ListTransactions(ctx, query)
	if err == nil {
		for _, tx := range allTxs {
			if tx.BlockNumber > maxBlockNumber {
				maxBlockNumber = tx.BlockNumber
			}
		}
	}

	return s.GetBlock(ctx, channelName, maxBlockNumber)
}

// ListBlocks lists blocks with pagination
func (s *Service) ListBlocks(
	ctx context.Context,
	channelName string,
	limit int,
	offset int,
) ([]*models.BlockInfo, int64, error) {
	s.logger.Info("Listing blocks",
		zap.String("channel", channelName),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	// Try to query from blocks table first (faster)
	if s.db != nil {
		blocks, total, err := QueryBlocksFromDB(ctx, s.db, channelName, limit, offset)
		if err == nil && len(blocks) > 0 {
			return blocks, total, nil
		}
		// If no blocks in DB, fall back to building from transactions
		s.logger.Debug("No blocks in database, building from transactions",
			zap.String("channel", channelName),
		)
	}

	// Fallback: Query all transactions to build block list
	query := &models.TransactionListQuery{
		ChannelName: channelName,
		Limit:        10000, // Get all transactions
		Offset:       0,
	}

	transactions, _, err := s.transactionService.ListTransactions(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Build block map
	blockMap := make(map[uint64]*models.BlockInfo)
	for _, tx := range transactions {
		if tx.BlockNumber == 0 {
			continue // Skip transactions without block number
		}

		block, exists := blockMap[tx.BlockNumber]
		if !exists {
			block = &models.BlockInfo{
				Number:           tx.BlockNumber,
				Hash:             tx.BlockHash,
				PreviousHash:     "",
				DataHash:         "",
				Timestamp:        tx.Timestamp.Format(time.RFC3339),
				Transactions:     []string{},
				TransactionCount: 0,
			}
			blockMap[tx.BlockNumber] = block
		}

		block.Transactions = append(block.Transactions, tx.TxID)
		block.TransactionCount++
	}

	// Convert map to slice and sort by block number (descending)
	blocks := make([]*models.BlockInfo, 0, len(blockMap))
	for _, block := range blockMap {
		blocks = append(blocks, block)
	}

	// Simple sort by block number (descending)
	for i := 0; i < len(blocks)-1; i++ {
		for j := i + 1; j < len(blocks); j++ {
			if blocks[i].Number < blocks[j].Number {
				blocks[i], blocks[j] = blocks[j], blocks[i]
			}
		}
	}

	// Apply pagination
	totalBlocks := int64(len(blocks))
	start := offset
	end := offset + limit
	if start > len(blocks) {
		start = len(blocks)
	}
	if end > len(blocks) {
		end = len(blocks)
	}

	if start < end {
		blocks = blocks[start:end]
	} else {
		blocks = []*models.BlockInfo{}
	}

	return blocks, totalBlocks, nil
}

// GetTransactionByBlock gets all transactions in a block
func (s *Service) GetTransactionByBlock(
	ctx context.Context,
	channelName string,
	blockNumber uint64,
) ([]*models.Transaction, error) {
	s.logger.Info("Getting transactions by block",
		zap.String("channel", channelName),
		zap.Uint64("blockNumber", blockNumber),
	)

	// Query transactions
	query := &models.TransactionListQuery{
		ChannelName: channelName,
		Limit:        1000,
		Offset:       0,
	}

	transactions, _, err := s.transactionService.ListTransactions(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Filter by block number
	var result []*models.Transaction
	for _, tx := range transactions {
		if tx.BlockNumber == blockNumber {
			result = append(result, tx)
		}
	}

	return result, nil
}

