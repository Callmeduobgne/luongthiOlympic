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

package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles CI/CD pipeline business logic
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new CI/CD service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreatePipelineRequest for creating a pipeline
type CreatePipelineRequest struct {
	Name              string
	Description       *string
	ChaincodeName     string
	ChannelName       string
	SourceType        string
	SourceRepository  *string
	SourceBranch      string
	SourcePath        *string
	BuildCommand      *string
	TestCommand       *string
	PackageCommand    *string
	AutoDeploy        bool
	DeployOnTags      bool
	DeployEnvironment string
	WebhookURL        *string
	WebhookSecret     *string
	CreatedBy         *uuid.UUID
}

// CreatePipeline creates a new CI/CD pipeline
func (s *Service) CreatePipeline(ctx context.Context, req *CreatePipelineRequest) (*Pipeline, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("pipeline name is required")
	}
	if req.ChaincodeName == "" {
		return nil, fmt.Errorf("chaincode name is required")
	}
	if req.ChannelName == "" {
		return nil, fmt.Errorf("channel name is required")
	}

	if req.SourceType == "" {
		req.SourceType = "git"
	}
	if req.SourceBranch == "" {
		req.SourceBranch = "main"
	}
	if req.DeployEnvironment == "" {
		req.DeployEnvironment = "production"
	}

	pipeline := &Pipeline{
		ID:                uuid.New(),
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
		IsActive:          true,
		CreatedBy:         req.CreatedBy,
	}

	if err := s.repo.CreatePipeline(ctx, pipeline); err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}

	s.logger.Info("Created CI/CD pipeline",
		zap.String("id", pipeline.ID.String()),
		zap.String("name", req.Name),
		zap.String("chaincode", req.ChaincodeName),
	)

	return pipeline, nil
}

// GetPipeline retrieves a pipeline by ID
func (s *Service) GetPipeline(ctx context.Context, id uuid.UUID) (*Pipeline, error) {
	return s.repo.GetPipelineByID(ctx, id)
}

// ListPipelines lists pipelines with filters
func (s *Service) ListPipelines(ctx context.Context, filters *PipelineFilters) ([]*Pipeline, error) {
	return s.repo.ListPipelines(ctx, filters)
}

// TriggerExecutionRequest for triggering a pipeline execution
type TriggerExecutionRequest struct {
	PipelineID    uuid.UUID
	TriggerType   string // webhook, manual, scheduled, api
	TriggerSource *string
	TriggeredBy   *uuid.UUID
	Metadata      map[string]interface{}
}

