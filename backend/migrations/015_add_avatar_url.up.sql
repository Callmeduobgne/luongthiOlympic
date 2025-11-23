-- Add avatar_url column to users table
ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- Add index for avatar_url (optional, for faster queries if needed)
CREATE INDEX IF NOT EXISTS idx_users_avatar_url ON auth.users(avatar_url) WHERE avatar_url IS NOT NULL;

