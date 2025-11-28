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

package approval

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

// Repository handles approval workflow data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new approval repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ApprovalRequest represents an approval request
type ApprovalRequest struct {
	ID                uuid.UUID
	ChaincodeVersionID uuid.UUID
	Operation         string
	Status            string
	RequestedBy       uuid.UUID
	RequestedAt       time.Time
	ExpiresAt         *time.Time
	Reason            *string
	Metadata          json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CreateRequest creates a new approval request
func (r *Repository) CreateRequest(ctx context.Context, req *ApprovalRequest) error {
	query := `
		INSERT INTO approval_requests (
			id, chaincode_version_id, operation, status, requested_by,
			requested_at, expires_at, reason, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at, updated_at
	`

	var metadataJSON interface{}
	if req.Metadata != nil {
		metadataJSON = req.Metadata
	}

	err := r.db.QueryRow(ctx, query,
		req.ID, req.ChaincodeVersionID, req.Operation, req.Status, req.RequestedBy,
		req.RequestedAt, req.ExpiresAt, req.Reason, metadataJSON,
	).Scan(&req.CreatedAt, &req.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create approval request: %w", err)
	}

	return nil
}

// GetRequestByID retrieves an approval request by ID
func (r *Repository) GetRequestByID(ctx context.Context, id uuid.UUID) (*ApprovalRequest, error) {
	query := `
		SELECT id, chaincode_version_id, operation, status, requested_by,
		       requested_at, expires_at, reason, metadata, created_at, updated_at
		FROM approval_requests
		WHERE id = $1
	`

	req := &ApprovalRequest{}
	var metadataJSON sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.ChaincodeVersionID, &req.Operation, &req.Status, &req.RequestedBy,
		&req.RequestedAt, &req.ExpiresAt, &req.Reason, &metadataJSON,
		&req.CreatedAt, &req.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("approval request not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval request: %w", err)
	}

	if metadataJSON.Valid {
		req.Metadata = json.RawMessage(metadataJSON.String)
	}

	return req, nil
}

// GetRequestByVersionAndOperation retrieves approval request for a version and operation
func (r *Repository) GetRequestByVersionAndOperation(ctx context.Context, versionID uuid.UUID, operation string) (*ApprovalRequest, error) {
	query := `
		SELECT id, chaincode_version_id, operation, status, requested_by,
		       requested_at, expires_at, reason, metadata, created_at, updated_at
		FROM approval_requests
		WHERE chaincode_version_id = $1 AND operation = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	req := &ApprovalRequest{}
	var metadataJSON sql.NullString

	err := r.db.QueryRow(ctx, query, versionID, operation).Scan(
		&req.ID, &req.ChaincodeVersionID, &req.Operation, &req.Status, &req.RequestedBy,
		&req.RequestedAt, &req.ExpiresAt, &req.Reason, &metadataJSON,
		&req.CreatedAt, &req.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("approval request not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval request: %w", err)
	}

	if metadataJSON.Valid {
		req.Metadata = json.RawMessage(metadataJSON.String)
	}

	return req, nil
}

// ListRequests lists approval requests with filters
func (r *Repository) ListRequests(ctx context.Context, filters *RequestFilters) ([]*ApprovalRequest, error) {
	query := `
		SELECT id, chaincode_version_id, operation, status, requested_by,
		       requested_at, expires_at, reason, metadata, created_at, updated_at
		FROM approval_requests
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.Operation != nil {
		query += fmt.Sprintf(" AND operation = $%d", argPos)
		args = append(args, *filters.Operation)
		argPos++
	}

	if filters.RequestedBy != nil {
		query += fmt.Sprintf(" AND requested_by = $%d", argPos)
		args = append(args, *filters.RequestedBy)
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
		return nil, fmt.Errorf("failed to list approval requests: %w", err)
	}
	defer rows.Close()

	var requests []*ApprovalRequest
	for rows.Next() {
		req := &ApprovalRequest{}
		var metadataJSON sql.NullString

		err := rows.Scan(
			&req.ID, &req.ChaincodeVersionID, &req.Operation, &req.Status, &req.RequestedBy,
			&req.RequestedAt, &req.ExpiresAt, &req.Reason, &metadataJSON,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval request: %w", err)
		}

		if metadataJSON.Valid {
			req.Metadata = json.RawMessage(metadataJSON.String)
		}

		requests = append(requests, req)
	}

	return requests, nil
}

