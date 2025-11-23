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

package handlers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ibn-network/admin-service/internal/services/chaincode"
	"go.uber.org/zap"
)

// ChaincodeHandler handles chaincode lifecycle operations
type ChaincodeHandler struct {
	service      *chaincode.Service
	logger       *zap.Logger
	uploadDir    string
	maxFileSize  int64 // in bytes
}

// NewChaincodeHandler creates a new chaincode handler
func NewChaincodeHandler(service *chaincode.Service, logger *zap.Logger) *ChaincodeHandler {
	// Create upload directory if it doesn't exist
	uploadDir := "/tmp/chaincode-uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Warn("Failed to create upload directory", zap.Error(err))
	}

	return &ChaincodeHandler{
		service:     service,
		logger:      logger,
		uploadDir:   uploadDir,
		maxFileSize: 100 * 1024 * 1024, // 100MB
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, code, message, detail string) {
	respondJSON(w, status, map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
			"detail":  detail,
		},
	})
}

// respondSuccess sends a success response
func respondSuccess(w http.ResponseWriter, status int, data interface{}) {
	respondJSON(w, status, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// UploadPackage handles file upload for chaincode package
// Accepts multipart/form-data with file field "package"
func (h *ChaincodeHandler) UploadPackage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(h.maxFileSize); err != nil {
		h.logger.Error("Failed to parse multipart form", zap.Error(err))
		respondError(w, http.StatusBadRequest, "BAD_REQUEST",
			"Failed to parse multipart form", err.Error())
		return
	}

	// Get file from form
	file, header, err := r.FormFile("package")
	if err != nil {
		h.logger.Error("Failed to get file from form", zap.Error(err))
		respondError(w, http.StatusBadRequest, "BAD_REQUEST",
			"File 'package' is required", err.Error())
		return
	}
	defer file.Close()

	// Validate file extension
	filename := header.Filename
	if !strings.HasSuffix(strings.ToLower(filename), ".tar.gz") &&
		!strings.HasSuffix(strings.ToLower(filename), ".gz") {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST",
			"Invalid file type. Only .tar.gz or .gz files are allowed", "")
		return
	}

	// Validate file size
	if header.Size > h.maxFileSize {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST",
			fmt.Sprintf("File size exceeds maximum allowed size of %d MB", h.maxFileSize/(1024*1024)), "")
		return
	}

	// Generate unique filename to avoid conflicts
	timestamp := time.Now().Format("20060102-150405")
	safeFilename := fmt.Sprintf("%s-%s", timestamp, filename)
	filePath := filepath.Join(h.uploadDir, safeFilename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("Failed to create destination file", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to save uploaded file", err.Error())
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, file); err != nil {
		h.logger.Error("Failed to copy file", zap.Error(err))
		os.Remove(filePath) // Cleanup on error
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to save uploaded file", err.Error())
		return
	}

	// Set file permissions
	if err := os.Chmod(filePath, 0644); err != nil {
		h.logger.Warn("Failed to set file permissions", zap.Error(err))
	}

	// Basic package validation: Check if file is a valid tar.gz
	// Full structure validation (metadata.json, code.tar.gz) will be done by peer CLI during install
	if err := h.validatePackageStructure(filePath); err != nil {
		h.logger.Warn("Package structure validation warning",
			zap.String("filepath", filePath),
			zap.Error(err),
		)
		// Don't fail upload, but log warning
		// Peer CLI will validate during install and return proper error
	}

	h.logger.Info("Chaincode package uploaded successfully",
		zap.String("filename", filename),
		zap.String("filepath", filePath),
		zap.Int64("size", header.Size),
	)

	// Return file path for installation
	respondSuccess(w, http.StatusCreated, map[string]string{
		"filePath": filePath,
		"filename": filename,
		"size":     fmt.Sprintf("%d", header.Size),
	})
}

