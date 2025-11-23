-- Approval Workflow Schema
-- Multi-signature approval system for chaincode deployment operations

-- Approval requests table
CREATE TABLE blockchain.approval_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    operation VARCHAR(50) NOT NULL CHECK (operation IN ('install', 'approve', 'commit')),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
    requested_by UUID NOT NULL REFERENCES auth.users(id),
    requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    reason TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Approval votes table
CREATE TABLE blockchain.approval_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    approval_request_id UUID NOT NULL REFERENCES blockchain.approval_requests(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL REFERENCES auth.users(id),
    vote VARCHAR(20) NOT NULL CHECK (vote IN ('approve', 'reject')),
    comment TEXT,
    voted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(approval_request_id, approver_id)
);

-- Approval policies table
CREATE TABLE blockchain.approval_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation VARCHAR(50) NOT NULL UNIQUE,
    required_approvals INTEGER NOT NULL DEFAULT 1 CHECK (required_approvals > 0),
    expiration_hours INTEGER DEFAULT 24 CHECK (expiration_hours > 0),
    is_active BOOLEAN DEFAULT TRUE,
    conditions JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for approval_requests
CREATE INDEX idx_approval_requests_version_id ON blockchain.approval_requests(chaincode_version_id);
CREATE INDEX idx_approval_requests_status ON blockchain.approval_requests(status) WHERE status = 'pending';
CREATE INDEX idx_approval_requests_requested_by ON blockchain.approval_requests(requested_by);
CREATE INDEX idx_approval_requests_operation ON blockchain.approval_requests(operation);
CREATE INDEX idx_approval_requests_expires_at ON blockchain.approval_requests(expires_at) WHERE status = 'pending';
CREATE INDEX idx_approval_requests_created_at ON blockchain.approval_requests(created_at DESC);

-- Composite index for common queries
CREATE INDEX idx_approval_requests_version_operation ON blockchain.approval_requests(chaincode_version_id, operation, status);

-- Indexes for approval_votes
CREATE INDEX idx_approval_votes_request_id ON blockchain.approval_votes(approval_request_id);
CREATE INDEX idx_approval_votes_approver_id ON blockchain.approval_votes(approver_id);
CREATE INDEX idx_approval_votes_vote ON blockchain.approval_votes(vote);
CREATE INDEX idx_approval_votes_voted_at ON blockchain.approval_votes(voted_at DESC);

-- Indexes for approval_policies
CREATE INDEX idx_approval_policies_operation ON blockchain.approval_policies(operation) WHERE is_active = TRUE;
CREATE INDEX idx_approval_policies_is_active ON blockchain.approval_policies(is_active) WHERE is_active = TRUE;

-- Triggers
CREATE TRIGGER update_approval_requests_updated_at BEFORE UPDATE ON blockchain.approval_requests
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_approval_policies_updated_at BEFORE UPDATE ON blockchain.approval_policies
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Function to check if approval request is approved
CREATE OR REPLACE FUNCTION blockchain.check_approval_status(request_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    required_count INTEGER;
    approval_count INTEGER;
    reject_count INTEGER;
BEGIN
    -- Get required approvals from policy
    SELECT ap.required_approvals INTO required_count
    FROM blockchain.approval_requests ar
    JOIN blockchain.approval_policies ap ON ar.operation = ap.operation
    WHERE ar.id = request_id AND ap.is_active = TRUE;
    
    -- If no policy found, default to 1
    IF required_count IS NULL THEN
        required_count := 1;
    END IF;
    
    -- Count approvals
    SELECT 
        COUNT(*) FILTER (WHERE vote = 'approve'),
        COUNT(*) FILTER (WHERE vote = 'reject')
    INTO approval_count, reject_count
    FROM blockchain.approval_votes
    WHERE approval_request_id = request_id;
    
    -- If any reject, return false
    IF reject_count > 0 THEN
        RETURN FALSE;
    END IF;
    
    -- Check if required approvals met
    RETURN approval_count >= required_count;
END;
$$ LANGUAGE plpgsql;

-- Function to update approval request status
CREATE OR REPLACE FUNCTION blockchain.update_approval_status()
RETURNS TRIGGER AS $$
DECLARE
    is_approved BOOLEAN;
    has_reject BOOLEAN;
    current_request RECORD;
BEGIN
    -- Get current request status
    SELECT status, expires_at INTO current_request
    FROM blockchain.approval_requests
    WHERE id = NEW.approval_request_id;
    
    -- Check if request is expired
    IF current_request.status = 'pending' AND current_request.expires_at IS NOT NULL AND current_request.expires_at < CURRENT_TIMESTAMP THEN
        UPDATE blockchain.approval_requests
        SET status = 'expired', updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.approval_request_id;
        RETURN NEW;
    END IF;
    
    -- Check approval status
    SELECT blockchain.check_approval_status(NEW.approval_request_id) INTO is_approved;
    
    -- Check for rejects
    SELECT EXISTS(
        SELECT 1 FROM blockchain.approval_votes
        WHERE approval_request_id = NEW.approval_request_id AND vote = 'reject'
    ) INTO has_reject;
    
    -- Update request status
    IF has_reject THEN
        UPDATE blockchain.approval_requests
        SET status = 'rejected', updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.approval_request_id AND status = 'pending';
    ELSIF is_approved THEN
        UPDATE blockchain.approval_requests
        SET status = 'approved', updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.approval_request_id AND status = 'pending';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER approval_vote_status_update AFTER INSERT OR UPDATE ON blockchain.approval_votes
    FOR EACH ROW EXECUTE FUNCTION blockchain.update_approval_status();

-- Insert default approval policies
INSERT INTO blockchain.approval_policies (operation, required_approvals, expiration_hours, is_active) VALUES
    ('install', 1, 24, TRUE),
    ('approve', 1, 24, TRUE),
    ('commit', 2, 48, TRUE)  -- Require 2 approvals for commit (more critical)
ON CONFLICT (operation) DO NOTHING;