// RequestFilters for querying approval requests
type RequestFilters struct {
	Status      *string
	Operation   *string
	RequestedBy *uuid.UUID
	Limit       int
	Offset      int
}

// ApprovalVote represents an approval vote
type ApprovalVote struct {
	ID                uuid.UUID
	ApprovalRequestID uuid.UUID
	ApproverID        uuid.UUID
	Vote              string
	Comment           *string
	VotedAt           time.Time
}

// CreateVote creates a new approval vote
func (r *Repository) CreateVote(ctx context.Context, vote *ApprovalVote) error {
	query := `
		INSERT INTO approval_votes (
			id, approval_request_id, approver_id, vote, comment, voted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING voted_at
	`

	err := r.db.QueryRow(ctx, query,
		vote.ID, vote.ApprovalRequestID, vote.ApproverID, vote.Vote, vote.Comment, vote.VotedAt,
	).Scan(&vote.VotedAt)

	if err != nil {
		return fmt.Errorf("failed to create approval vote: %w", err)
	}

	return nil
}

// GetVotesByRequest retrieves all votes for an approval request
func (r *Repository) GetVotesByRequest(ctx context.Context, requestID uuid.UUID) ([]*ApprovalVote, error) {
	query := `
		SELECT id, approval_request_id, approver_id, vote, comment, voted_at
		FROM approval_votes
		WHERE approval_request_id = $1
		ORDER BY voted_at DESC
	`

	rows, err := r.db.Query(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval votes: %w", err)
	}
	defer rows.Close()

	var votes []*ApprovalVote
	for rows.Next() {
		vote := &ApprovalVote{}
		err := rows.Scan(
			&vote.ID, &vote.ApprovalRequestID, &vote.ApproverID, &vote.Vote, &vote.Comment, &vote.VotedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval vote: %w", err)
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

// CheckApprovalStatus checks if an approval request is approved (using database function)
func (r *Repository) CheckApprovalStatus(ctx context.Context, requestID uuid.UUID) (bool, error) {
	query := `SELECT check_approval_status($1)`

	var isApproved bool
	err := r.db.QueryRow(ctx, query, requestID).Scan(&isApproved)
	if err != nil {
		return false, fmt.Errorf("failed to check approval status: %w", err)
	}

	return isApproved, nil
}

// ApprovalPolicy represents an approval policy
type ApprovalPolicy struct {
	ID                uuid.UUID
	Operation         string
	RequiredApprovals int
	ExpirationHours   int
	IsActive          bool
	Conditions        json.RawMessage
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GetPolicyByOperation retrieves approval policy for an operation
func (r *Repository) GetPolicyByOperation(ctx context.Context, operation string) (*ApprovalPolicy, error) {
	query := `
		SELECT id, operation, required_approvals, expiration_hours, is_active,
		       conditions, created_at, updated_at
		FROM approval_policies
		WHERE operation = $1 AND is_active = TRUE
	`

	policy := &ApprovalPolicy{}
	var conditionsJSON sql.NullString

	err := r.db.QueryRow(ctx, query, operation).Scan(
		&policy.ID, &policy.Operation, &policy.RequiredApprovals, &policy.ExpirationHours,
		&policy.IsActive, &conditionsJSON, &policy.CreatedAt, &policy.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		// Return default policy
		return &ApprovalPolicy{
			Operation:         operation,
			RequiredApprovals: 1,
			ExpirationHours:   24,
			IsActive:          true,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get approval policy: %w", err)
	}

	if conditionsJSON.Valid {
		policy.Conditions = json.RawMessage(conditionsJSON.String)
	}

	return policy, nil
}

// UpdateRequestStatus updates the status of an approval request
func (r *Repository) UpdateRequestStatus(ctx context.Context, requestID uuid.UUID, status string) error {
	query := `
		UPDATE approval_requests
		SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, requestID)
	if err != nil {
		return fmt.Errorf("failed to update approval request status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("approval request not found")
	}

	return nil
}

