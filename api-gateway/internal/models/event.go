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

package models

import (
	"time"
)

// SubscriptionType represents the type of event subscription
type SubscriptionType string

const (
	SubscriptionTypeWebSocket SubscriptionType = "websocket"
	SubscriptionTypeSSE       SubscriptionType = "sse"
	SubscriptionTypeWebhook   SubscriptionType = "webhook"
)

// EventSubscription represents an event subscription
type EventSubscription struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"userId,omitempty"`
	ApiKeyID     string                 `json:"apiKeyId,omitempty"`
	Name         string                 `json:"name"`
	Type         SubscriptionType       `json:"type"`
	ChannelName  string                 `json:"channelName"`
	ChaincodeName string                `json:"chaincodeName,omitempty"`
	EventName    string                 `json:"eventName,omitempty"`
	WebhookURL   string                 `json:"webhookUrl,omitempty"`
	WebhookSecret string                `json:"webhookSecret,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Active       bool                   `json:"active"`
	CreatedAt    time.Time              `json:"createdAt"`
	UpdatedAt    time.Time              `json:"updatedAt"`
}

// CreateSubscriptionRequest represents a request to create an event subscription
type CreateSubscriptionRequest struct {
	Name         string                 `json:"name" validate:"required"`
	Type         SubscriptionType       `json:"type" validate:"required,oneof=websocket sse webhook"`
	ChannelName  string                 `json:"channelName" validate:"required"`
	ChaincodeName string                `json:"chaincodeName,omitempty"`
	EventName    string                 `json:"eventName,omitempty"`
	WebhookURL   string                 `json:"webhookUrl,omitempty" validate:"required_if=Type webhook"`
	WebhookSecret string                `json:"webhookSecret,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
}

// UpdateSubscriptionRequest represents a request to update an event subscription
type UpdateSubscriptionRequest struct {
	Name         *string                `json:"name,omitempty"`
	Active       *bool                  `json:"active,omitempty"`
	WebhookURL   *string                `json:"webhookUrl,omitempty"`
	WebhookSecret *string               `json:"webhookSecret,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
}

// ChaincodeEvent represents a chaincode event from Fabric
type ChaincodeEvent struct {
	EventName     string                 `json:"eventName"`
	ChaincodeName string                 `json:"chaincodeName"`
	ChannelName   string                 `json:"channelName"`
	TransactionID string                 `json:"transactionId"`
	BlockNumber   uint64                 `json:"blockNumber"`
	Payload       map[string]interface{} `json:"payload"`
	Timestamp     time.Time              `json:"timestamp"`
}

// WebhookDelivery represents a webhook delivery record
type WebhookDelivery struct {
	ID            string                 `json:"id"`
	SubscriptionID string                `json:"subscriptionId"`
	EventID       string                 `json:"eventId"`
	WebhookURL    string                 `json:"webhookUrl"`
	Payload       map[string]interface{} `json:"payload"`
	Status        string                 `json:"status"` // pending, success, failed
	StatusCode    int                    `json:"statusCode,omitempty"`
	ResponseBody  string                 `json:"responseBody,omitempty"`
	ErrorMessage  string                 `json:"errorMessage,omitempty"`
	Attempts      int                    `json:"attempts"`
	DeliveredAt   *time.Time             `json:"deliveredAt,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection struct {
	ID            string     `json:"id"`
	SubscriptionID string   `json:"subscriptionId"`
	ConnectionID  string    `json:"connectionId"`
	UserID        string    `json:"userId"`
	IPAddress     string    `json:"ipAddress,omitempty"`
	UserAgent     string    `json:"userAgent,omitempty"`
	ConnectedAt   time.Time `json:"connectedAt"`
	DisconnectedAt *time.Time `json:"disconnectedAt,omitempty"`
	LastPingAt    *time.Time `json:"lastPingAt,omitempty"`
}

// SubscriptionListQuery represents query parameters for listing subscriptions
type SubscriptionListQuery struct {
	ChannelName   string `json:"channelName,omitempty"`
	ChaincodeName string `json:"chaincodeName,omitempty"`
	Type          SubscriptionType `json:"type,omitempty"`
	Active        *bool  `json:"active,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
}

// SubscriptionListResponse represents a paginated list of subscriptions
type SubscriptionListResponse struct {
	Subscriptions []*EventSubscription `json:"subscriptions"`
	Total         int64                `json:"total"`
	Limit         int                   `json:"limit"`
	Offset        int                   `json:"offset"`
}

