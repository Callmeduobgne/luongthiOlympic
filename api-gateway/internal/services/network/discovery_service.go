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
	"net"
	"strings"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// DiscoveryService provides network discovery capabilities
type DiscoveryService struct {
	gateway *client.Gateway
	config  *config.FabricConfig
	logger  *zap.Logger
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(gateway *client.Gateway, cfg *config.FabricConfig, logger *zap.Logger) *DiscoveryService {
	return &DiscoveryService{
		gateway: gateway,
		config:  cfg,
		logger:  logger,
	}
}

// ListPeers lists all peers in the network
func (s *DiscoveryService) ListPeers(ctx context.Context) ([]*models.PeerInfo, error) {
	s.logger.Info("Listing peers")

	peers := []*models.PeerInfo{}

	// Get primary peer from config
	primaryPeer := &models.PeerInfo{
		Name:       s.config.PeerHostOverride,
		Address:    s.config.PeerEndpoint,
		MSPID:      s.config.MSPId,
		Channels:   []string{s.config.Channel},
		Chaincodes: []string{s.config.Chaincode},
		Status:     "connected",
	}

	// Try to get block height for primary peer
	network := s.gateway.GetNetwork(s.config.Channel)
	if network != nil {
		blockHeight, err := s.getBlockHeight(ctx, network)
		if err == nil {
			primaryPeer.BlockHeight = blockHeight
		}
	}

	peers = append(peers, primaryPeer)

	// Add additional peers from config
	for _, peerEndpoint := range s.config.AdditionalPeers {
		if peerEndpoint == "" {
			continue
		}

		// Parse peer endpoint (format: host:port or hostname:port)
		parts := strings.Split(peerEndpoint, ":")
		if len(parts) != 2 {
			s.logger.Warn("Invalid peer endpoint format", zap.String("endpoint", peerEndpoint))
			continue
		}

		peerName := parts[0]
		peerAddress := peerEndpoint

		// Check if peer is healthy
		status := "unknown"
		if s.checkPeerConnection(ctx, peerAddress) {
			status = "connected"
		} else {
			status = "disconnected"
		}

		peerInfo := &models.PeerInfo{
			Name:        peerName,
			Address:     peerAddress,
			MSPID:       s.config.MSPId,
			Channels:    []string{s.config.Channel},
			Chaincodes:  []string{s.config.Chaincode},
			Status:      status,
			BlockHeight: 0, // Would need to query from peer
		}

		peers = append(peers, peerInfo)
	}

	return peers, nil
}

// getBlockHeight gets the current block height of a channel
func (s *DiscoveryService) getBlockHeight(ctx context.Context, network *client.Network) (uint64, error) {
	// Try to query a simple chaincode function to get block info
	// For now, we'll use a workaround: try to get latest block info
	// This is a simplified version - full implementation would query QSCC

	// Return 0 as placeholder - in production, would query from peer
	return 0, nil
}

// checkPeerConnection checks if peer is reachable via gRPC
func (s *DiscoveryService) checkPeerConnection(ctx context.Context, address string) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		s.logger.Debug("Peer dial failed", zap.String("address", address), zap.Error(err))
		return false
	}
	defer conn.Close()
	return true
}

// GetPeer gets peer details by ID/name
func (s *DiscoveryService) GetPeer(ctx context.Context, peerID string) (*models.PeerInfo, error) {
	s.logger.Info("Getting peer", zap.String("peerId", peerID))

	peers, err := s.ListPeers(ctx)
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		if peer.Name == peerID || peer.Address == peerID {
			return peer, nil
		}
	}

	return nil, fmt.Errorf("peer '%s' not found", peerID)
}

// ListOrderers lists all orderers in the network
func (s *DiscoveryService) ListOrderers(ctx context.Context) ([]*models.OrdererInfo, error) {
	s.logger.Info("Listing orderers")

	orderers := []*models.OrdererInfo{}

	// Get orderers from config
	ordererList := s.config.Orderers

	// If no orderers in config, use defaults from IBN network
	if len(ordererList) == 0 {
		ordererList = []string{
			"orderer.ibn.vn:7050",
			"orderer1.ibn.vn:8050",
			"orderer2.ibn.vn:9050",
		}
	}

	// Build orderer info list
	for i, ordererEndpoint := range ordererList {
		if ordererEndpoint == "" {
			continue
		}

		// Parse orderer endpoint (format: host:port)
		parts := strings.Split(ordererEndpoint, ":")
		if len(parts) != 2 {
			s.logger.Warn("Invalid orderer endpoint format", zap.String("endpoint", ordererEndpoint))
			continue
		}

		ordererName := parts[0]
		ordererAddress := ordererEndpoint

		// Check if orderer is healthy
		status := "unknown"
		if s.checkOrdererConnection(ctx, ordererAddress) {
			status = "healthy"
		} else {
			status = "unhealthy"
		}

		// First orderer is typically the leader in Raft
		isLeader := (i == 0)

		ordererInfo := &models.OrdererInfo{
			Name:     ordererName,
			Address:  ordererAddress,
			MSPID:    "OrdererMSP", // Default orderer MSP
			Status:   status,
			IsLeader: isLeader,
		}

		orderers = append(orderers, ordererInfo)
	}

	return orderers, nil
}

