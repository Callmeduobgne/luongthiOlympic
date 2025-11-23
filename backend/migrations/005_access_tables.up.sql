-- Access schema tables

-- ACL policies table
CREATE TABLE access.acl_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    resource_type VARCHAR(100) NOT NULL, -- channel, chaincode, transaction, etc.
    resource_id VARCHAR(255), -- Specific resource ID (nullable for wildcard)
    actions TEXT[] NOT NULL, -- Array of allowed actions
    effect VARCHAR(20) NOT NULL DEFAULT 'allow', -- allow, deny
    conditions JSONB, -- Additional conditions (time-based, IP-based, etc.)
    priority INT DEFAULT 0, -- Higher priority takes precedence
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES auth.users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- ACL permissions table (resource-action mapping)
CREATE TABLE access.acl_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id UUID NOT NULL REFERENCES access.acl_policies(id) ON DELETE CASCADE,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    granted BOOLEAN DEFAULT TRUE,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (policy_id, resource, action)
);

-- User permissions table (direct user-policy assignments)
CREATE TABLE access.user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES access.acl_policies(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES auth.users(id),
    granted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, policy_id)
);

-- Role permissions table (role-based access)
CREATE TABLE access.role_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role VARCHAR(50) NOT NULL, -- admin, operator, user, etc.
    policy_id UUID NOT NULL REFERENCES access.acl_policies(id) ON DELETE CASCADE,
    granted_by UUID REFERENCES auth.users(id),
    granted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (role, policy_id)
);

-- Indexes for access schema
CREATE INDEX idx_acl_policies_name ON access.acl_policies(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_acl_policies_resource_type ON access.acl_policies(resource_type) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_acl_policies_is_active ON access.acl_policies(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_acl_policies_priority ON access.acl_policies(priority DESC) WHERE is_active = TRUE AND deleted_at IS NULL;

CREATE INDEX idx_acl_permissions_policy_id ON access.acl_permissions(policy_id);
CREATE INDEX idx_acl_permissions_resource ON access.acl_permissions(resource);
CREATE INDEX idx_acl_permissions_action ON access.acl_permissions(action);

CREATE INDEX idx_user_permissions_user_id ON access.user_permissions(user_id) WHERE is_active = TRUE;
CREATE INDEX idx_user_permissions_policy_id ON access.user_permissions(policy_id) WHERE is_active = TRUE;
CREATE INDEX idx_user_permissions_expires_at ON access.user_permissions(expires_at) WHERE is_active = TRUE AND expires_at IS NOT NULL;

CREATE INDEX idx_role_permissions_role ON access.role_permissions(role) WHERE is_active = TRUE;
CREATE INDEX idx_role_permissions_policy_id ON access.role_permissions(policy_id) WHERE is_active = TRUE;

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_acl_policies_updated_at BEFORE UPDATE ON access.acl_policies
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

