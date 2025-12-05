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
	"os"
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
	env := os.Environ()
	
	// Build environment variables for peer CLI
	peerEnv := []string{
		fmt.Sprintf("CORE_PEER_LOCALMSPID=%s", s.config.MSPId),
		fmt.Sprintf("CORE_PEER_MSPCONFIGPATH=%s", s.orgMSPPath),
		fmt.Sprintf("CORE_PEER_ADDRESS=%s", s.config.PeerEndpoint),
		"CORE_PEER_TLS_ENABLED=true",
		fmt.Sprintf("CORE_PEER_TLS_ROOTCERT_FILE=%s", s.config.PeerTLSCAPath),
	}
	
	// Set FABRIC_CFG_PATH if not already set
	hasFabricCfgPath := false
	for _, e := range env {
		if strings.HasPrefix(e, "FABRIC_CFG_PATH=") {
			hasFabricCfgPath = true
			break
		}
	}
	if !hasFabricCfgPath {
		peerEnv = append(peerEnv, "FABRIC_CFG_PATH=/app/fabric-config")
	}
	
	// Validate MSP path is set
	if s.orgMSPPath == "" {
		s.logger.Error("MSP path is empty! Cannot execute peer command",
			zap.String("userCertPath", s.config.UserCertPath),
		)
	}
	
	return append(env, peerEnv...)
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
	
	// Check if chaincode is already installed (idempotent operation)
	// Peer CLI may return error or success, but output contains "already" message
	cleanOutput := stripANSI(output)
	isAlreadyInstalled := strings.Contains(strings.ToLower(cleanOutput), "already") && 
		(strings.Contains(strings.ToLower(cleanOutput), "installed") || 
		 strings.Contains(strings.ToLower(cleanOutput), "successfully"))
	
	// If command failed but output suggests already installed, try to extract package ID
	if err != nil && !isAlreadyInstalled {
		return "", fmt.Errorf("failed to install chaincode: %w", err)
	}

	// Parse package ID from output (works for both new install and already installed)
	// Output format can be:
	// 1. "Installed chaincode with package ID 'basic_1.0:abc123...'"
	// 2. "Installed remotely: response: <status:200 payload:"\nOteaTraceCC_1.0:hash..." (with ANSI codes)
	// 3. "chaincode already successfully installed (package ID 'teaTraceCC_1.0:hash...')"
	packageID := s.extractPackageID(output)
	if packageID == "" {
		s.logger.Warn("Failed to extract package ID using primary method, trying alternative methods",
			zap.String("output", output),
		)
		// Try alternative extraction methods
		packageID = s.extractPackageIDAlternative(output)
	}

	if packageID == "" {
		// If already installed but can't extract package ID, this is still an error
		if isAlreadyInstalled {
			s.logger.Warn("Chaincode appears to be already installed but package ID extraction failed",
				zap.String("output", output),
			)
			return "", fmt.Errorf("chaincode already installed but failed to extract package ID. Please check admin-service logs")
		}
		// Log full output for debugging
		s.logger.Error("Failed to extract package ID from output",
			zap.String("output", output),
			zap.String("output_length", fmt.Sprintf("%d", len(output))),
		)
		return "", fmt.Errorf("failed to extract package ID from output. Please check admin-service logs for full output")
	}

	// If already installed, log as info (not error) - this is a valid idempotent operation
	if isAlreadyInstalled {
		s.logger.Info("Chaincode already installed (idempotent operation)",
			zap.String("packageId", packageID),
			zap.String("package", packagePath),
		)
	} else {
		s.logger.Info("Chaincode installed successfully",
			zap.String("packageId", packageID),
		)
	}

	return packageID, nil
}

// stripANSI removes ANSI color codes from string
func stripANSI(s string) string {
	// Remove ANSI escape sequences: \x1b[ or \033[ followed by numbers and letters
	// Pattern: \x1b\[[0-9;]*[a-zA-Z] or \033\[[0-9;]*[a-zA-Z]
	var result strings.Builder
	inEscape := false
	escapeSeq := false
	
	for _, r := range s {
		if !inEscape && r == '\x1b' || r == '\033' {
			inEscape = true
			escapeSeq = false
			continue
		}
		if inEscape {
			if r == '[' {
				escapeSeq = true
				continue
			}
			if escapeSeq {
				// Check if this is the end of escape sequence (letter)
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					inEscape = false
					escapeSeq = false
					continue
				}
				// Still in escape sequence (numbers, semicolons)
				continue
			}
			// Not an escape sequence, reset
			inEscape = false
		}
		result.WriteRune(r)
	}
	return result.String()
}

