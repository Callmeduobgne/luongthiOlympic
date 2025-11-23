-- CI/CD Integration Schema
-- Tracks CI/CD pipelines, builds, and automated deployments

-- CI/CD pipelines table
CREATE TABLE blockchain.cicd_pipelines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    
    -- Pipeline configuration
    chaincode_name VARCHAR(255) NOT NULL,
    channel_name VARCHAR(100) NOT NULL,
    source_type VARCHAR(50) NOT NULL DEFAULT 'git' CHECK (source_type IN ('git', 'manual', 'api')),
    source_repository TEXT, -- Git repository URL
    source_branch VARCHAR(255) DEFAULT 'main',
    source_path TEXT, -- Path to chaincode in repository
    
    -- Build configuration
    build_command TEXT, -- e.g., 'npm install && npm run build'
    test_command TEXT, -- e.g., 'npm test'
    package_command TEXT, -- e.g., 'peer lifecycle chaincode package'
    
    -- Deployment configuration
    auto_deploy BOOLEAN DEFAULT FALSE, -- Auto deploy after successful build
    deploy_on_tags BOOLEAN DEFAULT TRUE, -- Deploy when git tag is created
    deploy_environment VARCHAR(50) DEFAULT 'production' CHECK (deploy_environment IN ('development', 'staging', 'production')),
    
    -- Webhook configuration
    webhook_url TEXT, -- Webhook URL for triggering pipeline
    webhook_secret TEXT, -- Secret for webhook authentication
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Pipeline executions table
CREATE TABLE blockchain.cicd_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES blockchain.cicd_pipelines(id) ON DELETE CASCADE,
    
    -- Execution details
    trigger_type VARCHAR(50) NOT NULL CHECK (trigger_type IN ('webhook', 'manual', 'scheduled', 'api')),
    trigger_source TEXT, -- Git commit hash, tag, or manual trigger info
    triggered_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled')),
    
    -- Execution stages
    build_status VARCHAR(50) DEFAULT 'pending',
    test_status VARCHAR(50) DEFAULT 'pending',
    package_status VARCHAR(50) DEFAULT 'pending',
    deploy_status VARCHAR(50) DEFAULT 'pending',
    
    -- Timestamps
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- Results
    build_output TEXT,
    test_output TEXT,
    package_path TEXT, -- Path to built package
    deployment_id UUID, -- Reference to chaincode version if deployed
    
    -- Error tracking
    error_message TEXT,
    error_stage VARCHAR(50), -- build, test, package, deploy
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Build artifacts table
CREATE TABLE blockchain.cicd_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES blockchain.cicd_executions(id) ON DELETE CASCADE,
    
    -- Artifact details
    artifact_type VARCHAR(50) NOT NULL CHECK (artifact_type IN ('package', 'build', 'test_report', 'log')),
    artifact_path TEXT NOT NULL, -- Path to artifact file
    artifact_size BIGINT, -- Size in bytes
    mime_type VARCHAR(100),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Webhook events table (for tracking webhook triggers)
CREATE TABLE blockchain.cicd_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES blockchain.cicd_pipelines(id) ON DELETE CASCADE,
    
    -- Event details
    event_type VARCHAR(50) NOT NULL, -- push, tag, pull_request, etc.
    payload JSONB NOT NULL, -- Full webhook payload
    signature TEXT, -- Webhook signature for verification
    
    -- Processing
    processed BOOLEAN DEFAULT FALSE,
    execution_id UUID REFERENCES blockchain.cicd_executions(id) ON DELETE SET NULL,
    
    -- Error tracking
    error_message TEXT,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMPTZ
);

-- Indexes for cicd_pipelines
CREATE INDEX idx_cicd_pipelines_name ON blockchain.cicd_pipelines(name);
CREATE INDEX idx_cicd_pipelines_chaincode ON blockchain.cicd_pipelines(chaincode_name, channel_name);
CREATE INDEX idx_cicd_pipelines_is_active ON blockchain.cicd_pipelines(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_cicd_pipelines_created_at ON blockchain.cicd_pipelines(created_at DESC);

-- Indexes for cicd_executions
CREATE INDEX idx_cicd_executions_pipeline_id ON blockchain.cicd_executions(pipeline_id);
CREATE INDEX idx_cicd_executions_status ON blockchain.cicd_executions(status);
CREATE INDEX idx_cicd_executions_trigger_type ON blockchain.cicd_executions(trigger_type);
CREATE INDEX idx_cicd_executions_triggered_by ON blockchain.cicd_executions(triggered_by);
CREATE INDEX idx_cicd_executions_started_at ON blockchain.cicd_executions(started_at DESC);
CREATE INDEX idx_cicd_executions_pipeline_status ON blockchain.cicd_executions(pipeline_id, status);

-- Indexes for cicd_artifacts
CREATE INDEX idx_cicd_artifacts_execution_id ON blockchain.cicd_artifacts(execution_id);
CREATE INDEX idx_cicd_artifacts_type ON blockchain.cicd_artifacts(artifact_type);
CREATE INDEX idx_cicd_artifacts_created_at ON blockchain.cicd_artifacts(created_at DESC);

-- Indexes for cicd_webhook_events
CREATE INDEX idx_cicd_webhook_events_pipeline_id ON blockchain.cicd_webhook_events(pipeline_id);
CREATE INDEX idx_cicd_webhook_events_processed ON blockchain.cicd_webhook_events(processed) WHERE processed = FALSE;
CREATE INDEX idx_cicd_webhook_events_type ON blockchain.cicd_webhook_events(event_type);
CREATE INDEX idx_cicd_webhook_events_created_at ON blockchain.cicd_webhook_events(created_at DESC);

-- Triggers
CREATE TRIGGER update_cicd_pipelines_updated_at BEFORE UPDATE ON blockchain.cicd_pipelines
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_cicd_executions_updated_at BEFORE UPDATE ON blockchain.cicd_executions
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Function to calculate execution duration
CREATE OR REPLACE FUNCTION blockchain.calculate_cicd_execution_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_ms = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at)) * 1000;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_cicd_execution_duration_trigger
BEFORE UPDATE ON blockchain.cicd_executions
FOR EACH ROW
EXECUTE FUNCTION blockchain.calculate_cicd_execution_duration();

-- Function to get latest execution for a pipeline
CREATE OR REPLACE FUNCTION blockchain.get_latest_pipeline_execution(p_pipeline_id UUID)
RETURNS UUID AS $$
DECLARE
    v_execution_id UUID;
BEGIN
    SELECT id INTO v_execution_id
    FROM blockchain.cicd_executions
    WHERE pipeline_id = p_pipeline_id
    ORDER BY created_at DESC
    LIMIT 1;
    
    RETURN v_execution_id;
END;
$$ LANGUAGE plpgsql;

