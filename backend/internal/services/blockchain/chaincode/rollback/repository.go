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

package rollback

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

// Repository handles rollback operations data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new rollback repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// RollbackOperation represents a rollback operation
type RollbackOperation struct {
	ID            uuid.UUID
	ChaincodeName string
	ChannelName   string
	FromVersionID uuid.UUID
	ToVersionID   uuid.UUID
	FromVersion   string
	ToVersion     string
	FromSequence  int
	ToSequence    int
	Status        string
	Reason        *string
	RollbackType  string
	StartedAt     *time.Time
	CompletedAt   *time.Time
	DurationMs    *int
	RequestedBy   uuid.UUID
	ExecutedBy    *uuid.UUID
	ErrorMessage  *string
	ErrorCode     *string
	Metadata      json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateRollbackOperation creates a new rollback operation
func (r *Repository) CreateRollbackOperation(ctx context.Context, op *RollbackOperation) error {
	query := `
		INSERT INTO blockchain.rollback_operations (
			id, chaincode_name, channel_name, from_version_id, to_version_id,
			from_version, to_version, from_sequence, to_sequence,
			status, reason, rollback_type, requested_by, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
		RETURNING created_at, updated_at
	`

	var metadataJSON interface{}
	if op.Metadata != nil {
		metadataJSON = op.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		op.ID, op.ChaincodeName, op.ChannelName, op.FromVersionID, op.ToVersionID,
		op.FromVersion, op.ToVersion, op.FromSequence, op.ToSequence,
		op.Status, op.Reason, op.RollbackType, op.RequestedBy, metadataJSON,
	).Scan(&op.CreatedAt, &op.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create rollback operation: %w", err)
	}

	return nil
}

// GetRollbackOperationByID retrieves a rollback operation by ID
func (r *Repository) GetRollbackOperationByID(ctx context.Context, id uuid.UUID) (*RollbackOperation, error) {
	query := `
		SELECT id, chaincode_name, channel_name, from_version_id, to_version_id,
		       from_version, to_version, from_sequence, to_sequence,
		       status, reason, rollback_type, started_at, completed_at, duration_ms,
		       requested_by, executed_by, error_message, error_code, metadata,
		       created_at, updated_at
		FROM blockchain.rollback_operations
		WHERE id = $1
	`

	op := &RollbackOperation{}
	var metadataJSON sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&op.ID, &op.ChaincodeName, &op.ChannelName, &op.FromVersionID, &op.ToVersionID,
		&op.FromVersion, &op.ToVersion, &op.FromSequence, &op.ToSequence,
		&op.Status, &op.Reason, &op.RollbackType, &op.StartedAt, &op.CompletedAt, &op.DurationMs,
		&op.RequestedBy, &op.ExecutedBy, &op.ErrorMessage, &op.ErrorCode, &metadataJSON,
		&op.CreatedAt, &op.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("rollback operation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback operation: %w", err)
	}

	if metadataJSON.Valid {
		op.Metadata = json.RawMessage(metadataJSON.String)
	}

	return op, nil
}

// UpdateRollbackStatus updates the status of a rollback operation
func (r *Repository) UpdateRollbackStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string, errorCode *string, executedBy *uuid.UUID) error {
	query := `
		UPDATE blockchain.rollback_operations
		SET status = $2, error_message = $3, error_code = $4, executed_by = $5,
		    started_at = CASE WHEN $2 = 'in_progress' AND started_at IS NULL THEN CURRENT_TIMESTAMP ELSE started_at END,
		    completed_at = CASE WHEN $2 IN ('completed', 'failed', 'cancelled') THEN CURRENT_TIMESTAMP ELSE completed_at END,
		    duration_ms = CASE WHEN $2 IN ('completed', 'failed', 'cancelled') AND started_at IS NOT NULL 
		                      THEN EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - started_at))::INTEGER * 1000 
		                      ELSE duration_ms END
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status, errorMsg, errorCode, executedBy)
	if err != nil {
		return fmt.Errorf("failed to update rollback status: %w", err)
	}

	return nil
}

// ListRollbackOperations lists rollback operations with filters
func (r *Repository) ListRollbackOperations(ctx context.Context, filters *RollbackFilters) ([]*RollbackOperation, error) {
	query := `
		SELECT id, chaincode_name, channel_name, from_version_id, to_version_id,
		       from_version, to_version, from_sequence, to_sequence,
		       status, reason, rollback_type, started_at, completed_at, duration_ms,
		       requested_by, executed_by, error_message, error_code, metadata,
		       created_at, updated_at
		FROM blockchain.rollback_operations
		WHERE 1=1
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
		return nil, fmt.Errorf("failed to list rollback operations: %w", err)
	}
	defer rows.Close()

	var operations []*RollbackOperation
	for rows.Next() {
		op := &RollbackOperation{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&op.ID, &op.ChaincodeName, &op.ChannelName, &op.FromVersionID, &op.ToVersionID,
			&op.FromVersion, &op.ToVersion, &op.FromSequence, &op.ToSequence,
			&op.Status, &op.Reason, &op.RollbackType, &op.StartedAt, &op.CompletedAt, &op.DurationMs,
			&op.RequestedBy, &op.ExecutedBy, &op.ErrorMessage, &op.ErrorCode, &metadataJSON,
			&op.CreatedAt, &op.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollback operation: %w", err)
		}

		if metadataJSON.Valid {
			op.Metadata = json.RawMessage(metadataJSON.String)
		}

		operations = append(operations, op)
	}

	return operations, nil
}