// extractPackageID extracts package ID from peer CLI output
// Supports multiple output formats:
// 1. "package ID 'name_version:hash'"
// 2. "payload:\"\\nOname_version:hash" (with ANSI codes)
// 3. "code package identifier: name_version:hash"
// 4. "already successfully installed (package ID 'name_version:hash')"
func (s *Service) extractPackageID(output string) string {
	// First, strip ANSI codes
	cleanOutput := stripANSI(output)
	
	// Method 1: Look for pattern: package ID 'xxx:yyy' or (package ID 'xxx:yyy')
	lines := strings.Split(cleanOutput, "\n")
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "package id") || strings.Contains(lowerLine, "package identifier") {
			// Try to find text between single quotes (most common format)
			// Look for pattern: package ID 'xxx:yyy' or (package ID 'xxx:yyy')
			start := strings.Index(line, "'")
			if start != -1 {
				end := strings.Index(line[start+1:], "'")
				if end != -1 {
					packageID := line[start+1 : start+1+end]
					if s.isValidPackageID(packageID) {
						return packageID
					}
				}
			}
			
			// Try to find text between double quotes
			start = strings.Index(line, "\"")
			if start != -1 {
				end := strings.Index(line[start+1:], "\"")
				if end != -1 {
					packageID := line[start+1 : start+1+end]
					// Remove escape sequences like \nO
					packageID = strings.ReplaceAll(packageID, "\\nO", "")
					packageID = strings.ReplaceAll(packageID, "\\n", "")
					packageID = strings.TrimSpace(packageID)
					if s.isValidPackageID(packageID) {
						return packageID
					}
				}
			}
		}
	}
	
	// Method 2: Look for pattern in parentheses: (package ID 'xxx:yyy')
	// This handles "already successfully installed (package ID 'xxx:yyy')" format
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "already") && strings.Contains(strings.ToLower(line), "installed") {
			// Find text in parentheses
			start := strings.Index(line, "(")
			if start != -1 {
				end := strings.Index(line[start:], ")")
				if end != -1 {
					parenContent := line[start+1 : start+end]
					// Look for package ID in parentheses
					quoteStart := strings.Index(parenContent, "'")
					if quoteStart != -1 {
						quoteEnd := strings.Index(parenContent[quoteStart+1:], "'")
						if quoteEnd != -1 {
							packageID := parenContent[quoteStart+1 : quoteStart+1+quoteEnd]
							if s.isValidPackageID(packageID) {
								return packageID
							}
						}
					}
				}
			}
		}
	}
	
	return ""
}

// extractPackageIDAlternative tries alternative methods to extract package ID
func (s *Service) extractPackageIDAlternative(output string) string {
	// Strip ANSI codes
	cleanOutput := stripANSI(output)
	
	// Method 1: Look for pattern in payload: payload:"\nOname_version:hash
	// Pattern: payload:" followed by optional escape sequences and then name_version:hash
	if idx := strings.Index(cleanOutput, "payload:"); idx != -1 {
		payloadPart := cleanOutput[idx:]
		// Find pattern: name_version:hash (e.g., teaTraceCC_1.0:579be867...)
		// Use regex-like approach: find colon followed by hex string
		parts := strings.Fields(payloadPart)
		for _, part := range parts {
			// Look for pattern containing colon and long hex string
			if strings.Contains(part, ":") {
				// Try to extract package ID pattern: name_version:hash
				colonIdx := strings.LastIndex(part, ":")
				if colonIdx > 0 && colonIdx < len(part)-1 {
					// Check if after colon is a hex string (at least 32 chars for hash)
					afterColon := part[colonIdx+1:]
					// Remove any trailing non-hex characters
					hashPart := ""
					for _, r := range afterColon {
						if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
							hashPart += string(r)
						} else {
							break
						}
					}
					if len(hashPart) >= 32 {
						// Found potential package ID
						beforeColon := part[:colonIdx]
						// Remove escape sequences and quotes
						beforeColon = strings.Trim(beforeColon, "\"\\nO")
						packageID := beforeColon + ":" + hashPart
						if s.isValidPackageID(packageID) {
							return packageID
						}
					}
				}
			}
		}
	}
	
	// Method 2: Look for pattern: name_version:hash directly (regex-like)
	// Pattern: alphanumeric_underscore:alphanumeric (at least 32 chars after colon)
	words := strings.Fields(cleanOutput)
	for _, word := range words {
		if strings.Contains(word, ":") {
			parts := strings.Split(word, ":")
			if len(parts) == 2 {
				namePart := strings.Trim(parts[0], "\"\\nO\t\r")
				hashPart := strings.Trim(parts[1], "\"\\nO\t\r")
				// Remove trailing non-hex from hash
				cleanHash := ""
				for _, r := range hashPart {
					if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
						cleanHash += string(r)
					} else {
						break
					}
				}
				if len(cleanHash) >= 32 && len(namePart) > 0 {
					packageID := namePart + ":" + cleanHash
					if s.isValidPackageID(packageID) {
						return packageID
					}
				}
			}
		}
	}
	
	return ""
}

