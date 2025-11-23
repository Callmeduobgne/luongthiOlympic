-- Remove account lockout fields
DROP INDEX IF EXISTS idx_users_locked_until;

ALTER TABLE users
DROP COLUMN IF EXISTS failed_login_attempts,
DROP COLUMN IF EXISTS locked_until;

