// Copyright 2024 IBN Network (ICTU Blockchain Network)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles test data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new testing repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// TestSuite represents a test suite
type TestSuite struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	Name              string
	Description       *string
	TestType          string
	Status            string
	StartedAt         *time.Time
	CompletedAt       *time.Time
	DurationMs        *int
	TotalTests        int
	PassedTests       int
	FailedTests       int
	SkippedTests      int
	Output            *string
	ErrorMessage      *string
	ErrorCode         *string
	Metadata          json.RawMessage
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// TestCase represents a test case
type TestCase struct {
	ID          uuid.UUID
	TestSuiteID uuid.UUID
	Name        string
	Description *string
	TestFunction *string
	Status      string
	StartedAt   *time.Time
	CompletedAt *time.Time
	DurationMs  *int
	Output      *string
	ErrorMessage *string
	ErrorStack  *string
	Assertions  json.RawMessage
	Metadata    json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TestConfiguration represents a test configuration
type TestConfiguration struct {
	ID                  uuid.UUID
	Name                string
	Description         *string
	TestType            string
	TestCommand         *string
	TestTimeoutSeconds  int
	RequiredToPass      bool
	EnvironmentVars     json.RawMessage
	TestData            json.RawMessage
	MinChaincodeVersion *string
	RequiredFunctions   []string
	IsActive            bool
	CreatedBy           *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

// CreateTestSuite creates a new test suite
func (r *Repository) CreateTestSuite(ctx context.Context, suite *TestSuite) error {
	query := `
		INSERT INTO blockchain.test_suites (
			id, chaincode_version_id, name, description, test_type,
			status, metadata, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING created_at, updated_at
	`

	var metadataJSON interface{}
	if suite.Metadata != nil {
		metadataJSON = suite.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		suite.ID, suite.ChaincodeVersionID, suite.Name, suite.Description,
		suite.TestType, suite.Status, metadataJSON, suite.CreatedBy,
	).Scan(&suite.CreatedAt, &suite.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create test suite: %w", err)
	}

	return nil
}

// GetTestSuiteByID retrieves a test suite by ID
func (r *Repository) GetTestSuiteByID(ctx context.Context, id uuid.UUID) (*TestSuite, error) {
	query := `
		SELECT id, chaincode_version_id, name, description, test_type,
		       status, started_at, completed_at, duration_ms,
		       total_tests, passed_tests, failed_tests, skipped_tests,
		       output, error_message, error_code, metadata,
		       created_by, created_at, updated_at
		FROM blockchain.test_suites
		WHERE id = $1
	`

	suite := &TestSuite{}
	var metadataJSON sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&suite.ID, &suite.ChaincodeVersionID, &suite.Name, &suite.Description, &suite.TestType,
		&suite.Status, &suite.StartedAt, &suite.CompletedAt, &suite.DurationMs,
		&suite.TotalTests, &suite.PassedTests, &suite.FailedTests, &suite.SkippedTests,
		&suite.Output, &suite.ErrorMessage, &suite.ErrorCode, &metadataJSON,
		&suite.CreatedBy, &suite.CreatedAt, &suite.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("test suite not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get test suite: %w", err)
	}

	if metadataJSON.Valid {
		suite.Metadata = json.RawMessage(metadataJSON.String)
	}

	return suite, nil
}

// UpdateTestSuiteStatus updates the status of a test suite
func (r *Repository) UpdateTestSuiteStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string, errorCode *string) error {
	query := `
		UPDATE blockchain.test_suites
		SET status = $2, error_message = $3, error_code = $4,
		    started_at = CASE WHEN $2 = 'running' AND started_at IS NULL THEN CURRENT_TIMESTAMP ELSE started_at END,
		    completed_at = CASE WHEN $2 IN ('passed', 'failed', 'skipped') THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, errorMsg, errorCode)
	if err != nil {
		return fmt.Errorf("failed to update test suite status: %w", err)
	}

	return nil
}

// ListTestSuites lists test suites with filters
func (r *Repository) ListTestSuites(ctx context.Context, filters *TestSuiteFilters) ([]*TestSuite, error) {
	query := `
		SELECT id, chaincode_version_id, name, description, test_type,
		       status, started_at, completed_at, duration_ms,
		       total_tests, passed_tests, failed_tests, skipped_tests,
		       output, error_message, error_code, metadata,
		       created_by, created_at, updated_at
		FROM blockchain.test_suites
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filters.ChaincodeVersionID != nil {
		query += fmt.Sprintf(" AND chaincode_version_id = $%d", argPos)
		args = append(args, *filters.ChaincodeVersionID)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.TestType != nil {
		query += fmt.Sprintf(" AND test_type = $%d", argPos)
		args = append(args, *filters.TestType)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filters.Offset)
		argPos++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list test suites: %w", err)
	}
	defer rows.Close()

	var suites []*TestSuite
	for rows.Next() {
		suite := &TestSuite{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&suite.ID, &suite.ChaincodeVersionID, &suite.Name, &suite.Description, &suite.TestType,
			&suite.Status, &suite.StartedAt, &suite.CompletedAt, &suite.DurationMs,
			&suite.TotalTests, &suite.PassedTests, &suite.FailedTests, &suite.SkippedTests,
			&suite.Output, &suite.ErrorMessage, &suite.ErrorCode, &metadataJSON,
			&suite.CreatedBy, &suite.CreatedAt, &suite.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test suite: %w", err)
		}

		if metadataJSON.Valid {
			suite.Metadata = json.RawMessage(metadataJSON.String)
		}

		suites = append(suites, suite)
	}

	return suites, nil
}

