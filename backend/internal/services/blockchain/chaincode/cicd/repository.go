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
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles CI/CD data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new CI/CD repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID                uuid.UUID
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
	IsActive          bool
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// Execution represents a pipeline execution
type Execution struct {
	ID            uuid.UUID
	PipelineID    uuid.UUID
	TriggerType   string
	TriggerSource *string
	TriggeredBy   *uuid.UUID
	Status        string
	BuildStatus   string
	TestStatus    string
	PackageStatus string
	DeployStatus  string
	StartedAt     *time.Time
	CompletedAt   *time.Time
	DurationMs    *int
	BuildOutput   *string
	TestOutput    *string
	PackagePath   *string
	DeploymentID  *uuid.UUID
	ErrorMessage  *string
	ErrorStage    *string
	Metadata      json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Artifact represents a build artifact
type Artifact struct {
	ID          uuid.UUID
	ExecutionID uuid.UUID
	ArtifactType string
	ArtifactPath string
	ArtifactSize *int64
	MimeType    *string
	Metadata    json.RawMessage
	CreatedAt   time.Time
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID          uuid.UUID
	PipelineID  uuid.UUID
	EventType   string
	Payload     json.RawMessage
	Signature   *string
	Processed   bool
	ExecutionID *uuid.UUID
	ErrorMessage *string
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// CreatePipeline creates a new CI/CD pipeline
func (r *Repository) CreatePipeline(ctx context.Context, pipeline *Pipeline) error {
	query := `
		INSERT INTO cicd_pipelines (
			id, name, description, chaincode_name, channel_name,
			source_type, source_repository, source_branch, source_path,
			build_command, test_command, package_command,
			auto_deploy, deploy_on_tags, deploy_environment,
			webhook_url, webhook_secret, is_active, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		pipeline.ID, pipeline.Name, pipeline.Description, pipeline.ChaincodeName, pipeline.ChannelName,
		pipeline.SourceType, pipeline.SourceRepository, pipeline.SourceBranch, pipeline.SourcePath,
		pipeline.BuildCommand, pipeline.TestCommand, pipeline.PackageCommand,
		pipeline.AutoDeploy, pipeline.DeployOnTags, pipeline.DeployEnvironment,
		pipeline.WebhookURL, pipeline.WebhookSecret, pipeline.IsActive, pipeline.CreatedBy,
	).Scan(&pipeline.CreatedAt, &pipeline.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	return nil
}

// GetPipelineByID retrieves a pipeline by ID
func (r *Repository) GetPipelineByID(ctx context.Context, id uuid.UUID) (*Pipeline, error) {
	query := `
		SELECT id, name, description, chaincode_name, channel_name,
		       source_type, source_repository, source_branch, source_path,
		       build_command, test_command, package_command,
		       auto_deploy, deploy_on_tags, deploy_environment,
		       webhook_url, webhook_secret, is_active, created_by,
		       created_at, updated_at, deleted_at
		FROM cicd_pipelines
		WHERE id = $1 AND deleted_at IS NULL
	`

	pipeline := &Pipeline{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pipeline.ID, &pipeline.Name, &pipeline.Description, &pipeline.ChaincodeName, &pipeline.ChannelName,
		&pipeline.SourceType, &pipeline.SourceRepository, &pipeline.SourceBranch, &pipeline.SourcePath,
		&pipeline.BuildCommand, &pipeline.TestCommand, &pipeline.PackageCommand,
		&pipeline.AutoDeploy, &pipeline.DeployOnTags, &pipeline.DeployEnvironment,
		&pipeline.WebhookURL, &pipeline.WebhookSecret, &pipeline.IsActive, &pipeline.CreatedBy,
		&pipeline.CreatedAt, &pipeline.UpdatedAt, &pipeline.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("pipeline not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}

	return pipeline, nil
}

// ListPipelines lists pipelines with filters
func (r *Repository) ListPipelines(ctx context.Context, filters *PipelineFilters) ([]*Pipeline, error) {
	query := `
		SELECT id, name, description, chaincode_name, channel_name,
		       source_type, source_repository, source_branch, source_path,
		       build_command, test_command, package_command,
		       auto_deploy, deploy_on_tags, deploy_environment,
		       webhook_url, webhook_secret, is_active, created_by,
		       created_at, updated_at, deleted_at
		FROM cicd_pipelines
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argPos := 1

	if filters.ChaincodeName != nil {
		query += fmt.Sprintf(" AND chaincode_name = $%d", argPos)
		args = append(args, *filters.ChaincodeName)
		argPos++
	}

	if filters.ChannelName != nil {
		query += fmt.Sprintf(" AND channel_name = $%d", argPos)
		args = append(args, *filters.ChannelName)
		argPos++
	}

	if filters.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filters.IsActive)
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
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	defer rows.Close()

	var pipelines []*Pipeline
	for rows.Next() {
		p := &Pipeline{}
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ChaincodeName, &p.ChannelName,
			&p.SourceType, &p.SourceRepository, &p.SourceBranch, &p.SourcePath,
			&p.BuildCommand, &p.TestCommand, &p.PackageCommand,
			&p.AutoDeploy, &p.DeployOnTags, &p.DeployEnvironment,
			&p.WebhookURL, &p.WebhookSecret, &p.IsActive, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pipeline: %w", err)
		}
		pipelines = append(pipelines, p)
	}

	return pipelines, nil
}

