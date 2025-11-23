-- Add ACL (Access Control List) system tables
-- Migration: 000005_add_acl_system

-- ACL Policies Table
CREATE TABLE IF NOT EXISTS acl_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type IN ('channel', 'chaincode', 'endpoint', 'all')),
    resource_pattern VARCHAR(255), -- Pattern for matching resources (e.g., 'ibnchannel/*', '/api/v1/transactions/*')
    actions TEXT[] NOT NULL, -- Array of allowed actions: ['read', 'write', 'invoke', 'query', 'admin']
    conditions JSONB DEFAULT '{}', -- Additional conditions (e.g., time-based, IP-based)
    priority INTEGER DEFAULT 0, -- Higher priority policies are checked first
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ACL Permissions Table (Predefined permissions)
CREATE TABLE IF NOT EXISTS acl_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- User Permissions Table (Links users to policies)
CREATE TABLE IF NOT EXISTS user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    policy_id UUID REFERENCES acl_policies(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL, -- Admin who granted this permission
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP, -- Optional expiration
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(user_id, policy_id)
);

-- Role Permissions Table (Links roles to policies for RBAC)
CREATE TABLE IF NOT EXISTS role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'farmer', 'verifier', 'admin')),
    policy_id UUID REFERENCES acl_policies(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(role, policy_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_acl_policies_resource_type ON acl_policies(resource_type);
CREATE INDEX IF NOT EXISTS idx_acl_policies_is_active ON acl_policies(is_active);
CREATE INDEX IF NOT EXISTS idx_acl_policies_priority ON acl_policies(priority DESC);
CREATE INDEX IF NOT EXISTS idx_acl_policies_name ON acl_policies(name);

CREATE INDEX IF NOT EXISTS idx_acl_permissions_resource_type ON acl_permissions(resource_type);
CREATE INDEX IF NOT EXISTS idx_acl_permissions_action ON acl_permissions(action);

CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_policy_id ON user_permissions(policy_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_is_active ON user_permissions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_permissions_expires_at ON user_permissions(expires_at) WHERE expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role);
CREATE INDEX IF NOT EXISTS idx_role_permissions_policy_id ON role_permissions(policy_id);

-- Insert default permissions
INSERT INTO acl_permissions (name, description, resource_type, action) VALUES
    ('channel:read', 'Read channel information', 'channel', 'read'),
    ('channel:write', 'Create or update channels', 'channel', 'write'),
    ('chaincode:invoke', 'Invoke chaincode functions', 'chaincode', 'invoke'),
    ('chaincode:query', 'Query chaincode functions', 'chaincode', 'query'),
    ('chaincode:install', 'Install chaincode', 'chaincode', 'install'),
    ('chaincode:approve', 'Approve chaincode', 'chaincode', 'approve'),
    ('chaincode:commit', 'Commit chaincode', 'chaincode', 'commit'),
    ('endpoint:read', 'Read endpoint data', 'endpoint', 'read'),
    ('endpoint:write', 'Write endpoint data', 'endpoint', 'write'),
    ('endpoint:admin', 'Admin endpoint operations', 'endpoint', 'admin')
ON CONFLICT (name) DO NOTHING;

-- Insert default admin policy (allows all actions)
INSERT INTO acl_policies (name, description, resource_type, resource_pattern, actions, priority, is_active) VALUES
    ('admin:all', 'Admin has access to all resources', 'all', '*', ARRAY['read', 'write', 'invoke', 'query', 'admin'], 100, true)
ON CONFLICT (name) DO NOTHING;

-- Grant admin policy to admin role
INSERT INTO role_permissions (role, policy_id)
SELECT 'admin', id FROM acl_policies WHERE name = 'admin:all'
ON CONFLICT (role, policy_id) DO NOTHING;

