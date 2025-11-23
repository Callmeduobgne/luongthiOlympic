-- Drop automated testing tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_test_suite_summary_trigger ON blockchain.test_cases;
DROP TRIGGER IF EXISTS calculate_test_case_duration_trigger ON blockchain.test_cases;
DROP TRIGGER IF EXISTS calculate_test_suite_duration_trigger ON blockchain.test_suites;
DROP TRIGGER IF EXISTS update_test_configurations_updated_at ON blockchain.test_configurations;
DROP TRIGGER IF EXISTS update_test_cases_updated_at ON blockchain.test_cases;
DROP TRIGGER IF EXISTS update_test_suites_updated_at ON blockchain.test_suites;

-- Drop functions
DROP FUNCTION IF EXISTS blockchain.update_test_suite_summary();
DROP FUNCTION IF EXISTS blockchain.calculate_test_case_duration();
DROP FUNCTION IF EXISTS blockchain.calculate_test_suite_duration();

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS blockchain.test_execution_history CASCADE;
DROP TABLE IF EXISTS blockchain.test_configurations CASCADE;
DROP TABLE IF EXISTS blockchain.test_cases CASCADE;
DROP TABLE IF EXISTS blockchain.test_suites CASCADE;

