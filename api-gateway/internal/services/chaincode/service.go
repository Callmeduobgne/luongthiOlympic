package chaincode

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ibn-network/api-gateway/internal/config"
	"go.uber.org/zap"
)

// Service manages chaincode lifecycle operations
// Note: Fabric Gateway SDK doesn't support lifecycle operations
// This service provides wrappers for peer CLI commands
type Service struct {
	config     *config.FabricConfig
	logger     *zap.Logger
	peerPath   string
	orgMSPPath string
}

// NewService creates a new chaincode service
func NewService(cfg *config.FabricConfig, logger *zap.Logger) (*Service, error) {
	// Default peer CLI path (should be in PATH or configured)
	peerPath := "peer"
	
	// Extract MSP path from user cert path
	// e.g., /app/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
	userCertPath := cfg.UserCertPath
	orgMSPPath := ""
	if userCertPath != "" {
		// Go up to organizations/peerOrganizations/org1.ibn.vn/msp
		parts := strings.Split(userCertPath, "/")
		for i, part := range parts {
			if part == "users" {
				orgMSPPath = filepath.Join(parts[:i]...)
				orgMSPPath = filepath.Join("/", orgMSPPath, "msp")
				break
			}
		}
	}

	return &Service{
		config:     cfg,
		logger:     logger,
		peerPath:   peerPath,
		orgMSPPath: orgMSPPath,
	}, nil
}

// ListInstalled, ListCommitted, GetCommittedInfo, Install, Approve, Commit
// are now implemented in lifecycle.go

// executePeerCommand executes a peer CLI command
func (s *Service) executePeerCommand(ctx context.Context, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, s.peerPath, args...)
	cmd.Env = s.getPeerEnv()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("peer command failed: %w, output: %s", err, string(output))
	}
	
	return string(output), nil
}

