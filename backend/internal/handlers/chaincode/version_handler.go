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
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/version"
	"go.uber.org/zap"
)

// VersionHandler handles version management operations
type VersionHandler struct {
	versionService *version.Service
	logger         *zap.Logger
}

// NewVersionHandler creates a new version handler
func NewVersionHandler(versionService *version.Service, logger *zap.Logger) *VersionHandler {
	return &VersionHandler{
		versionService: versionService,
		logger:         logger,
	}
}

// CreateTagRequest represents the request body for creating a tag
type CreateTagRequest struct {
	ChaincodeVersionID uuid.UUID `json:"chaincode_version_id"`
	TagName            string    `json:"tag_name"`
	TagType            string    `json:"tag_type,omitempty"`
	Description        *string   `json:"description,omitempty"`
}

// CreateTag creates a new version tag
func (h *VersionHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	var userID *uuid.UUID
	if userIDVal != nil {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.ChaincodeVersionID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "chaincode_version_id is required")
		return
	}
	if req.TagName == "" {
		h.respondError(w, http.StatusBadRequest, "tag_name is required")
		return
	}

	tagReq := &version.CreateTagRequest{
		ChaincodeVersionID: req.ChaincodeVersionID,
		TagName:            req.TagName,
		TagType:            req.TagType,
		Description:        req.Description,
		CreatedBy:          userID,
	}

	tag, err := h.versionService.CreateTag(r.Context(), tagReq)
	if err != nil {
		h.logger.Error("Failed to create tag", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to create tag: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, tag)
}

// GetTags retrieves tags for a version
func (h *VersionHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	versionIDStr := chi.URLParam(r, "version_id")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid version ID")
		return
	}

	tags, err := h.versionService.GetTags(r.Context(), versionID)
	if err != nil {
		h.logger.Error("Failed to get tags", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get tags: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"tags": tags,
		"count": len(tags),
	})
}

// GetVersionByTag retrieves a version by tag name
func (h *VersionHandler) GetVersionByTag(w http.ResponseWriter, r *http.Request) {
	chaincodeName := chi.URLParam(r, "chaincode_name")
	channelName := r.URL.Query().Get("channel")
	tagName := chi.URLParam(r, "tag_name")

	if chaincodeName == "" || channelName == "" || tagName == "" {
		h.respondError(w, http.StatusBadRequest, "chaincode_name, channel, and tag_name are required")
		return
	}

	versionID, err := h.versionService.GetVersionByTag(r.Context(), chaincodeName, channelName, tagName)
	if err != nil {
		h.logger.Error("Failed to get version by tag", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "version not found for tag: "+tagName)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"version_id": versionID,
		"tag_name":   tagName,
	})
}

