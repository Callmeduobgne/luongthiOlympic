-- RBAC/ABAC Schema Tables
-- Migration: 007_rbac_abac_tables
-- Description: Create tables for Role-Based Access Control (RBAC) and Attribute-Based Access Control (ABAC)

-- Roles table (hierarchical roles)
CREATE TABLE IF NOT EXISTS auth.roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    
    -- Hierarchy support
    parent_role_id UUID REFERENCES auth.roles(id) ON DELETE SET NULL,
    level INTEGER DEFAULT 0, -- Hierarchy level (0 = root)
    
    -- Metadata
    organization_id UUID, -- Optional: organization-specific roles
    is_system_role BOOLEAN DEFAULT FALSE, -- System roles cannot be deleted
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Permissions table (resource-action mapping with ABAC support)
CREATE TABLE IF NOT EXISTS auth.permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Resource-Action pattern
    resource_type VARCHAR(50) NOT NULL,  -- "batch", "transaction", "channel", etc.
    resource_id VARCHAR(255),            -- Specific resource or NULL for all
    action VARCHAR(50) NOT NULL,         -- "read", "write", "delete", "approve", etc.
    
    -- Scope
    scope VARCHAR(50) DEFAULT 'organization', -- "global", "organization", "self", "channel", "public"
    
    -- ABAC Conditions (stored as JSONB)
    conditions JSONB, -- e.g., {"user_attributes": {"certification_level": {"gte": 3}}, "resource_attributes": {"value": {"lte": 10000}}}
    
    -- Effect
    effect VARCHAR(10) DEFAULT 'allow',  -- "allow", "deny"
    priority INTEGER DEFAULT 0,          -- Higher priority takes precedence
    
    -- Metadata
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    -- Unique constraint
    UNIQUE (resource_type, resource_id, action, scope)
);

-- Role-Permissions mapping (RBAC)
CREATE TABLE IF NOT EXISTS auth.role_permissions (
    role_id UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
    
    -- Override options
    effect VARCHAR(10) DEFAULT 'allow', -- Can override permission effect for this role
    
    -- Timestamps
    granted_by UUID REFERENCES auth.users(id),
    granted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (role_id, permission_id)
);

-- User-Roles mapping (RBAC)
CREATE TABLE IF NOT EXISTS auth.user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES auth.roles(id) ON DELETE CASCADE,
    
    -- Scope limitations
    organization_id UUID, -- Optional: limit role to specific organization
    channel_name VARCHAR(100), -- Optional: limit role to specific channel
    
    -- Time-based
    valid_from TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMPTZ, -- Optional expiration
    
    -- Metadata
    granted_by UUID REFERENCES auth.users(id),
    granted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Unique constraint
    UNIQUE (user_id, role_id, organization_id, channel_name)
);

-- User-Permissions mapping (Direct permissions - override roles)
CREATE TABLE IF NOT EXISTS auth.user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES auth.permissions(id) ON DELETE CASCADE,
    
    -- Override
    effect VARCHAR(10) DEFAULT 'allow', -- Can override permission effect
    
    -- Time-based
    valid_from TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMPTZ, -- Optional expiration
    
    -- Metadata
    granted_by UUID REFERENCES auth.users(id),
    granted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Unique constraint
    UNIQUE (user_id, permission_id)
);

-- Indexes for performance
CREATE INDEX idx_roles_name ON auth.roles(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_parent_role_id ON auth.roles(parent_role_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_is_system_role ON auth.roles(is_system_role) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_organization_id ON auth.roles(organization_id) WHERE deleted_at IS NULL AND organization_id IS NOT NULL;

CREATE INDEX idx_permissions_resource_type ON auth.permissions(resource_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_action ON auth.permissions(action) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_scope ON auth.permissions(scope) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_priority ON auth.permissions(priority DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_permissions_effect ON auth.permissions(effect) WHERE deleted_at IS NULL;

CREATE INDEX idx_role_permissions_role_id ON auth.role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON auth.role_permissions(permission_id);

CREATE INDEX idx_user_roles_user_id ON auth.user_roles(user_id) WHERE is_active = TRUE;
CREATE INDEX idx_user_roles_role_id ON auth.user_roles(role_id) WHERE is_active = TRUE;
CREATE INDEX idx_user_roles_organization_id ON auth.user_roles(organization_id) WHERE is_active = TRUE AND organization_id IS NOT NULL;
CREATE INDEX idx_user_roles_valid_until ON auth.user_roles(valid_until) WHERE is_active = TRUE AND valid_until IS NOT NULL;

CREATE INDEX idx_user_permissions_user_id ON auth.user_permissions(user_id) WHERE is_active = TRUE;
CREATE INDEX idx_user_permissions_permission_id ON auth.user_permissions(permission_id) WHERE is_active = TRUE;
CREATE INDEX idx_user_permissions_valid_until ON auth.user_permissions(valid_until) WHERE is_active = TRUE AND valid_until IS NOT NULL;

-- Triggers to update updated_at timestamp
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON auth.roles
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON auth.permissions
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

