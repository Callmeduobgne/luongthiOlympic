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

package certificate

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/utils"
	"go.uber.org/zap"
)

// Service handles certificate operations with encryption
type Service struct {
	repo      *Repository
	masterKey string
	logger    *zap.Logger
}

// NewService creates a new certificate service
func NewService(repo *Repository, masterKey string, logger *zap.Logger) (*Service, error) {
	// Validate master key
	if err := utils.ValidateMasterKey(masterKey); err != nil {
		return nil, fmt.Errorf("invalid master key: %w", err)
	}

	return &Service{
		repo:      repo,
		masterKey: masterKey,
		logger:    logger,
	}, nil
}

// CreateCertificate creates a new certificate with encrypted private key
func (s *Service) CreateCertificate(ctx context.Context, req *CreateCertificateRequest, createdBy uuid.UUID) (*UserCertificate, error) {
	// Encrypt private key
	encryptedKey, err := utils.EncryptPrivateKey(req.PrivateKey, s.masterKey)
	if err != nil {
		s.logger.Error("Failed to encrypt private key", zap.Error(err))
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	// Create certificate record
	cert := &UserCertificate{
		ID:                  uuid.New(),
		UserID:              req.UserID,
		Certificate:         req.Certificate,
		EncryptedPrivateKey: encryptedKey,
		EncryptionKeyID:     "master-v1", // Can be rotated later
		MSPID:               req.MSPID,
		CAName:              req.CAName,
		SerialNumber:        req.SerialNumber,
		Issuer:              req.Issuer,
		IssuedAt:            req.IssuedAt,
		ExpiresAt:           req.ExpiresAt,
		IsRevoked:           false,
		IsActive:            true,
		CreatedBy:           &createdBy,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Save to database
	if err := s.repo.CreateCertificate(ctx, cert); err != nil {
		s.logger.Error("Failed to create certificate", zap.Error(err))
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	s.logger.Info("Certificate created",
		zap.String("cert_id", cert.ID.String()),
		zap.String("user_id", cert.UserID.String()),
		zap.String("msp_id", cert.MSPID),
	)

	return cert, nil
}

// GetActiveCertificateWithKey retrieves active certificate with decrypted private key
// WARNING: Private key is decrypted in memory - handle with care
func (s *Service) GetActiveCertificateWithKey(ctx context.Context, userID uuid.UUID) (*CertificateWithKey, error) {
	// Get certificate from database
	cert, err := s.repo.GetActiveCertificateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Decrypt private key
	privateKey, err := utils.DecryptPrivateKey(cert.EncryptedPrivateKey, s.masterKey)
	if err != nil {
		s.logger.Error("Failed to decrypt private key",
			zap.String("cert_id", cert.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// Return certificate with decrypted key (in memory only)
	return &CertificateWithKey{
		UserCertificate: cert,
		PrivateKey:      privateKey,
	}, nil
}

// GetActiveCertificate retrieves active certificate without private key
func (s *Service) GetActiveCertificate(ctx context.Context, userID uuid.UUID) (*UserCertificate, error) {
	return s.repo.GetActiveCertificateByUserID(ctx, userID)
}

// GetCertificateByID retrieves a certificate by ID
func (s *Service) GetCertificateByID(ctx context.Context, id uuid.UUID) (*UserCertificate, error) {
	return s.repo.GetCertificateByID(ctx, id)
}

// RevokeCertificate revokes a certificate
func (s *Service) RevokeCertificate(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	if err := s.repo.RevokeCertificate(ctx, id, revokedBy); err != nil {
		return err
	}

	s.logger.Info("Certificate revoked",
		zap.String("cert_id", id.String()),
		zap.String("revoked_by", revokedBy.String()),
	)

	return nil
}

// ListCertificates lists all certificates for a user
func (s *Service) ListCertificates(ctx context.Context, userID uuid.UUID) ([]*UserCertificate, error) {
	return s.repo.ListCertificatesByUserID(ctx, userID)
}

// RotateCertificate creates a new certificate and deactivates the old one
func (s *Service) RotateCertificate(ctx context.Context, userID uuid.UUID, req *CreateCertificateRequest, createdBy uuid.UUID) (*UserCertificate, error) {
	// Get current active certificate
	currentCert, err := s.repo.GetActiveCertificateByUserID(ctx, userID)
	if err != nil && err != ErrCertificateNotFound {
		return nil, err
	}

	// Create new certificate
	newCert, err := s.CreateCertificate(ctx, req, createdBy)
	if err != nil {
		return nil, err
	}

	// If there was a previous certificate, link it
	if currentCert != nil {
		newCert.PreviousCertID = &currentCert.ID
		// Update previous cert to link to new one (optional, for audit trail)
		// The trigger will automatically deactivate the old cert
	}

	s.logger.Info("Certificate rotated",
		zap.String("new_cert_id", newCert.ID.String()),
		zap.String("user_id", userID.String()),
		zap.String("previous_cert_id", func() string {
			if currentCert != nil {
				return currentCert.ID.String()
			}
			return "none"
		}()),
	)

	return newCert, nil
}

