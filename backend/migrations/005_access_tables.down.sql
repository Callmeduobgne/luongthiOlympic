-- Drop access tables
DROP TRIGGER IF EXISTS update_acl_policies_updated_at ON access.acl_policies;

DROP TABLE IF EXISTS access.role_permissions;
DROP TABLE IF EXISTS access.user_permissions;
DROP TABLE IF EXISTS access.acl_permissions;
DROP TABLE IF EXISTS access.acl_policies;

