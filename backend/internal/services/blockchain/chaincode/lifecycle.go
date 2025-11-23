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

package chaincode

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/infrastructure/admin"
	"github.com/ibn-network/backend/internal/services/analytics/audit"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/approval"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/registry"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/testing"
	"go.uber.org/zap"
)

// LifecycleService handles chaincode lifecycle operations via Admin Service
type LifecycleService struct {
	adminClient     *admin.Client
	registryService *registry.Service
	auditService    *audit.Service
	approvalService *approval.Service
	testingService  *testing.Service
	logger          *zap.Logger
}

// NewLifecycleService creates a new chaincode lifecycle service
func NewLifecycleService(
	adminClient *admin.Client,
	registryService *registry.Service,
	auditService *audit.Service,
	approvalService *approval.Service,
	testingService *testing.Service,
	logger *zap.Logger,
) *LifecycleService {
	return &LifecycleService{
		adminClient:     adminClient,
		registryService: registryService,
		auditService:    auditService,
		approvalService: approvalService,
		testingService:  testingService,
		logger:          logger,
	}
}

// getUserIDFromContext extracts user ID from context
func (s *LifecycleService) getUserIDFromContext(ctx context.Context) *uuid.UUID {
	userIDVal := ctx.Value("user_id")
	if userIDVal == nil {
		return nil
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		var err error
		userID, err = uuid.Parse(v)
		if err != nil {
			return nil
		}
	default:
		return nil
	}

	return &userID
}

// InstallChaincode installs a chaincode package
func (s *LifecycleService) InstallChaincode(ctx context.Context, packagePath, label string, ipAddress, userAgent *string) (string, error) {
	startTime := time.Now()
	userID := s.getUserIDFromContext(ctx)
	
	s.logger.Info("Installing chaincode via Admin Service",
		zap.String("package", packagePath),
		zap.String("label", label),
	)

	// Call Admin Service
	packageID, err := s.adminClient.InstallChaincode(ctx, packagePath, label)
	durationMs := int(time.Since(startTime).Milliseconds())

	// Log audit event
	if s.auditService != nil && userID != nil {
		status := "success"
		var errorMsg *string
		if err != nil {
			status = "failed"
			errStr := err.Error()
			errorMsg = &errStr
		}
		s.auditService.LogBlockchainEvent(
			ctx,
			"chaincode.install",
			*userID,
			"chaincode",
			&packageID,
			status,
			&durationMs,
			errorMsg,
		)
	}

	if err != nil {
		s.logger.Error("Failed to install chaincode", zap.Error(err))
		return "", fmt.Errorf("failed to install chaincode: %w", err)
	}

	s.logger.Info("Chaincode installed successfully",
		zap.String("packageId", packageID),
	)

	return packageID, nil
}

