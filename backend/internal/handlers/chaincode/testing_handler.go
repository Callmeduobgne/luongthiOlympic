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

package chaincode

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/testing"
	"go.uber.org/zap"
)

// TestingHandler handles test operations
type TestingHandler struct {
	testingService *testing.Service
	logger         *zap.Logger
}

// NewTestingHandler creates a new testing handler
func NewTestingHandler(testingService *testing.Service, logger *zap.Logger) *TestingHandler {
	return &TestingHandler{
		testingService: testingService,
		logger:         logger,
	}
}

// RunTestSuiteRequest represents the request body for running tests
type RunTestSuiteRequest struct {
	ChaincodeVersionID uuid.UUID             `json:"chaincode_version_id"`
	TestType           string                 `json:"test_type"` // unit, integration, e2e
	TestCommand        *string                `json:"test_command,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// RunTestSuite runs a test suite
func (h *TestingHandler) RunTestSuite(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	var userID *uuid.UUID
	if userIDVal != nil {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Parse request body
	var req RunTestSuiteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Validate required fields
	if req.ChaincodeVersionID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "chaincode_version_id is required")
		return
	}
	if req.TestType == "" {
		h.respondError(w, http.StatusBadRequest, "test_type is required")
		return
	}

	// Run test suite
	testReq := &testing.RunTestSuiteRequest{
		ChaincodeVersionID: req.ChaincodeVersionID,
		TestType:           req.TestType,
		TestCommand:        req.TestCommand,
		TriggeredBy:        userID,
		Metadata:           req.Metadata,
	}

	suite, err := h.testingService.RunTestSuite(r.Context(), testReq)
	if err != nil {
		h.logger.Error("Failed to run test suite", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to run test suite: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, suite)
}

// GetTestSuite retrieves a test suite by ID
func (h *TestingHandler) GetTestSuite(w http.ResponseWriter, r *http.Request) {
	suiteIDStr := chi.URLParam(r, "id")
	suiteID, err := uuid.Parse(suiteIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid suite ID")
		return
	}

	suite, err := h.testingService.GetTestSuite(r.Context(), suiteID)
	if err != nil {
		h.logger.Error("Failed to get test suite", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "test suite not found")
		return
	}

	h.respondJSON(w, http.StatusOK, suite)
}

// ListTestSuites lists test suites with filters
func (h *TestingHandler) ListTestSuites(w http.ResponseWriter, r *http.Request) {
	filters := &testing.TestSuiteFilters{}

	if versionIDStr := r.URL.Query().Get("chaincode_version_id"); versionIDStr != "" {
		if versionID, err := uuid.Parse(versionIDStr); err == nil {
			filters.ChaincodeVersionID = &versionID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	if testType := r.URL.Query().Get("test_type"); testType != "" {
		filters.TestType = &testType
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 50
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	suites, err := h.testingService.ListTestSuites(r.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list test suites", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to list test suites: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"suites": suites,
		"count":  len(suites),
	})
}

// GetTestCases retrieves test cases for a suite
func (h *TestingHandler) GetTestCases(w http.ResponseWriter, r *http.Request) {
	suiteIDStr := chi.URLParam(r, "id")
	suiteID, err := uuid.Parse(suiteIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid suite ID")
		return
	}

	testCases, err := h.testingService.GetTestCases(r.Context(), suiteID)
	if err != nil {
		h.logger.Error("Failed to get test cases", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get test cases: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"test_cases": testCases,
		"count":     len(testCases),
	})
}

// Helper methods
func (h *TestingHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TestingHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

