-- Rollback Mechanisms Schema
-- Tracks rollback operations for chaincode deployments

-- Rollback operations table
CREATE TABLE blockchain.rollback_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_name VARCHAR(255) NOT NULL,
    channel_name VARCHAR(100) NOT NULL,
    from_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE RESTRICT,
    to_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE RESTRICT,
    from_version VARCHAR(50) NOT NULL,
    to_version VARCHAR(50) NOT NULL,
    from_sequence INTEGER NOT NULL,
    to_sequence INTEGER NOT NULL,
    
    -- Rollback status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'cancelled')),
    
    -- Rollback details
    reason TEXT,
    rollback_type VARCHAR(50) NOT NULL DEFAULT 'version' CHECK (rollback_type IN ('version', 'sequence')),
    
    -- Execution tracking
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- User tracking
    requested_by UUID NOT NULL REFERENCES auth.users(id),
    executed_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    
    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(50),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Rollback history table (tracks what was rolled back)
CREATE TABLE blockchain.rollback_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rollback_operation_id UUID NOT NULL REFERENCES blockchain.rollback_operations(id) ON DELETE CASCADE,
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    
    -- What was rolled back
    operation VARCHAR(50) NOT NULL, -- 'install', 'approve', 'commit'
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    
    -- Details
    details JSONB DEFAULT '{}',
    
    -- Timestamp
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for rollback_operations
CREATE INDEX idx_rollback_operations_chaincode ON blockchain.rollback_operations(chaincode_name, channel_name);
CREATE INDEX idx_rollback_operations_status ON blockchain.rollback_operations(status) WHERE status IN ('pending', 'in_progress');
CREATE INDEX idx_rollback_operations_from_version ON blockchain.rollback_operations(from_version_id);
CREATE INDEX idx_rollback_operations_to_version ON blockchain.rollback_operations(to_version_id);
CREATE INDEX idx_rollback_operations_requested_by ON blockchain.rollback_operations(requested_by);
CREATE INDEX idx_rollback_operations_created_at ON blockchain.rollback_operations(created_at DESC);

-- Composite index for common queries
CREATE INDEX idx_rollback_operations_chaincode_channel ON blockchain.rollback_operations(chaincode_name, channel_name, created_at DESC);

-- Indexes for rollback_history
CREATE INDEX idx_rollback_history_operation_id ON blockchain.rollback_history(rollback_operation_id);
CREATE INDEX idx_rollback_history_version_id ON blockchain.rollback_history(chaincode_version_id);
CREATE INDEX idx_rollback_history_created_at ON blockchain.rollback_history(created_at DESC);

-- Triggers
CREATE TRIGGER update_rollback_operations_updated_at BEFORE UPDATE ON blockchain.rollback_operations
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Function to get previous active version for rollback
CREATE OR REPLACE FUNCTION blockchain.get_previous_active_version(
    p_chaincode_name VARCHAR,
    p_channel_name VARCHAR,
    p_current_sequence INTEGER
)
RETURNS UUID AS $$
DECLARE
    v_version_id UUID;
BEGIN
    SELECT id INTO v_version_id
    FROM blockchain.chaincode_versions
    WHERE name = p_chaincode_name
      AND channel_name = p_channel_name
      AND sequence < p_current_sequence
      AND commit_status = 'committed'
      AND deleted_at IS NULL
    ORDER BY sequence DESC
    LIMIT 1;
    
    RETURN v_version_id;
END;
$$ LANGUAGE plpgsql;

-- Function to check if rollback is safe (no pending operations)
CREATE OR REPLACE FUNCTION blockchain.is_rollback_safe(
    p_chaincode_name VARCHAR,
    p_channel_name VARCHAR
)
RETURNS BOOLEAN AS $$
DECLARE
    v_pending_count INTEGER;
BEGIN
    -- Check for pending approval requests
    SELECT COUNT(*) INTO v_pending_count
    FROM blockchain.approval_requests ar
    JOIN blockchain.chaincode_versions cv ON ar.chaincode_version_id = cv.id
    WHERE cv.name = p_chaincode_name
      AND cv.channel_name = p_channel_name
      AND ar.status = 'pending'
      AND ar.operation = 'commit';
    
    IF v_pending_count > 0 THEN
        RETURN FALSE;
    END IF;
    
    -- Check for in-progress rollback operations
    SELECT COUNT(*) INTO v_pending_count
    FROM blockchain.rollback_operations
    WHERE chaincode_name = p_chaincode_name
      AND channel_name = p_channel_name
      AND status IN ('pending', 'in_progress');
    
    IF v_pending_count > 0 THEN
        RETURN FALSE;
    END IF;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