// ApproveChaincode approves a chaincode definition
func (s *LifecycleService) ApproveChaincode(ctx context.Context, req *ApproveChaincodeRequest, ipAddress, userAgent *string) error {
	startTime := time.Now()
	userID := s.getUserIDFromContext(ctx)
	
	s.logger.Info("Approving chaincode via Admin Service",
		zap.String("channel", req.ChannelName),
		zap.String("name", req.Name),
		zap.String("version", req.Version),
	)

	// Find or create chaincode version record
	var versionID uuid.UUID
	filters := &registry.VersionFilters{
		Name:        &req.Name,
		ChannelName: &req.ChannelName,
		Limit:       1,
	}
	versions, listErr := s.registryService.ListVersions(ctx, filters)
	if listErr == nil && len(versions) > 0 {
		// Find matching version
		for _, v := range versions {
			if v.Version == req.Version && v.Sequence == int(req.Sequence) {
				versionID = v.ID
				break
			}
		}
	}

	// Create version if not found
	if versionID == uuid.Nil {
		collectionsJSON, _ := json.Marshal(req.Collections)
		versionReq := &registry.CreateVersionRequest{
			Name:              req.Name,
			Version:           req.Version,
			Sequence:          int(req.Sequence),
			PackageID:         &req.PackageID,
			ChannelName:       req.ChannelName,
			InitRequired:      req.InitRequired,
			EndorsementPlugin: req.EndorsementPlugin,
			ValidationPlugin:  req.ValidationPlugin,
			Collections:       collectionsJSON,
			CreatedBy:         userID,
		}
		version, createErr := s.registryService.CreateVersion(ctx, versionReq)
		if createErr != nil {
			s.logger.Warn("Failed to create version record", zap.Error(createErr))
		} else {
			versionID = version.ID
		}
	}

	// Create deployment log
	var deploymentLogID uuid.UUID
	if versionID != uuid.Nil && s.registryService != nil {
		requestData, _ := json.Marshal(req)
		logReq := &registry.CreateDeploymentLogRequest{
			ChaincodeVersionID: versionID,
			Operation:         "approve",
			RequestData:       requestData,
			StartedAt:         startTime,
			PerformedBy:       userID,
			IPAddress:         ipAddress,
			UserAgent:         userAgent,
		}
		log, logErr := s.registryService.CreateDeploymentLog(ctx, logReq)
		if logErr == nil {
			deploymentLogID = log.ID
		}
	}

	adminReq := &admin.ApproveChaincodeRequest{
		ChannelName:         req.ChannelName,
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		PackageID:           req.PackageID,
		InitRequired:        req.InitRequired,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		Collections:         req.Collections,
	}

	err := s.adminClient.ApproveChaincode(ctx, adminReq)
	durationMs := int(time.Since(startTime).Milliseconds())

	// Update deployment log and version status
	if versionID != uuid.Nil {
		status := "success"
		var errorMsg *string
		if err != nil {
			status = "failed"
			errStr := err.Error()
			errorMsg = &errStr
		}

		// Update version status
		s.registryService.UpdateVersionStatus(ctx, versionID, "approve", status, errorMsg, userID)

		// Update deployment log
		if deploymentLogID != uuid.Nil {
			responseData, _ := json.Marshal(map[string]interface{}{
				"status": status,
			})
			s.registryService.UpdateDeploymentLog(ctx, deploymentLogID, status, responseData, errorMsg, nil)
		}
	}

	// Log audit event
	if s.auditService != nil && userID != nil {
		status := "success"
		var errorMsg *string
		if err != nil {
			status = "failed"
			errStr := err.Error()
			errorMsg = &errStr
		}
		resourceID := fmt.Sprintf("%s:%s:%s", req.ChannelName, req.Name, req.Version)
		s.auditService.LogBlockchainEvent(
			ctx,
			"chaincode.approve",
			*userID,
			"chaincode",
			&resourceID,
			status,
			&durationMs,
			errorMsg,
		)
	}

	if err != nil {
		s.logger.Error("Failed to approve chaincode", zap.Error(err))
		return fmt.Errorf("failed to approve chaincode: %w", err)
	}

	s.logger.Info("Chaincode approved successfully",
		zap.String("name", req.Name),
		zap.String("version", req.Version),
	)

	return nil
}

