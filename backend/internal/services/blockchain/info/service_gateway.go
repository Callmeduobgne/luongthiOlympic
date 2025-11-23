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

// GatewayClient interface for Gateway queries
// This matches transaction.GatewayClient interface
type GatewayClient interface {
	QueryChaincode(ctx context.Context, channelName, chaincodeName, functionName string, args []string) ([]byte, error)
}

// ServiceViaGateway handles blockchain info queries via Gateway (REQUIRED)
type ServiceViaGateway struct {
	gatewayClient GatewayClient
	channelID     string
	logger        *zap.Logger
}

// NewServiceViaGateway creates a new blockchain info service using Gateway
// NOTE: Backend MUST use Gateway for all blockchain operations
func NewServiceViaGateway(gatewayClient GatewayClient, channelID string, logger *zap.Logger) *ServiceViaGateway {
	if gatewayClient == nil {
		logger.Fatal("Gateway client is required - Backend cannot connect directly to Fabric")
	}
	return &ServiceViaGateway{
		gatewayClient: gatewayClient,
		channelID:     channelID,
		logger:        logger,
	}
}

// GetBlockByNumber retrieves block by number (returns hex-encoded raw block) via Gateway
func (s *ServiceViaGateway) GetBlockByNumber(ctx context.Context, blockNumber uint64) (*BlockInfo, error) {
	// Use qscc to get block via Gateway
	result, err := s.gatewayClient.QueryChaincode(
		ctx,
		s.channelID,
		"qscc",
		"GetBlockByNumber",
		[]string{s.channelID, fmt.Sprintf("%d", blockNumber)},
	)
	if err != nil {
		s.logger.Error("Failed to get block by number via Gateway",
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

// GetChannelInfo retrieves channel information (returns hex-encoded blockchain info) via Gateway
func (s *ServiceViaGateway) GetChannelInfo(ctx context.Context) (*ChannelInfo, error) {
	// Use qscc to get blockchain info via Gateway
	result, err := s.gatewayClient.QueryChaincode(
		ctx,
		s.channelID,
		"qscc",
		"GetChainInfo",
		[]string{s.channelID},
	)
	if err != nil {
		s.logger.Error("Failed to get channel info via Gateway", zap.Error(err))
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}

	return &ChannelInfo{
		ChannelID: s.channelID,
		RawInfo:   hex.EncodeToString(result),
		Size:      len(result),
	}, nil
}

// GetBlockByTxID retrieves block by transaction ID (returns hex-encoded raw block) via Gateway
func (s *ServiceViaGateway) GetBlockByTxID(ctx context.Context, txID string) (*BlockInfo, error) {
	result, err := s.gatewayClient.QueryChaincode(
		ctx,
		s.channelID,
		"qscc",
		"GetBlockByTxID",
		[]string{s.channelID, txID},
	)
	if err != nil {
		s.logger.Error("Failed to get block by txID via Gateway",
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

// GetTransactionByID retrieves transaction by ID (returns hex-encoded transaction) via Gateway
func (s *ServiceViaGateway) GetTransactionByID(ctx context.Context, txID string) (string, error) {
	result, err := s.gatewayClient.QueryChaincode(
		ctx,
		s.channelID,
		"qscc",
		"GetTransactionByID",
		[]string{s.channelID, txID},
	)
	if err != nil {
		s.logger.Error("Failed to get transaction by ID via Gateway",
			zap.String("tx_id", txID),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get transaction: %w", err)
	}

	return hex.EncodeToString(result), nil
}

