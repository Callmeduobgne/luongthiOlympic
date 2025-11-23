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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/models"
)

// QueryBlocksFromDB queries blocks from database
func QueryBlocksFromDB(
	ctx context.Context,
	db *pgxpool.Pool,
	channelName string,
	limit int,
	offset int,
) ([]*models.BlockInfo, int64, error) {
	// Query blocks from database
	rows, err := db.Query(ctx, `
		SELECT 
			number, hash, previous_hash, data_hash, transaction_count,
			channel_name, timestamp
		FROM blocks
		WHERE channel_name = $1
		ORDER BY number DESC
		LIMIT $2 OFFSET $3
	`, channelName, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*models.BlockInfo
	for rows.Next() {
		var block models.BlockInfo
		var timestamp time.Time
		var channelName string
		err := rows.Scan(
			&block.Number,
			&block.Hash,
			&block.PreviousHash,
			&block.DataHash,
			&block.TransactionCount,
			&channelName,
			&timestamp,
		)
		if err != nil {
			continue
		}
		block.Timestamp = timestamp.Format(time.RFC3339)
		block.Transactions = []string{} // Will be populated from transactions table if needed
		blocks = append(blocks, &block)
	}

	// Get total count
	var total int64
	err = db.QueryRow(ctx, `
		SELECT COUNT(*) FROM blocks
		WHERE channel_name = $1
	`, channelName).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count blocks: %w", err)
	}

	return blocks, total, nil
}

// GetBlockFromDB gets a single block from database
func GetBlockFromDB(
	ctx context.Context,
	db *pgxpool.Pool,
	channelName string,
	blockNumber uint64,
) (*models.BlockInfo, error) {
	var block models.BlockInfo
	var timestamp time.Time
	var dbChannelName string
	err := db.QueryRow(ctx, `
		SELECT 
			number, hash, previous_hash, data_hash, transaction_count,
			channel_name, timestamp
		FROM blocks
		WHERE channel_name = $1 AND number = $2
	`, channelName, blockNumber).Scan(
		&block.Number,
		&block.Hash,
		&block.PreviousHash,
		&block.DataHash,
		&block.TransactionCount,
		&dbChannelName,
		&timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	block.Timestamp = timestamp.Format(time.RFC3339)
	block.Transactions = []string{} // Will be populated from transactions table if needed

	return &block, nil
}

