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
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles audit logging business logic
type Service struct {
	repo       *Repository
	logger     *zap.Logger
	logQueue   chan *Log
	batchSize  int
	flushTimer *time.Timer
	mu         sync.Mutex
}

// NewService creates a new audit service with write-behind pattern
func NewService(repo *Repository, logger *zap.Logger) *Service {
	s := &Service{
		repo:      repo,
		logger:    logger,
		logQueue:  make(chan *Log, 1000), // Buffer 1000 logs
		batchSize: 100,
	}

	// Start background worker for batch writes
	go s.processBatchWrites()

	return s
}

// Log creates an audit log entry (async)
func (s *Service) Log(ctx context.Context, req *CreateLogRequest) error {
	log := &Log{
		ID:           uuid.New(),
		UserID:       req.UserID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Status:       req.Status,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		RequestID:    req.RequestID,
		Method:       req.Method,
		Path:         req.Path,
		DurationMs:   req.DurationMs,
		ErrorMessage: req.ErrorMessage,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
	}

	// Send to queue (non-blocking)
	select {
	case s.logQueue <- log:
		return nil
	default:
		// Queue full, log directly to avoid data loss
		s.logger.Warn("Audit log queue full, writing directly to database")
		return s.repo.CreateLog(ctx, log)
	}
}

// LogSync creates an audit log entry (synchronous)
func (s *Service) LogSync(ctx context.Context, req *CreateLogRequest) error {
	log := &Log{
		ID:           uuid.New(),
		UserID:       req.UserID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Status:       req.Status,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		RequestID:    req.RequestID,
		Method:       req.Method,
		Path:         req.Path,
		DurationMs:   req.DurationMs,
		ErrorMessage: req.ErrorMessage,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
	}

	return s.repo.CreateLog(ctx, log)
}

// processBatchWrites processes audit logs in batches
func (s *Service) processBatchWrites() {
	batch := make([]*Log, 0, s.batchSize)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case log := <-s.logQueue:
			batch = append(batch, log)

			if len(batch) >= s.batchSize {
				s.flushBatch(batch)
				batch = make([]*Log, 0, s.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = make([]*Log, 0, s.batchSize)
			}
		}
	}
}

// flushBatch writes a batch of logs to database
func (s *Service) flushBatch(batch []*Log) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.repo.BatchCreateLogs(ctx, batch); err != nil {
		s.logger.Error("Failed to flush audit log batch",
			zap.Int("batch_size", len(batch)),
			zap.Error(err),
		)
		// Retry individual inserts
		for _, log := range batch {
			if err := s.repo.CreateLog(ctx, log); err != nil {
				s.logger.Error("Failed to write audit log",
					zap.String("log_id", log.ID.String()),
					zap.Error(err),
				)
			}
		}
	} else {
		s.logger.Debug("Flushed audit log batch",
			zap.Int("batch_size", len(batch)),
		)
	}
}

// QueryLogs queries audit logs with filters
func (s *Service) QueryLogs(ctx context.Context, req *QueryLogsRequest) ([]*Log, error) {
	// Set defaults
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	return s.repo.QueryLogs(ctx, req)
}

// SearchLogs performs full-text search on audit logs
func (s *Service) SearchLogs(ctx context.Context, searchTerm string, limit int) ([]*Log, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	return s.repo.SearchLogs(ctx, searchTerm, limit)
}

// GetSecurityEvents retrieves security-related events
func (s *Service) GetSecurityEvents(ctx context.Context, limit int) ([]*Log, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	return s.repo.GetSecurityEvents(ctx, limit)
}

// GetFailedAttempts retrieves failed authentication/authorization attempts
func (s *Service) GetFailedAttempts(ctx context.Context, userID *uuid.UUID, limit int) ([]*Log, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	return s.repo.GetFailedAttempts(ctx, userID, limit)
}

// LogAuthEvent is a helper to log authentication events
func (s *Service) LogAuthEvent(ctx context.Context, action string, userID *uuid.UUID, status string, ipAddress, userAgent *string, errorMsg *string) {
	req := &CreateLogRequest{
		UserID:       userID,
		Action:       action,
		ResourceType: "auth",
		Status:       status,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		ErrorMessage: errorMsg,
	}

	if err := s.Log(ctx, req); err != nil {
		s.logger.Error("Failed to log auth event", zap.Error(err))
	}
}

// LogACLEvent is a helper to log ACL events
func (s *Service) LogACLEvent(ctx context.Context, action string, userID uuid.UUID, resourceType string, resourceID *string, status string) {
	req := &CreateLogRequest{
		UserID:       &userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       status,
	}

	if err := s.Log(ctx, req); err != nil {
		s.logger.Error("Failed to log ACL event", zap.Error(err))
	}
}

// LogBlockchainEvent is a helper to log blockchain events
func (s *Service) LogBlockchainEvent(ctx context.Context, action string, userID uuid.UUID, resourceType string, resourceID *string, status string, durationMs *int, errorMsg *string) {
	req := &CreateLogRequest{
		UserID:       &userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Status:       status,
		DurationMs:   durationMs,
		ErrorMessage: errorMsg,
	}

	if err := s.Log(ctx, req); err != nil {
		s.logger.Error("Failed to log blockchain event", zap.Error(err))
	}
}

// Flush forces a flush of pending logs (useful for shutdown)
func (s *Service) Flush() {
	s.logger.Info("Flushing pending audit logs...")
	
	// Collect all pending logs
	batch := make([]*Log, 0)
	timeout := time.After(5 * time.Second)

	for {
		select {
		case log := <-s.logQueue:
			batch = append(batch, log)
		case <-timeout:
			s.flushBatch(batch)
			return
		default:
			if len(batch) > 0 {
				s.flushBatch(batch)
			}
			return
		}
	}
}

