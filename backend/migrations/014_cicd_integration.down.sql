-- Drop CI/CD integration tables

-- Drop triggers first
DROP TRIGGER IF EXISTS calculate_cicd_execution_duration_trigger ON blockchain.cicd_executions;
DROP TRIGGER IF EXISTS update_cicd_executions_updated_at ON blockchain.cicd_executions;
DROP TRIGGER IF EXISTS update_cicd_pipelines_updated_at ON blockchain.cicd_pipelines;

-- Drop functions
DROP FUNCTION IF EXISTS blockchain.get_latest_pipeline_execution(UUID);
DROP FUNCTION IF EXISTS blockchain.calculate_cicd_execution_duration();

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS blockchain.cicd_webhook_events CASCADE;
DROP TABLE IF EXISTS blockchain.cicd_artifacts CASCADE;
DROP TABLE IF EXISTS blockchain.cicd_executions CASCADE;
DROP TABLE IF EXISTS blockchain.cicd_pipelines CASCADE;