// CreateDependencyRequest represents the request body for creating a dependency
type CreateDependencyRequest struct {
	ChaincodeVersionID uuid.UUID              `json:"chaincode_version_id"`
	DependencyName     string                 `json:"dependency_name"`
	DependencyVersion  string                 `json:"dependency_version"`
	DependencyType     string                 `json:"dependency_type,omitempty"`
	IsRequired         bool                   `json:"is_required,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDependency creates a new version dependency
func (h *VersionHandler) CreateDependency(w http.ResponseWriter, r *http.Request) {
	var req CreateDependencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.ChaincodeVersionID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "chaincode_version_id is required")
		return
	}
	if req.DependencyName == "" {
		h.respondError(w, http.StatusBadRequest, "dependency_name is required")
		return
	}
	if req.DependencyVersion == "" {
		h.respondError(w, http.StatusBadRequest, "dependency_version is required")
		return
	}

	depReq := &version.CreateDependencyRequest{
		ChaincodeVersionID: req.ChaincodeVersionID,
		DependencyName:     req.DependencyName,
		DependencyVersion:  req.DependencyVersion,
		DependencyType:     req.DependencyType,
		IsRequired:         req.IsRequired,
		Metadata:           req.Metadata,
	}

	dep, err := h.versionService.CreateDependency(r.Context(), depReq)
	if err != nil {
		h.logger.Error("Failed to create dependency", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to create dependency: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, dep)
}

// GetDependencies retrieves dependencies for a version
func (h *VersionHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	versionIDStr := chi.URLParam(r, "version_id")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid version ID")
		return
	}

	deps, err := h.versionService.GetDependencies(r.Context(), versionID)
	if err != nil {
		h.logger.Error("Failed to get dependencies", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get dependencies: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"dependencies": deps,
		"count":       len(deps),
	})
}

// CreateReleaseNoteRequest represents the request body for creating release notes
type CreateReleaseNoteRequest struct {
	ChaincodeVersionID uuid.UUID `json:"chaincode_version_id"`
	Title              string    `json:"title"`
	Content            string    `json:"content"`
	ReleaseType        string    `json:"release_type,omitempty"`
	BreakingChanges    []string  `json:"breaking_changes,omitempty"`
	NewFeatures        []string  `json:"new_features,omitempty"`
	BugFixes           []string  `json:"bug_fixes,omitempty"`
	Improvements       []string  `json:"improvements,omitempty"`
}

// CreateReleaseNote creates a new release note
func (h *VersionHandler) CreateReleaseNote(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	var userID *uuid.UUID
	if userIDVal != nil {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	var req CreateReleaseNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.ChaincodeVersionID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "chaincode_version_id is required")
		return
	}
	if req.Title == "" {
		h.respondError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Content == "" {
		h.respondError(w, http.StatusBadRequest, "content is required")
		return
	}

	noteReq := &version.CreateReleaseNoteRequest{
		ChaincodeVersionID: req.ChaincodeVersionID,
		Title:             req.Title,
		Content:           req.Content,
		ReleaseType:       req.ReleaseType,
		BreakingChanges:   req.BreakingChanges,
		NewFeatures:       req.NewFeatures,
		BugFixes:          req.BugFixes,
		Improvements:      req.Improvements,
		CreatedBy:         userID,
	}

	note, err := h.versionService.CreateReleaseNote(r.Context(), noteReq)
	if err != nil {
		h.logger.Error("Failed to create release note", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to create release note: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, note)
}

// GetReleaseNote retrieves release note for a version
func (h *VersionHandler) GetReleaseNote(w http.ResponseWriter, r *http.Request) {
	versionIDStr := chi.URLParam(r, "version_id")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid version ID")
		return
	}

	note, err := h.versionService.GetReleaseNote(r.Context(), versionID)
	if err != nil {
		h.logger.Error("Failed to get release note", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "release note not found")
		return
	}

	h.respondJSON(w, http.StatusOK, note)
}

// CompareVersionsRequest represents the request body for comparing versions
type CompareVersionsRequest struct {
	FromVersionID uuid.UUID `json:"from_version_id"`
	ToVersionID   uuid.UUID `json:"to_version_id"`
}

// CompareVersions compares two versions
func (h *VersionHandler) CompareVersions(w http.ResponseWriter, r *http.Request) {
	var req CompareVersionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.FromVersionID == uuid.Nil || req.ToVersionID == uuid.Nil {
		h.respondError(w, http.StatusBadRequest, "from_version_id and to_version_id are required")
		return
	}

	compReq := &version.CompareVersionRequest{
		FromVersionID: req.FromVersionID,
		ToVersionID:   req.ToVersionID,
	}

	comparison, err := h.versionService.CompareVersionsByID(r.Context(), compReq)
	if err != nil {
		h.logger.Error("Failed to compare versions", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to compare versions: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, comparison)
}

// GetLatestVersion gets the latest version for a chaincode
func (h *VersionHandler) GetLatestVersion(w http.ResponseWriter, r *http.Request) {
	chaincodeName := chi.URLParam(r, "chaincode_name")
	channelName := r.URL.Query().Get("channel")

	if chaincodeName == "" || channelName == "" {
		h.respondError(w, http.StatusBadRequest, "chaincode_name and channel are required")
		return
	}

	versionID, err := h.versionService.GetLatestVersion(r.Context(), chaincodeName, channelName)
	if err != nil {
		h.logger.Error("Failed to get latest version", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "latest version not found")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"version_id": versionID,
	})
}

// GetVersionHistory retrieves version history for a chaincode
func (h *VersionHandler) GetVersionHistory(w http.ResponseWriter, r *http.Request) {
	chaincodeName := chi.URLParam(r, "chaincode_name")
	channelName := r.URL.Query().Get("channel")

	if chaincodeName == "" || channelName == "" {
		h.respondError(w, http.StatusBadRequest, "chaincode_name and channel are required")
		return
	}

	versions, err := h.versionService.GetVersionHistory(r.Context(), chaincodeName, channelName)
	if err != nil {
		h.logger.Error("Failed to get version history", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get version history: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"versions": versions,
		"count":   len(versions),
	})
}

// GetVersionComparisons retrieves comparisons for a version
func (h *VersionHandler) GetVersionComparisons(w http.ResponseWriter, r *http.Request) {
	versionIDStr := chi.URLParam(r, "version_id")
	versionID, err := uuid.Parse(versionIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid version ID")
		return
	}

	comparisons, err := h.versionService.GetVersionComparisons(r.Context(), versionID)
	if err != nil {
		h.logger.Error("Failed to get version comparisons", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "failed to get version comparisons: "+err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"comparisons": comparisons,
		"count":      len(comparisons),
	})
}

// Helper methods
func (h *VersionHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *VersionHandler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
	})
}

