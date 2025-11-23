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
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/backend/internal/infrastructure/admin"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode"
	"go.uber.org/zap"
)

// LifecycleHandler handles chaincode lifecycle operations
type LifecycleHandler struct {
	lifecycleService *chaincode.LifecycleService
	adminClient     *admin.Client
	logger          *zap.Logger
}

// NewLifecycleHandler creates a new chaincode lifecycle handler
func NewLifecycleHandler(lifecycleService *chaincode.LifecycleService, adminClient *admin.Client, logger *zap.Logger) *LifecycleHandler {
	return &LifecycleHandler{
		lifecycleService: lifecycleService,
		adminClient:     adminClient,
		logger:          logger,
	}
}

// respondJSON sends a JSON response
func (h *LifecycleHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *LifecycleHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// respondSuccess sends a success response
func (h *LifecycleHandler) respondSuccess(w http.ResponseWriter, status int, data interface{}) {
	h.respondJSON(w, status, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// Install installs a chaincode package
// @Summary Install chaincode
// @Description Install a chaincode package via Admin Service
// @Tags chaincode
// @Accept json
// @Produce json
// @Param request body InstallRequest true "Install request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/install [post]
func (h *LifecycleHandler) Install(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PackagePath string `json:"packagePath"`
		Label       string `json:"label,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if req.PackagePath == "" {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "packagePath is required")
		return
	}

	// Extract IP address and user agent
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	packageID, err := h.lifecycleService.InstallChaincode(r.Context(), req.PackagePath, req.Label, &ipAddress, &userAgent)
	if err != nil {
		h.logger.Error("Failed to install chaincode", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusCreated, map[string]string{
		"packageId": packageID,
	})
}

// Approve approves a chaincode definition
// @Summary Approve chaincode
// @Description Approve a chaincode definition via Admin Service
// @Tags chaincode
// @Accept json
// @Produce json
// @Param request body ApproveRequest true "Approve request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/approve [post]
func (h *LifecycleHandler) Approve(w http.ResponseWriter, r *http.Request) {
	var req chaincode.ApproveChaincodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if req.ChannelName == "" || req.Name == "" || req.Version == "" || req.Sequence == 0 {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "channelName, name, version, and sequence are required")
		return
	}

	// Extract IP address and user agent
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	if err := h.lifecycleService.ApproveChaincode(r.Context(), &req, &ipAddress, &userAgent); err != nil {
		h.logger.Error("Failed to approve chaincode", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]string{
		"message": "Chaincode approved successfully",
	})
}

// Commit commits a chaincode definition
// @Summary Commit chaincode
// @Description Commit a chaincode definition via Admin Service
// @Tags chaincode
// @Accept json
// @Produce json
// @Param request body CommitRequest true "Commit request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/commit [post]
func (h *LifecycleHandler) Commit(w http.ResponseWriter, r *http.Request) {
	var req chaincode.CommitChaincodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if req.ChannelName == "" || req.Name == "" || req.Version == "" || req.Sequence == 0 {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "channelName, name, version, and sequence are required")
		return
	}

	// Extract IP address and user agent
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = forwarded
	}
	userAgent := r.Header.Get("User-Agent")

	if err := h.lifecycleService.CommitChaincode(r.Context(), &req, &ipAddress, &userAgent); err != nil {
		h.logger.Error("Failed to commit chaincode", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]string{
		"message": "Chaincode committed successfully",
	})
}

// ListInstalled lists installed chaincodes
// @Summary List installed chaincodes
// @Description List all installed chaincodes via Admin Service
// @Tags chaincode
// @Produce json
// @Param peer query string false "Peer address"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/installed [get]
func (h *LifecycleHandler) ListInstalled(w http.ResponseWriter, r *http.Request) {
	peer := r.URL.Query().Get("peer")

	chaincodes, err := h.lifecycleService.ListInstalled(r.Context(), peer)
	if err != nil {
		h.logger.Error("Failed to list installed chaincodes", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, chaincodes)
}

// ListCommitted lists committed chaincodes
// @Summary List committed chaincodes
// @Description List all committed chaincodes on a channel via Admin Service
// @Tags chaincode
// @Produce json
// @Param channel query string false "Channel name"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/committed [get]
func (h *LifecycleHandler) ListCommitted(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		channel = "ibnchannel" // Default
	}

	chaincodes, err := h.lifecycleService.ListCommitted(r.Context(), channel)
	if err != nil {
		h.logger.Error("Failed to list committed chaincodes", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, chaincodes)
}

// GetCommittedInfo gets information about a committed chaincode
// @Summary Get committed chaincode info
// @Description Get detailed information about a committed chaincode via Admin Service
// @Tags chaincode
// @Produce json
// @Param name query string true "Chaincode name"
// @Param channel query string false "Channel name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/committed/{name} [get]
func (h *LifecycleHandler) GetCommittedInfo(w http.ResponseWriter, r *http.Request) {
	chaincodeName := chi.URLParam(r, "name")
	if chaincodeName == "" {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "chaincode name is required")
		return
	}

	channel := r.URL.Query().Get("channel")
	if channel == "" {
		channel = "ibnchannel" // Default
	}

	info, err := h.lifecycleService.GetCommittedInfo(r.Context(), channel, chaincodeName)
	if err != nil {
		if err.Error() == "chaincode not found" {
			h.respondError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		} else {
			h.logger.Error("Failed to get committed chaincode info", zap.Error(err))
			h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
		return
	}

	h.respondSuccess(w, http.StatusOK, info)
}

// UploadPackage handles file upload for chaincode package
// Proxies multipart/form-data request to Admin Service
func (h *LifecycleHandler) UploadPackage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Failed to parse multipart form: "+err.Error())
		return
	}

	// Get file from form
	file, header, err := r.FormFile("package")
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "File 'package' is required: "+err.Error())
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to read file: "+err.Error())
		return
	}

	// Upload to Admin Service
	filePath, err := h.adminClient.UploadPackage(r.Context(), fileData, header.Filename)
	if err != nil {
		h.logger.Error("Failed to upload package to admin service", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload package: "+err.Error())
		return
	}

	h.respondSuccess(w, http.StatusCreated, map[string]string{
		"filePath": filePath,
		"filename": header.Filename,
		"size":     fmt.Sprintf("%d", header.Size),
	})
}

