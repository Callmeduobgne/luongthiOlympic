-- Drop auth tables
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON auth.api_keys;
DROP TRIGGER IF EXISTS update_users_updated_at ON auth.users;
DROP FUNCTION IF EXISTS auth.update_updated_at_column();

DROP TABLE IF EXISTS auth.refresh_tokens;
DROP TABLE IF EXISTS auth.api_keys;
DROP TABLE IF EXISTS auth.users;

