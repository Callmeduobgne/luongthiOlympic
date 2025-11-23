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
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/ibn-network/admin-service/internal/config"
	"go.uber.org/zap"
)

// Service manages chaincode lifecycle operations using peer CLI
type Service struct {
	config     *config.FabricConfig
	logger     *zap.Logger
	peerPath   string
	orgMSPPath string
}

// NewService creates a new chaincode service
func NewService(cfg *config.FabricConfig, logger *zap.Logger) (*Service, error) {
	// Peer CLI path (should be in PATH after Dockerfile installation)
	peerPath := "peer"

	// Extract MSP path from user cert path
	// e.g., /app/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
	userCertPath := cfg.UserCertPath
	orgMSPPath := ""
	if userCertPath != "" {
		// Example: /app/organizations/.../users/Admin@org1.ibn.vn/msp/signcerts/cert.pem
		// We need the user's MSP directory (.../users/Admin@org1.ibn.vn/msp)
		mspDir := filepath.Dir(filepath.Dir(userCertPath))
		orgMSPPath = mspDir
		
		// Log MSP path for debugging
		logger.Info("Extracted MSP path from user cert",
			zap.String("userCertPath", userCertPath),
			zap.String("orgMSPPath", orgMSPPath),
		)
	} else {
		logger.Error("User cert path is empty! Cannot extract MSP path")
		return nil, fmt.Errorf("user cert path is required to extract MSP path")
	}

	return &Service{
		config:     cfg,
		logger:     logger,
		peerPath:   peerPath,
		orgMSPPath: orgMSPPath,
	}, nil
}

// executePeerCommand executes a peer CLI command
func (s *Service) executePeerCommand(ctx context.Context, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, s.peerPath, args...)
	cmd.Env = s.getPeerEnv()
	// Set working directory to /app (where fabric-config is located)
	cmd.Dir = "/app"

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("peer command failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}
