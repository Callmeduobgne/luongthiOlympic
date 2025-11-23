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

package registry

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

// Repository handles chaincode registry data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new chaincode registry repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ChaincodeVersion represents a chaincode version record
type ChaincodeVersion struct {
	ID                uuid.UUID
	Name              string
	Version           string
	Sequence          int
	PackageID         *string
	Label             *string
	Path              *string
	PackagePath       *string
	ChannelName       string
	InstallStatus     string
	ApproveStatus     string
	CommitStatus      string
	InitRequired      bool
	EndorsementPlugin string
	ValidationPlugin  string
	Collections       json.RawMessage
	InstalledAt       *time.Time
	ApprovedAt        *time.Time
	CommittedAt       *time.Time
	InstalledBy       *uuid.UUID
	ApprovedBy        *uuid.UUID
	CommittedBy       *uuid.UUID
	InstallError      *string
	ApproveError      *string
	CommitError       *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// CreateVersion creates a new chaincode version record
func (r *Repository) CreateVersion(ctx context.Context, version *ChaincodeVersion) error {
	query := `
		INSERT INTO blockchain.chaincode_versions (
			id, name, version, sequence, package_id, label, path, package_path,
			channel_name, install_status, approve_status, commit_status,
			init_required, endorsement_plugin, validation_plugin, collections,
			installed_by, approved_by, committed_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		RETURNING created_at, updated_at
	`

	var collectionsJSON interface{}
	if version.Collections != nil {
		collectionsJSON = version.Collections
	}

	err := r.db.QueryRow(ctx, query,
		version.ID, version.Name, version.Version, version.Sequence,
		version.PackageID, version.Label, version.Path, version.PackagePath,
		version.ChannelName, version.InstallStatus, version.ApproveStatus, version.CommitStatus,
		version.InitRequired, version.EndorsementPlugin, version.ValidationPlugin, collectionsJSON,
		version.InstalledBy, version.ApprovedBy, version.CommittedBy,
	).Scan(&version.CreatedAt, &version.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create chaincode version: %w", err)
	}

	return nil
}

// GetVersionByID retrieves a chaincode version by ID
func (r *Repository) GetVersionByID(ctx context.Context, id uuid.UUID) (*ChaincodeVersion, error) {
	query := `
		SELECT id, name, version, sequence, package_id, label, path, package_path,
		       channel_name, install_status, approve_status, commit_status,
		       init_required, endorsement_plugin, validation_plugin, collections,
		       installed_at, approved_at, committed_at,
		       installed_by, approved_by, committed_by,
		       install_error, approve_error, commit_error,
		       created_at, updated_at, deleted_at
		FROM blockchain.chaincode_versions
		WHERE id = $1 AND deleted_at IS NULL
	`

	version := &ChaincodeVersion{}
	var collectionsJSON sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&version.ID, &version.Name, &version.Version, &version.Sequence,
		&version.PackageID, &version.Label, &version.Path, &version.PackagePath,
		&version.ChannelName, &version.InstallStatus, &version.ApproveStatus, &version.CommitStatus,
		&version.InitRequired, &version.EndorsementPlugin, &version.ValidationPlugin, &collectionsJSON,
		&version.InstalledAt, &version.ApprovedAt, &version.CommittedAt,
		&version.InstalledBy, &version.ApprovedBy, &version.CommittedBy,
		&version.InstallError, &version.ApproveError, &version.CommitError,
		&version.CreatedAt, &version.UpdatedAt, &version.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("chaincode version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chaincode version: %w", err)
	}

	if collectionsJSON.Valid {
		version.Collections = json.RawMessage(collectionsJSON.String)
	}

	return version, nil
}