// TestSuiteFilters for querying test suites
type TestSuiteFilters struct {
	ChaincodeVersionID *uuid.UUID
	Status              *string
	TestType            *string
	Limit               int
	Offset              int
}

// CreateTestCase creates a new test case
func (r *Repository) CreateTestCase(ctx context.Context, testCase *TestCase) error {
	query := `
		INSERT INTO blockchain.test_cases (
			id, test_suite_id, name, description, test_function,
			status, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at, updated_at
	`

	var metadataJSON interface{}
	if testCase.Metadata != nil {
		metadataJSON = testCase.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		testCase.ID, testCase.TestSuiteID, testCase.Name, testCase.Description,
		testCase.TestFunction, testCase.Status, metadataJSON,
	).Scan(&testCase.CreatedAt, &testCase.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create test case: %w", err)
	}

	return nil
}

// UpdateTestCaseStatus updates the status of a test case
func (r *Repository) UpdateTestCaseStatus(ctx context.Context, id uuid.UUID, status string, output *string, errorMsg *string, errorStack *string, assertions json.RawMessage) error {
	query := `
		UPDATE blockchain.test_cases
		SET status = $2, output = $3, error_message = $4, error_stack = $5, assertions = $6,
		    started_at = CASE WHEN $2 = 'running' AND started_at IS NULL THEN CURRENT_TIMESTAMP ELSE started_at END,
		    completed_at = CASE WHEN $2 IN ('passed', 'failed', 'skipped') THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = $1
	`

	var assertionsJSON interface{}
	if assertions != nil {
		assertionsJSON = assertions
	}

	_, err := r.db.Exec(ctx, query, id, status, output, errorMsg, errorStack, assertionsJSON)
	if err != nil {
		return fmt.Errorf("failed to update test case status: %w", err)
	}

	return nil
}

// GetTestCasesBySuiteID retrieves test cases for a suite
func (r *Repository) GetTestCasesBySuiteID(ctx context.Context, suiteID uuid.UUID) ([]*TestCase, error) {
	query := `
		SELECT id, test_suite_id, name, description, test_function,
		       status, started_at, completed_at, duration_ms,
		       output, error_message, error_stack, assertions, metadata,
		       created_at, updated_at
		FROM blockchain.test_cases
		WHERE test_suite_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, suiteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}
	defer rows.Close()

	var testCases []*TestCase
	for rows.Next() {
		tc := &TestCase{}
		var assertionsJSON sql.NullString
		var metadataJSON sql.NullString

		err := rows.Scan(
			&tc.ID, &tc.TestSuiteID, &tc.Name, &tc.Description, &tc.TestFunction,
			&tc.Status, &tc.StartedAt, &tc.CompletedAt, &tc.DurationMs,
			&tc.Output, &tc.ErrorMessage, &tc.ErrorStack, &assertionsJSON, &metadataJSON,
			&tc.CreatedAt, &tc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test case: %w", err)
		}

		if assertionsJSON.Valid {
			tc.Assertions = json.RawMessage(assertionsJSON.String)
		}
		if metadataJSON.Valid {
			tc.Metadata = json.RawMessage(metadataJSON.String)
		}

		testCases = append(testCases, tc)
	}

	return testCases, nil
}

// GetActiveTestConfiguration retrieves active test configuration by type
func (r *Repository) GetActiveTestConfiguration(ctx context.Context, testType string) (*TestConfiguration, error) {
	query := `
		SELECT id, name, description, test_type, test_command, test_timeout_seconds,
		       required_to_pass, environment_vars, test_data,
		       min_chaincode_version, required_functions, is_active,
		       created_by, created_at, updated_at, deleted_at
		FROM blockchain.test_configurations
		WHERE test_type = $1 AND is_active = TRUE AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	config := &TestConfiguration{}
	var envVarsJSON sql.NullString
	var testDataJSON sql.NullString
	var requiredFunctions []string

	err := r.db.QueryRow(ctx, query, testType).Scan(
		&config.ID, &config.Name, &config.Description, &config.TestType, &config.TestCommand,
		&config.TestTimeoutSeconds, &config.RequiredToPass, &envVarsJSON, &testDataJSON,
		&config.MinChaincodeVersion, &requiredFunctions, &config.IsActive,
		&config.CreatedBy, &config.CreatedAt, &config.UpdatedAt, &config.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no active test configuration found for type: %s", testType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get test configuration: %w", err)
	}

	if envVarsJSON.Valid {
		config.EnvironmentVars = json.RawMessage(envVarsJSON.String)
	}
	if testDataJSON.Valid {
		config.TestData = json.RawMessage(testDataJSON.String)
	}
	config.RequiredFunctions = requiredFunctions

	return config, nil
}

// CreateTestExecutionHistory creates a test execution history entry
func (r *Repository) CreateTestExecutionHistory(ctx context.Context, versionID uuid.UUID, suiteID *uuid.UUID, configID *uuid.UUID, executionType string, status string, triggeredBy *uuid.UUID, metadata json.RawMessage) error {
	query := `
		INSERT INTO blockchain.test_execution_history (
			chaincode_version_id, test_suite_id, test_configuration_id,
			execution_type, overall_status, triggered_by, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`

	var metadataJSON interface{}
	if metadata != nil {
		metadataJSON = metadata
	}

	_, err := r.db.Exec(ctx, query, versionID, suiteID, configID, executionType, status, triggeredBy, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to create test execution history: %w", err)
	}

	return nil
}

