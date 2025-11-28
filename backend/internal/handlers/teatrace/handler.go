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

package teatrace

import (
	"encoding/json"
	"net/http"

	"github.com/ibn-network/backend/internal/services/teatrace"
	"go.uber.org/zap"
)

// Handler handles Tea Traceability HTTP requests
type Handler struct {
	service *teatrace.Service
	logger  *zap.Logger
}

// NewHandler creates a new Tea Traceability handler
func NewHandler(service *teatrace.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// VerifyRequest represents the request body for verification
type VerifyRequest struct {
	Hash string `json:"hash"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// VerifyByHash verifies an entity by its hash
func (h *Handler) VerifyByHash(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	if req.Hash == "" {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "BAD_REQUEST",
				"message": "Hash is required",
			},
		})
		return
	}

	result, err := h.service.VerifyByHash(r.Context(), req.Hash)
	if err != nil {
		h.logger.Error("Failed to verify by hash",
			zap.String("hash", req.Hash),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}
