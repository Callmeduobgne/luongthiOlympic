-- Automated Testing Schema
-- Tracks test results for chaincode before deployment

-- Test suites table
CREATE TABLE blockchain.test_suites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    test_type VARCHAR(50) NOT NULL DEFAULT 'unit' CHECK (test_type IN ('unit', 'integration', 'e2e')),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'passed', 'failed', 'skipped')),
    
    -- Test execution details
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- Test results summary
    total_tests INTEGER DEFAULT 0,
    passed_tests INTEGER DEFAULT 0,
    failed_tests INTEGER DEFAULT 0,
    skipped_tests INTEGER DEFAULT 0,
    
    -- Test output
    output TEXT,
    error_message TEXT,
    error_code VARCHAR(50),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- User tracking
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Test cases table
CREATE TABLE blockchain.test_cases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_suite_id UUID NOT NULL REFERENCES blockchain.test_suites(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    test_function VARCHAR(255), -- Function name being tested
    
    -- Test result
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'passed', 'failed', 'skipped')),
    
    -- Execution details
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    
    -- Test output
    output TEXT,
    error_message TEXT,
    error_stack TEXT,
    assertions JSONB, -- Store assertion results
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Test configurations table
CREATE TABLE blockchain.test_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    
    -- Test settings
    test_type VARCHAR(50) NOT NULL DEFAULT 'unit',
    test_command TEXT, -- Command to run tests
    test_timeout_seconds INTEGER DEFAULT 300,
    required_to_pass BOOLEAN DEFAULT TRUE, -- Block deployment if tests fail
    
    -- Test environment
    environment_vars JSONB DEFAULT '{}',
    test_data JSONB DEFAULT '{}',
    
    -- Chaincode requirements
    min_chaincode_version VARCHAR(50),
    required_functions TEXT[], -- Functions that must be tested
    
    -- Metadata
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Test execution history
CREATE TABLE blockchain.test_execution_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chaincode_version_id UUID NOT NULL REFERENCES blockchain.chaincode_versions(id) ON DELETE CASCADE,
    test_suite_id UUID REFERENCES blockchain.test_suites(id) ON DELETE SET NULL,
    test_configuration_id UUID REFERENCES blockchain.test_configurations(id) ON DELETE SET NULL,
    
    -- Execution context
    execution_type VARCHAR(50) NOT NULL CHECK (execution_type IN ('pre_install', 'pre_approve', 'pre_commit', 'manual')),
    triggered_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    
    -- Result
    overall_status VARCHAR(50) NOT NULL CHECK (overall_status IN ('passed', 'failed', 'skipped', 'error')),
    
    -- Timestamps
    executed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- Indexes for test_suites
CREATE INDEX idx_test_suites_version_id ON blockchain.test_suites(chaincode_version_id);
CREATE INDEX idx_test_suites_status ON blockchain.test_suites(status);
CREATE INDEX idx_test_suites_test_type ON blockchain.test_suites(test_type);
CREATE INDEX idx_test_suites_created_at ON blockchain.test_suites(created_at DESC);
CREATE INDEX idx_test_suites_version_status ON blockchain.test_suites(chaincode_version_id, status);

-- Indexes for test_cases
CREATE INDEX idx_test_cases_suite_id ON blockchain.test_cases(test_suite_id);
CREATE INDEX idx_test_cases_status ON blockchain.test_cases(status);
CREATE INDEX idx_test_cases_function ON blockchain.test_cases(test_function);
CREATE INDEX idx_test_cases_suite_status ON blockchain.test_cases(test_suite_id, status);

-- Indexes for test_configurations
CREATE INDEX idx_test_configurations_name ON blockchain.test_configurations(name);
CREATE INDEX idx_test_configurations_test_type ON blockchain.test_configurations(test_type);
CREATE INDEX idx_test_configurations_is_active ON blockchain.test_configurations(is_active) WHERE is_active = TRUE;

-- Indexes for test_execution_history
CREATE INDEX idx_test_execution_history_version_id ON blockchain.test_execution_history(chaincode_version_id);
CREATE INDEX idx_test_execution_history_suite_id ON blockchain.test_execution_history(test_suite_id);
CREATE INDEX idx_test_execution_history_execution_type ON blockchain.test_execution_history(execution_type);
CREATE INDEX idx_test_execution_history_status ON blockchain.test_execution_history(overall_status);
CREATE INDEX idx_test_execution_history_executed_at ON blockchain.test_execution_history(executed_at DESC);
CREATE INDEX idx_test_execution_history_version_type ON blockchain.test_execution_history(chaincode_version_id, execution_type);

-- Triggers
CREATE TRIGGER update_test_suites_updated_at BEFORE UPDATE ON blockchain.test_suites
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_test_cases_updated_at BEFORE UPDATE ON blockchain.test_cases
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

CREATE TRIGGER update_test_configurations_updated_at BEFORE UPDATE ON blockchain.test_configurations
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

-- Function to calculate test suite duration
CREATE OR REPLACE FUNCTION blockchain.calculate_test_suite_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_ms = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at)) * 1000;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_test_suite_duration_trigger
BEFORE UPDATE ON blockchain.test_suites
FOR EACH ROW
EXECUTE FUNCTION blockchain.calculate_test_suite_duration();

-- Function to calculate test case duration
CREATE OR REPLACE FUNCTION blockchain.calculate_test_case_duration()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.completed_at IS NOT NULL AND NEW.started_at IS NOT NULL THEN
        NEW.duration_ms = EXTRACT(EPOCH FROM (NEW.completed_at - NEW.started_at)) * 1000;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_test_case_duration_trigger
BEFORE UPDATE ON blockchain.test_cases
FOR EACH ROW
EXECUTE FUNCTION blockchain.calculate_test_case_duration();

-- Function to update test suite summary from test cases
CREATE OR REPLACE FUNCTION blockchain.update_test_suite_summary()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE blockchain.test_suites
    SET 
        total_tests = (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id),
        passed_tests = (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id AND status = 'passed'),
        failed_tests = (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id AND status = 'failed'),
        skipped_tests = (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id AND status = 'skipped'),
        status = CASE 
            WHEN (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id AND status = 'failed') > 0 THEN 'failed'
            WHEN (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id AND status = 'running') > 0 THEN 'running'
            WHEN (SELECT COUNT(*) FROM blockchain.test_cases WHERE test_suite_id = NEW.test_suite_id AND status = 'pending') > 0 THEN 'pending'
            ELSE 'passed'
        END
    WHERE id = NEW.test_suite_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_test_suite_summary_trigger
AFTER INSERT OR UPDATE ON blockchain.test_cases
FOR EACH ROW
EXECUTE FUNCTION blockchain.update_test_suite_summary();

-- Default test configuration
INSERT INTO blockchain.test_configurations (name, description, test_type, test_command, required_to_pass, is_active) VALUES
('Default Unit Tests', 'Default unit test configuration for chaincode', 'unit', 'npm test', TRUE, TRUE),
('Default Integration Tests', 'Default integration test configuration', 'integration', 'npm run test:integration', FALSE, TRUE);