// TriggerExecution triggers a pipeline execution
func (s *Service) TriggerExecution(ctx context.Context, req *TriggerExecutionRequest) (*Execution, error) {
	// Get pipeline
	pipeline, err := s.repo.GetPipelineByID(ctx, req.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	if !pipeline.IsActive {
		return nil, fmt.Errorf("pipeline is not active")
	}

	// Create execution
	exec := &Execution{
		ID:            uuid.New(),
		PipelineID:    req.PipelineID,
		TriggerType:   req.TriggerType,
		TriggerSource: req.TriggerSource,
		TriggeredBy:   req.TriggeredBy,
		Status:        "pending",
		BuildStatus:   "pending",
		TestStatus:    "pending",
		PackageStatus: "pending",
		DeployStatus:  "pending",
	}

	if req.Metadata != nil {
		exec.Metadata, _ = json.Marshal(req.Metadata)
	}

	if err := s.repo.CreateExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Start execution asynchronously
	go s.executePipeline(context.Background(), exec.ID, pipeline)

	s.logger.Info("Triggered pipeline execution",
		zap.String("execution_id", exec.ID.String()),
		zap.String("pipeline_id", req.PipelineID.String()),
		zap.String("trigger_type", req.TriggerType),
	)

	return exec, nil
}

// executePipeline executes a pipeline (simulated for now)
func (s *Service) executePipeline(ctx context.Context, executionID uuid.UUID, pipeline *Pipeline) {
	// Update status to running
	if err := s.repo.UpdateExecutionStatus(ctx, executionID, "running", nil, nil, nil, nil); err != nil {
		s.logger.Error("Failed to update execution status", zap.Error(err))
		return
	}

	startTime := time.Now()

	// Stage 1: Build
	if pipeline.BuildCommand != nil {
		s.logger.Info("Executing build stage", zap.String("execution_id", executionID.String()))
		// In production, this would execute the build command
		buildOutput := fmt.Sprintf("Build completed successfully for %s", pipeline.ChaincodeName)
		if err := s.repo.UpdateExecutionStatus(ctx, executionID, "running", strPtr("build"), &buildOutput, nil, nil); err != nil {
			s.logger.Error("Failed to update build status", zap.Error(err))
		}
		time.Sleep(2 * time.Second) // Simulate build time
	}

	// Stage 2: Test
	if pipeline.TestCommand != nil {
		s.logger.Info("Executing test stage", zap.String("execution_id", executionID.String()))
		// In production, this would execute the test command
		testOutput := "All tests passed"
		if err := s.repo.UpdateExecutionStatus(ctx, executionID, "running", strPtr("test"), &testOutput, nil, nil); err != nil {
			s.logger.Error("Failed to update test status", zap.Error(err))
		}
		time.Sleep(1 * time.Second) // Simulate test time
	}

	// Stage 3: Package
	if pipeline.PackageCommand != nil {
		s.logger.Info("Executing package stage", zap.String("execution_id", executionID.String()))
		// In production, this would execute the package command
		packagePath := fmt.Sprintf("/artifacts/%s-%s.tar.gz", pipeline.ChaincodeName, time.Now().Format("20060102-150405"))
		if err := s.repo.UpdateExecutionStatus(ctx, executionID, "running", strPtr("package"), nil, nil, nil); err != nil {
			s.logger.Error("Failed to update package status", zap.Error(err))
		}
		
		// Update package path
		if err := s.repo.UpdateExecutionPackagePath(ctx, executionID, packagePath); err != nil {
			s.logger.Warn("Failed to update package path", zap.Error(err))
		}
		
		time.Sleep(1 * time.Second) // Simulate package time
	}

	// Stage 4: Deploy (if auto_deploy is enabled)
	if pipeline.AutoDeploy {
		s.logger.Info("Executing deploy stage", zap.String("execution_id", executionID.String()))
		// In production, this would trigger deployment via lifecycle service
		// For now, we'll just mark it as success
		if err := s.repo.UpdateExecutionStatus(ctx, executionID, "running", strPtr("deploy"), nil, nil, nil); err != nil {
			s.logger.Error("Failed to update deploy status", zap.Error(err))
		}
		time.Sleep(2 * time.Second) // Simulate deploy time
	}

	// Mark as completed
	durationMs := int(time.Since(startTime).Milliseconds())
	if err := s.repo.UpdateExecutionStatus(ctx, executionID, "success", nil, nil, nil, nil); err != nil {
		s.logger.Error("Failed to update execution status", zap.Error(err))
	}

	s.logger.Info("Pipeline execution completed",
		zap.String("execution_id", executionID.String()),
		zap.Int("duration_ms", durationMs),
	)
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// GetExecution retrieves an execution by ID
func (s *Service) GetExecution(ctx context.Context, id uuid.UUID) (*Execution, error) {
	return s.repo.GetExecutionByID(ctx, id)
}

// ListExecutions lists executions with filters
func (s *Service) ListExecutions(ctx context.Context, filters *ExecutionFilters) ([]*Execution, error) {
	return s.repo.ListExecutions(ctx, filters)
}

// GetArtifacts retrieves artifacts for an execution
func (s *Service) GetArtifacts(ctx context.Context, executionID uuid.UUID) ([]*Artifact, error) {
	return s.repo.GetArtifactsByExecutionID(ctx, executionID)
}

// ProcessWebhookEvent processes a webhook event
func (s *Service) ProcessWebhookEvent(ctx context.Context, pipelineID uuid.UUID, eventType string, payload json.RawMessage, signature *string) (*WebhookEvent, error) {
	// Create webhook event
	event := &WebhookEvent{
		ID:         uuid.New(),
		PipelineID: pipelineID,
		EventType:  eventType,
		Payload:    payload,
		Signature:  signature,
		Processed:  false,
	}

	if err := s.repo.CreateWebhookEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create webhook event: %w", err)
	}

	// Trigger execution if conditions are met
	pipeline, err := s.repo.GetPipelineByID(ctx, pipelineID)
	if err != nil {
		return nil, fmt.Errorf("pipeline not found: %w", err)
	}

	// Check if should trigger (e.g., on push, tag, etc.)
	shouldTrigger := false
	if eventType == "push" && pipeline.SourceBranch != "" {
		shouldTrigger = true
	} else if eventType == "tag" && pipeline.DeployOnTags {
		shouldTrigger = true
	}

	if shouldTrigger {
		triggerReq := &TriggerExecutionRequest{
			PipelineID:    pipelineID,
			TriggerType:   "webhook",
			TriggerSource: strPtr(eventType),
			Metadata: map[string]interface{}{
				"webhook_event_id": event.ID.String(),
			},
		}

		exec, err := s.TriggerExecution(ctx, triggerReq)
		if err != nil {
			errorMsg := err.Error()
			s.repo.MarkWebhookEventProcessed(ctx, event.ID, nil, &errorMsg)
			return nil, fmt.Errorf("failed to trigger execution: %w", err)
		}

		s.repo.MarkWebhookEventProcessed(ctx, event.ID, &exec.ID, nil)
	} else {
		s.repo.MarkWebhookEventProcessed(ctx, event.ID, nil, nil)
	}

	return event, nil
}

