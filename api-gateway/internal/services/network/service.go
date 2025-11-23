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

package network

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// Service provides network information and query capabilities
type Service struct {
	gateway *client.Gateway
	config  *config.FabricConfig
	logger  *zap.Logger
}

// NewService creates a new network service
func NewService(gateway *client.Gateway, cfg *config.FabricConfig, logger *zap.Logger) *Service {
	return &Service{
		gateway: gateway,
		config:  cfg,
		logger:  logger,
	}
}

// GetChannelInfo gets information about a channel
func (s *Service) GetChannelInfo(ctx context.Context, channelName string) (*models.ChannelInfo, error) {
	s.logger.Info("Getting channel info", zap.String("channel", channelName))

	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Get block height (latest block number)
	blockHeight, err := s.getBlockHeight(ctx, network)
	if err != nil {
		s.logger.Warn("Failed to get block height", zap.Error(err))
		blockHeight = 0
	}

	// Get chaincodes on channel (from config or query)
	chaincodes := []string{s.config.Chaincode} // Default from config

	return &models.ChannelInfo{
		Name:        channelName,
		Peers:       []string{s.config.PeerEndpoint}, // From config
		Orderers:    []string{},                      // Would need to query
		Chaincodes:  chaincodes,
		BlockHeight: blockHeight,
	}, nil
}

// ListChannels lists all channels accessible by the gateway
// Note: Gateway SDK doesn't have direct method to list channels
// This returns the configured channel
func (s *Service) ListChannels(ctx context.Context) ([]*models.ChannelInfo, error) {
	s.logger.Info("Listing channels")

	// Gateway SDK doesn't support listing channels directly
	// Return the configured channel
	channelInfo, err := s.GetChannelInfo(ctx, s.config.Channel)
	if err != nil {
		return nil, err
	}

	return []*models.ChannelInfo{channelInfo}, nil
}

// GetChannelConfig gets channel configuration
// Note: This is a simplified version - full config requires querying system chaincode
func (s *Service) GetChannelConfig(ctx context.Context, channelName string) (*models.ChannelConfig, error) {
	s.logger.Info("Getting channel config", zap.String("channel", channelName))

	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Simplified config - would need to query system chaincode for full config
	return &models.ChannelConfig{
		Name:          channelName,
		Version:       "1.0",
		Consortium:    "SampleConsortium",
		Organizations: []string{s.config.MSPId},
		Capabilities: map[string]interface{}{
			"V2_0": true,
		},
		Policies: map[string]interface{}{},
		OrdererConfig: map[string]interface{}{},
		ApplicationConfig: map[string]interface{}{
			"capabilities": map[string]interface{}{
				"V2_0": true,
			},
		},
	}, nil
}

// GetNetworkInfo gets overall network information
func (s *Service) GetNetworkInfo(ctx context.Context) (*models.NetworkInfo, error) {
	s.logger.Info("Getting network info")

	channels, err := s.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to ChannelInfo slice
	channelInfos := make([]models.ChannelInfo, len(channels))
	for i, ch := range channels {
		channelInfos[i] = *ch
	}

	// Get peer info (from config)
	peers := []models.PeerInfo{
		{
			Name:     s.config.PeerHostOverride,
			Address:  s.config.PeerEndpoint,
			MSPID:    s.config.MSPId,
			Channels: []string{s.config.Channel},
			Chaincodes: []string{s.config.Chaincode},
			Status:   "connected",
		},
	}

	// Orderers (would need to query or get from config)
	orderers := []models.OrdererInfo{}

	return &models.NetworkInfo{
		Channels: channelInfos,
		Peers:    peers,
		Orderers: orderers,
		MSPs:     []string{s.config.MSPId},
	}, nil
}

// GetBlockInfo gets block information by block number
// Note: This uses transaction data from database
// For full block data with QSCC, peer direct access would be required
func (s *Service) GetBlockInfo(ctx context.Context, channelName string, blockNumber uint64) (*models.BlockInfo, error) {
	s.logger.Info("Getting block info",
		zap.String("channel", channelName),
		zap.Uint64("blockNumber", blockNumber),
	)

	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Note: Full implementation would use QSCC or peer direct access
	// For now, return basic info - this should be called from ExplorerService
	return nil, fmt.Errorf("use ExplorerService.GetBlock instead")
}

// GetTransactionInfo gets transaction information by transaction ID
func (s *Service) GetTransactionInfo(ctx context.Context, channelName, txID string) (*models.TransactionInfo, error) {
	s.logger.Info("Getting transaction info",
		zap.String("channel", channelName),
		zap.String("txId", txID),
	)

	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Note: Fabric Gateway SDK doesn't have direct method to get transaction by ID
	// This would require querying peer directly or using admin API
	// For now, return placeholder
	return nil, fmt.Errorf("getting transaction by ID requires peer query or admin API - not implemented yet")
}

// GetChaincodeInfoOnChannel gets chaincode information on a channel
func (s *Service) GetChaincodeInfoOnChannel(ctx context.Context, channelName, chaincodeName string) (*models.ChaincodeInfoOnChannel, error) {
	s.logger.Info("Getting chaincode info on channel",
		zap.String("channel", channelName),
		zap.String("chaincode", chaincodeName),
	)

	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Note: Would need to query lifecycle chaincode or use peer CLI
	// For now, return basic info from config
	return &models.ChaincodeInfoOnChannel{
		Name:                chaincodeName,
		Version:             "1.0", // Would need to query
		Sequence:            1,      // Would need to query
		EndorsementPlugin:   "escc",
		ValidationPlugin:    "vscc",
		InitRequired:        false,
		ApprovedOrganizations: []string{s.config.MSPId},
	}, nil
}

// getBlockHeight gets the current block height of a channel
func (s *Service) getBlockHeight(ctx context.Context, network *client.Network) (uint64, error) {
	// Try to query a simple chaincode function to get block info
	// Or use peer discovery to get block height
	// For now, return 0 as placeholder
	return 0, nil
}