// CommitChaincode commits a chaincode definition
func (s *LifecycleService) CommitChaincode(ctx context.Context, req *CommitChaincodeRequest, ipAddress, userAgent *string) error {
	startTime := time.Now()
	userID := s.getUserIDFromContext(ctx)
	
	s.logger.Info("Committing chaincode via Admin Service",
		zap.String("channel", req.ChannelName),
		zap.String("name", req.Name),
		zap.String("version", req.Version),
	)

	// Find or create chaincode version record
	var versionID uuid.UUID
	filters := &registry.VersionFilters{
		Name:        &req.Name,
		ChannelName: &req.ChannelName,
		Limit:       1,
	}
	versions, listErr := s.registryService.ListVersions(ctx, filters)
	if listErr == nil && len(versions) > 0 {
		// Find matching version
		for _, v := range versions {
			if v.Version == req.Version && v.Sequence == int(req.Sequence) {
				versionID = v.ID
				break
			}
		}
	}

	// Create version if not found
	if versionID == uuid.Nil {
		collectionsJSON, _ := json.Marshal(req.Collections)
		versionReq := &registry.CreateVersionRequest{
			Name:              req.Name,
			Version:           req.Version,
			Sequence:          int(req.Sequence),
			ChannelName:       req.ChannelName,
			InitRequired:      req.InitRequired,
			EndorsementPlugin: req.EndorsementPlugin,
			ValidationPlugin:  req.ValidationPlugin,
			Collections:       collectionsJSON,
			CreatedBy:         userID,
		}
		version, createErr := s.registryService.CreateVersion(ctx, versionReq)
		if createErr != nil {
			s.logger.Warn("Failed to create version record", zap.Error(createErr))
		} else {
			versionID = version.ID
		}
	}

	// Check tests before commit (if testing service is available)
	if s.testingService != nil && versionID != uuid.Nil {
		testsPassed, testErr := s.testingService.CheckTestsPassed(ctx, versionID, "unit")
		if testErr != nil {
			s.logger.Warn("Test check failed, proceeding anyway", zap.Error(testErr))
		} else if !testsPassed {
			// Get latest test suite for details
			latestSuite, suiteErr := s.testingService.GetLatestTestSuite(ctx, versionID, "unit")
			if suiteErr == nil {
				return fmt.Errorf("tests failed: cannot commit chaincode with failed tests (test suite: %s, failed tests: %d)", latestSuite.ID.String(), latestSuite.FailedTests)
			}
			return fmt.Errorf("tests failed: cannot commit chaincode with failed tests")
		}
		s.logger.Info("Tests passed, proceeding with commit")
	}

	// Check approval before commit (if approval service is available)
	if s.approvalService != nil && versionID != uuid.Nil && userID != nil {
		isApproved, err := s.approvalService.CheckApproval(ctx, versionID, "commit")
		if err != nil {
			s.logger.Warn("Approval check failed, proceeding anyway", zap.Error(err))
			// For backward compatibility, allow if approval check fails
			// In production, you might want to return error here
		} else if !isApproved {
			// Create approval request if it doesn't exist
			approvalReq, err := s.approvalService.GetRequestByVersionAndOperation(ctx, versionID, "commit")
			if err != nil {
				// Request doesn't exist, create it
				createReq := &approval.CreateRequestRequest{
					ChaincodeVersionID: versionID,
					Operation:          "commit",
					RequestedBy:        *userID,
					Reason:             nil,
					Metadata: map[string]interface{}{
						"channel": req.ChannelName,
						"name":    req.Name,
						"version": req.Version,
					},
				}
				approvalReq, err = s.approvalService.CreateRequest(ctx, createReq)
				if err != nil {
					s.logger.Error("Failed to create approval request", zap.Error(err))
					return fmt.Errorf("approval required but request creation failed: %w", err)
				}
			}
			
			return fmt.Errorf("approval required: approval request %s is not approved (status: %s)", approvalReq.ID.String(), approvalReq.Status)
		}
	}

	// Create deployment log
	var deploymentLogID uuid.UUID
	if versionID != uuid.Nil && s.registryService != nil {
		requestData, _ := json.Marshal(req)
		logReq := &registry.CreateDeploymentLogRequest{
			ChaincodeVersionID: versionID,
			Operation:         "commit",
			RequestData:       requestData,
			StartedAt:         startTime,
			PerformedBy:       userID,
			IPAddress:         ipAddress,
			UserAgent:         userAgent,
		}
		log, logErr := s.registryService.CreateDeploymentLog(ctx, logReq)
		if logErr == nil {
			deploymentLogID = log.ID
		}
	}

	adminReq := &admin.CommitChaincodeRequest{
		ChannelName:         req.ChannelName,
		Name:                req.Name,
		Version:             req.Version,
		Sequence:            req.Sequence,
		InitRequired:        req.InitRequired,
		EndorsementPlugin:   req.EndorsementPlugin,
		ValidationPlugin:    req.ValidationPlugin,
		Collections:         req.Collections,
	}

	err := s.adminClient.CommitChaincode(ctx, adminReq)
	durationMs := int(time.Since(startTime).Milliseconds())

	// Update deployment log and version status
	if versionID != uuid.Nil {
		status := "success"
		var errorMsg *string
		if err != nil {
			status = "failed"
			errStr := err.Error()
			errorMsg = &errStr
		}

		// Update version status
		s.registryService.UpdateVersionStatus(ctx, versionID, "commit", status, errorMsg, userID)

		// Update deployment log
		if deploymentLogID != uuid.Nil {
			responseData, _ := json.Marshal(map[string]interface{}{
				"status": status,
			})
			s.registryService.UpdateDeploymentLog(ctx, deploymentLogID, status, responseData, errorMsg, nil)
		}
	}

	// Log audit event
	if s.auditService != nil && userID != nil {
		status := "success"
		var errorMsg *string
		if err != nil {
			status = "failed"
			errStr := err.Error()
			errorMsg = &errStr
		}
		resourceID := fmt.Sprintf("%s:%s:%s", req.ChannelName, req.Name, req.Version)
		s.auditService.LogBlockchainEvent(
			ctx,
			"chaincode.commit",
			*userID,
			"chaincode",
			&resourceID,
			status,
			&durationMs,
			errorMsg,
		)
	}

	if err != nil {
		s.logger.Error("Failed to commit chaincode", zap.Error(err))
		return fmt.Errorf("failed to commit chaincode: %w", err)
	}

	s.logger.Info("Chaincode committed successfully",
		zap.String("name", req.Name),
		zap.String("version", req.Version),
	)

	return nil
}

