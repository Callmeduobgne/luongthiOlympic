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

package channel

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// Service provides channel management operations
type Service struct {
	gateway      *client.Gateway
	config       *config.FabricConfig
	logger       *zap.Logger
	corePath     string // Path to core directory with peer CLI and configtx
	organizationsPath string // Path to organizations directory
}

// NewService creates a new channel service
func NewService(gateway *client.Gateway, cfg *config.FabricConfig, logger *zap.Logger) *Service {
	// Default paths - can be configured via environment variables
	corePath := "/home/exp2/ibn/core"
	orgsPath := filepath.Join(corePath, "organizations")

	return &Service{
		gateway:           gateway,
		config:            cfg,
		logger:            logger,
		corePath:          corePath,
		organizationsPath: orgsPath,
	}
}

// CreateChannel creates a new channel
// Note: This requires peer CLI and configtxgen to be available
func (s *Service) CreateChannel(ctx context.Context, req *models.CreateChannelRequest) (*models.CreateChannelResponse, error) {
	s.logger.Info("Creating channel",
		zap.String("name", req.Name),
		zap.String("consortium", req.Consortium),
	)

	// Validate channel doesn't already exist
	network := s.gateway.GetNetwork(req.Name)
	if network != nil {
		return nil, fmt.Errorf("channel '%s' already exists", req.Name)
	}

	// Note: Full implementation would:
	// 1. Generate channel config using configtxgen
	// 2. Create genesis block
	// 3. Submit channel creation transaction to orderer
	// 4. Wait for channel to be created

	// For now, return a response indicating the operation would be performed
	// In production, this would execute:
	// - configtxgen -profile <profile> -outputCreateChannelTx <channel>.tx -channelID <name>
	// - peer channel create -o <orderer> -c <name> -f <channel>.tx --tls --cafile <ca>

	response := &models.CreateChannelResponse{
		Name:    req.Name,
		Status:  "pending",
		Message: fmt.Sprintf("Channel creation requires peer CLI. Channel '%s' would be created with consortium '%s'", req.Name, req.Consortium),
	}

	s.logger.Warn("Channel creation not fully implemented - requires peer CLI",
		zap.String("channel", req.Name),
		zap.String("message", "Use peer CLI: peer channel create -o <orderer> -c <name> -f <channel>.tx"),
	)

	return response, nil
}

// UpdateChannelConfig updates channel configuration
// Note: This requires configtxlator and peer CLI
func (s *Service) UpdateChannelConfig(ctx context.Context, channelName string, req *models.UpdateChannelConfigRequest) (*models.UpdateChannelConfigResponse, error) {
	s.logger.Info("Updating channel config",
		zap.String("channel", channelName),
	)

	// Validate channel exists
	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Note: Full implementation would:
	// 1. Fetch current channel config using peer CLI or QSCC
	// 2. Decode config using configtxlator
	// 3. Modify config JSON with updates
	// 4. Encode updated config using configtxlator
	// 5. Create config update transaction
	// 6. Sign and submit config update
	// 7. Wait for config update to be committed

	// For now, return a response indicating the operation would be performed
	response := &models.UpdateChannelConfigResponse{
		ChannelName: channelName,
		Status:      "pending",
		Message:     fmt.Sprintf("Channel config update requires configtxlator and peer CLI. Config for channel '%s' would be updated", channelName),
	}

	s.logger.Warn("Channel config update not fully implemented - requires configtxlator and peer CLI",
		zap.String("channel", channelName),
		zap.String("message", "Use configtxlator and peer CLI to update channel config"),
	)

	return response, nil
}

