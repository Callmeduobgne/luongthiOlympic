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
	"context"
	"encoding/json"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/services/fabric"
	"go.uber.org/zap"
)

// ListenerService listens to blockchain events and dispatches them
type ListenerService struct {
	gateway    *fabric.GatewayService
	dispatcher *EventDispatcher
	logger     *zap.Logger
	listeners  map[string]context.CancelFunc // subscription ID -> cancel function
}

// NewListenerService creates a new event listener service
func NewListenerService(
	gateway *fabric.GatewayService,
	dispatcher *EventDispatcher,
	logger *zap.Logger,
) *ListenerService {
	return &ListenerService{
	gateway:    gateway,
	dispatcher: dispatcher,
	logger:     logger,
	listeners:  make(map[string]context.CancelFunc),
	}
}

// StartListening starts listening to chaincode events for a subscription
func (s *ListenerService) StartListening(ctx context.Context, subscription *models.EventSubscription) error {
	// Check if already listening
	if _, exists := s.listeners[subscription.ID]; exists {
		s.logger.Warn("Already listening to subscription",
			zap.String("subscription_id", subscription.ID),
		)
		return nil
	}

	// Create context for this listener
	listenerCtx, cancel := context.WithCancel(ctx)
	s.listeners[subscription.ID] = cancel

	// Start listening in goroutine
	go s.listenToEvents(listenerCtx, subscription)

	s.logger.Info("Started listening to events",
		zap.String("subscription_id", subscription.ID),
		zap.String("channel", subscription.ChannelName),
		zap.String("chaincode", subscription.ChaincodeName),
	)

	return nil
}

// StopListening stops listening to events for a subscription
func (s *ListenerService) StopListening(subscriptionID string) {
	cancel, exists := s.listeners[subscriptionID]
	if !exists {
		return
	}

	cancel()
	delete(s.listeners, subscriptionID)

	s.logger.Info("Stopped listening to events",
		zap.String("subscription_id", subscriptionID),
	)
}

// listenToEvents listens to chaincode events and dispatches them
func (s *ListenerService) listenToEvents(ctx context.Context, subscription *models.EventSubscription) {
	// Get gateway client
	gatewayClient := s.gateway.GetGatewayClient()
	if gatewayClient == nil {
		s.logger.Error("Gateway client is nil")
		return
	}

	// Get network
	network := gatewayClient.GetNetwork(subscription.ChannelName)
	if network == nil {
		s.logger.Error("Network not found",
			zap.String("channel", subscription.ChannelName),
		)
		return
	}

	// Get chaincode events
	// ChaincodeEvents returns a channel of events
	events, err := network.ChaincodeEvents(ctx, subscription.ChaincodeName)
	if err != nil {
		s.logger.Error("Failed to get chaincode events",
			zap.String("subscription_id", subscription.ID),
			zap.String("chaincode", subscription.ChaincodeName),
			zap.Error(err),
		)
		return
	}

	// Process events from channel
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping event listener",
				zap.String("subscription_id", subscription.ID),
			)
			return
		case event, ok := <-events:
			if !ok {
				// Channel closed
				s.logger.Info("Event channel closed",
					zap.String("subscription_id", subscription.ID),
				)
				return
			}

			// Check if event matches subscription filter
			if !s.matchesFilter(event, subscription) {
				continue
			}

			// Parse payload
			payload, err := s.parsePayload(event.Payload)
			if err != nil {
				s.logger.Warn("Failed to parse event payload",
					zap.String("subscription_id", subscription.ID),
					zap.Error(err),
				)
				// Continue with empty payload
				payload = make(map[string]interface{})
			}

			// Create chaincode event model
			chaincodeEvent := &models.ChaincodeEvent{
				EventName:     event.EventName,
				ChaincodeName: subscription.ChaincodeName,
				ChannelName:   subscription.ChannelName,
				TransactionID: event.TransactionID,
				BlockNumber:   uint64(event.BlockNumber),
				Payload:       payload,
			}

			// Dispatch to all matching subscriptions
			s.dispatcher.Dispatch(subscription.ID, chaincodeEvent)
		}
	}
}

// matchesFilter checks if an event matches the subscription filter
func (s *ListenerService) matchesFilter(event *client.ChaincodeEvent, subscription *models.EventSubscription) bool {
	// Check event name filter
	if subscription.EventName != "" && event.EventName != subscription.EventName {
		return false
	}

	// Additional filters can be added here based on subscription.Filters
	// For now, we just check event name

	return true
}

// parsePayload parses event payload JSON
func (s *ListenerService) parsePayload(payload []byte) (map[string]interface{}, error) {
	if len(payload) == 0 {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(payload, &result); err != nil {
		// If not JSON, return as string
		return map[string]interface{}{
			"raw": string(payload),
		}, nil
	}

	return result, nil
}

// StopAll stops all active listeners
func (s *ListenerService) StopAll() {
	for subscriptionID, cancel := range s.listeners {
		cancel()
		s.logger.Info("Stopped listener",
			zap.String("subscription_id", subscriptionID),
		)
	}
	s.listeners = make(map[string]context.CancelFunc)
}

