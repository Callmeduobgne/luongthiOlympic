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
	"encoding/json"
	"fmt"

	"github.com/ibn-network/backend/internal/infrastructure/gateway"
	"go.uber.org/zap"
)

// DiscoveryService handles network discovery operations via Gateway
type DiscoveryService struct {
	gatewayClient *gateway.Client
	logger        *zap.Logger
}

// NetworkInfo represents network information from Gateway
type NetworkInfo struct {
	Channels []ChannelInfo  `json:"channels"`
	Peers    []PeerInfo     `json:"peers"`
	Orderers []OrdererInfo  `json:"orderers"`
	MSPs     []string       `json:"msps"`
}

// ChannelInfo represents channel information
type ChannelInfo struct {
	Name        string   `json:"name"`
	Peers       []string `json:"peers"`
	Orderers    []string `json:"orderers"`
	Chaincodes  []string `json:"chaincodes"`
	BlockHeight uint64   `json:"blockHeight"`
}

// PeerInfo represents peer information
type PeerInfo struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	MSPID       string   `json:"mspId"`
	Channels    []string `json:"channels"`
	Chaincodes  []string `json:"chaincodes"`
	Status      string   `json:"status"`
	BlockHeight uint64   `json:"blockHeight,omitempty"`
}

// OrdererInfo represents orderer information
type OrdererInfo struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	MSPID    string `json:"mspId"`
	Status   string `json:"status"`
	IsLeader bool   `json:"isLeader"`
}

// NetworkTopology represents network topology
type NetworkTopology struct {
	Peers    []PeerInfo     `json:"peers"`
	Orderers []OrdererInfo  `json:"orderers"`
	CAs      []interface{}  `json:"cas"`
	Channels []string       `json:"channels"`
	MSPs     []string       `json:"msps"`
}

// NewDiscoveryService creates a new network discovery service
func NewDiscoveryService(gatewayClient *gateway.Client, logger *zap.Logger) *DiscoveryService {
	return &DiscoveryService{
		gatewayClient: gatewayClient,
		logger:        logger,
	}
}

// GetNetworkInfo gets overall network information from Gateway
func (s *DiscoveryService) GetNetworkInfo(ctx context.Context) (*NetworkInfo, error) {
	s.logger.Info("Getting network info from Gateway")

	body, err := s.gatewayClient.Get(ctx, "/api/v1/network/info")
	if err != nil {
		return nil, fmt.Errorf("failed to get network info from Gateway: %w", err)
	}

	var apiResp struct {
		Success bool        `json:"success"`
		Data    NetworkInfo `json:"data"`
		Error   interface{} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gateway response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("Gateway returned error: %v", apiResp.Error)
	}

	return &apiResp.Data, nil
}

// ListPeers lists all peers from Gateway
func (s *DiscoveryService) ListPeers(ctx context.Context) ([]PeerInfo, error) {
	s.logger.Info("Listing peers from Gateway")

	body, err := s.gatewayClient.Get(ctx, "/api/v1/network/peers")
	if err != nil {
		return nil, fmt.Errorf("failed to list peers from Gateway: %w", err)
	}

	var apiResp struct {
		Success bool       `json:"success"`
		Data    []PeerInfo `json:"data"`
		Error   interface{} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gateway response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("Gateway returned error: %v", apiResp.Error)
	}

	return apiResp.Data, nil
}

// ListOrderers lists all orderers from Gateway
func (s *DiscoveryService) ListOrderers(ctx context.Context) ([]OrdererInfo, error) {
	s.logger.Info("Listing orderers from Gateway")

	body, err := s.gatewayClient.Get(ctx, "/api/v1/network/orderers")
	if err != nil {
		return nil, fmt.Errorf("failed to list orderers from Gateway: %w", err)
	}

	var apiResp struct {
		Success bool         `json:"success"`
		Data    []OrdererInfo `json:"data"`
		Error   interface{}  `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gateway response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("Gateway returned error: %v", apiResp.Error)
	}

	return apiResp.Data, nil
}

// ListChannels lists all channels from Gateway
func (s *DiscoveryService) ListChannels(ctx context.Context) ([]ChannelInfo, error) {
	s.logger.Info("Listing channels from Gateway")

	body, err := s.gatewayClient.Get(ctx, "/api/v1/network/channels")
	if err != nil {
		return nil, fmt.Errorf("failed to list channels from Gateway: %w", err)
	}

	var apiResp struct {
		Success bool         `json:"success"`
		Data    []ChannelInfo `json:"data"`
		Error   interface{}  `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gateway response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("Gateway returned error: %v", apiResp.Error)
	}

	return apiResp.Data, nil
}

// GetChannelInfo gets channel information by name from Gateway
func (s *DiscoveryService) GetChannelInfo(ctx context.Context, name string) (*ChannelInfo, error) {
	s.logger.Info("Getting channel info from Gateway", zap.String("channel", name))

	body, err := s.gatewayClient.Get(ctx, fmt.Sprintf("/api/v1/network/channels/%s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info from Gateway: %w", err)
	}

	var apiResp struct {
		Success bool        `json:"success"`
		Data    ChannelInfo `json:"data"`
		Error   interface{} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gateway response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("Gateway returned error: %v", apiResp.Error)
	}

	return &apiResp.Data, nil
}

// GetTopology gets network topology from Gateway
func (s *DiscoveryService) GetTopology(ctx context.Context) (*NetworkTopology, error) {
	s.logger.Info("Getting network topology from Gateway")

	body, err := s.gatewayClient.Get(ctx, "/api/v1/network/topology")
	if err != nil {
		return nil, fmt.Errorf("failed to get topology from Gateway: %w", err)
	}

	var apiResp struct {
		Success bool           `json:"success"`
		Data    NetworkTopology `json:"data"`
		Error   interface{}    `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gateway response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("Gateway returned error: %v", apiResp.Error)
	}

	return &apiResp.Data, nil
}

