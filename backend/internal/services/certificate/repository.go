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
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for user certificates
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new certificate repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateCertificate creates a new user certificate
func (r *Repository) CreateCertificate(ctx context.Context, cert *UserCertificate) error {
	query := `
		INSERT INTO auth.user_certificates (
			id, user_id, certificate, encrypted_private_key, encryption_key_id,
			msp_id, ca_name, serial_number, issuer,
			issued_at, expires_at, is_revoked, is_active, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	_, err := r.db.Exec(ctx, query,
		cert.ID,
		cert.UserID,
		cert.Certificate,
		cert.EncryptedPrivateKey,
		cert.EncryptionKeyID,
		cert.MSPID,
		cert.CAName,
		cert.SerialNumber,
		cert.Issuer,
		cert.IssuedAt,
		cert.ExpiresAt,
		cert.IsRevoked,
		cert.IsActive,
		cert.CreatedBy,
	)

	return err
}

// GetActiveCertificateByUserID retrieves the active certificate for a user
func (r *Repository) GetActiveCertificateByUserID(ctx context.Context, userID uuid.UUID) (*UserCertificate, error) {
	query := `
		SELECT 
			id, user_id, certificate, encrypted_private_key, encryption_key_id,
			msp_id, ca_name, serial_number, issuer,
			issued_at, expires_at, revoked_at, is_revoked,
			previous_certificate_id, is_active, created_by,
			created_at, updated_at, deleted_at
		FROM auth.user_certificates
		WHERE user_id = $1 
		  AND is_active = TRUE 
		  AND is_revoked = FALSE
		  AND deleted_at IS NULL
		  AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	cert := &UserCertificate{}
	var revokedAt, deletedAt sql.NullTime
	var previousCertID, createdBy sql.NullString
	var caName, serialNumber, issuer sql.NullString

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&cert.ID,
		&cert.UserID,
		&cert.Certificate,
		&cert.EncryptedPrivateKey,
		&cert.EncryptionKeyID,
		&cert.MSPID,
		&caName,
		&serialNumber,
		&issuer,
		&cert.IssuedAt,
		&cert.ExpiresAt,
		&revokedAt,
		&cert.IsRevoked,
		&previousCertID,
		&cert.IsActive,
		&createdBy,
		&cert.CreatedAt,
		&cert.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// Handle nullable fields
	if caName.Valid {
		cert.CAName = &caName.String
	}
	if serialNumber.Valid {
		cert.SerialNumber = &serialNumber.String
	}
	if issuer.Valid {
		cert.Issuer = &issuer.String
	}
	if revokedAt.Valid {
		cert.RevokedAt = &revokedAt.Time
	}
	if previousCertID.Valid {
		prevID, _ := uuid.Parse(previousCertID.String)
		cert.PreviousCertID = &prevID
	}
	if createdBy.Valid {
		createdByID, _ := uuid.Parse(createdBy.String)
		cert.CreatedBy = &createdByID
	}

	return cert, nil
}

// GetCertificateByID retrieves a certificate by ID
func (r *Repository) GetCertificateByID(ctx context.Context, id uuid.UUID) (*UserCertificate, error) {
	query := `
		SELECT 
			id, user_id, certificate, encrypted_private_key, encryption_key_id,
			msp_id, ca_name, serial_number, issuer,
			issued_at, expires_at, revoked_at, is_revoked,
			previous_certificate_id, is_active, created_by,
			created_at, updated_at, deleted_at
		FROM auth.user_certificates
		WHERE id = $1 AND deleted_at IS NULL
	`

	cert := &UserCertificate{}
	var revokedAt, deletedAt sql.NullTime
	var previousCertID, createdBy sql.NullString
	var caName, serialNumber, issuer sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&cert.ID,
		&cert.UserID,
		&cert.Certificate,
		&cert.EncryptedPrivateKey,
		&cert.EncryptionKeyID,
		&cert.MSPID,
		&caName,
		&serialNumber,
		&issuer,
		&cert.IssuedAt,
		&cert.ExpiresAt,
		&revokedAt,
		&cert.IsRevoked,
		&previousCertID,
		&cert.IsActive,
		&createdBy,
		&cert.CreatedAt,
		&cert.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCertificateNotFound
		}
		return nil, err
	}

	// Handle nullable fields
	if caName.Valid {
		cert.CAName = &caName.String
	}
	if serialNumber.Valid {
		cert.SerialNumber = &serialNumber.String
	}
	if issuer.Valid {
		cert.Issuer = &issuer.String
	}
	if revokedAt.Valid {
		cert.RevokedAt = &revokedAt.Time
	}
	if previousCertID.Valid {
		prevID, _ := uuid.Parse(previousCertID.String)
		cert.PreviousCertID = &prevID
	}
	if createdBy.Valid {
		createdByID, _ := uuid.Parse(createdBy.String)
		cert.CreatedBy = &createdByID
	}

	return cert, nil
}

// RevokeCertificate revokes a certificate
func (r *Repository) RevokeCertificate(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	query := `
		UPDATE auth.user_certificates
		SET is_revoked = TRUE,
		    revoked_at = NOW(),
		    is_active = FALSE,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCertificateNotFound
	}

	return nil
}

// ListCertificatesByUserID lists all certificates for a user
func (r *Repository) ListCertificatesByUserID(ctx context.Context, userID uuid.UUID) ([]*UserCertificate, error) {
	query := `
		SELECT 
			id, user_id, certificate, encrypted_private_key, encryption_key_id,
			msp_id, ca_name, serial_number, issuer,
			issued_at, expires_at, revoked_at, is_revoked,
			previous_certificate_id, is_active, created_by,
			created_at, updated_at, deleted_at
		FROM auth.user_certificates
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certificates []*UserCertificate
	for rows.Next() {
		cert := &UserCertificate{}
		var revokedAt, deletedAt sql.NullTime
		var previousCertID, createdBy sql.NullString
		var caName, serialNumber, issuer sql.NullString

		err := rows.Scan(
			&cert.ID,
			&cert.UserID,
			&cert.Certificate,
			&cert.EncryptedPrivateKey,
			&cert.EncryptionKeyID,
			&cert.MSPID,
			&caName,
			&serialNumber,
			&issuer,
			&cert.IssuedAt,
			&cert.ExpiresAt,
			&revokedAt,
			&cert.IsRevoked,
			&previousCertID,
			&cert.IsActive,
			&createdBy,
			&cert.CreatedAt,
			&cert.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if caName.Valid {
			cert.CAName = &caName.String
		}
		if serialNumber.Valid {
			cert.SerialNumber = &serialNumber.String
		}
		if issuer.Valid {
			cert.Issuer = &issuer.String
		}
		if revokedAt.Valid {
			cert.RevokedAt = &revokedAt.Time
		}
		if previousCertID.Valid {
			prevID, _ := uuid.Parse(previousCertID.String)
			cert.PreviousCertID = &prevID
		}
		if createdBy.Valid {
			createdByID, _ := uuid.Parse(createdBy.String)
			cert.CreatedBy = &createdByID
		}

		certificates = append(certificates, cert)
	}

	return certificates, rows.Err()
}

// Errors
var (
	ErrCertificateNotFound = &CertificateError{Message: "certificate not found"}
)

// CertificateError represents a certificate-related error
type CertificateError struct {
	Message string
}

func (e *CertificateError) Error() string {
	return e.Message
}

