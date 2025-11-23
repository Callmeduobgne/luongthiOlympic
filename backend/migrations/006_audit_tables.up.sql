-- Audit schema tables

-- Audit logs table (partitioned by month)
CREATE TABLE audit.audit_logs (
    id UUID DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    status VARCHAR(50) NOT NULL, -- success, failure
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(255),
    method VARCHAR(10),
    path TEXT,
    duration_ms INT,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions for current and next 3 months
CREATE TABLE audit.audit_logs_2024_11 PARTITION OF audit.audit_logs
    FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');

CREATE TABLE audit.audit_logs_2024_12 PARTITION OF audit.audit_logs
    FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');

CREATE TABLE audit.audit_logs_2025_01 PARTITION OF audit.audit_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE TABLE audit.audit_logs_2025_02 PARTITION OF audit.audit_logs
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');

-- Indexes on partitioned table
CREATE INDEX idx_audit_logs_user_id ON audit.audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit.audit_logs(action, created_at DESC);
CREATE INDEX idx_audit_logs_resource_type ON audit.audit_logs(resource_type, created_at DESC);
CREATE INDEX idx_audit_logs_status ON audit.audit_logs(status, created_at DESC);
CREATE INDEX idx_audit_logs_request_id ON audit.audit_logs(request_id);
CREATE INDEX idx_audit_logs_created_at ON audit.audit_logs(created_at DESC);

-- Full-text search index for searching in action and metadata
CREATE INDEX idx_audit_logs_search ON audit.audit_logs USING GIN(
    to_tsvector('english', action || ' ' || COALESCE(error_message, ''))
);

-- Function to automatically create new partitions
CREATE OR REPLACE FUNCTION audit.create_monthly_partition()
RETURNS void AS $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    -- Get next month
    start_date := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '1 month');
    end_date := start_date + INTERVAL '1 month';
    partition_name := 'audit_logs_' || TO_CHAR(start_date, 'YYYY_MM');
    
    -- Create partition if not exists
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS audit.%I PARTITION OF audit.audit_logs
         FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
    
    RAISE NOTICE 'Created partition %', partition_name;
END;
$$ LANGUAGE plpgsql;

-- Scheduled job to create partitions (needs pg_cron extension or external scheduler)
-- For now, manually run: SELECT audit.create_monthly_partition();

