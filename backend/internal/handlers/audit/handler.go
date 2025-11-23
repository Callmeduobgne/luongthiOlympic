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

package audit

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ibn-network/backend/internal/services/analytics/audit"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles audit-related HTTP requests
type Handler struct {
	service *audit.Service
	logger  *zap.Logger
}

// NewHandler creates a new audit handler
func NewHandler(service *audit.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// QueryLogs handles querying audit logs
func (h *Handler) QueryLogs(w http.ResponseWriter, r *http.Request) {
	var req audit.QueryLogsRequest

	// Parse query parameters
	query := r.URL.Query()
	
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		req.UserID = &userID
	}

	if action := query.Get("action"); action != "" {
		req.Action = &action
	}

	if resourceType := query.Get("resource_type"); resourceType != "" {
		req.ResourceType = &resourceType
	}

	if status := query.Get("status"); status != "" {
		req.Status = &status
	}

	if startDateStr := query.Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format (use RFC3339)", http.StatusBadRequest)
			return
		}
		req.StartDate = &startDate
	}

	if endDateStr := query.Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format (use RFC3339)", http.StatusBadRequest)
			return
		}
		req.EndDate = &endDate
	}

	// Default pagination
	req.Limit = 100
	req.Offset = 0

	logs, err := h.service.QueryLogs(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to query audit logs", zap.Error(err))
		http.Error(w, "Failed to query logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// SearchLogs handles full-text search on audit logs
func (h *Handler) SearchLogs(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		http.Error(w, "Search term 'q' is required", http.StatusBadRequest)
		return
	}

	logs, err := h.service.SearchLogs(r.Context(), searchTerm, 100)
	if err != nil {
		h.logger.Error("Failed to search audit logs", zap.Error(err))
		http.Error(w, "Failed to search logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// GetSecurityEvents handles retrieving security-related events
func (h *Handler) GetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.GetSecurityEvents(r.Context(), 100)
	if err != nil {
		h.logger.Error("Failed to get security events", zap.Error(err))
		http.Error(w, "Failed to get security events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": logs,
		"count":  len(logs),
	})
}

// GetFailedAttempts handles retrieving failed authentication/authorization attempts
func (h *Handler) GetFailedAttempts(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
		userID = &parsedID
	}

	logs, err := h.service.GetFailedAttempts(r.Context(), userID, 100)
	if err != nil {
		h.logger.Error("Failed to get failed attempts", zap.Error(err))
		http.Error(w, "Failed to get failed attempts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"attempts": logs,
		"count":    len(logs),
	})
}

