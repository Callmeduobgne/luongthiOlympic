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

package ca

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ibn-network/api-gateway/internal/config"
	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// Service provides Fabric CA operations
type Service struct {
	client *Client
	config *config.CAConfig
	logger *zap.Logger
	mspDir string
}

// NewService creates a new CA service
func NewService(cfg *config.CAConfig, logger *zap.Logger) (*Service, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("CA URL is not configured")
	}

	client, err := NewClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA client: %w", err)
	}

	mspDir := cfg.MSPDir
	if mspDir == "" {
		mspDir = "/app/organizations/peerOrganizations/org1.ibn.vn/msp"
	}

	return &Service{
		client: client,
		config: cfg,
		logger: logger,
		mspDir: mspDir,
	}, nil
}

// Enroll enrolls a user with Fabric CA
func (s *Service) Enroll(ctx context.Context, req *models.CAEnrollRequest) (*models.CAEnrollResponse, error) {
	s.logger.Info("Enrolling user", zap.String("username", req.Username))

	enrollReq := &EnrollRequest{
		Name:   req.Username,
		Secret: req.Password,
	}

	enrollResp, err := s.client.Enroll(enrollReq)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll user: %w", err)
	}

	// Parse certificate
	certPEM := enrollResp.Result.Cert
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate")
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Extract MSP ID from certificate or use config
	mspID := s.config.MSPID
	if mspID == "" {
		mspID = "Org1MSP" // Default
	}

	// Store certificate in wallet (optional)
	if err := s.storeCertificate(req.Username, certPEM, ""); err != nil {
		s.logger.Warn("Failed to store certificate", zap.Error(err))
		// Continue anyway
	}

	return &models.CAEnrollResponse{
		Username:    req.Username,
		Certificate: certPEM,
		PrivateKey:  "", // Private key is not returned by CA API for security
		MSPID:       mspID,
	}, nil
}

// Register registers a new identity with Fabric CA
func (s *Service) Register(ctx context.Context, req *models.CARegisterRequest) (*models.CARegisterResponse, error) {
	s.logger.Info("Registering user", zap.String("username", req.Username))

	// Convert attributes
	attrs := make(map[string]string)
	for _, attr := range req.Attributes {
		// Parse attribute (format: "key=value")
		// For now, just use as-is
		attrs[attr] = ""
	}

	registerReq := &RegisterRequest{
		Name:           req.Username,
		Type:           req.Type,
		Affiliation:    req.Affiliation,
		Attributes:     attrs,
		MaxEnrollments: -1, // Unlimited
	}

	registerResp, err := s.client.Register(registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return &models.CARegisterResponse{
		Username: req.Username,
		Secret:   registerResp.Result.Secret,
	}, nil
}

// Reenroll re-enrolls a user (renews certificate)
func (s *Service) Reenroll(ctx context.Context, username string) (*models.CAEnrollResponse, error) {
	s.logger.Info("Re-enrolling user", zap.String("username", username))

	// Load user's existing certificate
	cert, key, err := s.loadUserCertificate(username)
	if err != nil {
		return nil, fmt.Errorf("failed to load user certificate: %w", err)
	}

	reenrollReq := &ReenrollRequest{}

	reenrollResp, err := s.client.Reenroll(reenrollReq, cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to re-enroll user: %w", err)
	}

	// Parse certificate
	certPEM := reenrollResp.Result.Cert
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate")
	}

	// Store new certificate
	if err := s.storeCertificate(username, certPEM, ""); err != nil {
		s.logger.Warn("Failed to store certificate", zap.Error(err))
	}

	mspID := s.config.MSPID
	if mspID == "" {
		mspID = "Org1MSP"
	}

	return &models.CAEnrollResponse{
		Username:    username,
		Certificate: certPEM,
		PrivateKey:  "", // Private key is not returned
		MSPID:       mspID,
	}, nil
}

// Revoke revokes a user certificate
func (s *Service) Revoke(ctx context.Context, username, reason string) error {
	s.logger.Info("Revoking user certificate", zap.String("username", username))

	revokeReq := &RevokeRequest{
		Name:   username,
		Reason: reason,
		GenCRL: false,
	}

	if _, err := s.client.Revoke(revokeReq); err != nil {
		return fmt.Errorf("failed to revoke certificate: %w", err)
	}

	return nil
}

// GetCertificate retrieves a user's certificate
func (s *Service) GetCertificate(ctx context.Context, username string) (string, error) {
	certPath := filepath.Join(s.mspDir, "users", username, "msp", "signcerts", fmt.Sprintf("%s-cert.pem", username))

	cert, err := os.ReadFile(certPath)
	if err != nil {
		return "", fmt.Errorf("certificate not found: %w", err)
	}

	return string(cert), nil
}

// Helper functions

func (s *Service) storeCertificate(username, certPEM, keyPEM string) error {
	// Store in wallet directory
	userDir := filepath.Join(s.mspDir, "users", username, "msp", "signcerts")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	certPath := filepath.Join(userDir, fmt.Sprintf("%s-cert.pem", username))
	if err := os.WriteFile(certPath, []byte(certPEM), 0644); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Store private key if provided
	if keyPEM != "" {
		keyDir := filepath.Join(s.mspDir, "users", username, "msp", "keystore")
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return fmt.Errorf("failed to create key directory: %w", err)
		}

		// Find existing key file or create new one
		keyFiles, err := filepath.Glob(filepath.Join(keyDir, "*_sk"))
		if err == nil && len(keyFiles) > 0 {
			// Update existing key
			if err := os.WriteFile(keyFiles[0], []byte(keyPEM), 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}
		} else {
			// Create new key file
			keyPath := filepath.Join(keyDir, fmt.Sprintf("%s_sk", username))
			if err := os.WriteFile(keyPath, []byte(keyPEM), 0600); err != nil {
				return fmt.Errorf("failed to write private key: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) loadUserCertificate(username string) ([]byte, []byte, error) {
	return loadIdentityCredentials(s.mspDir, userCandidatePath(username))
}
