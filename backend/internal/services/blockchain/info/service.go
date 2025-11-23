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

package info

import (
	"context"
	"encoding/hex"
	"fmt"

	"go.uber.org/zap"
)

// FabricClient interface for blockchain queries (using raw bytes, no protobuf parsing)
type FabricClient interface {
	// Query using qscc (Query System Chaincode)
	EvaluateTransaction(ctx context.Context, chaincodeName, functionName string, args ...string) ([]byte, error)
}

// Service handles blockchain info queries without protobuf parsing
type Service struct {
	client    FabricClient
	channelID string
	logger    *zap.Logger
}

// NewService creates a new blockchain info service
func NewService(client FabricClient, channelID string, logger *zap.Logger) *Service {
	return &Service{
		client:    client,
		channelID: channelID,
		logger:    logger,
	}
}

// GetBlockByNumber retrieves block by number (returns hex-encoded raw block)
func (s *Service) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*BlockInfo, error) {
	// Use qscc to get block
	result, err := s.client.EvaluateTransaction(
		ctx,
		"qscc",
		"GetBlockByNumber",
		s.channelID,
		fmt.Sprintf("%d", blockNumber),
	)
	if err != nil {
		s.logger.Error("Failed to get block by number",
			zap.Uint64("block_number", blockNumber),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get block by number: %w", err)
	}

	return &BlockInfo{
		BlockNumber: blockNumber,
		RawBlock:    hex.EncodeToString(result),
		Size:        len(result),
	}, nil
}

// GetChannelInfo retrieves channel information (returns hex-encoded blockchain info)
func (s *Service) GetChannelInfo(ctx context.Context) (*ChannelInfo, error) {
	// Use qscc to get blockchain info
	result, err := s.client.EvaluateTransaction(
		ctx,
		"qscc",
		"GetChainInfo",
		s.channelID,
	)
	if err != nil {
		s.logger.Error("Failed to get channel info", zap.Error(err))
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}

	return &ChannelInfo{
		ChannelID: s.channelID,
		RawInfo:   hex.EncodeToString(result),
		Size:      len(result),
	}, nil
}

// GetBlockByTxID retrieves block by transaction ID (returns hex-encoded raw block)
func (s *Service) GetBlockByTxID(ctx context.Context, txID string) (*BlockInfo, error) {
	result, err := s.client.EvaluateTransaction(
		ctx,
		"qscc",
		"GetBlockByTxID",
		s.channelID,
		txID,
	)
	if err != nil {
		s.logger.Error("Failed to get block by txID",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get block by txID: %w", err)
	}

	return &BlockInfo{
		BlockNumber: 0, // Unknown without parsing
		RawBlock:    hex.EncodeToString(result),
		Size:        len(result),
	}, nil
}

// GetTransactionByID retrieves transaction by ID (returns hex-encoded transaction)
func (s *Service) GetTransactionByID(ctx context.Context, txID string) (string, error) {
	result, err := s.client.EvaluateTransaction(
		ctx,
		"qscc",
		"GetTransactionByID",
		s.channelID,
		txID,
	)
	if err != nil {
		s.logger.Error("Failed to get transaction by ID",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get transaction: %w", err)
	}

	return hex.EncodeToString(result), nil
}