// ListInstalled lists installed chaincodes
func (s *LifecycleService) ListInstalled(ctx context.Context, peer string) ([]InstalledChaincode, error) {
	s.logger.Info("Listing installed chaincodes via Admin Service", zap.String("peer", peer))

	chaincodes, err := s.adminClient.ListInstalled(ctx, peer)
	if err != nil {
		s.logger.Error("Failed to list installed chaincodes", zap.Error(err))
		return nil, fmt.Errorf("failed to list installed chaincodes: %w", err)
	}

	// Convert to internal format
	result := make([]InstalledChaincode, 0, len(chaincodes))
	for _, cc := range chaincodes {
		result = append(result, InstalledChaincode{
			PackageID: cc.PackageID,
			Label:     cc.Label,
			Chaincode: ChaincodeInfo{
				Name:    cc.Chaincode.Name,
				Version: cc.Chaincode.Version,
				Path:    cc.Chaincode.Path,
			},
		})
	}

	return result, nil
}

// ListCommitted lists committed chaincodes
func (s *LifecycleService) ListCommitted(ctx context.Context, channel string) ([]CommittedChaincode, error) {
	s.logger.Info("Listing committed chaincodes via Admin Service", zap.String("channel", channel))

	chaincodes, err := s.adminClient.ListCommitted(ctx, channel)
	if err != nil {
		s.logger.Error("Failed to list committed chaincodes", zap.Error(err))
		return nil, fmt.Errorf("failed to list committed chaincodes: %w", err)
	}

	// Convert to internal format
	result := make([]CommittedChaincode, 0, len(chaincodes))
	for _, cc := range chaincodes {
		result = append(result, CommittedChaincode{
			Name:                 cc.Name,
			Version:              cc.Version,
			Sequence:             cc.Sequence,
			EndorsementPlugin:    cc.EndorsementPlugin,
			ValidationPlugin:     cc.ValidationPlugin,
			InitRequired:         cc.InitRequired,
			Collections:          cc.Collections,
			ApprovedOrganizations: cc.ApprovedOrganizations,
		})
	}

	return result, nil
}

// GetCommittedInfo gets information about a committed chaincode
func (s *LifecycleService) GetCommittedInfo(ctx context.Context, channel, name string) (*CommittedChaincode, error) {
	s.logger.Info("Getting committed chaincode info via Admin Service",
		zap.String("channel", channel),
		zap.String("name", name),
	)

	cc, err := s.adminClient.GetCommittedInfo(ctx, channel, name)
	if err != nil {
		s.logger.Error("Failed to get committed chaincode info", zap.Error(err))
		return nil, fmt.Errorf("failed to get committed chaincode info: %w", err)
	}

	return &CommittedChaincode{
		Name:                 cc.Name,
		Version:              cc.Version,
		Sequence:             cc.Sequence,
		EndorsementPlugin:    cc.EndorsementPlugin,
		ValidationPlugin:     cc.ValidationPlugin,
		InitRequired:         cc.InitRequired,
		Collections:          cc.Collections,
		ApprovedOrganizations: cc.ApprovedOrganizations,
	}, nil
}

// Request types
type ApproveChaincodeRequest struct {
	ChannelName         string
	Name                string
	Version             string
	Sequence            int64
	PackageID           string
	InitRequired        bool
	EndorsementPlugin   string
	ValidationPlugin    string
	Collections         []string
}

type CommitChaincodeRequest struct {
	ChannelName         string
	Name                string
	Version             string
	Sequence            int64
	InitRequired        bool
	EndorsementPlugin   string
	ValidationPlugin    string
	Collections         []string
}

// Response types
type InstalledChaincode struct {
	PackageID string        `json:"packageId"`
	Label     string        `json:"label"`
	Chaincode ChaincodeInfo  `json:"chaincode"`
}

type ChaincodeInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

type CommittedChaincode struct {
	Name                 string
	Version              string
	Sequence             int64
	EndorsementPlugin    string
	ValidationPlugin     string
	InitRequired         bool
	Collections          []string
	ApprovedOrganizations []string
}

