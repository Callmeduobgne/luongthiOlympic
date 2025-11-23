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
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// checkPeerCLIAvailable checks if peer CLI is available
func (s *Service) checkPeerCLIAvailable(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, s.peerPath, "version")
	cmd.Env = s.getPeerEnv()
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("peer CLI not available: %w (peer command not found in PATH)", err)
	}
	return nil
}

// getPeerEnv returns environment variables for peer CLI
func (s *Service) getPeerEnv() []string {
	return []string{
		fmt.Sprintf("CORE_PEER_LOCALMSPID=%s", s.config.MSPId),
		fmt.Sprintf("CORE_PEER_MSPCONFIGPATH=%s", s.orgMSPPath),
		fmt.Sprintf("CORE_PEER_ADDRESS=%s", s.config.PeerEndpoint),
		fmt.Sprintf("CORE_PEER_TLS_ENABLED=true"),
		fmt.Sprintf("CORE_PEER_TLS_ROOTCERT_FILE=%s", s.config.PeerTLSCAPath),
	}
}

// Install installs a chaincode package using peer CLI
func (s *Service) Install(ctx context.Context, packagePath, label string) (string, error) {
	s.logger.Info("Installing chaincode",
		zap.String("package", packagePath),
		zap.String("label", label),
	)

	// Check if peer CLI is available
	if err := s.checkPeerCLIAvailable(ctx); err != nil {
		return "", fmt.Errorf("peer CLI not available: %w", err)
	}

	// Execute: peer lifecycle chaincode install <package>
	args := []string{
		"lifecycle", "chaincode", "install",
		packagePath,
	}

	output, err := s.executePeerCommand(ctx, args)
	if err != nil {
		return "", fmt.Errorf("failed to install chaincode: %w", err)
	}

	// Parse package ID from output
	// Output format: "2024-01-01 12:00:00.000 UTC [chaincodeCmd] installChaincode -> INFO 001 Installed chaincode with package ID 'basic_1.0:abc123...'"
	packageID := s.extractPackageID(output)
	if packageID == "" {
		// Try to extract from different output format
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "package ID") {
				parts := strings.Split(line, "'")
				if len(parts) >= 2 {
					packageID = parts[1]
					break
				}
			}
		}
	}

	if packageID == "" {
		return "", fmt.Errorf("failed to extract package ID from output: %s", output)
	}

	s.logger.Info("Chaincode installed successfully",
		zap.String("packageId", packageID),
	)

	return packageID, nil
}

// extractPackageID extracts package ID from peer CLI output
func (s *Service) extractPackageID(output string) string {
	// Look for pattern: package ID 'xxx:yyy'
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "package ID") {
			// Try to find text between quotes
			start := strings.Index(line, "'")
			if start != -1 {
				end := strings.Index(line[start+1:], "'")
				if end != -1 {
					return line[start+1 : start+1+end]
				}
			}
		}
	}
	return ""
}

// Approve approves a chaincode definition using peer CLI
func (s *Service) Approve(ctx context.Context, req *ApproveChaincodeRequest) error {
	s.logger.Info("Approving chaincode",
		zap.String("channel", req.ChannelName),
		zap.String("name", req.Name),
		zap.String("version", req.Version),
		zap.Int64("sequence", req.Sequence),
	)

	// Check if peer CLI is available
	if err := s.checkPeerCLIAvailable(ctx); err != nil {
		return fmt.Errorf("peer CLI not available: %w", err)
	}

	// Use default channel if not provided
	channelName := req.ChannelName
	if channelName == "" {
		channelName = s.config.Channel
	}

	// Build command: peer lifecycle chaincode approveformyorg
	args := []string{
		"lifecycle", "chaincode", "approveformyorg",
		"--channelID", channelName,
		"--name", req.Name,
		"--version", req.Version,
		"--sequence", strconv.FormatInt(req.Sequence, 10),
		"--tls",
		"--cafile", s.config.PeerTLSCAPath,
	}

	// Add package ID if provided
	if req.PackageID != "" {
		args = append(args, "--package-id", req.PackageID)
	}

	// Add init required flag
	if req.InitRequired {
		args = append(args, "--init-required")
	}

	// Add endorsement plugin if provided
	if req.EndorsementPlugin != "" {
		args = append(args, "--signature-policy", req.EndorsementPlugin)
	}

	// Add validation plugin if provided
	if req.ValidationPlugin != "" {
		args = append(args, "--validation-plugin", req.ValidationPlugin)
	}

	// Execute command
	_, err := s.executePeerCommand(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to approve chaincode: %w", err)
	}

	s.logger.Info("Chaincode approved successfully",
		zap.String("name", req.Name),
		zap.String("version", req.Version),
	)

	return nil
}

