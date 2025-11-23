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

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/auth"
	"go.uber.org/zap"
)

// AuthHandler handles authentication operations
type AuthHandler struct {
	authService    *auth.Service
	logger         *zap.Logger
	backendBaseURL string
	httpClient     *http.Client
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service, backendBaseURL string, logger *zap.Logger) *AuthHandler {
	if backendBaseURL == "" {
		backendBaseURL = "http://ibn-backend:8080"
	}
	backendBaseURL = strings.TrimRight(backendBaseURL, "/")

	return &AuthHandler{
		authService:    authService,
		logger:         logger,
		backendBaseURL: backendBaseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Login godoc
// @Summary Login
// @Description Authenticate user and get JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Username and password are required",
			nil,
		))
		return
	}

	// Login
	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		// Check if it's invalid credentials
		if strings.Contains(err.Error(), "invalid credentials") {
			h.logger.Debug("Login failed: invalid credentials", zap.String("username", req.Username))
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid username or password",
				nil,
			))
			return
		}

		// Check if account is inactive
		if strings.Contains(err.Error(), "inactive") {
			respondJSON(w, http.StatusForbidden, models.NewErrorResponse(
				models.ErrCodeForbidden,
				"User account is inactive",
				nil,
			))
			return
		}

		h.logger.Error("Login failed", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to login",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// Register godoc
// @Summary Register
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register request"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Username, email, and password are required",
			nil,
		))
		return
	}

	// Register
	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		// Check if it's duplicate email/username
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			respondJSON(w, http.StatusConflict, models.NewErrorResponse(
				models.ErrCodeConflict,
				"Username or email already exists",
				nil,
			))
			return
		}

		h.logger.Error("Registration failed", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to register user",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(user))
}

// RefreshToken godoc
// @Summary Refresh token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Validate request
	if req.RefreshToken == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Refresh token is required",
			nil,
		))
		return
	}

	// Refresh token
	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "expired") {
			respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
				models.ErrCodeUnauthorized,
				"Invalid or expired refresh token",
				nil,
			))
			return
		}

		h.logger.Error("Refresh token failed", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to refresh token",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Retrieve profile information for the authenticated user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("userID")
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"Unauthorized",
			nil,
		))
		return
	}

	profile, err := h.authService.GetProfile(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get profile", zap.String("user_id", userID), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to get profile",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(profile))
}

// UploadAvatar godoc
// @Summary Upload avatar
// @Description Upload user avatar and proxy request to backend service
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Router /auth/avatar [post]
func (h *AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"Unauthorized",
			nil,
		))
		return
	}

	if err := r.ParseMultipartForm(6 << 20); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid multipart form data",
			err.Error(),
		))
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	file, header, err := r.FormFile("avatar")
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Avatar file is required",
			err.Error(),
		))
		return
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar", header.Filename)
	if err != nil {
		h.logger.Error("Failed to create multipart form file", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to process avatar",
			err.Error(),
		))
		return
	}

	if _, err := io.Copy(part, file); err != nil {
		h.logger.Error("Failed to copy avatar file", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to process avatar",
			err.Error(),
		))
		return
	}

	if err := writer.Close(); err != nil {
		h.logger.Error("Failed to finalize multipart writer", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to process avatar",
			err.Error(),
		))
		return
	}

	url := fmt.Sprintf("%s/api/v1/auth/avatar", h.backendBaseURL)
	proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, url, &body)
	if err != nil {
		h.logger.Error("Failed to create backend avatar request", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to upload avatar",
			err.Error(),
		))
		return
	}

	proxyReq.Header.Set("Content-Type", writer.FormDataContentType())
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		proxyReq.Header.Set("Authorization", authHeader)
	}

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		h.logger.Error("Failed to proxy avatar upload", zap.Error(err))
		respondJSON(w, http.StatusBadGateway, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to upload avatar",
			err.Error(),
		))
		return
	}
	defer resp.Body.Close()

	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		h.logger.Warn("Failed to write avatar response", zap.Error(err))
	}
}

// GenerateAPIKey godoc
// @Summary Generate API key
// @Description Create a new API key for authenticated user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.APIKeyRequest true "API key request"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Security BearerAuth
// @Router /auth/api-keys [post]
func (h *AuthHandler) GenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
			nil,
		))
		return
	}

	var req models.APIKeyRequest
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
			"API key name is required",
			nil,
		))
		return
	}

	// Generate API key
	response, err := h.authService.GenerateAPIKey(r.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to generate API key", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to generate API key",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(response))
}

// ListAPIKeys godoc
// @Summary List API keys
// @Description List all API keys for authenticated user
// @Tags auth
// @Produce json
// @Success 200 {object} models.APIResponse
// @Security BearerAuth
// @Router /auth/api-keys [get]
func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
			nil,
		))
		return
	}

	// List API keys
	keys, err := h.authService.ListAPIKeys(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to list API keys", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to list API keys",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(keys))
}

// RevokeAPIKey godoc
// @Summary Revoke API key
// @Description Revoke an API key
// @Tags auth
// @Param id path string true "API Key ID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /auth/api-keys/{id} [delete]
func (h *AuthHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(string)
	if !ok || userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.NewErrorResponse(
			models.ErrCodeUnauthorized,
			"User not authenticated",
			nil,
		))
		return
	}

	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"API key ID is required",
			nil,
		))
		return
	}

	// Revoke API key
	err := h.authService.RevokeAPIKey(r.Context(), userID, keyID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
				models.ErrCodeNotFound,
				"API key not found",
				nil,
			))
			return
		}

		h.logger.Error("Failed to revoke API key", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalServer,
			"Failed to revoke API key",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]string{
		"message": "API key revoked successfully",
	}))
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