// PipelineFilters for querying pipelines
type PipelineFilters struct {
	ChaincodeName *string
	ChannelName   *string
	IsActive      *bool
	Limit         int
	Offset        int
}

// CreateExecution creates a new pipeline execution
func (r *Repository) CreateExecution(ctx context.Context, exec *Execution) error {
	query := `
		INSERT INTO cicd_executions (
			id, pipeline_id, trigger_type, trigger_source, triggered_by,
			status, build_status, test_status, package_status, deploy_status,
			metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		RETURNING created_at, updated_at
	`

	var metadataJSON interface{}
	if exec.Metadata != nil {
		metadataJSON = exec.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		exec.ID, exec.PipelineID, exec.TriggerType, exec.TriggerSource, exec.TriggeredBy,
		exec.Status, exec.BuildStatus, exec.TestStatus, exec.PackageStatus, exec.DeployStatus,
		metadataJSON,
	).Scan(&exec.CreatedAt, &exec.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	return nil
}

// UpdateExecutionStatus updates execution status
func (r *Repository) UpdateExecutionStatus(ctx context.Context, id uuid.UUID, status string, stage *string, output *string, errorMsg *string, errorStage *string) error {
	query := `
		UPDATE cicd_executions
		SET status = $2,
		    build_status = CASE WHEN $3 = 'build' THEN $2 ELSE build_status END,
		    test_status = CASE WHEN $3 = 'test' THEN $2 ELSE test_status END,
		    package_status = CASE WHEN $3 = 'package' THEN $2 ELSE package_status END,
		    deploy_status = CASE WHEN $3 = 'deploy' THEN $2 ELSE deploy_status END,
		    build_output = CASE WHEN $3 = 'build' THEN $4 ELSE build_output END,
		    test_output = CASE WHEN $3 = 'test' THEN $4 ELSE test_output END,
		    error_message = $5,
		    error_stage = $6,
		    started_at = CASE WHEN $2 = 'running' AND started_at IS NULL THEN CURRENT_TIMESTAMP ELSE started_at END,
		    completed_at = CASE WHEN $2 IN ('success', 'failed', 'cancelled') THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, stage, output, errorMsg, errorStage)
	if err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	return nil
}

// GetExecutionByID retrieves an execution by ID
func (r *Repository) GetExecutionByID(ctx context.Context, id uuid.UUID) (*Execution, error) {
	query := `
		SELECT id, pipeline_id, trigger_type, trigger_source, triggered_by,
		       status, build_status, test_status, package_status, deploy_status,
		       started_at, completed_at, duration_ms,
		       build_output, test_output, package_path, deployment_id,
		       error_message, error_stage, metadata,
		       created_at, updated_at
		FROM cicd_executions
		WHERE id = $1
	`

	exec := &Execution{}
	var metadataJSON sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&exec.ID, &exec.PipelineID, &exec.TriggerType, &exec.TriggerSource, &exec.TriggeredBy,
		&exec.Status, &exec.BuildStatus, &exec.TestStatus, &exec.PackageStatus, &exec.DeployStatus,
		&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs,
		&exec.BuildOutput, &exec.TestOutput, &exec.PackagePath, &exec.DeploymentID,
		&exec.ErrorMessage, &exec.ErrorStage, &metadataJSON,
		&exec.CreatedAt, &exec.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("execution not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	if metadataJSON.Valid {
		exec.Metadata = json.RawMessage(metadataJSON.String)
	}

	return exec, nil
}

// ListExecutions lists executions with filters
func (r *Repository) ListExecutions(ctx context.Context, filters *ExecutionFilters) ([]*Execution, error) {
	query := `
		SELECT id, pipeline_id, trigger_type, trigger_source, triggered_by,
		       status, build_status, test_status, package_status, deploy_status,
		       started_at, completed_at, duration_ms,
		       build_output, test_output, package_path, deployment_id,
		       error_message, error_stage, metadata,
		       created_at, updated_at
		FROM cicd_executions
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filters.PipelineID != nil {
		query += fmt.Sprintf(" AND pipeline_id = $%d", argPos)
		args = append(args, *filters.PipelineID)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
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
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}
	defer rows.Close()

	var executions []*Execution
	for rows.Next() {
		exec := &Execution{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&exec.ID, &exec.PipelineID, &exec.TriggerType, &exec.TriggerSource, &exec.TriggeredBy,
			&exec.Status, &exec.BuildStatus, &exec.TestStatus, &exec.PackageStatus, &exec.DeployStatus,
			&exec.StartedAt, &exec.CompletedAt, &exec.DurationMs,
			&exec.BuildOutput, &exec.TestOutput, &exec.PackagePath, &exec.DeploymentID,
			&exec.ErrorMessage, &exec.ErrorStage, &metadataJSON,
			&exec.CreatedAt, &exec.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		if metadataJSON.Valid {
			exec.Metadata = json.RawMessage(metadataJSON.String)
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// ExecutionFilters for querying executions
type ExecutionFilters struct {
	PipelineID *uuid.UUID
	Status     *string
	Limit      int
	Offset     int
}

// CreateArtifact creates a new artifact
func (r *Repository) CreateArtifact(ctx context.Context, artifact *Artifact) error {
	query := `
		INSERT INTO cicd_artifacts (
			id, execution_id, artifact_type, artifact_path,
			artifact_size, mime_type, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at
	`

	var metadataJSON interface{}
	if artifact.Metadata != nil {
		metadataJSON = artifact.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		artifact.ID, artifact.ExecutionID, artifact.ArtifactType, artifact.ArtifactPath,
		artifact.ArtifactSize, artifact.MimeType, metadataJSON,
	).Scan(&artifact.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create artifact: %w", err)
	}

	return nil
}

// GetArtifactsByExecutionID retrieves artifacts for an execution
func (r *Repository) GetArtifactsByExecutionID(ctx context.Context, executionID uuid.UUID) ([]*Artifact, error) {
	query := `
		SELECT id, execution_id, artifact_type, artifact_path,
		       artifact_size, mime_type, metadata, created_at
		FROM cicd_artifacts
		WHERE execution_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []*Artifact
	for rows.Next() {
		art := &Artifact{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&art.ID, &art.ExecutionID, &art.ArtifactType, &art.ArtifactPath,
			&art.ArtifactSize, &art.MimeType, &metadataJSON, &art.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan artifact: %w", err)
		}

		if metadataJSON.Valid {
			art.Metadata = json.RawMessage(metadataJSON.String)
		}

		artifacts = append(artifacts, art)
	}

	return artifacts, nil
}

// CreateWebhookEvent creates a new webhook event
func (r *Repository) CreateWebhookEvent(ctx context.Context, event *WebhookEvent) error {
	query := `
		INSERT INTO cicd_webhook_events (
			id, pipeline_id, event_type, payload, signature
		) VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING created_at
	`

	var payloadJSON interface{}
	if event.Payload != nil {
		payloadJSON = event.Payload
	}

	err := r.db.QueryRow(ctx, query,
		event.ID, event.PipelineID, event.EventType, payloadJSON, event.Signature,
	).Scan(&event.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create webhook event: %w", err)
	}

	return nil
}

// GetUnprocessedWebhookEvents retrieves unprocessed webhook events
func (r *Repository) GetUnprocessedWebhookEvents(ctx context.Context, limit int) ([]*WebhookEvent, error) {
	query := `
		SELECT id, pipeline_id, event_type, payload, signature,
		       processed, execution_id, error_message, created_at, processed_at
		FROM cicd_webhook_events
		WHERE processed = FALSE
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook events: %w", err)
	}
	defer rows.Close()

	var events []*WebhookEvent
	for rows.Next() {
		event := &WebhookEvent{}
		var payloadJSON sql.NullString

		err := rows.Scan(
			&event.ID, &event.PipelineID, &event.EventType, &payloadJSON, &event.Signature,
			&event.Processed, &event.ExecutionID, &event.ErrorMessage, &event.CreatedAt, &event.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook event: %w", err)
		}

		if payloadJSON.Valid {
			event.Payload = json.RawMessage(payloadJSON.String)
		}

		events = append(events, event)
	}

	return events, nil
}

// MarkWebhookEventProcessed marks a webhook event as processed
func (r *Repository) MarkWebhookEventProcessed(ctx context.Context, eventID uuid.UUID, executionID *uuid.UUID, errorMsg *string) error {
	query := `
		UPDATE cicd_webhook_events
		SET processed = TRUE, execution_id = $2, error_message = $3, processed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, eventID, executionID, errorMsg)
	if err != nil {
		return fmt.Errorf("failed to mark webhook event as processed: %w", err)
	}

	return nil
}

// UpdateExecutionPackagePath updates the package path for an execution
func (r *Repository) UpdateExecutionPackagePath(ctx context.Context, executionID uuid.UUID, packagePath string) error {
	query := `UPDATE cicd_executions SET package_path = $1 WHERE id = $2`

	_, err := r.db.Exec(ctx, query, packagePath, executionID)
	if err != nil {
		return fmt.Errorf("failed to update package path: %w", err)
	}

	return nil
}

