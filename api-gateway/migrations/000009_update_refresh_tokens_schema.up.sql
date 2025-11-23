-- Update refresh_tokens table schema
-- Change revoked column to is_revoked and add revoked_at column

DO $$ 
BEGIN
    -- Rename revoked to is_revoked if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='refresh_tokens' AND column_name='revoked') THEN
        ALTER TABLE refresh_tokens RENAME COLUMN revoked TO is_revoked;
    END IF;
    
    -- Add revoked_at column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='refresh_tokens' AND column_name='revoked_at') THEN
        ALTER TABLE refresh_tokens ADD COLUMN revoked_at TIMESTAMP;
    END IF;
END $$;

