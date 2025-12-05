-- Drop chaincode registry tables

-- Drop triggers first
DROP TRIGGER IF EXISTS sync_active_chaincode_trigger ON blockchain.chaincode_versions;
DROP TRIGGER IF EXISTS calculate_deployment_duration_trigger ON blockchain.deployment_logs;
DROP TRIGGER IF EXISTS update_active_chaincodes_updated_at ON blockchain.active_chaincodes;
DROP TRIGGER IF EXISTS update_chaincode_versions_updated_at ON blockchain.chaincode_versions;

-- Drop functions
DROP FUNCTION IF EXISTS blockchain.sync_active_chaincode();
DROP FUNCTION IF EXISTS blockchain.calculate_deployment_duration();

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS blockchain.active_chaincodes CASCADE;
DROP TABLE IF EXISTS blockchain.deployment_logs CASCADE;
DROP TABLE IF EXISTS blockchain.chaincode_versions CASCADE;

