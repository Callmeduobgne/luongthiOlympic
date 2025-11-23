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

package event

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/event"
	"go.uber.org/zap"
)

// EventHandler handles event subscription operations
type EventHandler struct {
	subscriptionService *event.SubscriptionService
	listenerService     *event.ListenerService
	dispatcher          *event.EventDispatcher
	wsManager          *event.WebSocketManager
	logger             *zap.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(
	subscriptionService *event.SubscriptionService,
	listenerService *event.ListenerService,
	dispatcher *event.EventDispatcher,
	wsManager *event.WebSocketManager,
	logger *zap.Logger,
) *EventHandler {
	return &EventHandler{
		subscriptionService: subscriptionService,
		listenerService:     listenerService,
		dispatcher:          dispatcher,
		wsManager:           wsManager,
		logger:              logger,
	}
}

// CreateSubscription godoc
// @Summary Create event subscription
// @Description Create a new event subscription (WebSocket, SSE, or Webhook)
// @Tags events
// @Accept json
// @Produce json
// @Param request body models.CreateSubscriptionRequest true "Subscription request"
// @Success 201 {object} models.APIResponse{data=models.EventSubscription}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /events/subscriptions [post]
func (h *EventHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	// Get user ID and API key ID from context
	userID := ""
	if uid, ok := r.Context().Value("userID").(string); ok {
		userID = uid
	}

	apiKeyID := ""
	if akid, ok := r.Context().Value("apiKeyID").(string); ok {
		apiKeyID = akid
	}

	// Create subscription
	subscription, err := h.subscriptionService.CreateSubscription(r.Context(), &req, userID, apiKeyID)
	if err != nil {
		h.logger.Error("Failed to create subscription",
			zap.String("type", string(req.Type)),
			zap.String("channel", req.ChannelName),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to create subscription",
			err.Error(),
		))
		return
	}

	// Register with dispatcher
	h.dispatcher.RegisterSubscription(subscription)

	// Start listening if active
	if subscription.Active {
		if err := h.listenerService.StartListening(r.Context(), subscription); err != nil {
			h.logger.Warn("Failed to start listening",
				zap.String("subscription_id", subscription.ID),
				zap.Error(err),
			)
		}
	}

	respondJSON(w, http.StatusCreated, models.NewSuccessResponse(subscription))
}

// ListSubscriptions godoc
// @Summary List event subscriptions
// @Description List event subscriptions with filters
// @Tags events
// @Produce json
// @Param channelName query string false "Filter by channel name"
// @Param chaincodeName query string false "Filter by chaincode name"
// @Param type query string false "Filter by subscription type (websocket, sse, webhook)"
// @Param active query bool false "Filter by active status"
// @Param limit query int false "Limit (default: 50, max: 100)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} models.APIResponse{data=models.SubscriptionListResponse}
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /events/subscriptions [get]
func (h *EventHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	query := &models.SubscriptionListQuery{
		ChannelName:   r.URL.Query().Get("channelName"),
		ChaincodeName: r.URL.Query().Get("chaincodeName"),
		Type:          models.SubscriptionType(r.URL.Query().Get("type")),
	}

	// Parse active filter
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		active, err := strconv.ParseBool(activeStr)
		if err == nil {
			query.Active = &active
		}
	}

	// Parse limit and offset
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			query.Offset = offset
		}
	}

	// Get user ID and API key ID from context
	userID := ""
	if uid, ok := r.Context().Value("userID").(string); ok {
		userID = uid
	}

	apiKeyID := ""
	if akid, ok := r.Context().Value("apiKeyID").(string); ok {
		apiKeyID = akid
	}

	// List subscriptions
	response, err := h.subscriptionService.ListSubscriptions(r.Context(), query, userID, apiKeyID)
	if err != nil {
		h.logger.Error("Failed to list subscriptions", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to list subscriptions",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(response))
}

// GetSubscription godoc
// @Summary Get subscription details
// @Description Get event subscription details by ID
// @Tags events
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.APIResponse{data=models.EventSubscription}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /events/subscriptions/{id} [get]
func (h *EventHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "id")
	if subscriptionID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription ID is required",
			"",
		))
		return
	}

	subscription, err := h.subscriptionService.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.Error("Failed to get subscription",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Subscription not found",
			err.Error(),
		))
		return
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(subscription))
}

