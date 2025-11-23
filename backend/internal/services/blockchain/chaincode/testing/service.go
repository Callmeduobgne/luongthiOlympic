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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles test execution business logic
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new testing service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// RunTestSuiteRequest for running a test suite
type RunTestSuiteRequest struct {
	ChaincodeVersionID uuid.UUID
	TestType           string // unit, integration, e2e
	TestCommand        *string // Optional: override default command
	TriggeredBy        *uuid.UUID
	Metadata           map[string]interface{}
}

// RunTestSuite runs a test suite for a chaincode version
func (s *Service) RunTestSuite(ctx context.Context, req *RunTestSuiteRequest) (*TestSuite, error) {
	// Get test configuration
	config, err := s.repo.GetActiveTestConfiguration(ctx, req.TestType)
	if err != nil {
		return nil, fmt.Errorf("failed to get test configuration: %w", err)
	}

	// Use provided command or default from config
	testCommand := req.TestCommand
	if testCommand == nil && config.TestCommand != nil {
		testCommand = config.TestCommand
	}
	if testCommand == nil {
		return nil, fmt.Errorf("test command is required")
	}

	// Create test suite
	suite := &TestSuite{
		ID:                uuid.New(),
		ChaincodeVersionID: req.ChaincodeVersionID,
		Name:              fmt.Sprintf("%s tests for version %s", req.TestType, req.ChaincodeVersionID.String()),
		TestType:          req.TestType,
		Status:            "pending",
		CreatedBy:         req.TriggeredBy,
	}

	if req.Metadata != nil {
		suite.Metadata, _ = json.Marshal(req.Metadata)
	}

	if err := s.repo.CreateTestSuite(ctx, suite); err != nil {
		return nil, fmt.Errorf("failed to create test suite: %w", err)
	}

	// Update status to running
	if err := s.repo.UpdateTestSuiteStatus(ctx, suite.ID, "running", nil, nil); err != nil {
		return nil, fmt.Errorf("failed to update test suite status: %w", err)
	}

	// Execute tests (async - in production, this would execute actual test commands)
	// For now, we'll simulate test execution
	go s.executeTests(ctx, suite.ID, *testCommand, config, req.ChaincodeVersionID)

	s.logger.Info("Test suite started",
		zap.String("id", suite.ID.String()),
		zap.String("type", req.TestType),
		zap.String("version_id", req.ChaincodeVersionID.String()),
	)

	return suite, nil
}

// executeTests executes the test command (simulated for now)
func (s *Service) executeTests(ctx context.Context, suiteID uuid.UUID, testCommand string, config *TestConfiguration, versionID uuid.UUID) {
	// In production, this would:
	// 1. Execute the test command (e.g., npm test, npm run test:integration)
	// 2. Parse test output
	// 3. Create test cases from results
	// 4. Update test suite status

	// For now, simulate test execution
	time.Sleep(2 * time.Second) // Simulate test execution time

	// Create mock test cases
	testCases := []struct {
		name   string
		status string
	}{
		{"Test function 1", "passed"},
		{"Test function 2", "passed"},
		{"Test function 3", "failed"},
	}

	for _, tc := range testCases {
		testCase := &TestCase{
			ID:          uuid.New(),
			TestSuiteID: suiteID,
			Name:        tc.name,
			Status:      tc.status,
		}

		if err := s.repo.CreateTestCase(ctx, testCase); err != nil {
			s.logger.Error("Failed to create test case", zap.Error(err))
			continue
		}

		// Update test case status
		output := fmt.Sprintf("Test %s completed", tc.name)
		var errorMsg *string
		if tc.status == "failed" {
			msg := "Test assertion failed"
			errorMsg = &msg
		}

		if err := s.repo.UpdateTestCaseStatus(ctx, testCase.ID, tc.status, &output, errorMsg, nil, nil); err != nil {
			s.logger.Error("Failed to update test case status", zap.Error(err))
		}
	}

	// Determine overall suite status
	hasFailed := false
	for _, tc := range testCases {
		if tc.status == "failed" {
			hasFailed = true
			break
		}
	}

	finalStatus := "passed"
	if hasFailed {
		finalStatus = "failed"
	}

	// Update test suite status
	if err := s.repo.UpdateTestSuiteStatus(ctx, suiteID, finalStatus, nil, nil); err != nil {
		s.logger.Error("Failed to update test suite status", zap.Error(err))
	}

	// Create execution history
	historyMetadata, _ := json.Marshal(map[string]interface{}{
		"test_command": testCommand,
		"config_id":    config.ID.String(),
	})

	if err := s.repo.CreateTestExecutionHistory(ctx, versionID, &suiteID, &config.ID, "manual", finalStatus, nil, historyMetadata); err != nil {
		s.logger.Warn("Failed to create test execution history", zap.Error(err))
	}

	s.logger.Info("Test suite completed",
		zap.String("id", suiteID.String()),
		zap.String("status", finalStatus),
	)
}

// GetTestSuite retrieves a test suite by ID
func (s *Service) GetTestSuite(ctx context.Context, id uuid.UUID) (*TestSuite, error) {
	return s.repo.GetTestSuiteByID(ctx, id)
}

// ListTestSuites lists test suites with filters
func (s *Service) ListTestSuites(ctx context.Context, filters *TestSuiteFilters) ([]*TestSuite, error) {
	return s.repo.ListTestSuites(ctx, filters)
}

// GetTestCases retrieves test cases for a suite
func (s *Service) GetTestCases(ctx context.Context, suiteID uuid.UUID) ([]*TestCase, error) {
	return s.repo.GetTestCasesBySuiteID(ctx, suiteID)
}

// CheckTestsPassed checks if tests passed for a chaincode version
func (s *Service) CheckTestsPassed(ctx context.Context, versionID uuid.UUID, testType string) (bool, error) {
	filters := &TestSuiteFilters{
		ChaincodeVersionID: &versionID,
		TestType:           &testType,
		Status:             nil, // Get all statuses
		Limit:              1,
	}

	suites, err := s.repo.ListTestSuites(ctx, filters)
	if err != nil {
		return false, fmt.Errorf("failed to check tests: %w", err)
	}

	if len(suites) == 0 {
		// No tests run yet
		return false, fmt.Errorf("no tests have been run for this version")
	}

	// Get the most recent test suite
	latestSuite := suites[0]

	// Check if tests passed
	if latestSuite.Status == "passed" {
		return true, nil
	}

	return false, fmt.Errorf("tests did not pass (status: %s)", latestSuite.Status)
}

// GetLatestTestSuite gets the latest test suite for a version
func (s *Service) GetLatestTestSuite(ctx context.Context, versionID uuid.UUID, testType string) (*TestSuite, error) {
	filters := &TestSuiteFilters{
		ChaincodeVersionID: &versionID,
		TestType:           &testType,
		Limit:              1,
	}

	suites, err := s.repo.ListTestSuites(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest test suite: %w", err)
	}

	if len(suites) == 0 {
		return nil, fmt.Errorf("no test suite found")
	}

	return suites[0], nil
}