// isValidPackageID validates if a string looks like a valid package ID
// Format: name_version:hash (hash should be at least 32 hex characters)
func (s *Service) isValidPackageID(packageID string) bool {
	if packageID == "" {
		return false
	}
	
	// Must contain colon
	if !strings.Contains(packageID, ":") {
		return false
	}
	
	parts := strings.Split(packageID, ":")
	if len(parts) != 2 {
		return false
	}
	
	namePart := parts[0]
	hashPart := parts[1]
	
	// Name part should not be empty
	if len(namePart) == 0 {
		return false
	}
	
	// Hash part should be at least 32 hex characters
	if len(hashPart) < 32 {
		return false
	}
	
	// Hash should be hex
	for _, r := range hashPart {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	
	return true
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
		"--cafile", s.config.OrdererTLSCAPath,
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
		args = append(args, "--endorsement-plugin", req.EndorsementPlugin)
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
		"--cafile", s.config.OrdererTLSCAPath,
	}

	// Add init required flag
	if req.InitRequired {
		args = append(args, "--init-required")
	}

	// Add endorsement plugin if provided
	if req.EndorsementPlugin != "" {
		args = append(args, "--endorsement-plugin", req.EndorsementPlugin)
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
	s.logger.Info("Listing installed chaincodes",
		zap.String("peer", peerAddress),
		zap.String("mspPath", s.orgMSPPath),
		zap.String("mspId", s.config.MSPId),
	)

	// Validate MSP path is set
	if s.orgMSPPath == "" {
		s.logger.Error("MSP path is empty! Cannot query installed chaincodes",
			zap.String("userCertPath", s.config.UserCertPath),
		)
		return nil, fmt.Errorf("MSP path is empty. User cert path: %s", s.config.UserCertPath)
	}

	// Check if peer CLI is available
	if err := s.checkPeerCLIAvailable(ctx); err != nil {
		s.logger.Error("Peer CLI not available",
			zap.Error(err),
			zap.String("note", "Peer CLI is required to query installed chaincodes. Please ensure peer CLI is installed and available in PATH."),
		)
		// Return error so frontend knows there's an issue
		return nil, fmt.Errorf("peer CLI not available: %w. Please check admin-service logs", err)
	}

	// Build command: peer lifecycle chaincode queryinstalled --output json
	args := []string{
		"lifecycle", "chaincode", "queryinstalled",
		"--output", "json",
	}

	if peerAddress != "" {
		args = append(args, "--peerAddresses", peerAddress)
		args = append(args, "--tlsRootCertFiles", s.config.PeerTLSCAPath)
	} else {
		// If no peer address specified, use default peer from config
		// Add TLS config for default peer
		args = append(args, "--peerAddresses", s.config.PeerEndpoint)
		args = append(args, "--tlsRootCertFiles", s.config.PeerTLSCAPath)
	}

	// Execute command
	output, err := s.executePeerCommand(ctx, args)
	if err != nil {
		s.logger.Error("Failed to query installed chaincodes",
			zap.Error(err),
			zap.String("output", output),
			zap.Strings("command", args),
		)
		// Return error with output for debugging
		return nil, fmt.Errorf("failed to query installed chaincodes: %w. Output: %s", err, output)
	}

	// Log raw output for debugging
	s.logger.Debug("Peer CLI queryinstalled output",
		zap.String("output", output),
		zap.Int("length", len(output)),
	)

	// Parse JSON output
	var result struct {
		InstalledChaincodes []struct {
			PackageID string `json:"package_id"`
			Label     string `json:"label"`
		} `json:"installed_chaincodes"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// If JSON parsing fails, log error and return it
		s.logger.Error("Failed to parse queryinstalled output as JSON",
			zap.Error(err),
			zap.String("output", output),
			zap.String("output_preview", func() string {
				if len(output) > 500 {
					return output[:500] + "..."
				}
				return output
			}()),
		)
		return nil, fmt.Errorf("failed to parse queryinstalled JSON output: %w. Raw output: %s", err, output)
	}

	s.logger.Info("Successfully parsed installed chaincodes",
		zap.Int("count", len(result.InstalledChaincodes)),
	)

	// Convert to InstalledChaincode slice
	chaincodes := make([]*InstalledChaincode, 0, len(result.InstalledChaincodes))
	for _, cc := range result.InstalledChaincodes {
		if cc.PackageID == "" {
			s.logger.Warn("Skipping chaincode with empty package ID",
				zap.String("label", cc.Label),
			)
			continue
		}
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

	s.logger.Info("Returning installed chaincodes",
		zap.Int("count", len(chaincodes)),
	)

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
			Name                  string   `json:"name"`
			Version               string   `json:"version"`
			Sequence              int64    `json:"sequence"`
			EndorsementPlugin     string   `json:"endorsement_plugin"`
			ValidationPlugin      string   `json:"validation_plugin"`
			InitRequired          bool     `json:"init_required"`
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
			Name:                  cc.Name,
			Version:               cc.Version,
			Sequence:              cc.Sequence,
			EndorsementPlugin:     cc.EndorsementPlugin,
			ValidationPlugin:      cc.ValidationPlugin,
			InitRequired:          cc.InitRequired,
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
