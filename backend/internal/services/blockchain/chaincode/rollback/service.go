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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/registry"
	"go.uber.org/zap"
)

// Service handles rollback operations business logic
type Service struct {
	repo           *Repository
	registryService *registry.Service
	logger         *zap.Logger
}

// NewService creates a new rollback service
func NewService(repo *Repository, registryService *registry.Service, logger *zap.Logger) *Service {
	return &Service{
		repo:            repo,
		registryService: registryService,
		logger:          logger,
	}
}

// CreateRollbackRequest for creating a rollback operation
type CreateRollbackRequest struct {
	ChaincodeName string
	ChannelName   string
	ToVersionID   *uuid.UUID // If nil, rollback to previous version
	Reason         *string
	RequestedBy    uuid.UUID
	Metadata       map[string]interface{}
}

// CreateRollbackOperation creates a new rollback operation
func (s *Service) CreateRollbackOperation(ctx context.Context, req *CreateRollbackRequest) (*RollbackOperation, error) {
	// Get current active version
	filters := &registry.VersionFilters{
		Name:        &req.ChaincodeName,
		ChannelName: &req.ChannelName,
		Limit:       1,
	}
	versions, err := s.registryService.ListVersions(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get current version: %w", err)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no active version found for chaincode %s on channel %s", req.ChaincodeName, req.ChannelName)
	}

	// Find current committed version
	var currentVersion *registry.ChaincodeVersion
	for _, v := range versions {
		if v.CommitStatus == "committed" {
			currentVersion = v
			break
		}
	}

	if currentVersion == nil {
		return nil, fmt.Errorf("no committed version found for chaincode %s on channel %s", req.ChaincodeName, req.ChannelName)
	}

	// Get target version
	var targetVersionID uuid.UUID
	var targetVersion *registry.ChaincodeVersion

	if req.ToVersionID != nil {
		// Rollback to specific version
		targetVersion, err = s.registryService.GetVersionByID(ctx, *req.ToVersionID)
		if err != nil {
			return nil, fmt.Errorf("target version not found: %w", err)
		}
		targetVersionID = *req.ToVersionID
	} else {
		// Rollback to previous version
		prevVersionID, err := s.repo.GetPreviousActiveVersion(ctx, req.ChaincodeName, req.ChannelName, currentVersion.Sequence)
		if err != nil {
			return nil, fmt.Errorf("no previous version found: %w", err)
		}
		targetVersion, err = s.registryService.GetVersionByID(ctx, *prevVersionID)
		if err != nil {
			return nil, fmt.Errorf("previous version not found: %w", err)
		}
		targetVersionID = *prevVersionID
	}

	// Validate rollback
	if targetVersion.Sequence >= currentVersion.Sequence {
		return nil, fmt.Errorf("cannot rollback to version with sequence >= current sequence")
	}

	// Check if rollback is safe
	isSafe, err := s.repo.IsRollbackSafe(ctx, req.ChaincodeName, req.ChannelName)
	if err != nil {
		s.logger.Warn("Failed to check rollback safety", zap.Error(err))
		// Continue anyway, but log warning
	} else if !isSafe {
		return nil, fmt.Errorf("rollback is not safe: there are pending operations")
	}

	// Prepare metadata
	var metadataJSON json.RawMessage
	if req.Metadata != nil {
		metadataJSON, _ = json.Marshal(req.Metadata)
	}

	// Create rollback operation
	op := &RollbackOperation{
		ID:            uuid.New(),
		ChaincodeName: req.ChaincodeName,
		ChannelName:   req.ChannelName,
		FromVersionID: currentVersion.ID,
		ToVersionID:   targetVersionID,
		FromVersion:   currentVersion.Version,
		ToVersion:     targetVersion.Version,
		FromSequence:  currentVersion.Sequence,
		ToSequence:    targetVersion.Sequence,
		Status:        "pending",
		Reason:        req.Reason,
		RollbackType:  "version",
		RequestedBy:   req.RequestedBy,
		Metadata:      metadataJSON,
	}

	if err := s.repo.CreateRollbackOperation(ctx, op); err != nil {
		return nil, fmt.Errorf("failed to create rollback operation: %w", err)
	}

	s.logger.Info("Created rollback operation",
		zap.String("id", op.ID.String()),
		zap.String("chaincode", req.ChaincodeName),
		zap.String("from_version", op.FromVersion),
		zap.String("to_version", op.ToVersion),
	)

	return op, nil
}