// RollbackFilters for querying rollback operations
type RollbackFilters struct {
	ChaincodeName *string
	ChannelName   *string
	Status        *string
	Limit         int
	Offset        int
}

// GetPreviousActiveVersion gets the previous active version for rollback
func (r *Repository) GetPreviousActiveVersion(ctx context.Context, chaincodeName, channelName string, currentSequence int) (*uuid.UUID, error) {
	query := `SELECT blockchain.get_previous_active_version($1, $2, $3)`

	var versionID uuid.UUID
	err := r.db.QueryRow(ctx, query, chaincodeName, channelName, currentSequence).Scan(&versionID)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("no previous version found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get previous version: %w", err)
	}

	return &versionID, nil
}

// IsRollbackSafe checks if rollback is safe (no pending operations)
func (r *Repository) IsRollbackSafe(ctx context.Context, chaincodeName, channelName string) (bool, error) {
	query := `SELECT blockchain.is_rollback_safe($1, $2)`

	var isSafe bool
	err := r.db.QueryRow(ctx, query, chaincodeName, channelName).Scan(&isSafe)
	if err != nil {
		return false, fmt.Errorf("failed to check rollback safety: %w", err)
	}

	return isSafe, nil
}

// CreateRollbackHistory creates a rollback history entry
func (r *Repository) CreateRollbackHistory(ctx context.Context, operationID, versionID uuid.UUID, operation, previousStatus, newStatus string, details json.RawMessage) error {
	query := `
		INSERT INTO blockchain.rollback_history (
			rollback_operation_id, chaincode_version_id, operation,
			previous_status, new_status, details
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	var detailsJSON interface{}
	if details != nil {
		detailsJSON = details
	}

	_, err := r.db.Exec(ctx, query, operationID, versionID, operation, previousStatus, newStatus, detailsJSON)
	if err != nil {
		return fmt.Errorf("failed to create rollback history: %w", err)
	}

	return nil
}

// GetRollbackHistory retrieves rollback history for an operation
func (r *Repository) GetRollbackHistory(ctx context.Context, operationID uuid.UUID) ([]*RollbackHistory, error) {
	query := `
		SELECT id, rollback_operation_id, chaincode_version_id, operation,
		       previous_status, new_status, details, created_at
		FROM blockchain.rollback_history
		WHERE rollback_operation_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rollback history: %w", err)
	}
	defer rows.Close()

	var history []*RollbackHistory
	for rows.Next() {
		h := &RollbackHistory{}
		var detailsJSON sql.NullString

		err := rows.Scan(
			&h.ID, &h.RollbackOperationID, &h.ChaincodeVersionID, &h.Operation,
			&h.PreviousStatus, &h.NewStatus, &detailsJSON, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollback history: %w", err)
		}

		if detailsJSON.Valid {
			h.Details = json.RawMessage(detailsJSON.String)
		}

		history = append(history, h)
	}

	return history, nil
}

// RollbackHistory represents a rollback history entry
type RollbackHistory struct {
	ID                uuid.UUID
	RollbackOperationID uuid.UUID
	ChaincodeVersionID uuid.UUID
	Operation          string
	PreviousStatus     *string
	NewStatus          *string
	Details            json.RawMessage
	CreatedAt          time.Time
}

