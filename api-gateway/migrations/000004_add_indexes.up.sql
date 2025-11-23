-- Additional performance indexes

-- Batch analytics cache table
CREATE TABLE batch_analytics (
    batch_id VARCHAR(255) PRIMARY KEY,
    farm_location VARCHAR(255),
    status VARCHAR(50),
    owner VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    verification_count INTEGER DEFAULT 0,
    query_count INTEGER DEFAULT 0,
    last_queried_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for batch analytics
CREATE INDEX idx_batch_analytics_status ON batch_analytics(status, created_at DESC);
CREATE INDEX idx_batch_analytics_location ON batch_analytics(farm_location);
CREATE INDEX idx_batch_analytics_owner ON batch_analytics(owner);
CREATE INDEX idx_batch_analytics_created_at ON batch_analytics(created_at DESC);

-- Sessions table for JWT refresh tokens
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) UNIQUE NOT NULL,
    user_agent TEXT,
    ip_address INET,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes on sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Add trigger for batch_analytics updated_at
CREATE TRIGGER update_batch_analytics_updated_at BEFORE UPDATE ON batch_analytics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