// UpdateVersionStatus updates the status of a chaincode version
func (r *Repository) UpdateVersionStatus(ctx context.Context, id uuid.UUID, operation string, status string, errorMsg *string, performedBy *uuid.UUID) error {
	var query string
	var args []interface{}

	switch operation {
	case "install":
		query = `
			UPDATE blockchain.chaincode_versions
			SET install_status = $2, install_error = $3, installed_by = $4, installed_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		args = []interface{}{id, status, errorMsg, performedBy}
	case "approve":
		query = `
			UPDATE blockchain.chaincode_versions
			SET approve_status = $2, approve_error = $3, approved_by = $4, approved_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		args = []interface{}{id, status, errorMsg, performedBy}
	case "commit":
		query = `
			UPDATE blockchain.chaincode_versions
			SET commit_status = $2, commit_error = $3, committed_by = $4, committed_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		args = []interface{}{id, status, errorMsg, performedBy}
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update version status: %w", err)
	}

	return nil
}

// ListVersions lists chaincode versions with filters
func (r *Repository) ListVersions(ctx context.Context, filters *VersionFilters) ([]*ChaincodeVersion, error) {
	query := `
		SELECT id, name, version, sequence, package_id, label, path, package_path,
		       channel_name, install_status, approve_status, commit_status,
		       init_required, endorsement_plugin, validation_plugin, collections,
		       installed_at, approved_at, committed_at,
		       installed_by, approved_by, committed_by,
		       install_error, approve_error, commit_error,
		       created_at, updated_at, deleted_at
		FROM blockchain.chaincode_versions
		WHERE deleted_at IS NULL
	`

	args := []interface{}{}
	argPos := 1

	if filters.Name != nil {
		query += fmt.Sprintf(" AND name = $%d", argPos)
		args = append(args, *filters.Name)
		argPos++
	}

	if filters.ChannelName != nil {
		query += fmt.Sprintf(" AND channel_name = $%d", argPos)
		args = append(args, *filters.ChannelName)
		argPos++
	}

	if filters.CommitStatus != nil {
		query += fmt.Sprintf(" AND commit_status = $%d", argPos)
		args = append(args, *filters.CommitStatus)
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
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}
	defer rows.Close()

	var versions []*ChaincodeVersion
	for rows.Next() {
		version := &ChaincodeVersion{}
		var collectionsJSON sql.NullString

		err := rows.Scan(
			&version.ID, &version.Name, &version.Version, &version.Sequence,
			&version.PackageID, &version.Label, &version.Path, &version.PackagePath,
			&version.ChannelName, &version.InstallStatus, &version.ApproveStatus, &version.CommitStatus,
			&version.InitRequired, &version.EndorsementPlugin, &version.ValidationPlugin, &collectionsJSON,
			&version.InstalledAt, &version.ApprovedAt, &version.CommittedAt,
			&version.InstalledBy, &version.ApprovedBy, &version.CommittedBy,
			&version.InstallError, &version.ApproveError, &version.CommitError,
			&version.CreatedAt, &version.UpdatedAt, &version.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}

		if collectionsJSON.Valid {
			version.Collections = json.RawMessage(collectionsJSON.String)
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// VersionFilters for querying versions
type VersionFilters struct {
	Name        *string
	ChannelName *string
	CommitStatus *string
	Limit       int
	Offset      int
}

// DeploymentLog represents a deployment operation log
type DeploymentLog struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	Operation         string
	Status            string
	RequestData       json.RawMessage
	ResponseData      json.RawMessage
	ErrorMessage      *string
	ErrorCode         *string
	StartedAt         *time.Time
	CompletedAt       *time.Time
	DurationMs        *int
	PerformedBy       *uuid.UUID
	IPAddress         *string
	UserAgent         *string
	CreatedAt         time.Time
}

// CreateDeploymentLog creates a new deployment log
func (r *Repository) CreateDeploymentLog(ctx context.Context, log *DeploymentLog) error {
	query := `
		INSERT INTO blockchain.deployment_logs (
			id, chaincode_version_id, operation, status, request_data, response_data,
			error_message, error_code, started_at, completed_at, duration_ms,
			performed_by, ip_address, user_agent
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		RETURNING created_at
	`

	var requestData, responseData interface{}
	if log.RequestData != nil {
		requestData = log.RequestData
	}
	if log.ResponseData != nil {
		responseData = log.ResponseData
	}

	err := r.db.QueryRow(ctx, query,
		log.ID, log.ChaincodeVersionID, log.Operation, log.Status,
		requestData, responseData,
		log.ErrorMessage, log.ErrorCode, log.StartedAt, log.CompletedAt, log.DurationMs,
		log.PerformedBy, log.IPAddress, log.UserAgent,
	).Scan(&log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create deployment log: %w", err)
	}

	return nil
}

// UpdateDeploymentLog updates a deployment log
func (r *Repository) UpdateDeploymentLog(ctx context.Context, id uuid.UUID, status string, responseData json.RawMessage, errorMsg *string, errorCode *string) error {
	query := `
		UPDATE blockchain.deployment_logs
		SET status = $2, response_data = $3, error_message = $4, error_code = $5,
		    completed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	var responseDataJSON interface{}
	if responseData != nil {
		responseDataJSON = responseData
	}

	_, err := r.db.Exec(ctx, query, id, status, responseDataJSON, errorMsg, errorCode)
	if err != nil {
		return fmt.Errorf("failed to update deployment log: %w", err)
	}

	return nil
}

// GetDeploymentLogs retrieves deployment logs for a chaincode version
func (r *Repository) GetDeploymentLogs(ctx context.Context, versionID uuid.UUID, limit int) ([]*DeploymentLog, error) {
	query := `
		SELECT id, chaincode_version_id, operation, status, request_data, response_data,
		       error_message, error_code, started_at, completed_at, duration_ms,
		       performed_by, ip_address, user_agent, created_at
		FROM blockchain.deployment_logs
		WHERE chaincode_version_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, versionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment logs: %w", err)
	}
	defer rows.Close()

	var logs []*DeploymentLog
	for rows.Next() {
		log := &DeploymentLog{}
		var requestDataJSON, responseDataJSON sql.NullString

		err := rows.Scan(
			&log.ID, &log.ChaincodeVersionID, &log.Operation, &log.Status,
			&requestDataJSON, &responseDataJSON,
			&log.ErrorMessage, &log.ErrorCode, &log.StartedAt, &log.CompletedAt, &log.DurationMs,
			&log.PerformedBy, &log.IPAddress, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment log: %w", err)
		}

		if requestDataJSON.Valid {
			log.RequestData = json.RawMessage(requestDataJSON.String)
		}
		if responseDataJSON.Valid {
			log.ResponseData = json.RawMessage(responseDataJSON.String)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// ActiveChaincode represents an active chaincode on a channel
type ActiveChaincode struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	Name              string
	Version           string
	Sequence          int
	ChannelName       string
	PackageID         *string
	InitRequired      bool
	EndorsementPlugin *string
	ValidationPlugin  *string
	Collections       json.RawMessage
	IsActive          bool
	ActivatedAt       time.Time
	DeactivatedAt     *time.Time
	Metadata          json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GetActiveChaincodes retrieves active chaincodes for a channel
func (r *Repository) GetActiveChaincodes(ctx context.Context, channelName *string) ([]*ActiveChaincode, error) {
	query := `
		SELECT id, chaincode_version_id, name, version, sequence, channel_name,
		       package_id, init_required, endorsement_plugin, validation_plugin,
		       collections, is_active, activated_at, deactivated_at, metadata,
		       created_at, updated_at
		FROM blockchain.active_chaincodes
		WHERE is_active = TRUE
	`

	args := []interface{}{}
	if channelName != nil {
		query += " AND channel_name = $1"
		args = append(args, *channelName)
	}

	query += " ORDER BY activated_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get active chaincodes: %w", err)
	}
	defer rows.Close()

	var chaincodes []*ActiveChaincode
	for rows.Next() {
		cc := &ActiveChaincode{}
		var collectionsJSON, metadataJSON sql.NullString

		err := rows.Scan(
			&cc.ID, &cc.ChaincodeVersionID, &cc.Name, &cc.Version, &cc.Sequence, &cc.ChannelName,
			&cc.PackageID, &cc.InitRequired, &cc.EndorsementPlugin, &cc.ValidationPlugin,
			&collectionsJSON, &cc.IsActive, &cc.ActivatedAt, &cc.DeactivatedAt, &metadataJSON,
			&cc.CreatedAt, &cc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan active chaincode: %w", err)
		}

		if collectionsJSON.Valid {
			cc.Collections = json.RawMessage(collectionsJSON.String)
		}
		if metadataJSON.Valid {
			cc.Metadata = json.RawMessage(metadataJSON.String)
		}

		chaincodes = append(chaincodes, cc)
	}

	return chaincodes, nil
}

