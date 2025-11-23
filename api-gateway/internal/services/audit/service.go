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
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/repository/db"
	"go.uber.org/zap"
)

// Service handles audit logging operations
type Service struct {
	db      *pgxpool.Pool
	queries *db.Queries
	logger  *zap.Logger
}

// NewService creates a new audit service
func NewService(dbPool *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{
		db:      dbPool,
		queries: db.New(),
		logger:  logger,
	}
}

// LogRequest represents an audit log entry for an API request
type LogRequest struct {
	UserID       string
	ApiKeyID     string
	Action       string
	ResourceType string
	ResourceID   string
	TxID         string
	Status       string
	Details      map[string]interface{}
	IpAddress    string
	UserAgent    string
}

// CreateLog creates an audit log entry
func (s *Service) CreateLog(ctx context.Context, req *LogRequest) error {
	// Convert user ID
	var userUUID pgtype.UUID
	if req.UserID != "" {
		userID, err := uuid.Parse(req.UserID)
		if err == nil {
			if err := userUUID.Scan(userID.String()); err != nil {
				s.logger.Warn("Failed to parse user ID", zap.String("userID", req.UserID), zap.Error(err))
			}
		}
	}

	// Convert API key ID
	var apiKeyUUID pgtype.UUID
	if req.ApiKeyID != "" {
		apiKeyID, err := uuid.Parse(req.ApiKeyID)
		if err == nil {
			if err := apiKeyUUID.Scan(apiKeyID.String()); err != nil {
				s.logger.Warn("Failed to parse API key ID", zap.String("apiKeyID", req.ApiKeyID), zap.Error(err))
			}
		}
	}

	// Convert resource type
	var resourceType pgtype.Text
	if req.ResourceType != "" {
		resourceType = pgtype.Text{String: req.ResourceType, Valid: true}
	}

	// Convert resource ID
	var resourceID pgtype.Text
	if req.ResourceID != "" {
		resourceID = pgtype.Text{String: req.ResourceID, Valid: true}
	}

	// Convert tx ID
	var txID pgtype.Text
	if req.TxID != "" {
		txID = pgtype.Text{String: req.TxID, Valid: true}
	}

	// Convert details to JSON
	var detailsBytes []byte
	if req.Details != nil {
		var err error
		detailsBytes, err = json.Marshal(req.Details)
		if err != nil {
			s.logger.Warn("Failed to marshal details", zap.Error(err))
			detailsBytes = []byte("{}")
		}
	} else {
		detailsBytes = []byte("{}")
	}

	// Convert IP address
	var ipAddr *netip.Addr
	if req.IpAddress != "" {
		if addr, err := netip.ParseAddr(req.IpAddress); err == nil {
			ipAddr = &addr
		}
	}

	// Convert user agent
	var userAgent pgtype.Text
	if req.UserAgent != "" {
		userAgent = pgtype.Text{String: req.UserAgent, Valid: true}
	}

	// Create audit log
	params := db.CreateAuditLogParams{
		UserID:       userUUID,
		ApiKeyID:     apiKeyUUID,
		Action:       req.Action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TxID:         txID,
		Status:       req.Status,
		Details:      detailsBytes,
		IpAddress:    ipAddr,
		UserAgent:    userAgent,
	}

	_, err := s.queries.CreateAuditLog(ctx, s.db, params)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetLog retrieves an audit log by ID
func (s *Service) GetLog(ctx context.Context, id int64) (*db.AuditLog, error) {
	log, err := s.queries.GetAuditLog(ctx, s.db, id)
	if err != nil {
		return nil, fmt.Errorf("audit log not found: %w", err)
	}

	return &log, nil
}

// ListLogs lists audit logs with pagination
func (s *Service) ListLogs(ctx context.Context, limit, offset int32) ([]*db.AuditLog, int64, error) {
	// Get logs
	logs, err := s.queries.ListAuditLogs(ctx, s.db, db.ListAuditLogsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	// Get total count
	total, err := s.queries.CountAuditLogs(ctx, s.db)
	if err != nil {
		s.logger.Warn("Failed to count audit logs", zap.Error(err))
		total = int64(len(logs))
	}

	// Convert to pointers
	result := make([]*db.AuditLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, total, nil
}

// ListLogsByUser lists audit logs for a specific user
func (s *Service) ListLogsByUser(ctx context.Context, userID string, limit, offset int32) ([]*db.AuditLog, int64, error) {
	// Convert user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(userUUID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to convert user ID: %w", err)
	}

	// Get logs
	logs, err := s.queries.ListAuditLogsByUser(ctx, s.db, db.ListAuditLogsByUserParams{
		UserID: pgUUID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs by user: %w", err)
	}

	// Get total count
	total, err := s.queries.CountAuditLogsByUser(ctx, s.db, pgUUID)
	if err != nil {
		s.logger.Warn("Failed to count audit logs by user", zap.Error(err))
		total = int64(len(logs))
	}

	// Convert to pointers
	result := make([]*db.AuditLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, total, nil
}

// ListLogsByAction lists audit logs for a specific action
func (s *Service) ListLogsByAction(ctx context.Context, action string, limit, offset int32) ([]*db.AuditLog, int64, error) {
	// Get logs
	logs, err := s.queries.ListAuditLogsByAction(ctx, s.db, db.ListAuditLogsByActionParams{
		Action: action,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs by action: %w", err)
	}

	// Count total (approximate - would need separate query for exact count)
	total := int64(len(logs))

	// Convert to pointers
	result := make([]*db.AuditLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, total, nil
}

// ListLogsByTxID lists audit logs for a specific transaction
func (s *Service) ListLogsByTxID(ctx context.Context, txID string) ([]*db.AuditLog, error) {
	txIDText := pgtype.Text{String: txID, Valid: true}

	// Get logs
	logs, err := s.queries.ListAuditLogsByTxID(ctx, s.db, txIDText)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs by tx ID: %w", err)
	}

	// Convert to pointers
	result := make([]*db.AuditLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, nil
}

// ListLogsByDateRange lists audit logs within a date range
func (s *Service) ListLogsByDateRange(
	ctx context.Context,
	startTime, endTime time.Time,
	limit, offset int32,
) ([]*db.AuditLog, int64, error) {
	start := pgtype.Timestamptz{Time: startTime, Valid: true}
	end := pgtype.Timestamptz{Time: endTime, Valid: true}

	// Get logs
	logs, err := s.queries.GetAuditLogsByDateRange(ctx, s.db, db.GetAuditLogsByDateRangeParams{
		CreatedAt:   start,
		CreatedAt_2: end,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs by date range: %w", err)
	}

	// Count total (approximate)
	total := int64(len(logs))

	// Convert to pointers
	result := make([]*db.AuditLog, len(logs))
	for i := range logs {
		result[i] = &logs[i]
	}

	return result, total, nil
}

