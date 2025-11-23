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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/auth"
	"github.com/ibn-network/backend/internal/utils"
	"go.uber.org/zap"
)

// Handler handles auth HTTP requests
type Handler struct {
	service *auth.Service
	logger  *zap.Logger
}

// NewHandler creates a new auth handler
func NewHandler(service *auth.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration request"
// @Success 201 {object} auth.User
// @Failure 400 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(&req); validationErrors != nil {
		h.logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.respondJSON(w, http.StatusCreated, user)
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login request"
// @Success 200 {object} auth.LoginResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if validationErrors := utils.ValidateStruct(&req); validationErrors != nil {
		h.logger.Warn("Validation failed", zap.Any("errors", validationErrors))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	response, err := h.service.Login(r.Context(), &req)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err))
		h.respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req auth.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("Token refresh failed", zap.Error(err))
		h.respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req auth.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		h.logger.Error("Logout failed", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// CreateAPIKey handles API key creation
func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req auth.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.service.CreateAPIKey(r.Context(), userID, &req)
	if err != nil {
		h.logger.Error("API key creation failed", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get authenticated user's profile information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} auth.User
// @Failure 401 {object} map[string]string
// @Router /profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "User not found")
		return
	}

	// Wrap response in success format
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    user,
	})
}

// UploadAvatar handles avatar upload
// @Summary Upload user avatar
// @Description Upload and update user's avatar image
// @Tags auth
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Avatar image file (JPG, PNG, WebP, max 5MB)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/avatar [post]
func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form (max 5MB)
	const maxSize = 5 << 20 // 5MB
	if err := r.ParseMultipartForm(maxSize); err != nil {
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Failed to parse form data")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("avatar")
	if err != nil {
		h.logger.Error("Failed to get file from form", zap.Error(err))
		h.respondError(w, http.StatusBadRequest, "Avatar file is required")
		return
	}
	defer file.Close()

	// Validate file type
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	if !validTypes[header.Header.Get("Content-Type")] {
		h.respondError(w, http.StatusBadRequest, "Invalid file type. Only JPG, PNG, and WebP are allowed")
		return
	}

	// Validate file size
	if header.Size > maxSize {
		h.respondError(w, http.StatusBadRequest, "File size exceeds 5MB limit")
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.logger.Error("Failed to create upload directory", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to save avatar")
		return
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s-%d%s", userID.String(), time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("Failed to create destination file", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Failed to save avatar")
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		h.logger.Error("Failed to copy file", zap.Error(err))
		os.Remove(filePath) // Cleanup on error
		h.respondError(w, http.StatusInternalServerError, "Failed to save avatar")
		return
	}

	// Generate avatar URL (relative path or full URL depending on your setup)
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)

	// Update user avatar in database
	if err := h.service.UpdateAvatar(r.Context(), userID, avatarURL); err != nil {
		h.logger.Error("Failed to update avatar", zap.Error(err))
		os.Remove(filePath) // Cleanup on error
		h.respondError(w, http.StatusInternalServerError, "Failed to update avatar")
		return
	}

	// Return success response
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"avatarUrl": avatarURL,
		},
	})
}

// Helper functions
func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, message string) {
	h.respondJSON(w, status, map[string]string{"error": message})
}

