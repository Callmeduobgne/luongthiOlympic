-- Drop rollback mechanisms tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_rollback_operations_updated_at ON blockchain.rollback_operations;

-- Drop functions
DROP FUNCTION IF EXISTS blockchain.is_rollback_safe(VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS blockchain.get_previous_active_version(VARCHAR, VARCHAR, INTEGER);

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS blockchain.rollback_history CASCADE;
DROP TABLE IF EXISTS blockchain.rollback_operations CASCADE;

