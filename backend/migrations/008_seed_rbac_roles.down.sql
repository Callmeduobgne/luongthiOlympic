-- Rollback Seed Data: Predefined RBAC Roles and Permissions
-- Migration: 008_seed_rbac_roles

-- Delete role-permission mappings
DELETE FROM auth.role_permissions;

-- Delete permissions
DELETE FROM auth.permissions;

-- Delete roles (only non-system roles if needed, but we'll delete all for rollback)
DELETE FROM auth.roles;

