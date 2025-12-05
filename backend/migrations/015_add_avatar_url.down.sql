-- Remove avatar_url column from users table
DROP INDEX IF EXISTS idx_users_avatar_url;
ALTER TABLE auth.users DROP COLUMN IF EXISTS avatar_url;

