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
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles chaincode registry business logic
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService creates a new chaincode registry service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateVersion creates a new chaincode version record
func (s *Service) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*ChaincodeVersion, error) {
	version := &ChaincodeVersion{
		ID:                uuid.New(),
		Name:              req.Name,
		Version:           req.Version,
		Sequence:         req.Sequence,
		PackageID:         req.PackageID,
		Label:             req.Label,
		Path:              req.Path,
		PackagePath:       req.PackagePath,
		ChannelName:       req.ChannelName,
		InstallStatus:     "pending",
		ApproveStatus:     "pending",
		CommitStatus:      "pending",
		InitRequired:      req.InitRequired,
		EndorsementPlugin: req.EndorsementPlugin,
		ValidationPlugin:  req.ValidationPlugin,
		Collections:       req.Collections,
		InstalledBy:       req.CreatedBy,
	}

	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	s.logger.Info("Created chaincode version",
		zap.String("id", version.ID.String()),
		zap.String("name", version.Name),
		zap.String("version", version.Version),
		zap.String("channel", version.ChannelName),
	)

	return version, nil
}

// CreateVersionRequest for creating a chaincode version
type CreateVersionRequest struct {
	Name              string
	Version           string
	Sequence          int
	PackageID         *string
	Label             *string
	Path              *string
	PackagePath       *string
	ChannelName       string
	InitRequired      bool
	EndorsementPlugin string
	ValidationPlugin  string
	Collections       json.RawMessage
	CreatedBy         *uuid.UUID
}

// UpdateVersionStatus updates the status of a chaincode version operation
func (s *Service) UpdateVersionStatus(ctx context.Context, versionID uuid.UUID, operation string, status string, errorMsg *string, performedBy *uuid.UUID) error {
	if err := s.repo.UpdateVersionStatus(ctx, versionID, operation, status, errorMsg, performedBy); err != nil {
		return fmt.Errorf("failed to update version status: %w", err)
	}

	s.logger.Info("Updated chaincode version status",
		zap.String("version_id", versionID.String()),
		zap.String("operation", operation),
		zap.String("status", status),
	)

	return nil
}

// CreateDeploymentLog creates a deployment log entry
func (s *Service) CreateDeploymentLog(ctx context.Context, req *CreateDeploymentLogRequest) (*DeploymentLog, error) {
	log := &DeploymentLog{
		ID:                uuid.New(),
		ChaincodeVersionID: req.ChaincodeVersionID,
		Operation:         req.Operation,
		Status:            "in_progress",
		RequestData:       req.RequestData,
		StartedAt:         &req.StartedAt,
		PerformedBy:       req.PerformedBy,
		IPAddress:         req.IPAddress,
		UserAgent:         req.UserAgent,
	}

	if err := s.repo.CreateDeploymentLog(ctx, log); err != nil {
		return nil, fmt.Errorf("failed to create deployment log: %w", err)
	}

	return log, nil
}

// CreateDeploymentLogRequest for creating a deployment log
type CreateDeploymentLogRequest struct {
	ChaincodeVersionID uuid.UUID
	Operation         string
	RequestData       json.RawMessage
	StartedAt         time.Time
	PerformedBy       *uuid.UUID
	IPAddress         *string
	UserAgent         *string
}

// UpdateDeploymentLog updates a deployment log with completion status
func (s *Service) UpdateDeploymentLog(ctx context.Context, logID uuid.UUID, status string, responseData json.RawMessage, errorMsg *string, errorCode *string) error {
	if err := s.repo.UpdateDeploymentLog(ctx, logID, status, responseData, errorMsg, errorCode); err != nil {
		return fmt.Errorf("failed to update deployment log: %w", err)
	}

	return nil
}

// GetVersionByID retrieves a chaincode version by ID
func (s *Service) GetVersionByID(ctx context.Context, id uuid.UUID) (*ChaincodeVersion, error) {
	return s.repo.GetVersionByID(ctx, id)
}

// ListVersions lists chaincode versions with filters
func (s *Service) ListVersions(ctx context.Context, filters *VersionFilters) ([]*ChaincodeVersion, error) {
	return s.repo.ListVersions(ctx, filters)
}

// GetDeploymentLogs retrieves deployment logs for a chaincode version
func (s *Service) GetDeploymentLogs(ctx context.Context, versionID uuid.UUID, limit int) ([]*DeploymentLog, error) {
	return s.repo.GetDeploymentLogs(ctx, versionID, limit)
}

// GetActiveChaincodes retrieves active chaincodes
func (s *Service) GetActiveChaincodes(ctx context.Context, channelName *string) ([]*ActiveChaincode, error) {
	return s.repo.GetActiveChaincodes(ctx, channelName)
}

// ExtractIPAddress extracts IP address from request
func ExtractIPAddress(ipStr string) *string {
	if ipStr == "" {
		return nil
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil
	}
	ipStr = ip.String()
	return &ipStr
}

