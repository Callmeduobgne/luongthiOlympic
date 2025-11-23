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
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ibn-network/backend/internal/services/blockchain/chaincode/approval"
	"go.uber.org/zap"
)

// ApprovalHandler handles approval workflow operations
type ApprovalHandler struct {
	approvalService *approval.Service
	logger          *zap.Logger
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(approvalService *approval.Service, logger *zap.Logger) *ApprovalHandler {
	return &ApprovalHandler{
		approvalService: approvalService,
		logger:          logger,
	}
}

// respondJSON sends a JSON response
func (h *ApprovalHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *ApprovalHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// respondSuccess sends a success response
func (h *ApprovalHandler) respondSuccess(w http.ResponseWriter, status int, data interface{}) {
	h.respondJSON(w, status, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// CreateRequest creates a new approval request
// @Summary Create approval request
// @Description Create a new approval request for chaincode operation
// @Tags chaincode
// @Accept json
// @Produce json
// @Param request body CreateApprovalRequestRequest true "Approval request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/approval/request [post]
func (h *ApprovalHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChaincodeVersionID string                 `json:"chaincodeVersionId"`
		Operation          string                 `json:"operation"`
		Reason             *string                `json:"reason,omitempty"`
		Metadata           map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		var err error
		userID, err = uuid.Parse(v)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID")
			return
		}
	default:
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID")
		return
	}

	// Parse chaincode version ID
	versionID, err := uuid.Parse(req.ChaincodeVersionID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid chaincode version ID")
		return
	}

	// Validate operation
	if req.Operation != "install" && req.Operation != "approve" && req.Operation != "commit" {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid operation. Must be install, approve, or commit")
		return
	}

	// Create approval request
	createReq := &approval.CreateRequestRequest{
		ChaincodeVersionID: versionID,
		Operation:          req.Operation,
		RequestedBy:        userID,
		Reason:             req.Reason,
		Metadata:           req.Metadata,
	}

	approvalReq, err := h.approvalService.CreateRequest(r.Context(), createReq)
	if err != nil {
		h.logger.Error("Failed to create approval request", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":                approvalReq.ID.String(),
		"chaincodeVersionId": approvalReq.ChaincodeVersionID.String(),
		"operation":         approvalReq.Operation,
		"status":            approvalReq.Status,
		"requestedBy":       approvalReq.RequestedBy.String(),
		"requestedAt":       approvalReq.RequestedAt,
		"expiresAt":         approvalReq.ExpiresAt,
	})
}

// Vote votes on an approval request
// @Summary Vote on approval request
// @Description Approve or reject an approval request
// @Tags chaincode
// @Accept json
// @Produce json
// @Param request body VoteRequest true "Vote request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/approval/vote [post]
func (h *ApprovalHandler) Vote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ApprovalRequestID string  `json:"approvalRequestId"`
		Vote              string  `json:"vote"` // "approve" or "reject"
		Comment           *string `json:"comment,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		h.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	var userID uuid.UUID
	switch v := userIDVal.(type) {
	case uuid.UUID:
		userID = v
	case string:
		var err error
		userID, err = uuid.Parse(v)
		if err != nil {
			h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID")
			return
		}
	default:
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid user ID")
		return
	}

	// Parse approval request ID
	requestID, err := uuid.Parse(req.ApprovalRequestID)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid approval request ID")
		return
	}

	// Validate vote
	if req.Vote != "approve" && req.Vote != "reject" {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid vote. Must be 'approve' or 'reject'")
		return
	}

	// Vote
	voteReq := &approval.VoteRequest{
		ApprovalRequestID: requestID,
		ApproverID:        userID,
		Vote:              req.Vote,
		Comment:           req.Comment,
	}

	if err := h.approvalService.Vote(r.Context(), voteReq); err != nil {
		h.logger.Error("Failed to vote on approval request", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.respondSuccess(w, http.StatusOK, map[string]string{
		"message": "Vote recorded successfully",
	})
}

// GetRequest gets approval request details
// @Summary Get approval request
// @Description Get details of an approval request including votes
// @Tags chaincode
// @Produce json
// @Param id path string true "Approval request ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/approval/request/{id} [get]
func (h *ApprovalHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	requestID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid approval request ID")
		return
	}

	// Get approval request
	approvalReq, err := h.approvalService.GetRequest(r.Context(), requestID)
	if err != nil {
		h.logger.Error("Failed to get approval request", zap.Error(err))
		h.respondError(w, http.StatusNotFound, "NOT_FOUND", "Approval request not found")
		return
	}

	// Get votes
	votes, err := h.approvalService.GetVotes(r.Context(), requestID)
	if err != nil {
		h.logger.Warn("Failed to get votes", zap.Error(err))
		votes = []*approval.ApprovalVote{} // Empty votes on error
	}

	// Format votes
	votesData := make([]map[string]interface{}, len(votes))
	for i, vote := range votes {
		votesData[i] = map[string]interface{}{
			"id":        vote.ID.String(),
			"approverId": vote.ApproverID.String(),
			"vote":      vote.Vote,
			"comment":   vote.Comment,
			"votedAt":   vote.VotedAt,
		}
	}

	h.respondSuccess(w, http.StatusOK, map[string]interface{}{
		"id":                approvalReq.ID.String(),
		"chaincodeVersionId": approvalReq.ChaincodeVersionID.String(),
		"operation":         approvalReq.Operation,
		"status":            approvalReq.Status,
		"requestedBy":       approvalReq.RequestedBy.String(),
		"requestedAt":       approvalReq.RequestedAt,
		"expiresAt":         approvalReq.ExpiresAt,
		"reason":            approvalReq.Reason,
		"metadata":          approvalReq.Metadata,
		"votes":             votesData,
	})
}

// ListRequests lists approval requests
// @Summary List approval requests
// @Description List approval requests with optional filters
// @Tags chaincode
// @Produce json
// @Param status query string false "Filter by status (pending, approved, rejected, expired)"
// @Param operation query string false "Filter by operation (install, approve, commit)"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/chaincode/approval/requests [get]
func (h *ApprovalHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	filters := &approval.RequestFilters{}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = &status
	}

	if operation := r.URL.Query().Get("operation"); operation != "" {
		filters.Operation = &operation
	}

	// Get user ID for filtering (optional)
	userIDVal := r.Context().Value("user_id")
	if userIDVal != nil {
		var userID uuid.UUID
		switch v := userIDVal.(type) {
		case uuid.UUID:
			userID = v
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				userID = parsed
				filters.RequestedBy = &userID
			}
		}
	}

	// Parse limit and offset
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var limit int
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		var offset int
		if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	requests, err := h.approvalService.ListRequests(r.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list approval requests", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	// Format requests
	requestsData := make([]map[string]interface{}, len(requests))
	for i, req := range requests {
		requestsData[i] = map[string]interface{}{
			"id":                req.ID.String(),
			"chaincodeVersionId": req.ChaincodeVersionID.String(),
			"operation":         req.Operation,
			"status":            req.Status,
			"requestedBy":       req.RequestedBy.String(),
			"requestedAt":       req.RequestedAt,
			"expiresAt":         req.ExpiresAt,
			"reason":            req.Reason,
		}
	}

	h.respondSuccess(w, http.StatusOK, map[string]interface{}{
		"requests": requestsData,
		"count":    len(requestsData),
	})
}