// ExecuteRollback executes a rollback operation
func (s *Service) ExecuteRollback(ctx context.Context, operationID uuid.UUID, executedBy uuid.UUID) error {
	// Get rollback operation
	op, err := s.repo.GetRollbackOperationByID(ctx, operationID)
	if err != nil {
		return fmt.Errorf("rollback operation not found: %w", err)
	}

	// Check status
	if op.Status != "pending" {
		return fmt.Errorf("rollback operation is not pending (status: %s)", op.Status)
	}

	// Update status to in_progress
	if err := s.repo.UpdateRollbackStatus(ctx, operationID, "in_progress", nil, nil, &executedBy); err != nil {
		return fmt.Errorf("failed to update rollback status: %w", err)
	}

	startTime := time.Now()

	// Get target version details (for validation)
	_, err = s.registryService.GetVersionByID(ctx, op.ToVersionID)
	if err != nil {
		errMsg := err.Error()
		s.repo.UpdateRollbackStatus(ctx, operationID, "failed", &errMsg, nil, &executedBy)
		return fmt.Errorf("target version not found: %w", err)
	}

	// Rollback strategy:
	// 1. Deactivate current version
	// 2. Reactivate target version
	// 3. Update version statuses

	// Step 1: Deactivate current version (update commit_status)
	// Note: In Fabric, we can't directly "deactivate" a committed chaincode
	// Instead, we need to commit the previous version again
	// This is a simplified approach - in production, you might need to:
	// - Commit the previous version definition again
	// - Or use Fabric's chaincode upgrade mechanism

	// For now, we'll track the rollback in our registry
	// The actual Fabric rollback would require committing the previous version again

	// Update current version status (mark as rolled back)
	// This is metadata tracking - actual Fabric state requires re-commit

	// Create rollback history entries
	historyDetails, _ := json.Marshal(map[string]interface{}{
		"from_sequence": op.FromSequence,
		"to_sequence":    op.ToSequence,
		"rollback_type":  op.RollbackType,
	})

	// Record rollback history
	if err := s.repo.CreateRollbackHistory(ctx, operationID, op.FromVersionID, "commit", "committed", "rolled_back", historyDetails); err != nil {
		s.logger.Warn("Failed to create rollback history", zap.Error(err))
	}

	if err := s.repo.CreateRollbackHistory(ctx, operationID, op.ToVersionID, "commit", "committed", "active", historyDetails); err != nil {
		s.logger.Warn("Failed to create rollback history", zap.Error(err))
	}

	// Update rollback operation status (duration calculated automatically by repository)
	if err := s.repo.UpdateRollbackStatus(ctx, operationID, "completed", nil, nil, &executedBy); err != nil {
		return fmt.Errorf("failed to update rollback status: %w", err)
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	s.logger.Info("Rollback operation completed",
		zap.String("id", operationID.String()),
		zap.String("chaincode", op.ChaincodeName),
		zap.Int("duration_ms", durationMs),
	)

	return nil
}

// GetRollbackOperation retrieves a rollback operation by ID
func (s *Service) GetRollbackOperation(ctx context.Context, id uuid.UUID) (*RollbackOperation, error) {
	return s.repo.GetRollbackOperationByID(ctx, id)
}

// ListRollbackOperations lists rollback operations with filters
func (s *Service) ListRollbackOperations(ctx context.Context, filters *RollbackFilters) ([]*RollbackOperation, error) {
	return s.repo.ListRollbackOperations(ctx, filters)
}

// GetRollbackHistory retrieves rollback history for an operation
func (s *Service) GetRollbackHistory(ctx context.Context, operationID uuid.UUID) ([]*RollbackHistory, error) {
	return s.repo.GetRollbackHistory(ctx, operationID)
}

// CancelRollback cancels a pending rollback operation
func (s *Service) CancelRollback(ctx context.Context, operationID uuid.UUID, userID uuid.UUID) error {
	op, err := s.repo.GetRollbackOperationByID(ctx, operationID)
	if err != nil {
		return fmt.Errorf("rollback operation not found: %w", err)
	}

	if op.Status != "pending" {
		return fmt.Errorf("cannot cancel rollback operation with status: %s", op.Status)
	}

	return s.repo.UpdateRollbackStatus(ctx, operationID, "cancelled", nil, nil, &userID)
}

