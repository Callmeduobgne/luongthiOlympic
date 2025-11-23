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

package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Service handles event subscription and delivery
type Service struct {
	repo               *Repository
	logger             *zap.Logger
	webhookQueue       chan *WebhookDelivery
	eventSubscriptions map[string][]*EventSubscription
	mu                 sync.RWMutex
}

// NewService creates a new event service
func NewService(repo *Repository, logger *zap.Logger) *Service {
	s := &Service{
		repo:               repo,
		logger:             logger,
		webhookQueue:       make(chan *WebhookDelivery, 1000),
		eventSubscriptions: make(map[string][]*EventSubscription),
	}

	// Load active subscriptions
	go s.loadSubscriptions()

	// Start webhook delivery worker
	go s.processWebhookDeliveries()

	return s
}

// CreateSubscription creates a new event subscription
func (s *Service) CreateSubscription(ctx context.Context, userID uuid.UUID, req *CreateSubscriptionRequest) (*EventSubscription, error) {
	sub := &EventSubscription{
		ID:          uuid.New(),
		UserID:      userID,
		ChannelID:   req.ChannelID,
		ChaincodeID: req.ChaincodeID,
		EventName:   req.EventName,
		FilterType:  req.FilterType,
		FilterValue: req.FilterValue,
		WebhookURL:  req.WebhookURL,
		IsActive:    true,
	}

	// Set default filter type
	if sub.FilterType == "" {
		sub.FilterType = FilterTypeExact
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		s.logger.Error("Failed to create subscription", zap.Error(err))
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Add to in-memory cache
	s.addSubscriptionToCache(sub)

	return sub, nil
}

// GetSubscription retrieves a subscription by ID
func (s *Service) GetSubscription(ctx context.Context, id uuid.UUID) (*EventSubscription, error) {
	return s.repo.GetSubscription(ctx, id)
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (s *Service) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]*EventSubscription, error) {
	return s.repo.GetUserSubscriptions(ctx, userID)
}

// UpdateSubscription updates a subscription
func (s *Service) UpdateSubscription(ctx context.Context, id uuid.UUID, isActive bool, webhookURL *string) error {
	sub, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}

	sub.IsActive = isActive
	sub.WebhookURL = webhookURL

	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}

	// Reload subscriptions
	go s.loadSubscriptions()

	return nil
}

// DeleteSubscription deletes a subscription
func (s *Service) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteSubscription(ctx, id); err != nil {
		return err
	}

	// Reload subscriptions
	go s.loadSubscriptions()

	return nil
}

// PublishEvent publishes an event to matching subscriptions
func (s *Service) PublishEvent(ctx context.Context, event *Event) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.getSubscriptionKey(event.ChannelID, event.ChaincodeID, event.EventName)
	subscriptions, ok := s.eventSubscriptions[key]
	if !ok {
		s.logger.Debug("No subscriptions for event", zap.String("key", key))
		return nil
	}

	// Match subscriptions with filters
	for _, sub := range subscriptions {
		if !sub.IsActive {
			continue
		}

		if s.matchesFilter(sub, event) {
			s.deliverEvent(sub, event)
		}
	}

	return nil
}

// matchesFilter checks if an event matches a subscription filter
func (s *Service) matchesFilter(sub *EventSubscription, event *Event) bool {
	if sub.FilterType == "" || sub.FilterValue == "" {
		return true
	}

	eventPayloadStr := fmt.Sprintf("%v", event.Payload)

	switch sub.FilterType {
	case FilterTypeExact:
		return strings.Contains(eventPayloadStr, sub.FilterValue)
	case FilterTypePrefix:
		return strings.HasPrefix(eventPayloadStr, sub.FilterValue)
	case FilterTypeRegex:
		matched, err := regexp.MatchString(sub.FilterValue, eventPayloadStr)
		if err != nil {
			s.logger.Error("Invalid regex filter", zap.Error(err))
			return false
		}
		return matched
	default:
		return true
	}
}

// deliverEvent delivers an event to a subscription
func (s *Service) deliverEvent(sub *EventSubscription, event *Event) {
	if sub.WebhookURL == nil {
		return
	}

	delivery := &WebhookDelivery{
		ID:             uuid.New(),
		SubscriptionID: sub.ID,
		EventID:        event.ID,
		EventName:      event.EventName,
		Payload:        event.Payload,
		WebhookURL:     *sub.WebhookURL,
		Status:         DeliveryStatusPending,
		AttemptCount:   0,
	}

	// Save to database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.CreateWebhookDelivery(ctx, delivery); err != nil {
		s.logger.Error("Failed to create webhook delivery", zap.Error(err))
		return
	}

	// Queue for delivery
	select {
	case s.webhookQueue <- delivery:
	default:
		s.logger.Warn("Webhook queue full, delivery may be delayed")
	}
}

