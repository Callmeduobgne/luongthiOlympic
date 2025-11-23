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

package acl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	aclservice "github.com/ibn-network/api-gateway/internal/services/acl"
	"go.uber.org/zap"
)

// ACLHandler handles ACL operations
type ACLHandler struct {
	aclService *aclservice.Service
	logger     *zap.Logger
}

// NewACLHandler creates a new ACL handler
func NewACLHandler(aclService *aclservice.Service, logger *zap.Logger) *ACLHandler {
	return &ACLHandler{
		aclService: aclService,
		logger:     logger,
	}
}

// ListPolicies godoc
// @Summary List ACL policies
// @Description Get a list of all ACL policies with pagination
// @Tags acl
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(50)
// @Param resourceType query string false "Filter by resource type"
// @Success 200 {object} models.APIResponse{data=models.ListPoliciesResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/policies [get]
func (h *ACLHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page := 1
	pageSize := 50
	resourceType := ""

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	resourceType = r.URL.Query().Get("resourceType")

	policies, err := h.aclService.ListPolicies(r.Context(), page, pageSize, resourceType)
	if err != nil {
		h.logger.Error("Failed to list policies", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list policies",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(policies))
}

// CreatePolicy godoc
// @Summary Create ACL policy
// @Description Create a new ACL policy
// @Tags acl
// @Accept json
// @Produce json
// @Param request body models.CreatePolicyRequest true "Policy creation request"
// @Success 201 {object} models.APIResponse{data=models.ACLPolicy}
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/policies [post]
func (h *ACLHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Policy name is required",
			nil,
		))
		return
	}

	if req.ResourceType == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Resource type is required",
			nil,
		))
		return
	}

	if len(req.Actions) == 0 {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"At least one action is required",
			nil,
		))
		return
	}

	policy, err := h.aclService.CreatePolicy(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create policy", zap.String("name", req.Name), zap.Error(err))
		if strings.Contains(err.Error(), "already exists") {
			respondJSON(w, http.StatusConflict, models.NewErrorResponse(
				models.ErrCodeConflict,
				fmt.Sprintf("Policy with name '%s' already exists", req.Name),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create policy",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(policy))
}

// GetPolicy godoc
// @Summary Get policy details
// @Description Get details of a specific ACL policy
// @Tags acl
// @Produce json
// @Param id path string true "Policy ID"
// @Success 200 {object} models.APIResponse{data=models.ACLPolicy}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/policies/{id} [get]
func (h *ACLHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	if policyID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Policy ID is required",
			nil,
		))
		return
	}

	policy, err := h.aclService.GetPolicy(r.Context(), policyID)
	if err != nil {
		h.logger.Error("Failed to get policy", zap.String("policyId", policyID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Policy '%s' not found", policyID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to get policy",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(policy))
}

// UpdatePolicy godoc
// @Summary Update ACL policy
// @Description Update an existing ACL policy
// @Tags acl
// @Accept json
// @Produce json
// @Param id path string true "Policy ID"
// @Param request body models.UpdatePolicyRequest true "Policy update request"
// @Success 200 {object} models.APIResponse{data=models.ACLPolicy}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/policies/{id} [patch]
func (h *ACLHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	if policyID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Policy ID is required",
			nil,
		))
		return
	}

	var req models.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	policy, err := h.aclService.UpdatePolicy(r.Context(), policyID, &req)
	if err != nil {
		h.logger.Error("Failed to update policy", zap.String("policyId", policyID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Policy '%s' not found", policyID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to update policy",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(policy))
}

// DeletePolicy godoc
// @Summary Delete ACL policy
// @Description Delete an ACL policy
// @Tags acl
// @Produce json
// @Param id path string true "Policy ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/policies/{id} [delete]
func (h *ACLHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	if policyID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Policy ID is required",
			nil,
		))
		return
	}

	err := h.aclService.DeletePolicy(r.Context(), policyID)
	if err != nil {
		h.logger.Error("Failed to delete policy", zap.String("policyId", policyID), zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				fmt.Sprintf("Policy '%s' not found", policyID),
				err.Error(),
			))
			return
		}
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to delete policy",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": "Policy deleted successfully",
	}))
}

// ListPermissions godoc
// @Summary List permissions
// @Description Get a list of all predefined permissions
// @Tags acl
// @Produce json
// @Param resourceType query string false "Filter by resource type"
// @Success 200 {object} models.APIResponse{data=models.ListPermissionsResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/permissions [get]
func (h *ACLHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	resourceType := r.URL.Query().Get("resourceType")

	permissions, err := h.aclService.ListPermissions(r.Context(), resourceType)
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list permissions",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(permissions))
}

// CheckPermission godoc
// @Summary Check permission
// @Description Check if a user has permission to perform an action on a resource
// @Tags acl
// @Accept json
// @Produce json
// @Param request body models.CheckPermissionRequest true "Permission check request"
// @Success 200 {object} models.APIResponse{data=models.CheckPermissionResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /acl/check [post]
func (h *ACLHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	var req models.CheckPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.UserID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"User ID is required",
			nil,
		))
		return
	}

	if req.ResourceType == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Resource type is required",
			nil,
		))
		return
	}

	if req.Resource == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Resource is required",
			nil,
		))
		return
	}

	if req.Action == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Action is required",
			nil,
		))
		return
	}

	result, err := h.aclService.CheckPermission(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to check permission", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to check permission",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(result))
}

// respondJSON is a helper function to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

