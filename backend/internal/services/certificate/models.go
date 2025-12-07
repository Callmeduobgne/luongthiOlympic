package certificate

import (
	"time"

	"github.com/google/uuid"
)

// UserCertificate represents a user's Fabric certificate
type UserCertificate struct {
	ID                  uuid.UUID  `json:"id"`
	UserID              uuid.UUID  `json:"user_id"`
	Certificate         string     `json:"certificate"`          // PEM format (public)
	EncryptedPrivateKey string     `json:"-"`                   // Encrypted private key (never expose)
	EncryptionKeyID     string     `json:"encryption_key_id"`    // Key ID for rotation
	MSPID               string     `json:"msp_id"`
	CAName              *string    `json:"ca_name,omitempty"`
	SerialNumber        *string    `json:"serial_number,omitempty"`
	Issuer              *string    `json:"issuer,omitempty"`
	IssuedAt            time.Time  `json:"issued_at"`
	ExpiresAt           time.Time  `json:"expires_at"`
	RevokedAt           *time.Time `json:"revoked_at,omitempty"`
	IsRevoked           bool       `json:"is_revoked"`
	PreviousCertID      *uuid.UUID `json:"previous_cert_id,omitempty"`
	IsActive            bool       `json:"is_active"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
}

// CreateCertificateRequest represents request to create a certificate
type CreateCertificateRequest struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	Certificate string    `json:"certificate" validate:"required"` // PEM format
	PrivateKey  string    `json:"private_key" validate:"required"` // PEM format (will be encrypted)
	MSPID       string    `json:"msp_id" validate:"required"`
	CAName      *string   `json:"ca_name,omitempty"`
	SerialNumber *string  `json:"serial_number,omitempty"`
	Issuer      *string   `json:"issuer,omitempty"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at" validate:"required"`
}

// CertificateWithKey represents certificate with decrypted private key (for internal use only)
type CertificateWithKey struct {
	*UserCertificate
	PrivateKey string `json:"-"` // Decrypted private key (only in memory, never persisted)
}

