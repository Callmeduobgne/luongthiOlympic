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

package identity

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ibn-network/api-gateway/internal/config"
	"go.uber.org/zap"
)

// Service manages Fabric identities
// Currently manages identities from file system (cryptogen)
// Can be extended to use Fabric CA when CA server is available
type Service struct {
	config *config.CAConfig
	logger *zap.Logger
	mspDir string
}

// NewService creates a new identity service
func NewService(cfg *config.CAConfig, logger *zap.Logger) (*Service, error) {
	// Use MSP directory from config or default
	mspDir := cfg.MSPDir
	if mspDir == "" {
		// Default to peer organizations MSP
		mspDir = "/app/organizations/peerOrganizations/org1.ibn.vn/msp"
	}

	// If MSPDir is still empty or invalid, return error
	if mspDir == "" {
		return nil, fmt.Errorf("MSP directory not configured")
	}

	return &Service{
		config: cfg,
		logger: logger,
		mspDir: mspDir,
	}, nil
}

// ListUsers lists all users in the MSP
func (s *Service) ListUsers(ctx context.Context, affiliation string) ([]*UserInfo, error) {
	s.logger.Info("Listing users", zap.String("affiliation", affiliation))

	usersDir := filepath.Join(s.mspDir, "users")
	if _, err := os.Stat(usersDir); os.IsNotExist(err) {
		return []*UserInfo{}, nil
	}

	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read users directory: %w", err)
	}

	users := make([]*UserInfo, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		username := entry.Name()
		userDir := filepath.Join(usersDir, username)

		// Check if user has valid certificate
		certPath := filepath.Join(userDir, "msp", "signcerts", fmt.Sprintf("%s-cert.pem", username))
		if _, err := os.Stat(certPath); os.IsNotExist(err) {
			continue
		}

		// Read certificate to get info
		cert, err := s.readCertificate(certPath)
		if err != nil {
			s.logger.Warn("Failed to read certificate", zap.String("user", username), zap.Error(err))
			continue
		}

		// Extract attributes from certificate
		attrs := s.extractAttributes(cert)

		users = append(users, &UserInfo{
			Username:    username,
			Type:        "client", // Default, can be determined from cert
			Affiliation: s.extractAffiliation(username),
			Attributes:  attrs,
			Revoked:     false, // Would need to check revocation list
		})
	}

	return users, nil
}

// GetUser gets user information
func (s *Service) GetUser(ctx context.Context, username string) (*UserInfo, error) {
	s.logger.Info("Getting user", zap.String("username", username))

	userDir := filepath.Join(s.mspDir, "users", username)
	certPath := filepath.Join(userDir, "msp", "signcerts", fmt.Sprintf("%s-cert.pem", username))

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("user %s not found", username)
	}

	cert, err := s.readCertificate(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	attrs := s.extractAttributes(cert)

	return &UserInfo{
		Username:    username,
		Type:        "client",
		Affiliation: s.extractAffiliation(username),
		Attributes:  attrs,
		Revoked:     false,
	}, nil
}

// readCertificate reads and parses X.509 certificate
func (s *Service) readCertificate(certPath string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// extractAttributes extracts attributes from certificate
func (s *Service) extractAttributes(cert *x509.Certificate) []string {
	attrs := make([]string, 0)

	// Extract from Subject
	if cert.Subject.OrganizationalUnit != nil {
		for _, ou := range cert.Subject.OrganizationalUnit {
			attrs = append(attrs, fmt.Sprintf("ou=%s", ou))
		}
	}

	// Extract from Subject Alternative Name
	for _, ext := range cert.Extensions {
		if ext.Id.String() == "2.5.29.17" { // SAN extension
			// Parse SAN if needed
		}
	}

	return attrs
}

// extractAffiliation extracts affiliation from username
func (s *Service) extractAffiliation(username string) string {
	// Username format: User1@org1.ibn.vn -> org1.department1
	parts := strings.Split(username, "@")
	if len(parts) > 1 {
		org := strings.Split(parts[1], ".")[0]
		return fmt.Sprintf("%s.department1", org)
	}
	return "org1.department1"
}

