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
	"fmt"
	"sync"

	"github.com/ibn-network/api-gateway/internal/models"
	"go.uber.org/zap"
)

// EventDispatcher dispatches events to subscribers
type EventDispatcher struct {
	subscriptions map[string][]*models.EventSubscription // key: channel+chaincode+event -> subscriptions
	webhookClient *WebhookClient
	wsManager     *WebSocketManager
	logger        *zap.Logger
	mu            sync.RWMutex
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(
	webhookClient *WebhookClient,
	wsManager *WebSocketManager,
	logger *zap.Logger,
) *EventDispatcher {
	return &EventDispatcher{
		subscriptions: make(map[string][]*models.EventSubscription),
		webhookClient: webhookClient,
		wsManager:     wsManager,
		logger:        logger,
	}
}

// RegisterSubscription registers a subscription for event dispatching
func (d *EventDispatcher) RegisterSubscription(subscription *models.EventSubscription) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := d.getSubscriptionKey(subscription)
	d.subscriptions[key] = append(d.subscriptions[key], subscription)

	d.logger.Info("Subscription registered",
		zap.String("subscription_id", subscription.ID),
		zap.String("key", key),
	)
}

// UnregisterSubscription unregisters a subscription
func (d *EventDispatcher) UnregisterSubscription(subscriptionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, subs := range d.subscriptions {
		for i, sub := range subs {
			if sub.ID == subscriptionID {
				// Remove subscription from slice
				d.subscriptions[key] = append(subs[:i], subs[i+1:]...)
				if len(d.subscriptions[key]) == 0 {
					delete(d.subscriptions, key)
				}
				d.logger.Info("Subscription unregistered",
					zap.String("subscription_id", subscriptionID),
					zap.String("key", key),
				)
				return
			}
		}
	}
}

// Dispatch sends event to all matching subscriptions
func (d *EventDispatcher) Dispatch(subscriptionID string, event *models.ChaincodeEvent) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Get all possible keys for this event
	keys := []string{
		d.getEventKey(event.ChannelName, event.ChaincodeName, event.EventName),
		d.getEventKey(event.ChannelName, event.ChaincodeName, ""), // All events for chaincode
		d.getEventKey(event.ChannelName, "", ""),                 // All events for channel
	}

	// Collect all matching subscriptions
	subsToNotify := make(map[string]*models.EventSubscription)
	for _, key := range keys {
		if subs, ok := d.subscriptions[key]; ok {
			for _, sub := range subs {
				if sub.Active && sub.ID == subscriptionID {
					subsToNotify[sub.ID] = sub
				}
			}
		}
	}

	// Dispatch to each subscription
	for _, sub := range subsToNotify {
		d.dispatchToSubscription(sub, event)
	}
}

// DispatchToAll sends event to all matching subscriptions (not just one subscription ID)
func (d *EventDispatcher) DispatchToAll(event *models.ChaincodeEvent) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Get all possible keys for this event
	keys := []string{
		d.getEventKey(event.ChannelName, event.ChaincodeName, event.EventName),
		d.getEventKey(event.ChannelName, event.ChaincodeName, ""), // All events for chaincode
		d.getEventKey(event.ChannelName, "", ""),                 // All events for channel
	}

	// Collect all matching subscriptions
	subsToNotify := make(map[string]*models.EventSubscription)
	for _, key := range keys {
		if subs, ok := d.subscriptions[key]; ok {
			for _, sub := range subs {
				if sub.Active {
					subsToNotify[sub.ID] = sub
				}
			}
		}
	}

	// Dispatch to each subscription
	for _, sub := range subsToNotify {
		d.dispatchToSubscription(sub, event)
	}
}

// dispatchToSubscription dispatches event to a specific subscription
func (d *EventDispatcher) dispatchToSubscription(subscription *models.EventSubscription, event *models.ChaincodeEvent) {
	switch subscription.Type {
	case models.SubscriptionTypeWebSocket:
		if err := d.wsManager.SendEvent(subscription.ID, event); err != nil {
			d.logger.Error("Failed to send WebSocket event",
				zap.String("subscription_id", subscription.ID),
				zap.Error(err),
			)
		}
	case models.SubscriptionTypeSSE:
		if err := d.wsManager.SendSSEEvent(subscription.ID, event); err != nil {
			d.logger.Error("Failed to send SSE event",
				zap.String("subscription_id", subscription.ID),
				zap.Error(err),
			)
		}
	case models.SubscriptionTypeWebhook:
		if err := d.webhookClient.Deliver(subscription, event); err != nil {
			d.logger.Error("Failed to deliver webhook",
				zap.String("subscription_id", subscription.ID),
				zap.Error(err),
			)
		}
	default:
		d.logger.Warn("Unknown subscription type",
			zap.String("subscription_id", subscription.ID),
			zap.String("type", string(subscription.Type)),
		)
	}
}

// getSubscriptionKey generates a key for subscription lookup
func (d *EventDispatcher) getSubscriptionKey(subscription *models.EventSubscription) string {
	return d.getEventKey(subscription.ChannelName, subscription.ChaincodeName, subscription.EventName)
}

// getEventKey generates a key for event matching
func (d *EventDispatcher) getEventKey(channel, chaincode, eventName string) string {
	return fmt.Sprintf("%s:%s:%s", channel, chaincode, eventName)
}