// JoinPeer joins a peer to a channel
// Note: This requires peer CLI
func (s *Service) JoinPeer(ctx context.Context, channelName string, req *models.JoinChannelRequest) (*models.JoinChannelResponse, error) {
	s.logger.Info("Joining peer to channel",
		zap.String("channel", channelName),
		zap.String("peer", req.PeerAddress),
	)

	// Validate channel exists
	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Determine block path
	blockPath := req.BlockPath
	if blockPath == "" {
		// Default genesis block path
		blockPath = filepath.Join(s.corePath, "channel-artifacts", fmt.Sprintf("%s.block", channelName))
	}

	// Note: Full implementation would:
	// 1. Check if genesis block exists
	// 2. Execute: peer channel join -b <block>
	// 3. Verify peer joined successfully
	// 4. Update peer's channel list

	// For now, try to check if we can access the block file
	// In production, this would execute:
	// - peer channel join -b <block> --tls --cafile <ca>

	response := &models.JoinChannelResponse{
		ChannelName: channelName,
		PeerAddress: req.PeerAddress,
		Status:      "pending",
		Message:     fmt.Sprintf("Peer join requires peer CLI. Peer '%s' would join channel '%s' using block '%s'", req.PeerAddress, channelName, blockPath),
	}

	s.logger.Warn("Peer join not fully implemented - requires peer CLI",
		zap.String("channel", channelName),
		zap.String("peer", req.PeerAddress),
		zap.String("block", blockPath),
		zap.String("message", "Use peer CLI: peer channel join -b <block>"),
	)

	return response, nil
}

// ListChannelMembers lists all members (organizations) in a channel
func (s *Service) ListChannelMembers(ctx context.Context, channelName string) (*models.ListChannelMembersResponse, error) {
	s.logger.Info("Listing channel members",
		zap.String("channel", channelName),
	)

	// Validate channel exists
	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Note: Full implementation would:
	// 1. Query channel config from peer
	// 2. Parse config to extract organizations
	// 3. Get peer information for each organization
	// 4. Build member list

	// For now, return basic info from config
	members := []models.ChannelMember{
		{
			MSPID:        s.config.MSPId,
			Organization: s.config.MSPId,
			Peers:        []string{s.config.PeerEndpoint},
		},
	}

	// Try to get additional peers from config
	if len(s.config.AdditionalPeers) > 0 {
		members[0].Peers = append(members[0].Peers, s.config.AdditionalPeers...)
	}

	response := &models.ListChannelMembersResponse{
		ChannelName: channelName,
		Members:     members,
		Total:       len(members),
	}

	return response, nil
}

// ListChannelPeers lists all peers in a channel
func (s *Service) ListChannelPeers(ctx context.Context, channelName string) (*models.ListChannelPeersResponse, error) {
	s.logger.Info("Listing peers in channel",
		zap.String("channel", channelName),
	)

	// Validate channel exists
	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Note: Full implementation would:
	// 1. Query discovery service for peers in channel
	// 2. Or query channel config to get organization peers
	// 3. Return list of peer addresses

	// For now, return peers from config
	peers := []string{s.config.PeerEndpoint}
	peers = append(peers, s.config.AdditionalPeers...)

	response := &models.ListChannelPeersResponse{
		ChannelName: channelName,
		Peers:       peers,
		Total:       len(peers),
	}

	return response, nil
}

// executePeerCommand executes a peer CLI command
// Helper function for executing peer commands
func (s *Service) executePeerCommand(ctx context.Context, args ...string) (string, error) {
	peerPath := filepath.Join(s.corePath, "bin", "peer")
	
	cmd := exec.CommandContext(ctx, peerPath, args...)
	
	// Set environment variables for peer command
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("CORE_PEER_LOCALMSPID=%s", s.config.MSPId),
		fmt.Sprintf("CORE_PEER_MSPCONFIGPATH=%s", s.organizationsPath),
		fmt.Sprintf("CORE_PEER_ADDRESS=%s", s.config.PeerEndpoint),
		fmt.Sprintf("CORE_PEER_TLS_ROOTCERT_FILE=%s", s.config.PeerTLSCAPath),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("peer command failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// checkPeerCLIAvailable checks if peer CLI is available
func (s *Service) checkPeerCLIAvailable() bool {
	peerPath := filepath.Join(s.corePath, "bin", "peer")
	cmd := exec.Command(peerPath, "version")
	err := cmd.Run()
	return err == nil
}

// checkConfigtxlatorAvailable checks if configtxlator is available
func (s *Service) checkConfigtxlatorAvailable() bool {
	configtxlatorPath := filepath.Join(s.corePath, "bin", "configtxlator")
	cmd := exec.Command(configtxlatorPath, "version")
	err := cmd.Run()
	return err == nil
}

