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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/repository/db"
	auditservice "github.com/ibn-network/api-gateway/internal/services/audit"
	"go.uber.org/zap"
)

// Handler handles audit log operations
type Handler struct {
	auditService *auditservice.Service
	logger       *zap.Logger
}

// NewHandler creates a new audit handler
func NewHandler(auditService *auditservice.Service, logger *zap.Logger) *Handler {
	return &Handler{
		auditService: auditService,
		logger:       logger,
	}
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ListLogs godoc
// @Summary List audit logs
// @Description List audit logs with filters and pagination
// @Tags audit
// @Produce json
// @Param userId query string false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param txId query string false "Filter by transaction ID"
// @Param startTime query string false "Start time (RFC3339)"
// @Param endTime query string false "End time (RFC3339)"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} models.APIResponse{data=object}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /audit/logs [get]
func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	query := &models.AuditLogListQuery{}

	// Parse query parameters
	if userId := r.URL.Query().Get("userId"); userId != "" {
		query.UserID = userId
	}
	if action := r.URL.Query().Get("action"); action != "" {
		query.Action = action
	}
	if txId := r.URL.Query().Get("txId"); txId != "" {
		query.TxID = txId
	}
	if startTimeStr := r.URL.Query().Get("startTime"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			query.StartTime = &t
		}
	}
	if endTimeStr := r.URL.Query().Get("endTime"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			query.EndTime = &t
		}
	}

	limit := int32(50)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}

	offset := int32(0)
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	var logs []*models.AuditLogResponse
	var total int64
	var err error

	// Route to appropriate query method
	if query.UserID != "" {
		logs, total, err = h.listLogsByUser(r.Context(), query.UserID, limit, offset)
	} else if query.Action != "" {
		logs, total, err = h.listLogsByAction(r.Context(), query.Action, limit, offset)
	} else if query.TxID != "" {
		logs, total, err = h.listLogsByTxID(r.Context(), query.TxID)
	} else if query.StartTime != nil && query.EndTime != nil {
		logs, total, err = h.listLogsByDateRange(r.Context(), *query.StartTime, *query.EndTime, limit, offset)
	} else {
		logs, total, err = h.listAllLogs(r.Context(), limit, offset)
	}

	if err != nil {
		h.logger.Error("Failed to list audit logs", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list audit logs",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}))
}

// Helper methods
func (h *Handler) listAllLogs(ctx context.Context, limit, offset int32) ([]*models.AuditLogResponse, int64, error) {
	dbLogs, total, err := h.auditService.ListLogs(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*models.AuditLogResponse, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = h.dbLogToResponse(dbLog)
	}

	return logs, total, nil
}

func (h *Handler) listLogsByUser(ctx context.Context, userID string, limit, offset int32) ([]*models.AuditLogResponse, int64, error) {
	dbLogs, total, err := h.auditService.ListLogsByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*models.AuditLogResponse, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = h.dbLogToResponse(dbLog)
	}

	return logs, total, nil
}

func (h *Handler) listLogsByAction(ctx context.Context, action string, limit, offset int32) ([]*models.AuditLogResponse, int64, error) {
	dbLogs, total, err := h.auditService.ListLogsByAction(ctx, action, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*models.AuditLogResponse, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = h.dbLogToResponse(dbLog)
	}

	return logs, total, nil
}

func (h *Handler) listLogsByTxID(ctx context.Context, txID string) ([]*models.AuditLogResponse, int64, error) {
	dbLogs, err := h.auditService.ListLogsByTxID(ctx, txID)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*models.AuditLogResponse, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = h.dbLogToResponse(dbLog)
	}

	return logs, int64(len(logs)), nil
}

func (h *Handler) listLogsByDateRange(ctx context.Context, startTime, endTime time.Time, limit, offset int32) ([]*models.AuditLogResponse, int64, error) {
	dbLogs, total, err := h.auditService.ListLogsByDateRange(ctx, startTime, endTime, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	logs := make([]*models.AuditLogResponse, len(dbLogs))
	for i, dbLog := range dbLogs {
		logs[i] = h.dbLogToResponse(dbLog)
	}

	return logs, total, nil
}

// dbLogToResponse converts database audit log to API response
func (h *Handler) dbLogToResponse(dbLog *db.AuditLog) *models.AuditLogResponse {
	log := &models.AuditLogResponse{
		ID:     dbLog.ID,
		Action: dbLog.Action,
		Status: dbLog.Status,
	}

	// Convert user ID
	if dbLog.UserID.Valid {
		log.UserID = uuidToString(dbLog.UserID)
	}

	// Convert API key ID
	if dbLog.ApiKeyID.Valid {
		log.ApiKeyID = uuidToString(dbLog.ApiKeyID)
	}

	// Convert resource type
	if dbLog.ResourceType.Valid {
		log.ResourceType = dbLog.ResourceType.String
	}

	// Convert resource ID
	if dbLog.ResourceID.Valid {
		log.ResourceID = dbLog.ResourceID.String
	}

	// Convert tx ID
	if dbLog.TxID.Valid {
		log.TxID = dbLog.TxID.String
	}

	// Convert details
	if len(dbLog.Details) > 0 {
		var details map[string]interface{}
		if err := json.Unmarshal(dbLog.Details, &details); err == nil {
			log.Details = details
		}
	}

	// Convert IP address
	if dbLog.IpAddress != nil {
		log.IpAddress = dbLog.IpAddress.String()
	}

	// Convert user agent
	if dbLog.UserAgent.Valid {
		log.UserAgent = dbLog.UserAgent.String
	}

	// Convert created at
	if dbLog.CreatedAt.Valid {
		log.CreatedAt = dbLog.CreatedAt.Time
	}

	return log
}

// uuidToString converts pgtype.UUID to string
func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid.Bytes[0:4],
		uuid.Bytes[4:6],
		uuid.Bytes[6:8],
		uuid.Bytes[8:10],
		uuid.Bytes[10:16],
	)
}

// GetLog godoc
// @Summary Get audit log by ID
// @Description Get a specific audit log entry by ID
// @Tags audit
// @Produce json
// @Param id path int true "Audit log ID"
// @Success 200 {object} models.APIResponse{data=models.AuditLogResponse}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /audit/logs/{id} [get]
func (h *Handler) GetLog(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid audit log ID",
			err.Error(),
		))
		return
	}

	dbLog, err := h.auditService.GetLog(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get audit log",
			zap.Int64("id", id),
			zap.Error(err),
		)
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Audit log not found",
			err.Error(),
		))
		return
	}

	log := h.dbLogToResponse(dbLog)
	respondJSON(w, http.StatusOK, models.NewSuccessResponse(log))
}

