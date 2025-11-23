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

package audit

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles audit data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new audit repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateLog creates a new audit log entry
func (r *Repository) CreateLog(ctx context.Context, log *Log) error {
	query := `
		INSERT INTO audit.audit_logs 
		(id, user_id, action, resource_type, resource_id, status, ip_address, user_agent,
		 request_id, method, path, duration_ms, error_message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
		log.Status, log.IPAddress, log.UserAgent, log.RequestID,
		log.Method, log.Path, log.DurationMs, log.ErrorMessage,
		log.Metadata, log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// BatchCreateLogs creates multiple audit log entries
func (r *Repository) BatchCreateLogs(ctx context.Context, logs []*Log) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO audit.audit_logs 
		(id, user_id, action, resource_type, resource_id, status, ip_address, user_agent,
		 request_id, method, path, duration_ms, error_message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	for _, log := range logs {
		_, err := tx.Exec(ctx, query,
			log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID,
			log.Status, log.IPAddress, log.UserAgent, log.RequestID,
			log.Method, log.Path, log.DurationMs, log.ErrorMessage,
			log.Metadata, log.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert audit log: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// QueryLogs queries audit logs with filters
func (r *Repository) QueryLogs(ctx context.Context, req *QueryLogsRequest) ([]*Log, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, status,
		       ip_address, user_agent, request_id, method, path,
		       duration_ms, error_message, metadata, created_at
		FROM audit.audit_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if req.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *req.UserID)
		argPos++
	}

	if req.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, *req.Action)
		argPos++
	}

	if req.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argPos)
		args = append(args, *req.ResourceType)
		argPos++
	}

	if req.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, *req.Status)
		argPos++
	}

	if req.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, *req.StartDate)
		argPos++
	}

	if req.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, *req.EndDate)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if req.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, req.Limit)
		argPos++
	}

	if req.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, req.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		log := &Log{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.Status, &log.IPAddress, &log.UserAgent,
			&log.RequestID, &log.Method, &log.Path, &log.DurationMs,
			&log.ErrorMessage, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// SearchLogs performs full-text search on audit logs
func (r *Repository) SearchLogs(ctx context.Context, searchTerm string, limit int) ([]*Log, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, status,
		       ip_address, user_agent, request_id, method, path,
		       duration_ms, error_message, metadata, created_at
		FROM audit.audit_logs
		WHERE to_tsvector('english', action || ' ' || COALESCE(error_message, '')) @@ plainto_tsquery('english', $1)
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, searchTerm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		log := &Log{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.Status, &log.IPAddress, &log.UserAgent,
			&log.RequestID, &log.Method, &log.Path, &log.DurationMs,
			&log.ErrorMessage, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetSecurityEvents retrieves security-related audit logs
func (r *Repository) GetSecurityEvents(ctx context.Context, limit int) ([]*Log, error) {
	securityActions := []string{
		ActionLogin,
		ActionLogout,
		ActionCreateAPIKey,
		ActionRevokeAPIKey,
		ActionCreatePolicy,
		ActionDeletePolicy,
		ActionAssignPermission,
		ActionRevokePermission,
	}

	placeholders := make([]string, len(securityActions))
	args := make([]interface{}, len(securityActions)+1)
	
	for i, action := range securityActions {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = action
	}
	args[len(securityActions)] = limit

	query := fmt.Sprintf(`
		SELECT id, user_id, action, resource_type, resource_id, status,
		       ip_address, user_agent, request_id, method, path,
		       duration_ms, error_message, metadata, created_at
		FROM audit.audit_logs
		WHERE action IN (%s)
		ORDER BY created_at DESC
		LIMIT $%d
	`, strings.Join(placeholders, ","), len(securityActions)+1)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		log := &Log{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.Status, &log.IPAddress, &log.UserAgent,
			&log.RequestID, &log.Method, &log.Path, &log.DurationMs,
			&log.ErrorMessage, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan security event: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// GetFailedAttempts retrieves failed authentication/authorization attempts
func (r *Repository) GetFailedAttempts(ctx context.Context, userID *uuid.UUID, limit int) ([]*Log, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, status,
		       ip_address, user_agent, request_id, method, path,
		       duration_ms, error_message, metadata, created_at
		FROM audit.audit_logs
		WHERE status = $1
	`

	args := []interface{}{StatusFailure}
	argPos := 2

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *userID)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argPos)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed attempts: %w", err)
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		log := &Log{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType,
			&log.ResourceID, &log.Status, &log.IPAddress, &log.UserAgent,
			&log.RequestID, &log.Method, &log.Path, &log.DurationMs,
			&log.ErrorMessage, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan failed attempt: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

