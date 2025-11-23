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

package indexer

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/repository/db"
	"go.uber.org/zap"
)

// Service provides block indexing capabilities
type Service struct {
	gateway    *client.Gateway
	config    *config.FabricConfig
	db        *pgxpool.Pool
	queries   *db.Queries
	logger    *zap.Logger
	isRunning bool
	stopCh    chan struct{}
}

// NewService creates a new block indexer service
func NewService(
	gateway *client.Gateway,
	cfg *config.FabricConfig,
	dbPool *pgxpool.Pool,
	logger *zap.Logger,
) *Service {
	return &Service{
		gateway: gateway,
		config:  cfg,
		db:      dbPool,
		queries: db.New(),
		logger:  logger,
		stopCh:  make(chan struct{}),
	}
}

// Start starts the block indexer service
func (s *Service) Start(ctx context.Context) error {
	if s.isRunning {
		s.logger.Warn("Block indexer is already running")
		return nil
	}

	s.isRunning = true
	s.logger.Info("Starting block indexer service")

	// Start background indexing goroutine
	go s.indexBlocks(ctx)

	// Start event listener for new blocks
	go s.listenToBlockEvents(ctx)

	return nil
}

// Stop stops the block indexer service
func (s *Service) Stop() {
	if !s.isRunning {
		return
	}

	s.logger.Info("Stopping block indexer service")
	close(s.stopCh)
	s.isRunning = false
}