// checkOrdererConnection checks if orderer is reachable via gRPC
func (s *DiscoveryService) checkOrdererConnection(ctx context.Context, address string) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		s.logger.Debug("Orderer dial failed", zap.String("address", address), zap.Error(err))
		return false
	}
	defer conn.Close()
	return true
}

// GetOrderer gets orderer details by ID/name
func (s *DiscoveryService) GetOrderer(ctx context.Context, ordererID string) (*models.OrdererInfo, error) {
	s.logger.Info("Getting orderer", zap.String("ordererId", ordererID))

	orderers, err := s.ListOrderers(ctx)
	if err != nil {
		return nil, err
	}

	for _, orderer := range orderers {
		if orderer.Name == ordererID || orderer.Address == ordererID {
			return orderer, nil
		}
	}

	return nil, fmt.Errorf("orderer '%s' not found", ordererID)
}

// ListCAs lists all Fabric CAs in the network
func (s *DiscoveryService) ListCAs(ctx context.Context) ([]*models.CAInfo, error) {
	s.logger.Info("Listing CAs")

	cas := []*models.CAInfo{}

	// Get CAs from config
	caList := s.config.CAEndpoints

	// If no CAs in config, use default from IBN network
	if len(caList) == 0 {
		caList = []string{
			"ca.org1.ibn.vn:7054",
		}
	}

	// Build CA info list
	for _, caEndpoint := range caList {
		if caEndpoint == "" {
			continue
		}

		// Parse CA endpoint (format: host:port)
		parts := strings.Split(caEndpoint, ":")
		if len(parts) != 2 {
			s.logger.Warn("Invalid CA endpoint format", zap.String("endpoint", caEndpoint))
			continue
		}

		caName := parts[0]
		caAddress := caEndpoint

		// Check if CA is healthy (HTTP endpoint for Fabric CA)
		status := "unknown"
		if s.checkCAConnection(ctx, caAddress) {
			status = "healthy"
		} else {
			status = "unhealthy"
		}

		caInfo := &models.CAInfo{
			Name:    caName,
			Address: caAddress,
			MSPID:   s.config.MSPId,
			Status:  status,
		}

		cas = append(cas, caInfo)
	}

	return cas, nil
}

// checkCAConnection checks if CA is reachable via HTTP
func (s *DiscoveryService) checkCAConnection(ctx context.Context, address string) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Fabric CA uses HTTP, try to connect to TCP port
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GetTopology gets network topology
func (s *DiscoveryService) GetTopology(ctx context.Context) (*models.NetworkTopology, error) {
	s.logger.Info("Getting network topology")

	// Get all components in parallel for better performance
	type result struct {
		peers    []*models.PeerInfo
		orderers []*models.OrdererInfo
		cas      []*models.CAInfo
		err      error
	}

	resultChan := make(chan result, 3)

	// Get peers
	go func() {
		peers, err := s.ListPeers(ctx)
		resultChan <- result{peers: peers, err: err}
	}()

	// Get orderers
	go func() {
		orderers, err := s.ListOrderers(ctx)
		resultChan <- result{orderers: orderers, err: err}
	}()

	// Get CAs
	go func() {
		cas, err := s.ListCAs(ctx)
		resultChan <- result{cas: cas, err: err}
	}()

	// Collect results
	var peers []*models.PeerInfo
	var orderers []*models.OrdererInfo
	var cas []*models.CAInfo

	for i := 0; i < 3; i++ {
		res := <-resultChan
		if res.err != nil {
			s.logger.Warn("Failed to get component", zap.Error(res.err))
		}
		if res.peers != nil {
			peers = res.peers
		}
		if res.orderers != nil {
			orderers = res.orderers
		}
		if res.cas != nil {
			cas = res.cas
		}
	}

	// Default to empty slices if nil
	if peers == nil {
		peers = []*models.PeerInfo{}
	}
	if orderers == nil {
		orderers = []*models.OrdererInfo{}
	}
	if cas == nil {
		cas = []*models.CAInfo{}
	}

	// Get channels
	network := s.gateway.GetNetwork(s.config.Channel)
	channels := []string{}
	if network != nil {
		channels = append(channels, s.config.Channel)
	}

	// Collect unique MSPs
	msps := make(map[string]bool)
	msps[s.config.MSPId] = true
	msps["OrdererMSP"] = true // Add orderer MSP
	for _, peer := range peers {
		if peer.MSPID != "" {
			msps[peer.MSPID] = true
		}
	}
	for _, ca := range cas {
		if ca.MSPID != "" {
			msps[ca.MSPID] = true
		}
	}

	mspList := make([]string, 0, len(msps))
	for msp := range msps {
		mspList = append(mspList, msp)
	}

	// Build topology
	topology := &models.NetworkTopology{
		Peers:    peers,
		Orderers: orderers,
		CAs:      cas,
		Channels: channels,
		MSPs:     mspList,
	}

	return topology, nil
}

