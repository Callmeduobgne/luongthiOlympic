-- Chaincode Registry Schema
-- Tracks chaincode versions, deployments, and active chaincodes

-- Chaincode versions table
-- Stores all chaincode versions that have been installed/approved/committed
CREATE TABLE blockchain.chaincode_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 1, -- Lifecycle sequence number
    package_id VARCHAR(255), -- Package ID from install
    label VARCHAR(255), -- Package label
    path TEXT, -- Chaincode path on server
    package_path TEXT, -- Full path to package file
    channel_name VARCHAR(100) NOT NULL,
    
    -- Lifecycle status
    install_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, installed, failed
    approve_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, failed
    commit_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, committed, failed
    
    -- Chaincode definition
    init_required BOOLEAN DEFAULT FALSE,
    endorsement_plugin VARCHAR(100) DEFAULT 'escc',
    validation_plugin VARCHAR(100) DEFAULT 'vscc',
    collections JSONB, -- Chaincode collections config
    
    -- Metadata
    installed_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    committed_at TIMESTAMPTZ,
    installed_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    committed_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    
    -- Error tracking
    install_error TEXT,
    approve_error TEXT,
    commit_error TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    UNIQUE(name, version, channel_name, sequence)
);

-- Deployment logs table
-- Detailed logs for each deployment operation (install/approve/commit)
CREATE TABLE blockchain.deployment_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    operation VARCHAR(50) NOT NULL CHECK (operation IN ('install', 'approve', 'commit')),
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'in_progress', 'success', 'failed')),
    
    -- Operation details
    request_data JSONB, -- Full request payload
    response_data JSONB, -- Response from peer CLI
    error_message TEXT,
    error_code VARCHAR(50),
    
    -- Execution metadata
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER, -- Duration in milliseconds
    
    -- User tracking
    performed_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Active chaincodes table
-- Tracks currently active (committed) chaincodes on each channel
CREATE TABLE blockchain.active_chaincodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    sequence INTEGER NOT NULL,
    channel_name VARCHAR(100) NOT NULL,
    
    -- Chaincode info from peer
    package_id VARCHAR(255),
    init_required BOOLEAN DEFAULT FALSE,
    endorsement_plugin VARCHAR(100),
    validation_plugin VARCHAR(100),
    collections JSONB,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    activated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMPTZ,
    
    -- Metadata
    metadata JSONB, -- Additional chaincode metadata from peer
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    UNIQUE(name, channel_name, sequence)
);

-- Indexes for chaincode_versions
CREATE INDEX idx_chaincode_versions_name ON blockchain.chaincode_versions(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_chaincode_versions_channel ON blockchain.chaincode_versions(channel_name) WHERE deleted_at IS NULL;
CREATE INDEX idx_chaincode_versions_status ON blockchain.chaincode_versions(commit_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chaincode_versions_name_version ON blockchain.chaincode_versions(name, version) WHERE deleted_at IS NULL;
CREATE INDEX idx_chaincode_versions_created_at ON blockchain.chaincode_versions(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_chaincode_versions_installed_by ON blockchain.chaincode_versions(installed_by) WHERE deleted_at IS NULL;

-- Composite index for common queries
CREATE INDEX idx_chaincode_versions_name_channel ON blockchain.chaincode_versions(name, channel_name, sequence DESC) WHERE deleted_at IS NULL;

-- Indexes for deployment_logs
CREATE INDEX idx_deployment_logs_version_id ON blockchain.deployment_logs(chaincode_version_id);
CREATE INDEX idx_deployment_logs_operation ON blockchain.deployment_logs(operation);
CREATE INDEX idx_deployment_logs_status ON blockchain.deployment_logs(status);
CREATE INDEX idx_deployment_logs_created_at ON blockchain.deployment_logs(created_at DESC);
CREATE INDEX idx_deployment_logs_performed_by ON blockchain.deployment_logs(performed_by);

-- Composite index for operation tracking
CREATE INDEX idx_deployment_logs_version_operation ON blockchain.deployment_logs(chaincode_version_id, operation, created_at DESC);

-- Indexes for active_chaincodes
CREATE INDEX idx_active_chaincodes_name ON blockchain.active_chaincodes(name);
CREATE INDEX idx_active_chaincodes_channel ON blockchain.active_chaincodes(channel_name);
CREATE INDEX idx_active_chaincodes_is_active ON blockchain.active_chaincodes(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_active_chaincodes_name_channel ON blockchain.active_chaincodes(name, channel_name);
CREATE INDEX idx_active_chaincodes_activated_at ON blockchain.active_chaincodes(activated_at DESC);

-- Trigger to update updated_at timestamp for chaincode_versions
CREATE TRIGGER update_chaincode_versions_updated_at BEFORE UPDATE ON blockchain.chaincode_versions
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Trigger to update updated_at timestamp for active_chaincodes
CREATE TRIGGER update_active_chaincodes_updated_at BEFORE UPDATE ON blockchain.active_chaincodes
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Function to automatically create/update active_chaincodes when commit succeeds
CREATE OR REPLACE FUNCTION blockchain.sync_active_chaincode()
RETURNS TRIGGER AS $$
BEGIN
    -- When commit_status changes to 'committed', create/update active_chaincode
    IF NEW.commit_status = 'committed' AND (OLD.commit_status IS NULL OR OLD.commit_status != 'committed') THEN
        INSERT INTO blockchain.active_chaincodes (
            chaincode_version_id, name, version, sequence, channel_name,
            package_id, init_required, endorsement_plugin, validation_plugin,
            collections, is_active, activated_at, metadata
        ) VALUES (
            NEW.id, NEW.name, NEW.version, NEW.sequence, NEW.channel_name,
            NEW.package_id, NEW.init_required, NEW.endorsement_plugin, NEW.validation_plugin,
            NEW.collections, TRUE, COALESCE(NEW.committed_at, CURRENT_TIMESTAMP),
            jsonb_build_object(
                'installed_at', NEW.installed_at,
                'approved_at', NEW.approved_at,
                'committed_at', NEW.committed_at
            )
        )
        ON CONFLICT (name, channel_name, sequence) 
        DO UPDATE SET
            chaincode_version_id = NEW.id,
            version = NEW.version,
            package_id = NEW.package_id,
            init_required = NEW.init_required,
            endorsement_plugin = NEW.endorsement_plugin,
            validation_plugin = NEW.validation_plugin,
            collections = NEW.collections,
            is_active = TRUE,
            activated_at = COALESCE(NEW.committed_at, CURRENT_TIMESTAMP),
            updated_at = CURRENT_TIMESTAMP;
    END IF;
    
    -- When commit_status changes from 'committed' to something else, deactivate
    IF OLD.commit_status = 'committed' AND NEW.commit_status != 'committed' THEN
        UPDATE blockchain.active_chaincodes
        SET is_active = FALSE,
            deactivated_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE chaincode_version_id = NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER sync_active_chaincode_trigger AFTER UPDATE ON blockchain.chaincode_versions
    FOR EACH ROW EXECUTE FUNCTION blockchain.sync_active_chaincode();

-- Function to calculate deployment duration
CREATE OR REPLACE FUNCTION blockchain.calculate_deployment_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_ms := EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at)) * 1000;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER calculate_deployment_duration_trigger BEFORE INSERT OR UPDATE ON blockchain.deployment_logs
    FOR EACH ROW EXECUTE FUNCTION blockchain.calculate_deployment_duration();

