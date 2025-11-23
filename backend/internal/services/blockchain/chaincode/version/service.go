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

package version

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/registry"
	"go.uber.org/zap"
)

// Service handles version management business logic
type Service struct {
	repo            *Repository
	registryService *registry.Service
	logger          *zap.Logger
}

// NewService creates a new version management service
func NewService(repo *Repository, registryService *registry.Service, logger *zap.Logger) *Service {
	return &Service{
		repo:            repo,
		registryService: registryService,
		logger:          logger,
	}
}

// CreateTagRequest for creating a version tag
type CreateTagRequest struct {
	ChaincodeVersionID uuid.UUID
	TagName            string
	TagType            string // version, alias, custom
	Description        *string
	CreatedBy          *uuid.UUID
}

// CreateTag creates a new version tag
func (s *Service) CreateTag(ctx context.Context, req *CreateTagRequest) (*VersionTag, error) {
	// Validate tag name
	if req.TagName == "" {
		return nil, fmt.Errorf("tag name is required")
	}

	// Check if tag already exists for this version
	tags, err := s.repo.GetVersionTagsByVersionID(ctx, req.ChaincodeVersionID)
	if err == nil {
		for _, tag := range tags {
			if tag.TagName == req.TagName && tag.IsActive {
				return nil, fmt.Errorf("tag %s already exists for this version", req.TagName)
			}
		}
	}

	tag := &VersionTag{
		ID:                uuid.New(),
		ChaincodeVersionID: req.ChaincodeVersionID,
		TagName:           req.TagName,
		TagType:           req.TagType,
		Description:       req.Description,
		IsActive:          true,
		CreatedBy:         req.CreatedBy,
	}

	if req.TagType == "" {
		tag.TagType = "version"
	}

	if err := s.repo.CreateVersionTag(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	s.logger.Info("Created version tag",
		zap.String("tag", req.TagName),
		zap.String("version_id", req.ChaincodeVersionID.String()),
	)

	return tag, nil
}

// GetTags retrieves tags for a version
func (s *Service) GetTags(ctx context.Context, versionID uuid.UUID) ([]*VersionTag, error) {
	return s.repo.GetVersionTagsByVersionID(ctx, versionID)
}

// GetVersionByTag retrieves a version by tag name
func (s *Service) GetVersionByTag(ctx context.Context, chaincodeName, channelName, tagName string) (*uuid.UUID, error) {
	return s.repo.GetVersionByTag(ctx, chaincodeName, channelName, tagName)
}

// CreateDependencyRequest for creating a version dependency
type CreateDependencyRequest struct {
	ChaincodeVersionID uuid.UUID
	DependencyName    string
	DependencyVersion  string
	DependencyType     string // chaincode, library, external
	IsRequired         bool
	Metadata           map[string]interface{}
}

// CreateDependency creates a new version dependency
func (s *Service) CreateDependency(ctx context.Context, req *CreateDependencyRequest) (*VersionDependency, error) {
	if req.DependencyName == "" {
		return nil, fmt.Errorf("dependency name is required")
	}
	if req.DependencyVersion == "" {
		return nil, fmt.Errorf("dependency version is required")
	}

	if req.DependencyType == "" {
		req.DependencyType = "chaincode"
	}

	var metadataJSON json.RawMessage
	if req.Metadata != nil {
		metadataJSON, _ = json.Marshal(req.Metadata)
	}

	dep := &VersionDependency{
		ID:                uuid.New(),
		ChaincodeVersionID: req.ChaincodeVersionID,
		DependencyName:    req.DependencyName,
		DependencyVersion: req.DependencyVersion,
		DependencyType:    req.DependencyType,
		IsRequired:        req.IsRequired,
		Metadata:          metadataJSON,
	}

	if err := s.repo.CreateVersionDependency(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed to create dependency: %w", err)
	}

	return dep, nil
}

// GetDependencies retrieves dependencies for a version
func (s *Service) GetDependencies(ctx context.Context, versionID uuid.UUID) ([]*VersionDependency, error) {
	return s.repo.GetVersionDependencies(ctx, versionID)
}

// CreateReleaseNoteRequest for creating release notes
type CreateReleaseNoteRequest struct {
	ChaincodeVersionID uuid.UUID
	Title             string
	Content           string
	ReleaseType       string // major, minor, patch, hotfix
	BreakingChanges   []string
	NewFeatures       []string
	BugFixes          []string
	Improvements      []string
	CreatedBy         *uuid.UUID
}

// CreateReleaseNote creates a new release note
func (s *Service) CreateReleaseNote(ctx context.Context, req *CreateReleaseNoteRequest) (*VersionReleaseNote, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if req.ReleaseType == "" {
		req.ReleaseType = "patch"
	}

	note := &VersionReleaseNote{
		ID:                uuid.New(),
		ChaincodeVersionID: req.ChaincodeVersionID,
		Title:             req.Title,
		Content:           req.Content,
		ReleaseType:       req.ReleaseType,
		BreakingChanges:   req.BreakingChanges,
		NewFeatures:       req.NewFeatures,
		BugFixes:          req.BugFixes,
		Improvements:      req.Improvements,
		CreatedBy:         req.CreatedBy,
	}

	if err := s.repo.CreateVersionReleaseNote(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create release note: %w", err)
	}

	return note, nil
}

// GetReleaseNote retrieves release note for a version
func (s *Service) GetReleaseNote(ctx context.Context, versionID uuid.UUID) (*VersionReleaseNote, error) {
	return s.repo.GetVersionReleaseNote(ctx, versionID)
}

// CompareVersions compares two versions
func (s *Service) CompareVersions(ctx context.Context, version1, version2 string) (int, error) {
	// Use database function for comparison
	result, err := s.repo.CompareVersions(ctx, version1, version2)
	if err != nil {
		return 0, fmt.Errorf("failed to compare versions: %w", err)
	}

	return result, nil
}

// CompareVersionRequest for comparing two versions
type CompareVersionRequest struct {
	FromVersionID uuid.UUID
	ToVersionID   uuid.UUID
}

// CompareVersionsByID compares two versions by their IDs
func (s *Service) CompareVersionsByID(ctx context.Context, req *CompareVersionRequest) (*VersionComparison, error) {
	// Get version details
	fromVersion, err := s.registryService.GetVersionByID(ctx, req.FromVersionID)
	if err != nil {
		return nil, fmt.Errorf("from version not found: %w", err)
	}

	toVersion, err := s.registryService.GetVersionByID(ctx, req.ToVersionID)
	if err != nil {
		return nil, fmt.Errorf("to version not found: %w", err)
	}

	// Compare semantic versions
	comparison, err := s.repo.CompareVersions(ctx, fromVersion.Version, toVersion.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to compare versions: %w", err)
	}

	// Determine comparison type
	var comparisonType string
	if comparison < 0 {
		comparisonType = "upgrade"
	} else if comparison > 0 {
		comparisonType = "downgrade"
	} else {
		comparisonType = "sidegrade"
	}

	// Get release notes for target version
	toNote, _ := s.repo.GetVersionReleaseNote(ctx, req.ToVersionID)

	// Calculate counts
	breakingChangesCount := 0
	newFeaturesCount := 0
	bugFixesCount := 0

	if toNote != nil {
		breakingChangesCount = len(toNote.BreakingChanges)
		newFeaturesCount = len(toNote.NewFeatures)
		bugFixesCount = len(toNote.BugFixes)
	}

	// Create comparison record
	comp := &VersionComparison{
		ID:                  uuid.New(),
		FromVersionID:       req.FromVersionID,
		ToVersionID:         req.ToVersionID,
		ComparisonType:      comparisonType,
		BreakingChangesCount: breakingChangesCount,
		NewFeaturesCount:    newFeaturesCount,
		BugFixesCount:       bugFixesCount,
	}

	changesSummary := fmt.Sprintf("From %s to %s (%s)", fromVersion.Version, toVersion.Version, comparisonType)
	comp.ChangesSummary = &changesSummary

	// Store metadata
	metadata := map[string]interface{}{
		"from_version": fromVersion.Version,
		"to_version":   toVersion.Version,
		"comparison":   comparison,
	}
	comp.Metadata, _ = json.Marshal(metadata)

	if err := s.repo.CreateVersionComparison(ctx, comp); err != nil {
		s.logger.Warn("Failed to create version comparison record", zap.Error(err))
		// Continue anyway
	}

	return comp, nil
}

// GetLatestVersion gets the latest version for a chaincode
func (s *Service) GetLatestVersion(ctx context.Context, chaincodeName, channelName string) (*uuid.UUID, error) {
	return s.repo.GetLatestVersion(ctx, chaincodeName, channelName)
}

// GetVersionComparisons retrieves comparisons for a version
func (s *Service) GetVersionComparisons(ctx context.Context, versionID uuid.UUID) ([]*VersionComparison, error) {
	return s.repo.GetVersionComparisons(ctx, versionID)
}

// GetVersionHistory retrieves version history for a chaincode
func (s *Service) GetVersionHistory(ctx context.Context, chaincodeName, channelName string) ([]*registry.ChaincodeVersion, error) {
	filters := &registry.VersionFilters{
		Name:        &chaincodeName,
		ChannelName: &channelName,
		Limit:       100, // Get all versions
	}

	versions, err := s.registryService.ListVersions(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get version history: %w", err)
	}

	// Sort by semantic version (using database function would be better, but this works)
	// For now, return as-is (database should handle ordering)
	return versions, nil
}

