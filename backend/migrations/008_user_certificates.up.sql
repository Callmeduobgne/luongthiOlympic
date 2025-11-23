-- User Certificates table
-- Migration: 008_user_certificates
-- Description: Store user Fabric certificates with encrypted private keys (Enterprise-grade security)

-- User certificates table (separate from users for security)
CREATE TABLE IF NOT EXISTS auth.user_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    
    -- Certificate (public - can be stored plaintext)
    certificate TEXT NOT NULL, -- PEM format certificate (public)
    
    -- Private key (MUST be encrypted)
    encrypted_private_key TEXT NOT NULL, -- AES-256-GCM encrypted private key
    encryption_key_id VARCHAR(100), -- Key ID for rotation (e.g., "master-v1")
    
    -- Certificate metadata
    msp_id VARCHAR(100) NOT NULL, -- Fabric MSP ID
    ca_name VARCHAR(100), -- CA that issued the certificate
    serial_number VARCHAR(255), -- Certificate serial number
    issuer VARCHAR(255), -- Certificate issuer
    
    -- Certificate lifecycle
    issued_at TIMESTAMPTZ NOT NULL, -- When certificate was issued
    expires_at TIMESTAMPTZ NOT NULL, -- Certificate expiration
    revoked_at TIMESTAMPTZ, -- When certificate was revoked
    is_revoked BOOLEAN DEFAULT FALSE,
    
    -- Key rotation support
    previous_certificate_id UUID REFERENCES auth.user_certificates(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE, -- Only one active cert per user
    
    -- Audit fields
    created_by UUID REFERENCES auth.users(id), -- Who created this certificate
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT unique_active_cert UNIQUE (user_id, is_active) DEFERRABLE INITIALLY DEFERRED
);

-- Indexes for performance
CREATE INDEX idx_user_certificates_user_id ON auth.user_certificates(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_certificates_user_id_active ON auth.user_certificates(user_id, is_active) WHERE deleted_at IS NULL AND is_active = TRUE;
CREATE INDEX idx_user_certificates_msp_id ON auth.user_certificates(msp_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_certificates_expires_at ON auth.user_certificates(expires_at) WHERE deleted_at IS NULL AND is_revoked = FALSE;
CREATE INDEX idx_user_certificates_serial_number ON auth.user_certificates(serial_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_certificates_is_revoked ON auth.user_certificates(is_revoked) WHERE deleted_at IS NULL;

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_user_certificates_updated_at BEFORE UPDATE ON auth.user_certificates
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Trigger to ensure only one active certificate per user
CREATE OR REPLACE FUNCTION auth.ensure_single_active_cert()
RETURNS TRIGGER AS $$
BEGIN
    -- If setting a cert as active, deactivate others
    IF NEW.is_active = TRUE THEN
        UPDATE auth.user_certificates
        SET is_active = FALSE
        WHERE user_id = NEW.user_id
          AND id != NEW.id
          AND is_active = TRUE
          AND deleted_at IS NULL;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER ensure_single_active_cert_trigger
    BEFORE INSERT OR UPDATE ON auth.user_certificates
    FOR EACH ROW
    WHEN (NEW.is_active = TRUE)
    EXECUTE FUNCTION auth.ensure_single_active_cert();

-- Comments for documentation
COMMENT ON TABLE auth.user_certificates IS 'Stores user Fabric certificates with encrypted private keys (Enterprise-grade security)';
COMMENT ON COLUMN auth.user_certificates.certificate IS 'PEM format certificate (public - stored plaintext)';
COMMENT ON COLUMN auth.user_certificates.encrypted_private_key IS 'AES-256-GCM encrypted private key (NEVER store plaintext)';
COMMENT ON COLUMN auth.user_certificates.encryption_key_id IS 'Key ID for encryption key rotation support';
COMMENT ON COLUMN auth.user_certificates.is_active IS 'Only one active certificate per user (enforced by trigger)';

