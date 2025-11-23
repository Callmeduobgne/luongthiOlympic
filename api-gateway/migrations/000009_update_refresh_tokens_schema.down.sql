-- Rollback refresh_tokens table schema changes

DO $$ 
BEGIN
    -- Rename is_revoked back to revoked if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='refresh_tokens' AND column_name='is_revoked') THEN
        ALTER TABLE refresh_tokens RENAME COLUMN is_revoked TO revoked;
    END IF;
    
    -- Remove revoked_at column if it exists
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='refresh_tokens' AND column_name='revoked_at') THEN
        ALTER TABLE refresh_tokens DROP COLUMN revoked_at;
    END IF;
END $$;