// processWebhookDeliveries processes webhook deliveries from the queue
func (s *Service) processWebhookDeliveries() {
	for delivery := range s.webhookQueue {
		s.sendWebhook(delivery)
	}
}

// sendWebhook sends a webhook delivery
func (s *Service) sendWebhook(delivery *WebhookDelivery) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	delivery.AttemptCount++

	// Prepare payload
	payloadBytes, err := json.Marshal(delivery.Payload)
	if err != nil {
		s.logger.Error("Failed to marshal webhook payload", zap.Error(err))
		errMsg := err.Error()
		delivery.Status = DeliveryStatusFailed
		delivery.ErrorMessage = &errMsg
		s.repo.UpdateWebhookDelivery(ctx, delivery)
		return
	}

	// Send HTTP POST request
	req, err := http.NewRequestWithContext(ctx, "POST", delivery.WebhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		s.logger.Error("Failed to create webhook request", zap.Error(err))
		s.handleWebhookError(ctx, delivery, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-ID", delivery.EventID)
	req.Header.Set("X-Event-Name", delivery.EventName)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send webhook", zap.Error(err))
		s.handleWebhookError(ctx, delivery, err)
		return
	}
	defer resp.Body.Close()

	delivery.ResponseCode = &resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		delivery.Status = DeliveryStatusDelivered
		now := time.Now()
		delivery.CompletedAt = &now
		s.repo.UpdateWebhookDelivery(ctx, delivery)
		s.logger.Info("Webhook delivered successfully",
			zap.String("delivery_id", delivery.ID.String()),
			zap.String("webhook_url", delivery.WebhookURL),
		)
	} else {
		// Failure - retry
		s.handleWebhookError(ctx, delivery, fmt.Errorf("HTTP %d", resp.StatusCode))
	}
}

// handleWebhookError handles webhook delivery errors
func (s *Service) handleWebhookError(ctx context.Context, delivery *WebhookDelivery, err error) {
	errMsg := err.Error()
	delivery.ErrorMessage = &errMsg

	if delivery.AttemptCount < 5 {
		// Retry with exponential backoff
		delivery.Status = DeliveryStatusRetrying
		retryDelay := time.Duration(delivery.AttemptCount*delivery.AttemptCount) * time.Minute
		nextRetry := time.Now().Add(retryDelay)
		delivery.NextRetryAt = &nextRetry

		s.logger.Info("Webhook delivery will be retried",
			zap.String("delivery_id", delivery.ID.String()),
			zap.Int("attempt", delivery.AttemptCount),
			zap.Time("next_retry", nextRetry),
		)
	} else {
		// Max retries reached
		delivery.Status = DeliveryStatusFailed
		now := time.Now()
		delivery.CompletedAt = &now

		s.logger.Error("Webhook delivery failed after max retries",
			zap.String("delivery_id", delivery.ID.String()),
			zap.String("webhook_url", delivery.WebhookURL),
		)
	}

	s.repo.UpdateWebhookDelivery(ctx, delivery)
}

// loadSubscriptions loads active subscriptions into memory
func (s *Service) loadSubscriptions() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	subscriptions, err := s.repo.GetActiveSubscriptions(ctx)
	if err != nil {
		s.logger.Error("Failed to load subscriptions", zap.Error(err))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing subscriptions
	s.eventSubscriptions = make(map[string][]*EventSubscription)

	// Group by channel, chaincode, and event name
	for _, sub := range subscriptions {
		key := s.getSubscriptionKey(sub.ChannelID, sub.ChaincodeID, sub.EventName)
		s.eventSubscriptions[key] = append(s.eventSubscriptions[key], sub)
	}

	s.logger.Info("Loaded event subscriptions",
		zap.Int("count", len(subscriptions)),
		zap.Int("groups", len(s.eventSubscriptions)),
	)
}

// addSubscriptionToCache adds a subscription to the in-memory cache
func (s *Service) addSubscriptionToCache(sub *EventSubscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.getSubscriptionKey(sub.ChannelID, sub.ChaincodeID, sub.EventName)
	s.eventSubscriptions[key] = append(s.eventSubscriptions[key], sub)
}

// getSubscriptionKey generates a key for subscription grouping
func (s *Service) getSubscriptionKey(channelID, chaincodeID, eventName string) string {
	return fmt.Sprintf("%s:%s:%s", channelID, chaincodeID, eventName)
}

