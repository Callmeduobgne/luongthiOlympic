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

package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/identity"
	"go.uber.org/zap"
)

// UserHandler handles user/CA management operations
type UserHandler struct {
	identityService *identity.Service
	caService       CAServiceInterface
	logger          *zap.Logger
}

// CAServiceInterface defines CA service interface
type CAServiceInterface interface {
	Enroll(ctx context.Context, req *models.CAEnrollRequest) (*models.CAEnrollResponse, error)
	Register(ctx context.Context, req *models.CARegisterRequest) (*models.CARegisterResponse, error)
	Reenroll(ctx context.Context, username string) (*models.CAEnrollResponse, error)
	Revoke(ctx context.Context, username, reason string) error
	GetCertificate(ctx context.Context, username string) (string, error)
}

// NewUserHandler creates a new user handler
func NewUserHandler(identityService *identity.Service, caService CAServiceInterface, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		identityService: identityService,
		caService:       caService,
		logger:          logger,
	}
}

// ListUsers godoc
// @Summary List users
// @Description List all users in MSP
// @Tags users
// @Accept json
// @Produce json
// @Param affiliation query string false "Filter by affiliation"
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	affiliation := r.URL.Query().Get("affiliation")

	users, err := h.identityService.ListUsers(r.Context(), affiliation)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to list users",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(users))
}

// GetUser godoc
// @Summary Get user information
// @Description Get user information by username
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (username)"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"User ID is required",
			nil,
		))
		return
	}

	user, err := h.identityService.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user", zap.String("userID", userID), zap.Error(err))
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			fmt.Sprintf("User '%s' not found", userID),
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(user))
}

// Enroll godoc
// @Summary Enroll user with Fabric CA
// @Description Enroll a user with Fabric CA server
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CAEnrollRequest true "Enrollment request"
// @Success 200 {object} models.APIResponse{data=models.CAEnrollResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/enroll [post]
func (h *UserHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	if h.caService == nil {
		respondJSON(w, http.StatusServiceUnavailable, models.NewErrorResponse(
			"SERVICE_UNAVAILABLE",
			"Fabric CA service is not configured",
			nil,
		))
		return
	}

	var req models.CAEnrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	response, err := h.caService.Enroll(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to enroll user", zap.String("username", req.Username), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to enroll user",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// Register godoc
// @Summary Register new user with Fabric CA
// @Description Register a new identity with Fabric CA server
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CARegisterRequest true "Registration request"
// @Success 201 {object} models.APIResponse{data=models.CARegisterResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if h.caService == nil {
		respondJSON(w, http.StatusServiceUnavailable, models.NewErrorResponse(
			"SERVICE_UNAVAILABLE",
			"Fabric CA service is not configured",
			nil,
		))
		return
	}

	var req models.CARegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	response, err := h.caService.Register(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to register user", zap.String("username", req.Username), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to register user",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(response))
}

// Revoke godoc
// @Summary Revoke user certificate
// @Description Revoke a user certificate with Fabric CA
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (username)"
// @Param request body models.CARevokeRequest false "Revocation request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/{id}/revoke [delete]
func (h *UserHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	if h.caService == nil {
		respondJSON(w, http.StatusServiceUnavailable, models.NewErrorResponse(
			"SERVICE_UNAVAILABLE",
			"Fabric CA service is not configured",
			nil,
		))
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"User ID is required",
			nil,
		))
		return
	}

	var req models.CARevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	reason := req.Reason
	if reason == "" {
		reason = "unspecified"
	}

	err := h.caService.Revoke(r.Context(), userID, reason)
	if err != nil {
		h.logger.Error("Failed to revoke certificate", zap.String("userID", userID), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to revoke certificate",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": fmt.Sprintf("Certificate for user '%s' has been revoked", userID),
	}))
}

// Reenroll godoc
// @Summary Re-enroll user certificate
// @Description Re-enroll (renew) a user certificate with Fabric CA
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (username)"
// @Success 200 {object} models.APIResponse{data=models.CAEnrollResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/{id}/reenroll [post]
func (h *UserHandler) Reenroll(w http.ResponseWriter, r *http.Request) {
	if h.caService == nil {
		respondJSON(w, http.StatusServiceUnavailable, models.NewErrorResponse(
			"SERVICE_UNAVAILABLE",
			"Fabric CA service is not configured",
			nil,
		))
		return
	}

	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"User ID is required",
			nil,
		))
		return
	}

	response, err := h.caService.Reenroll(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to re-enroll user", zap.String("userID", userID), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeTransactionFailed,
			"Failed to re-enroll user",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetCertificate godoc
// @Summary Get user certificate
// @Description Get a user's certificate
// @Tags users
// @Produce json
// @Param id path string true "User ID (username)"
// @Success 200 {object} models.APIResponse{data=object}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /users/{id}/certificate [get]
func (h *UserHandler) GetCertificate(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"User ID is required",
			nil,
		))
		return
	}

	// Try CA service first, fallback to identity service
	var cert string
	var err error

	if h.caService != nil {
		cert, err = h.caService.GetCertificate(r.Context(), userID)
	} else {
		// Fallback to file system
		respondJSON(w, http.StatusServiceUnavailable, models.NewErrorResponse(
			"SERVICE_UNAVAILABLE",
			"Certificate retrieval requires CA service or file system access",
			nil,
		))
		return
	}

	if err != nil {
		h.logger.Error("Failed to get certificate", zap.String("userID", userID), zap.Error(err))
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Certificate not found",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]string{
		"username":    userID,
		"certificate": cert,
	}))
}

// respondJSON is a helper function to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

