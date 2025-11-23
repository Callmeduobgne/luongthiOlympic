-- Rollback ACL system migration
-- Migration: 000005_add_acl_system

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_permissions;
DROP TABLE IF EXISTS acl_permissions;
DROP TABLE IF EXISTS acl_policies;