// Commit commits a chaincode definition using peer CLI
func (s *Service) Commit(ctx context.Context, req *CommitChaincodeRequest) error {
	s.logger.Info("Committing chaincode",
		zap.String("channel", req.ChannelName),
		zap.String("name", req.Name),
		zap.String("version", req.Version),
		zap.Int64("sequence", req.Sequence),
	)

	// Check if peer CLI is available
	if err := s.checkPeerCLIAvailable(ctx); err != nil {
		return fmt.Errorf("peer CLI not available: %w", err)
	}

	// Use default channel if not provided
	channelName := req.ChannelName
	if channelName == "" {
		channelName = s.config.Channel
	}

	// Build command: peer lifecycle chaincode commit
	args := []string{
		"lifecycle", "chaincode", "commit",
		"--channelID", channelName,
		"--name", req.Name,
		"--version", req.Version,
		"--sequence", strconv.FormatInt(req.Sequence, 10),
		"--tls",
		"--cafile", s.config.PeerTLSCAPath,
	}

	// Add init required flag
	if req.InitRequired {
		args = append(args, "--init-required")
	}

	// Add endorsement plugin if provided
	if req.EndorsementPlugin != "" {
		args = append(args, "--signature-policy", req.EndorsementPlugin)
	}

	// Add validation plugin if provided
	if req.ValidationPlugin != "" {
		args = append(args, "--validation-plugin", req.ValidationPlugin)
	}

	// Execute command
	_, err := s.executePeerCommand(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to commit chaincode: %w", err)
	}

	s.logger.Info("Chaincode committed successfully",
		zap.String("name", req.Name),
		zap.String("version", req.Version),
	)

	return nil
}

// ListInstalled lists installed chaincodes using peer CLI
func (s *Service) ListInstalled(ctx context.Context, peerAddress string) ([]*InstalledChaincode, error) {
	s.logger.Info("Listing installed chaincodes", zap.String("peer", peerAddress))

	// Check if peer CLI is available
	if err := s.checkPeerCLIAvailable(ctx); err != nil {
		s.logger.Warn("Peer CLI not available, returning empty list",
			zap.Error(err),
			zap.String("note", "Peer CLI is required to query installed chaincodes. Please ensure peer CLI is installed and available in PATH."),
		)
		// Return empty list instead of error to allow UI to display empty state
		return []*InstalledChaincode{}, nil
	}

	// Build command: peer lifecycle chaincode queryinstalled --output json
	args := []string{
		"lifecycle", "chaincode", "queryinstalled",
		"--output", "json",
	}

	if peerAddress != "" {
		args = append(args, "--peerAddresses", peerAddress)
		args = append(args, "--tlsRootCertFiles", s.config.PeerTLSCAPath)
	}

	// Execute command
	output, err := s.executePeerCommand(ctx, args)
	if err != nil {
		s.logger.Warn("Failed to query installed chaincodes, returning empty list",
			zap.Error(err),
			zap.String("output", output),
		)
		// Return empty list instead of error to allow UI to display empty state
		return []*InstalledChaincode{}, nil
	}

	// Parse JSON output
	var result struct {
		InstalledChaincodes []struct {
			PackageID string `json:"package_id"`
			Label     string `json:"label"`
		} `json:"installed_chaincodes"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// If JSON parsing fails, return empty list with warning
		s.logger.Warn("Failed to parse queryinstalled output as JSON",
			zap.Error(err),
			zap.String("output", output),
		)
		return []*InstalledChaincode{}, nil
	}

	// Convert to InstalledChaincode slice
	chaincodes := make([]*InstalledChaincode, 0, len(result.InstalledChaincodes))
	for _, cc := range result.InstalledChaincodes {
		chaincodes = append(chaincodes, &InstalledChaincode{
			PackageID: cc.PackageID,
			Label:     cc.Label,
			Chaincode: ChaincodeInfo{
				Name:    cc.Label, // Label often contains name
				Version: "",       // Version not in queryinstalled output
				Path:    "",
			},
		})
	}

	return chaincodes, nil
}

// ListCommitted lists committed chaincodes using peer CLI
func (s *Service) ListCommitted(ctx context.Context, channelName string) ([]*CommittedChaincode, error) {
	s.logger.Info("Listing committed chaincodes", zap.String("channel", channelName))

	// Check if peer CLI is available
	if err := s.checkPeerCLIAvailable(ctx); err != nil {
		s.logger.Warn("Peer CLI not available, returning empty list",
			zap.Error(err),
			zap.String("channel", channelName),
			zap.String("note", "Peer CLI is required to query committed chaincodes. Please ensure peer CLI is installed and available in PATH."),
		)
		// Return empty list instead of error to allow UI to display empty state
		return []*CommittedChaincode{}, nil
	}

	// Use default channel if not provided
	if channelName == "" {
		channelName = s.config.Channel
	}

	// Build command: peer lifecycle chaincode querycommitted --channelID <channel> --output json
	args := []string{
		"lifecycle", "chaincode", "querycommitted",
		"--channelID", channelName,
		"--output", "json",
		"--peerAddresses", s.config.PeerEndpoint,
		"--tlsRootCertFiles", s.config.PeerTLSCAPath,
	}

	// Execute command
	output, err := s.executePeerCommand(ctx, args)
	if err != nil {
		s.logger.Warn("Failed to query committed chaincodes, returning empty list",
			zap.Error(err),
			zap.String("channel", channelName),
			zap.String("output", output),
		)
		// Return empty list instead of error to allow UI to display empty state
		return []*CommittedChaincode{}, nil
	}

	// Parse JSON output
	var result struct {
		ChaincodeDefinitions []struct {
			Name                string   `json:"name"`
			Version             string   `json:"version"`
			Sequence            int64    `json:"sequence"`
			EndorsementPlugin   string   `json:"endorsement_plugin"`
			ValidationPlugin    string   `json:"validation_plugin"`
			InitRequired        bool     `json:"init_required"`
			ApprovedOrganizations []string `json:"approved_organizations"`
		} `json:"chaincode_definitions"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// If JSON parsing fails, return empty list with warning
		s.logger.Warn("Failed to parse querycommitted output as JSON",
			zap.Error(err),
			zap.String("output", output),
		)
		return []*CommittedChaincode{}, nil
	}

	// Convert to CommittedChaincode slice
	chaincodes := make([]*CommittedChaincode, 0, len(result.ChaincodeDefinitions))
	for _, cc := range result.ChaincodeDefinitions {
		chaincodes = append(chaincodes, &CommittedChaincode{
			Name:                cc.Name,
			Version:             cc.Version,
			Sequence:            cc.Sequence,
			EndorsementPlugin:   cc.EndorsementPlugin,
			ValidationPlugin:    cc.ValidationPlugin,
			InitRequired:        cc.InitRequired,
			ApprovedOrganizations: cc.ApprovedOrganizations,
		})
	}

	return chaincodes, nil
}

