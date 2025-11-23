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
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/cicd"
	"go.uber.org/zap"
)

// CICDHandler handles CI/CD pipeline operations
type CICDHandler struct {
	cicdService *cicd.Service
	logger      *zap.Logger
}

// NewCICDHandler creates a new CI/CD handler
func NewCICDHandler(cicdService *cicd.Service, logger *zap.Logger) *CICDHandler {
	return &CICDHandler{
		cicdService: cicdService,
		logger:      logger,
	}
}

// CreatePipelineRequest represents the request body for creating a pipeline
type CreatePipelineRequest struct {
	Name              string                 `json:"name"`
	Description       *string                `json:"description,omitempty"`
	ChaincodeName     string                 `json:"chaincode_name"`
	ChannelName       string                 `json:"channel_name"`
	SourceType        string                 `json:"source_type,omitempty"`
	SourceRepository  *string                `json:"source_repository,omitempty"`
	SourceBranch      string                 `json:"source_branch,omitempty"`
	SourcePath        *string                `json:"source_path,omitempty"`
	BuildCommand      *string                `json:"build_command,omitempty"`
	TestCommand       *string                `json:"test_command,omitempty"`
	PackageCommand    *string                `json:"package_command,omitempty"`
	AutoDeploy        bool                   `json:"auto_deploy,omitempty"`
	DeployOnTags      bool                   `json:"deploy_on_tags,omitempty"`
	DeployEnvironment string                 `json:"deploy_environment,omitempty"`
	WebhookURL        *string                `json:"webhook_url,omitempty"`
	WebhookSecret     *string                `json:"webhook_secret,omitempty"`
}

// CreatePipeline creates a new CI/CD pipeline
func (h *CICDHandler) CreatePipeline(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	var userID *uuid.UUID
	if userIDVal != nil {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	var req CreatePipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.ChaincodeName == "" {
		h.respondError(w, http.StatusBadRequest, "chaincode_name is required")
		return
	}
	if req.ChannelName == "" {
		h.respondError(w, http.StatusBadRequest, "channel_name is required")
		return
	}

	pipelineReq := &cicd.CreatePipelineRequest{
		Name:              req.Name,
		Description:       req.Description,
		ChaincodeName:     req.ChaincodeName,
		ChannelName:       req.ChannelName,
		SourceType:        req.SourceType,
		SourceRepository:  req.SourceRepository,
		SourceBranch:      req.SourceBranch,
		SourcePath:        req.SourcePath,
		BuildCommand:      req.BuildCommand,
		TestCommand:       req.TestCommand,
		PackageCommand:    req.PackageCommand,
		AutoDeploy:        req.AutoDeploy,
		DeployOnTags:      req.DeployOnTags,
		DeployEnvironment: req.DeployEnvironment,
		WebhookURL:        req.WebhookURL,
		WebhookSecret:     req.WebhookSecret,
		CreatedBy:         userID,
	}

	pipeline, err := h.cicdService.CreatePipeline(r.Context(), pipelineReq)
	if err != nil {
		h.logger.Error("Failed to create pipeline", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to create pipeline: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, pipeline)
}

// GetPipeline retrieves a pipeline by ID
func (h *CICDHandler) GetPipeline(w http.ResponseWriter, r *http.Request) {
	pipelineIDStr := chi.URLParam(r, "id")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	pipeline, err := h.cicdService.GetPipeline(r.Context(), pipelineID)
	if err != nil {
		h.logger.Error("Failed to get pipeline", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "pipeline not found")
		return
	}

	h.respondJSON(w, http.StatusOK, pipeline)
}

// ListPipelines lists pipelines with filters
func (h *CICDHandler) ListPipelines(w http.ResponseWriter, r *http.Request) {
	filters := &cicd.PipelineFilters{}

	if chaincodeName := r.URL.Query().Get("chaincode_name"); chaincodeName != "" {
		filters.ChaincodeName = &chaincodeName
	}

	if channelName := r.URL.Query().Get("channel_name"); channelName != "" {
		filters.ChannelName = &channelName
	}

	if isActiveStr := r.URL.Query().Get("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filters.IsActive = &isActive
		}
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

	pipelines, err := h.cicdService.ListPipelines(r.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list pipelines", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to list pipelines: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"pipelines": pipelines,
		"count":    len(pipelines),
	})
}

// TriggerExecutionRequest represents the request body for triggering an execution
type TriggerExecutionRequest struct {
	PipelineID    uuid.UUID              `json:"pipeline_id"`
	TriggerType   string                 `json:"trigger_type,omitempty"`
	TriggerSource *string                `json:"trigger_source,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TriggerExecution triggers a pipeline execution
func (h *CICDHandler) TriggerExecution(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	var userID *uuid.UUID
	if userIDVal != nil {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	var req TriggerExecutionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.PipelineID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "pipeline_id is required")
		return
	}

	if req.TriggerType == "" {
		req.TriggerType = "manual"
	}

	triggerReq := &cicd.TriggerExecutionRequest{
		PipelineID:    req.PipelineID,
		TriggerType:   req.TriggerType,
		TriggerSource: req.TriggerSource,
		TriggeredBy:   userID,
		Metadata:      req.Metadata,
	}

	exec, err := h.cicdService.TriggerExecution(r.Context(), triggerReq)
	if err != nil {
		h.logger.Error("Failed to trigger execution", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to trigger execution: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, exec)
}

// GetExecution retrieves an execution by ID
func (h *CICDHandler) GetExecution(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "id")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	exec, err := h.cicdService.GetExecution(r.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to get execution", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "execution not found")
		return
	}

	h.respondJSON(w, http.StatusOK, exec)
}

// ListExecutions lists executions with filters
func (h *CICDHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	filters := &cicd.ExecutionFilters{}

	if pipelineIDStr := r.URL.Query().Get("pipeline_id"); pipelineIDStr != "" {
		if pipelineID, err := uuid.Parse(pipelineIDStr); err == nil {
			filters.PipelineID = &pipelineID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
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

	executions, err := h.cicdService.ListExecutions(r.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list executions", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to list executions: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"executions": executions,
		"count":     len(executions),
	})
}

// GetArtifacts retrieves artifacts for an execution
func (h *CICDHandler) GetArtifacts(w http.ResponseWriter, r *http.Request) {
	executionIDStr := chi.URLParam(r, "id")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid execution ID")
		return
	}

	artifacts, err := h.cicdService.GetArtifacts(r.Context(), executionID)
	if err != nil {
		h.logger.Error("Failed to get artifacts", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get artifacts: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"artifacts": artifacts,
		"count":    len(artifacts),
	})
}

// ProcessWebhook processes a webhook event
func (h *CICDHandler) ProcessWebhook(w http.ResponseWriter, r *http.Request) {
	pipelineIDStr := chi.URLParam(r, "pipeline_id")
	pipelineID, err := uuid.Parse(pipelineIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid pipeline ID")
		return
	}

	// Get event type from header or query
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		eventType = r.URL.Query().Get("event_type")
	}
	if eventType == "" {
		eventType = "push" // Default
	}

	// Get signature from header
	signature := r.Header.Get("X-Hub-Signature-256")
	var sigPtr *string
	if signature != "" {
		sigPtr = &signature
	}

	// Read payload
	var payload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid payload: "+err.Error())
		return
	}

	event, err := h.cicdService.ProcessWebhookEvent(r.Context(), pipelineID, eventType, payload, sigPtr)
	if err != nil {
		h.logger.Error("Failed to process webhook", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to process webhook: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, event)
}

// Helper methods
func (h *CICDHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *CICDHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

