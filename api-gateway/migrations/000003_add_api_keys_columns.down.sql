-- Remove index
DROP INDEX IF EXISTS idx_api_keys_allowed_ips;

-- Remove columns
ALTER TABLE api_keys 
DROP COLUMN IF EXISTS allowed_ips,
DROP COLUMN IF EXISTS description;

