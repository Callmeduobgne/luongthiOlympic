-- Rollback blocks table
DROP TRIGGER IF EXISTS update_blocks_updated_at ON blocks;
DROP FUNCTION IF EXISTS update_blocks_updated_at();
DROP INDEX IF EXISTS idx_blocks_channel_number;
DROP INDEX IF EXISTS idx_blocks_timestamp;
DROP INDEX IF EXISTS idx_blocks_number;
DROP INDEX IF EXISTS idx_blocks_channel;
DROP TABLE IF EXISTS blocks;

