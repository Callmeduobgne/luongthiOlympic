-- Add additional columns to api_keys if needed
ALTER TABLE api_keys 
ADD COLUMN IF NOT EXISTS description TEXT,
ADD COLUMN IF NOT EXISTS allowed_ips INET[];

-- Create index for IP filtering
CREATE INDEX IF NOT EXISTS idx_api_keys_allowed_ips ON api_keys USING GIN(allowed_ips);