// UpdateSubscription godoc
// @Summary Update subscription
// @Description Update an event subscription
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param request body models.UpdateSubscriptionRequest true "Update request"
// @Success 200 {object} models.APIResponse{data=models.EventSubscription}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /events/subscriptions/{id} [patch]
func (h *EventHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "id")
	if subscriptionID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription ID is required",
			"",
		))
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Invalid request body",
			err.Error(),
		))
		return
	}

	subscription, err := h.subscriptionService.UpdateSubscription(r.Context(), subscriptionID, &req)
	if err != nil {
		h.logger.Error("Failed to update subscription",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to update subscription",
			err.Error(),
		))
		return
	}

	// Update dispatcher registration
	h.dispatcher.RegisterSubscription(subscription)

	// Start/stop listening based on active status
	if subscription.Active {
		if err := h.listenerService.StartListening(r.Context(), subscription); err != nil {
			h.logger.Warn("Failed to start listening",
				zap.String("subscription_id", subscription.ID),
				zap.Error(err),
			)
		}
	} else {
		h.listenerService.StopListening(subscription.ID)
	}

	respondJSON(w, http.StatusOK, models.NewSuccessResponse(subscription))
}

// DeleteSubscription godoc
// @Summary Delete subscription
// @Description Delete an event subscription
// @Tags events
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /events/subscriptions/{id} [delete]
func (h *EventHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "id")
	if subscriptionID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription ID is required",
			"",
		))
		return
	}

	// Stop listening
	h.listenerService.StopListening(subscriptionID)

	// Unregister from dispatcher
	h.dispatcher.UnregisterSubscription(subscriptionID)

	// Delete subscription
	err := h.subscriptionService.DeleteSubscription(r.Context(), subscriptionID)
	if err != nil {
		h.logger.Error("Failed to delete subscription",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		respondJSON(w, http.StatusInternalServerError, models.NewErrorResponse(
			models.ErrCodeInternalError,
			"Failed to delete subscription",
			err.Error(),
		))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WebSocketHandler handles WebSocket connections for event streaming
func (h *EventHandler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionId")
	if subscriptionID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription ID is required",
			"",
		))
		return
	}

	// Get subscription
	subscription, err := h.subscriptionService.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Subscription not found",
			err.Error(),
		))
		return
	}

	// Verify subscription type
	if subscription.Type != models.SubscriptionTypeWebSocket {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription is not a WebSocket subscription",
			"",
		))
		return
	}

	// Upgrade to WebSocket
	upgrader := h.wsManager.GetUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		return
	}

	// Register connection
	h.wsManager.RegisterWebSocketConnection(subscriptionID, conn)

	h.logger.Info("WebSocket connection established",
		zap.String("subscription_id", subscriptionID),
	)
}

// SSEHandler handles Server-Sent Events connections
func (h *EventHandler) SSEHandler(w http.ResponseWriter, r *http.Request) {
	subscriptionID := chi.URLParam(r, "subscriptionId")
	if subscriptionID == "" {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription ID is required",
			"",
		))
		return
	}

	// Get subscription
	subscription, err := h.subscriptionService.GetSubscription(r.Context(), subscriptionID)
	if err != nil {
		respondJSON(w, http.StatusNotFound, models.NewErrorResponse(
			models.ErrCodeNotFound,
			"Subscription not found",
			err.Error(),
		))
		return
	}

	// Verify subscription type
	if subscription.Type != models.SubscriptionTypeSSE {
		respondJSON(w, http.StatusBadRequest, models.NewErrorResponse(
			models.ErrCodeBadRequest,
			"Subscription is not an SSE subscription",
			"",
		))
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Register SSE client
	client := h.wsManager.RegisterSSEClient(subscriptionID)
	defer h.wsManager.UnregisterSSEClient(subscriptionID)

	// Send events
	for {
		select {
		case event, ok := <-client.Channel:
			if !ok {
				return
			}

			// Format as SSE
			eventJSON, err := json.Marshal(event)
			if err != nil {
				h.logger.Error("Failed to marshal event", zap.Error(err))
				continue
			}

			// Write SSE format: data: {json}\n\n
			if _, err := w.Write([]byte("data: " + string(eventJSON) + "\n\n")); err != nil {
				h.logger.Error("Failed to write SSE event", zap.Error(err))
				return
			}

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case <-client.Done:
			return
		case <-r.Context().Done():
			return
		}
	}
}

// respondJSON is a helper function to send JSON responses
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

