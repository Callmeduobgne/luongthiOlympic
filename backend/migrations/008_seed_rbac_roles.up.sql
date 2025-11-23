-- Seed Data: Predefined RBAC Roles and Permissions
-- Migration: 008_seed_rbac_roles
-- Description: Insert system roles and basic permissions

-- Insert System Roles
INSERT INTO auth.roles (name, description, is_system_role, level) VALUES
    ('system:admin', 'System Administrator - Full access to all resources', TRUE, 0),
    ('system:auditor', 'System Auditor - Read-only access for auditing', TRUE, 0),
    ('org:admin', 'Organization Administrator - Full access within organization', TRUE, 1),
    ('org:member', 'Organization Member - Basic member access', TRUE, 1),
    ('supplier', 'Supplier - Can create and manage batches', FALSE, 2),
    ('manufacturer', 'Manufacturer - Can process batches', FALSE, 2),
    ('distributor', 'Distributor - Can ship batches', FALSE, 2),
    ('retailer', 'Retailer - Can sell batches', FALSE, 2),
    ('consumer', 'Consumer - Read-only access to batches', FALSE, 2),
    ('quality:inspector', 'Quality Inspector - Can verify and approve batches', FALSE, 2),
    ('compliance:officer', 'Compliance Officer - Audit access', FALSE, 2),
    ('analyst', 'Analyst - Analytics and reporting access', FALSE, 2)
ON CONFLICT (name) DO NOTHING;

-- Insert Basic Permissions
INSERT INTO auth.permissions (resource_type, resource_id, action, scope, effect, priority, description) VALUES
    -- System Admin Permissions
    ('*', NULL, '*', 'global', 'allow', 100, 'Full access to all resources'),
    
    -- Batch Permissions
    ('batch', NULL, 'create', 'organization', 'allow', 10, 'Create batches'),
    ('batch', NULL, 'read', 'organization', 'allow', 10, 'Read batches in organization'),
    ('batch', NULL, 'read', 'channel', 'allow', 10, 'Read batches in channel'),
    ('batch', NULL, 'read', 'public', 'allow', 10, 'Read public batches'),
    ('batch', NULL, 'update', 'organization', 'allow', 10, 'Update batches in organization'),
    ('batch', NULL, 'update', 'self', 'allow', 10, 'Update own batches'),
    ('batch', NULL, 'delete', 'organization', 'allow', 10, 'Delete batches in organization'),
    ('batch', NULL, 'process', 'organization', 'allow', 10, 'Process batches'),
    ('batch', NULL, 'ship', 'organization', 'allow', 10, 'Ship batches'),
    ('batch', NULL, 'sell', 'organization', 'allow', 10, 'Sell batches'),
    ('batch', NULL, 'verify', 'channel', 'allow', 10, 'Verify batches'),
    ('batch', NULL, 'approve', 'organization', 'allow', 10, 'Approve batches'),
    ('batch', NULL, 'reject', 'organization', 'allow', 10, 'Reject batches'),
    
    -- Transaction Permissions
    ('transaction', NULL, 'read', 'organization', 'allow', 10, 'Read transactions'),
    ('transaction', NULL, 'submit', 'organization', 'allow', 10, 'Submit transactions'),
    ('transaction', NULL, 'query', 'organization', 'allow', 10, 'Query transactions'),
    
    -- Channel Permissions
    ('channel', NULL, 'read', 'global', 'allow', 10, 'Read channel information'),
    ('channel', NULL, 'join', 'global', 'allow', 10, 'Join channels'),
    
    -- Analytics Permissions
    ('analytics', NULL, 'read', 'organization', 'allow', 10, 'Read analytics'),
    ('analytics', NULL, 'query', 'organization', 'allow', 10, 'Query analytics'),
    ('analytics', NULL, 'analyze', 'organization', 'allow', 10, 'Analyze data')
ON CONFLICT (resource_type, resource_id, action, scope) DO NOTHING;

-- Map Roles to Permissions
-- System Admin: All permissions
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'system:admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- System Auditor: Read-only permissions
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'system:auditor'
  AND p.action IN ('read', 'query')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Org Admin: Organization-level permissions
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'org:admin'
  AND p.scope IN ('organization', 'global')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Supplier: Create and manage batches
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'supplier'
  AND p.resource_type = 'batch'
  AND p.action IN ('create', 'read', 'update', 'delete')
  AND p.scope = 'organization'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Manufacturer: Process batches
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'manufacturer'
  AND p.resource_type = 'batch'
  AND p.action IN ('read', 'update', 'process')
  AND p.scope IN ('organization', 'channel')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Distributor: Ship batches
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'distributor'
  AND p.resource_type = 'batch'
  AND p.action IN ('read', 'update', 'ship')
  AND p.scope IN ('organization', 'channel')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Retailer: Sell batches
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'retailer'
  AND p.resource_type = 'batch'
  AND p.action IN ('read', 'sell')
  AND p.scope IN ('organization', 'channel')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Consumer: Read-only access
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'consumer'
  AND p.resource_type = 'batch'
  AND p.action = 'read'
  AND p.scope = 'public'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Quality Inspector: Verify and approve
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'quality:inspector'
  AND p.resource_type = 'batch'
  AND p.action IN ('read', 'verify', 'approve', 'reject')
  AND p.scope IN ('organization', 'channel')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Compliance Officer: Audit access
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'compliance:officer'
  AND p.action IN ('read', 'query')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Analyst: Analytics access
INSERT INTO auth.role_permissions (role_id, permission_id, effect)
SELECT r.id, p.id, 'allow'
FROM auth.roles r, auth.permissions p
WHERE r.name = 'analyst'
  AND p.resource_type IN ('batch', 'transaction', 'analytics')
  AND p.action IN ('read', 'query', 'analyze')
  AND p.scope IN ('organization', 'channel')
ON CONFLICT (role_id, permission_id) DO NOTHING;