// GetCommittedInfo gets information about a committed chaincode
func (s *Service) GetCommittedInfo(ctx context.Context, channelName, chaincodeName string) (*CommittedChaincode, error) {
	s.logger.Info("Getting committed chaincode info",
		zap.String("channel", channelName),
		zap.String("chaincode", chaincodeName),
	)

	// List all committed chaincodes
	chaincodes, err := s.ListCommitted(ctx, channelName)
	if err != nil {
		return nil, err
	}

	// Find the specific chaincode
	for _, cc := range chaincodes {
		if cc.Name == chaincodeName {
			return cc, nil
		}
	}

	return nil, fmt.Errorf("chaincode '%s' not found on channel '%s'", chaincodeName, channelName)
}

// Upgrade upgrades a chaincode (approve + commit with new version/sequence)
func (s *Service) Upgrade(ctx context.Context, channelName, name, version string, sequence int64) error {
	s.logger.Info("Upgrading chaincode",
		zap.String("channel", channelName),
		zap.String("name", name),
		zap.String("version", version),
		zap.Int64("sequence", sequence),
	)

	// Upgrade is essentially approve + commit with new version
	// First approve
	approveReq := &ApproveChaincodeRequest{
		ChannelName: channelName,
		Name:        name,
		Version:     version,
		Sequence:    sequence,
	}

	if err := s.Approve(ctx, approveReq); err != nil {
		return fmt.Errorf("failed to approve chaincode for upgrade: %w", err)
	}

	// Then commit
	commitReq := &CommitChaincodeRequest{
		ChannelName: channelName,
		Name:        name,
		Version:     version,
		Sequence:    sequence,
	}

	if err := s.Commit(ctx, commitReq); err != nil {
		return fmt.Errorf("failed to commit chaincode for upgrade: %w", err)
	}

	s.logger.Info("Chaincode upgraded successfully",
		zap.String("name", name),
		zap.String("version", version),
	)

	return nil
}