// IndexHistoricalBlocks indexes historical blocks from transactions in database
// Note: We can't directly query blocks from Fabric Gateway SDK, so we build blocks from transactions
func (s *Service) IndexHistoricalBlocks(ctx context.Context, channelName string) error {
	s.logger.Info("Indexing historical blocks from transactions", zap.String("channel", channelName))

	// Get all transactions from database
	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT block_number, block_hash, timestamp
		FROM transactions
		WHERE channel_name = $1 AND block_number IS NOT NULL AND block_number > 0
		ORDER BY block_number ASC
	`, channelName)
	if err != nil {
		return fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	// Index blocks from transactions
	for rows.Next() {
		var blockNumber sql.NullInt64
		var blockHash sql.NullString
		var timestamp time.Time

		if err := rows.Scan(&blockNumber, &blockHash, &timestamp); err != nil {
			continue
		}

		if !blockNumber.Valid {
			continue
		}

		// Check if block already exists
		exists, err := s.blockExists(ctx, channelName, uint64(blockNumber.Int64))
		if err != nil {
			continue
		}
		if exists {
			continue // Block already indexed
		}

		// Get transaction count for this block
		var txCount int
		err = s.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM transactions
			WHERE channel_name = $1 AND block_number = $2
		`, channelName, blockNumber.Int64).Scan(&txCount)
		if err != nil {
			txCount = 0
		}

		// Get previous block hash
		var previousHash string
		if blockNumber.Int64 > 0 {
			var prevHash sql.NullString
			err := s.db.QueryRow(ctx, `
				SELECT hash FROM blocks
				WHERE channel_name = $1 AND number = $2
			`, channelName, blockNumber.Int64-1).Scan(&prevHash)
			if err == nil && prevHash.Valid {
				previousHash = prevHash.String
			}
		}

		// Save block to database
		hash := ""
		if blockHash.Valid {
			hash = blockHash.String
		} else {
			hash = fmt.Sprintf("block_%d", blockNumber.Int64) // Fallback hash
		}

		_, err = s.db.Exec(ctx, `
			INSERT INTO blocks (
				number, hash, previous_hash, data_hash, transaction_count,
				channel_name, timestamp
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (channel_name, number) DO UPDATE SET
				hash = EXCLUDED.hash,
				previous_hash = EXCLUDED.previous_hash,
				transaction_count = EXCLUDED.transaction_count,
				timestamp = EXCLUDED.timestamp,
				updated_at = CURRENT_TIMESTAMP
		`, blockNumber.Int64, hash, previousHash, hash, txCount, channelName, timestamp)

		if err != nil {
			s.logger.Warn("Failed to save block",
				zap.Uint64("block_number", uint64(blockNumber.Int64)),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

// indexBlocks periodically indexes new blocks
func (s *Service) indexBlocks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Index every 30 seconds
	defer ticker.Stop()

	// Index initial historical blocks
	channelName := s.config.Channel
	if channelName != "" {
		if err := s.IndexHistoricalBlocks(ctx, channelName); err != nil {
			s.logger.Error("Failed to index initial historical blocks",
				zap.String("channel", channelName),
				zap.Error(err),
			)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			// Index blocks for configured channel
			channelName := s.config.Channel
			if channelName != "" {
				if err := s.IndexHistoricalBlocks(ctx, channelName); err != nil {
					s.logger.Error("Failed to index blocks",
						zap.String("channel", channelName),
						zap.Error(err),
					)
				}
			}
		}
	}
}

// listenToBlockEvents listens to new block events
func (s *Service) listenToBlockEvents(ctx context.Context) {
	channelName := s.config.Channel
	if channelName != "" {
		go s.listenToChannelBlockEvents(ctx, channelName)
	}
}

// listenToChannelBlockEvents listens to block events for a specific channel
func (s *Service) listenToChannelBlockEvents(ctx context.Context, channelName string) {
	network := s.gateway.GetNetwork(channelName)

	// Create block event listener (returns a channel)
	events, err := network.BlockEvents(ctx)
	if err != nil {
		s.logger.Error("Failed to create block event listener",
			zap.String("channel", channelName),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("Listening to block events", zap.String("channel", channelName))

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case event, ok := <-events:
			if !ok {
				// Channel closed
				s.logger.Info("Block event channel closed",
					zap.String("channel", channelName),
				)
				return
			}

			// Process block event
			if err := s.processBlockEvent(ctx, channelName, event); err != nil {
				s.logger.Error("Failed to process block event",
					zap.String("channel", channelName),
					zap.Error(err),
				)
			}
		}
	}
}

// processBlockEvent processes a block event
func (s *Service) processBlockEvent(ctx context.Context, channelName string, event *common.Block) error {
	// Extract block number from block header
	blockNumber := event.Header.Number
	s.logger.Info("Processing block event",
		zap.String("channel", channelName),
		zap.Uint64("block_number", blockNumber),
	)

	return s.indexBlockFromEvent(ctx, channelName, event)
}

// indexBlockFromEvent indexes a block from a block event
func (s *Service) indexBlockFromEvent(ctx context.Context, channelName string, block *common.Block) error {
	blockNumber := block.Header.Number

	// Check if block already exists
	exists, err := s.blockExists(ctx, channelName, blockNumber)
	if err != nil {
		return fmt.Errorf("failed to check if block exists: %w", err)
	}
	if exists {
		return nil // Block already indexed
	}

	// Parse block
	blockHash := fmt.Sprintf("%x", block.Header.DataHash)
	previousHash := ""
	if blockNumber > 0 {
		// Get previous block hash from database
		var prevHash sql.NullString
		err := s.db.QueryRow(ctx, `
			SELECT hash FROM blocks
			WHERE channel_name = $1 AND number = $2
		`, channelName, blockNumber-1).Scan(&prevHash)
		if err == nil && prevHash.Valid {
			previousHash = prevHash.String
		}
	}

	// Count transactions
	txCount := len(block.Data.Data)

	// Get timestamp (use current time as Fabric blocks don't have direct timestamp)
	timestamp := time.Now()

	// Save block to database
	_, err = s.db.Exec(ctx, `
		INSERT INTO blocks (
			number, hash, previous_hash, data_hash, transaction_count,
			channel_name, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (channel_name, number) DO UPDATE SET
			hash = EXCLUDED.hash,
			previous_hash = EXCLUDED.previous_hash,
			data_hash = EXCLUDED.data_hash,
			transaction_count = EXCLUDED.transaction_count,
			timestamp = EXCLUDED.timestamp,
			updated_at = CURRENT_TIMESTAMP
	`, blockNumber, blockHash, previousHash, fmt.Sprintf("%x", block.Header.DataHash), txCount, channelName, timestamp)

	if err != nil {
		return fmt.Errorf("failed to save block to database: %w", err)
	}

	s.logger.Debug("Indexed block",
		zap.String("channel", channelName),
		zap.Uint64("block_number", blockNumber),
		zap.Int("transaction_count", txCount),
	)

	return nil
}

// blockExists checks if a block exists in database
func (s *Service) blockExists(ctx context.Context, channelName string, blockNumber uint64) (bool, error) {
	var count int64
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM blocks
		WHERE channel_name = $1 AND number = $2
	`, channelName, blockNumber).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// getBlockchainHeight gets the current blockchain height from database
// Note: We can't directly query blocks from Fabric Gateway SDK
func (s *Service) getBlockchainHeight(ctx context.Context, channelName string) (uint64, error) {
	return s.getDBBlockHeight(ctx, channelName)
}

// getDBBlockHeight gets the current database block height
func (s *Service) getDBBlockHeight(ctx context.Context, channelName string) (uint64, error) {
	var maxBlock sql.NullInt64
	err := s.db.QueryRow(ctx, `
		SELECT MAX(number) FROM blocks
		WHERE channel_name = $1
	`, channelName).Scan(&maxBlock)

	if err != nil {
		return 0, err
	}

	if !maxBlock.Valid {
		return 0, nil
	}

	return uint64(maxBlock.Int64), nil
}

// GetSyncStatus returns the sync status
func (s *Service) GetSyncStatus(ctx context.Context, channelName string) (map[string]interface{}, error) {
	// Get database height (we can't directly query blockchain height from SDK)
	dbHeight, err := s.getDBBlockHeight(ctx, channelName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database height: %w", err)
	}

	// Get transaction count to estimate blockchain activity
	var txCount int64
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM transactions
		WHERE channel_name = $1
	`, channelName).Scan(&txCount)
	if err != nil {
		txCount = 0
	}

	status := "healthy"
	if dbHeight == 0 {
		status = "initializing"
	}

	return map[string]interface{}{
		"channel_name":     channelName,
		"db_height":        dbHeight,
		"transaction_count": txCount,
		"status":           status,
		"is_running":       s.isRunning,
	}, nil
}