// Install installs a chaincode package
// Supports both file upload (via UploadPackage) and direct path
func (h *ChaincodeHandler) Install(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PackagePath string `json:"packagePath"`
		Label       string `json:"label,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	if req.PackagePath == "" {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "packagePath is required", "")
		return
	}

	// Verify file exists
	if _, err := os.Stat(req.PackagePath); os.IsNotExist(err) {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST",
			"Package file not found", fmt.Sprintf("File does not exist: %s", req.PackagePath))
		return
	}

	packageID, err := h.service.Install(r.Context(), req.PackagePath, req.Label)
	if err != nil {
		h.logger.Error("Failed to install chaincode", zap.Error(err))
		if strings.Contains(err.Error(), "peer CLI not available") {
			respondError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"Peer CLI is not available. Please ensure admin-service container has peer CLI installed.", err.Error())
		} else if strings.Contains(err.Error(), "already installed") {
			// Chaincode already installed is a valid idempotent operation
			// Extract package ID from error message if possible
			// Error format: "chaincode already installed but failed to extract package ID..."
			// Or we might have package ID in the error
			if packageID == "" {
				// Try to extract from error message
				if idx := strings.Index(err.Error(), "package ID"); idx != -1 {
					// Try to extract package ID from error message
					parts := strings.Split(err.Error(), "'")
					if len(parts) >= 2 {
						packageID = parts[1]
					}
				}
			}
			// If we still don't have package ID, return error
			if packageID == "" {
				respondError(w, http.StatusBadRequest, "ALREADY_INSTALLED",
					"Chaincode already installed but package ID could not be extracted", err.Error())
				return
			}
			// Return success with package ID (idempotent operation)
			respondSuccess(w, http.StatusOK, map[string]string{
				"packageId": packageID,
				"message":    "Chaincode already installed (idempotent operation)",
			})
			return
		} else {
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
				"Failed to install chaincode", err.Error())
		}
		return
	}

	// Cleanup uploaded file after successful install (async, non-blocking)
	// Keep file for 5 minutes to allow for potential retries or debugging
	if strings.HasPrefix(req.PackagePath, h.uploadDir) {
		go func() {
			time.Sleep(5 * time.Minute) // Wait 5 minutes before cleanup
			if err := os.Remove(req.PackagePath); err != nil {
				h.logger.Warn("Failed to cleanup uploaded file",
					zap.String("path", req.PackagePath),
					zap.Error(err),
				)
			} else {
				h.logger.Info("Cleaned up uploaded chaincode package",
					zap.String("path", req.PackagePath),
				)
			}
		}()
	}

	respondSuccess(w, http.StatusCreated, map[string]string{
		"packageId": packageID,
	})
}

// Approve approves a chaincode definition
func (h *ChaincodeHandler) Approve(w http.ResponseWriter, r *http.Request) {
	var req chaincode.ApproveChaincodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	if err := h.service.Approve(r.Context(), &req); err != nil {
		h.logger.Error("Failed to approve chaincode", zap.Error(err))
		if strings.Contains(err.Error(), "peer CLI not available") {
			respondError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"Peer CLI is not available", err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
				"Failed to approve chaincode", err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Chaincode %s version %s approved successfully", req.Name, req.Version),
	})
}

// Commit commits a chaincode definition
func (h *ChaincodeHandler) Commit(w http.ResponseWriter, r *http.Request) {
	var req chaincode.CommitChaincodeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body", err.Error())
		return
	}

	if err := h.service.Commit(r.Context(), &req); err != nil {
		h.logger.Error("Failed to commit chaincode", zap.Error(err))
		if strings.Contains(err.Error(), "peer CLI not available") {
			respondError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"Peer CLI is not available", err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
				"Failed to commit chaincode", err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Chaincode %s version %s committed successfully", req.Name, req.Version),
	})
}

// ListInstalled lists installed chaincodes
func (h *ChaincodeHandler) ListInstalled(w http.ResponseWriter, r *http.Request) {
	peerAddress := r.URL.Query().Get("peer")

	chaincodes, err := h.service.ListInstalled(r.Context(), peerAddress)
	if err != nil {
		h.logger.Error("Failed to list installed chaincodes", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to list installed chaincodes", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, chaincodes)
}

// ListCommitted lists committed chaincodes
func (h *ChaincodeHandler) ListCommitted(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channel")
	if channelName == "" {
		channelName = "ibnchannel" // Default
	}

	chaincodes, err := h.service.ListCommitted(r.Context(), channelName)
	if err != nil {
		h.logger.Error("Failed to list committed chaincodes", zap.Error(err))
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
			"Failed to list committed chaincodes", err.Error())
		return
	}

	respondSuccess(w, http.StatusOK, chaincodes)
}

// GetCommittedInfo gets information about a committed chaincode
func (h *ChaincodeHandler) GetCommittedInfo(w http.ResponseWriter, r *http.Request) {
	chaincodeName := r.URL.Query().Get("name")
	if chaincodeName == "" {
		respondError(w, http.StatusBadRequest, "BAD_REQUEST", "chaincode name is required", "")
		return
	}

	channelName := r.URL.Query().Get("channel")
	if channelName == "" {
		channelName = "ibnchannel" // Default
	}

	info, err := h.service.GetCommittedInfo(r.Context(), channelName, chaincodeName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondError(w, http.StatusNotFound, "NOT_FOUND",
				fmt.Sprintf("Chaincode '%s' not found on channel '%s'", chaincodeName, channelName), err.Error())
		} else {
			h.logger.Error("Failed to get committed chaincode info", zap.Error(err))
			respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR",
				"Failed to get committed chaincode info", err.Error())
		}
		return
	}

	respondSuccess(w, http.StatusOK, info)
}

// validatePackageStructure performs basic validation on chaincode package
// Checks if file is a valid tar.gz and contains expected structure
// Note: Full validation (metadata.json, code.tar.gz) is done by peer CLI during install
func (h *ChaincodeHandler) validatePackageStructure(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Try to open as gzip
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("file is not a valid gzip archive: %w", err)
	}
	defer gzReader.Close()

	// Try to open as tar
	tarReader := tar.NewReader(gzReader)
	
	hasMetadata := false
	hasCodeTar := false
	entryCount := 0
	maxEntries := 100 // Safety limit

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		entryCount++
		if entryCount > maxEntries {
			return fmt.Errorf("too many entries in package (max %d)", maxEntries)
		}

		// Check for metadata.json
		if header.Name == "metadata.json" {
			hasMetadata = true
		}

		// Check for code.tar.gz
		if header.Name == "code.tar.gz" || header.Name == "code.tar" {
			hasCodeTar = true
		}

		// Check for invalid directory entries (common error)
		if header.Typeflag == tar.TypeDir && !strings.HasSuffix(header.Name, "/") {
			h.logger.Warn("Package contains directory entry that may cause issues",
				zap.String("entry", header.Name),
				zap.String("filepath", filePath),
			)
		}
	}

	// Warn if missing expected files (but don't fail - peer CLI will validate)
	if !hasMetadata {
		h.logger.Warn("Package may be missing metadata.json",
			zap.String("filepath", filePath),
		)
	}
	if !hasCodeTar {
		h.logger.Warn("Package may be missing code.tar.gz",
			zap.String("filepath", filePath),
		)
	}

	return nil
}
