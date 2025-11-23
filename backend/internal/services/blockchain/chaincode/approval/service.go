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
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles approval workflow business logic
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new approval service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateRequestRequest for creating an approval request
type CreateRequestRequest struct {
	ChaincodeVersionID uuid.UUID
	Operation          string
	RequestedBy        uuid.UUID
	Reason             *string
	Metadata           map[string]interface{}
}

// CreateRequest creates a new approval request
func (s *Service) CreateRequest(ctx context.Context, req *CreateRequestRequest) (*ApprovalRequest, error) {
	// Get policy for operation
	policy, err := s.repo.GetPolicyByOperation(ctx, req.Operation)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval policy: %w", err)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(policy.ExpirationHours) * time.Hour)

	// Prepare metadata
	var metadataJSON json.RawMessage
	if req.Metadata != nil {
		metadataJSON, _ = json.Marshal(req.Metadata)
	}

	approvalReq := &ApprovalRequest{
		ID:                uuid.New(),
		ChaincodeVersionID: req.ChaincodeVersionID,
		Operation:         req.Operation,
		Status:            "pending",
		RequestedBy:       req.RequestedBy,
		RequestedAt:       time.Now(),
		ExpiresAt:         &expiresAt,
		Reason:            req.Reason,
		Metadata:          metadataJSON,
	}

	if err := s.repo.CreateRequest(ctx, approvalReq); err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	s.logger.Info("Created approval request",
		zap.String("id", approvalReq.ID.String()),
		zap.String("operation", approvalReq.Operation),
		zap.String("version_id", approvalReq.ChaincodeVersionID.String()),
	)

	return approvalReq, nil
}

// VoteRequest for voting on an approval request
type VoteRequest struct {
	ApprovalRequestID uuid.UUID
	ApproverID        uuid.UUID
	Vote              string // "approve" or "reject"
	Comment           *string
}

// Vote votes on an approval request
func (s *Service) Vote(ctx context.Context, req *VoteRequest) error {
	// Get approval request
	approvalReq, err := s.repo.GetRequestByID(ctx, req.ApprovalRequestID)
	if err != nil {
		return fmt.Errorf("approval request not found: %w", err)
	}

	// Check if already voted
	votes, err := s.repo.GetVotesByRequest(ctx, req.ApprovalRequestID)
	if err == nil {
		for _, vote := range votes {
			if vote.ApproverID == req.ApproverID {
				return fmt.Errorf("approver has already voted on this request")
			}
		}
	}

	// Check if request is still pending
	if approvalReq.Status != "pending" {
		return fmt.Errorf("approval request is not pending (status: %s)", approvalReq.Status)
	}

	// Check if expired
	if approvalReq.ExpiresAt != nil && approvalReq.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("approval request has expired")
	}

	// Prevent self-approval
	if approvalReq.RequestedBy == req.ApproverID {
		return fmt.Errorf("cannot approve your own request")
	}

	// Create vote
	vote := &ApprovalVote{
		ID:                uuid.New(),
		ApprovalRequestID: req.ApprovalRequestID,
		ApproverID:        req.ApproverID,
		Vote:              req.Vote,
		Comment:           req.Comment,
		VotedAt:           time.Now(),
	}

	if err := s.repo.CreateVote(ctx, vote); err != nil {
		return fmt.Errorf("failed to create vote: %w", err)
	}

	s.logger.Info("Vote created",
		zap.String("request_id", req.ApprovalRequestID.String()),
		zap.String("approver_id", req.ApproverID.String()),
		zap.String("vote", req.Vote),
	)

	// Check if vote is a rejection
	if req.Vote == "reject" {
		// Update status to rejected immediately
		if err := s.repo.UpdateRequestStatus(ctx, req.ApprovalRequestID, "rejected"); err != nil {
			s.logger.Error("Failed to update request status to rejected", zap.Error(err))
		} else {
			s.logger.Info("Approval request rejected",
				zap.String("request_id", req.ApprovalRequestID.String()),
			)
		}
		return nil
	}

	// For approve votes, check if request is now fully approved
	isApproved, err := s.repo.CheckApprovalStatus(ctx, req.ApprovalRequestID)
	if err != nil {
		s.logger.Error("Failed to check approval status after vote", zap.Error(err))
		return nil // Don't fail the vote, just log
	}

	if isApproved {
		// Update status to approved
		if err := s.repo.UpdateRequestStatus(ctx, req.ApprovalRequestID, "approved"); err != nil {
			s.logger.Error("Failed to update request status to approved", zap.Error(err))
		} else {
			s.logger.Info("Approval request approved",
				zap.String("request_id", req.ApprovalRequestID.String()),
			)
		}
	}

	return nil
}

// CheckApproval checks if an approval request is approved
func (s *Service) CheckApproval(ctx context.Context, versionID uuid.UUID, operation string) (bool, error) {
	// Get approval request
	approvalReq, err := s.repo.GetRequestByVersionAndOperation(ctx, versionID, operation)
	if err != nil {
		// No approval request found - for backward compatibility, allow operation
		// In production, you might want to require approval
		s.logger.Warn("No approval request found, allowing operation",
			zap.String("version_id", versionID.String()),
			zap.String("operation", operation),
		)
		return true, nil // Allow if no approval required
	}

	// Check if expired
	if approvalReq.ExpiresAt != nil && approvalReq.ExpiresAt.Before(time.Now()) {
		if approvalReq.Status == "pending" {
			// Update status to expired (trigger will handle this, but check anyway)
			return false, fmt.Errorf("approval request has expired")
		}
	}

	// Check status
	if approvalReq.Status == "approved" {
		return true, nil
	}

	if approvalReq.Status == "rejected" {
		return false, fmt.Errorf("approval request was rejected")
	}

	if approvalReq.Status == "expired" {
		return false, fmt.Errorf("approval request has expired")
	}

	// If pending, check if it's actually approved (votes might have been cast)
	if approvalReq.Status == "pending" {
		isApproved, err := s.repo.CheckApprovalStatus(ctx, approvalReq.ID)
		if err != nil {
			return false, fmt.Errorf("failed to check approval status: %w", err)
		}
		return isApproved, nil
	}

	return false, fmt.Errorf("approval request status is invalid: %s", approvalReq.Status)
}

// GetRequest retrieves an approval request by ID
func (s *Service) GetRequest(ctx context.Context, id uuid.UUID) (*ApprovalRequest, error) {
	return s.repo.GetRequestByID(ctx, id)
}

// GetRequestByVersionAndOperation retrieves approval request for a version and operation
func (s *Service) GetRequestByVersionAndOperation(ctx context.Context, versionID uuid.UUID, operation string) (*ApprovalRequest, error) {
	return s.repo.GetRequestByVersionAndOperation(ctx, versionID, operation)
}

// ListRequests lists approval requests with filters
func (s *Service) ListRequests(ctx context.Context, filters *RequestFilters) ([]*ApprovalRequest, error) {
	return s.repo.ListRequests(ctx, filters)
}

// GetVotes retrieves all votes for an approval request
func (s *Service) GetVotes(ctx context.Context, requestID uuid.UUID) ([]*ApprovalVote, error) {
	return s.repo.GetVotesByRequest(ctx, requestID)
}

// GetPolicy retrieves approval policy for an operation
func (s *Service) GetPolicy(ctx context.Context, operation string) (*ApprovalPolicy, error) {
	return s.repo.GetPolicyByOperation(ctx, operation)
}