// GetPeersInChannel gets peers in a specific channel
func (s *DiscoveryService) GetPeersInChannel(ctx context.Context, channelName string) ([]*models.PeerInfo, error) {
	s.logger.Info("Getting peers in channel", zap.String("channel", channelName))

	network := s.gateway.GetNetwork(channelName)
	if network == nil {
		return nil, fmt.Errorf("channel '%s' not found", channelName)
	}

	// Get all peers
	peers, err := s.ListPeers(ctx)
	if err != nil {
		return nil, err
	}

	// Filter peers that are in this channel
	channelPeers := []*models.PeerInfo{}
	for _, peer := range peers {
		for _, ch := range peer.Channels {
			if ch == channelName {
				channelPeers = append(channelPeers, peer)
				break
			}
		}
	}

	return channelPeers, nil
}

// CheckPeerHealth checks peer health status
func (s *DiscoveryService) CheckPeerHealth(ctx context.Context, peerID string) (*models.HealthStatus, error) {
	s.logger.Info("Checking peer health", zap.String("peerId", peerID))

	peer, err := s.GetPeer(ctx, peerID)
	if err != nil {
		return nil, err
	}

	// Try to ping peer gRPC endpoint
	status := "unhealthy"
	details := make(map[string]interface{})

	if s.checkPeerConnection(ctx, peer.Address) {
		status = "healthy"
		details["address"] = peer.Address
		details["mspId"] = peer.MSPID
		details["channels"] = peer.Channels
	} else {
		details["address"] = peer.Address
		details["error"] = "Connection timeout or refused"
	}

	// Also check if we can get network from gateway (for primary peer)
	if peer.Address == s.config.PeerEndpoint {
		network := s.gateway.GetNetwork(s.config.Channel)
		if network != nil {
			details["gateway_connected"] = true
		} else {
			details["gateway_connected"] = false
		}
	}

	return &models.HealthStatus{
		Component: "peer",
		ID:        peerID,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Details:   details,
	}, nil
}

// CheckOrdererHealth checks orderer health status
func (s *DiscoveryService) CheckOrdererHealth(ctx context.Context, ordererID string) (*models.HealthStatus, error) {
	s.logger.Info("Checking orderer health", zap.String("ordererId", ordererID))

	orderer, err := s.GetOrderer(ctx, ordererID)
	if err != nil {
		return nil, err
	}

	// Try to ping orderer gRPC endpoint
	status := "unhealthy"
	details := make(map[string]interface{})

	if s.checkOrdererConnection(ctx, orderer.Address) {
		status = "healthy"
		details["address"] = orderer.Address
		details["mspId"] = orderer.MSPID
		details["isLeader"] = orderer.IsLeader
	} else {
		details["address"] = orderer.Address
		details["error"] = "Connection timeout or refused"
	}

	return &models.HealthStatus{
		Component: "orderer",
		ID:        ordererID,
		Status:    status,
		Timestamp: time.Now().Format(time.RFC3339),
		Details:   details,
	}, nil
}

// CheckAllPeersHealth checks health of all peers
func (s *DiscoveryService) CheckAllPeersHealth(ctx context.Context) ([]*models.HealthStatus, error) {
	s.logger.Info("Checking all peers health")

	peers, err := s.ListPeers(ctx)
	if err != nil {
		return nil, err
	}

	healthStatuses := []*models.HealthStatus{}
	for _, peer := range peers {
		health, err := s.CheckPeerHealth(ctx, peer.Name)
		if err != nil {
			s.logger.Warn("Failed to check peer health", zap.String("peer", peer.Name), zap.Error(err))
			continue
		}
		healthStatuses = append(healthStatuses, health)
	}

	return healthStatuses, nil
}

// CheckAllOrderersHealth checks health of all orderers
func (s *DiscoveryService) CheckAllOrderersHealth(ctx context.Context) ([]*models.HealthStatus, error) {
	s.logger.Info("Checking all orderers health")

	orderers, err := s.ListOrderers(ctx)
	if err != nil {
		return nil, err
	}

	healthStatuses := []*models.HealthStatus{}
	for _, orderer := range orderers {
		health, err := s.CheckOrdererHealth(ctx, orderer.Name)
		if err != nil {
			s.logger.Warn("Failed to check orderer health", zap.String("orderer", orderer.Name), zap.Error(err))
			continue
		}
		healthStatuses = append(healthStatuses, health)
	}

	return healthStatuses, nil
}
