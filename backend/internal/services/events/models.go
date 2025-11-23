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
	"time"

	"github.com/google/uuid"
)

// EventSubscription represents a subscription to blockchain events
type EventSubscription struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ChannelID   string    `json:"channel_id"`
	ChaincodeID string    `json:"chaincode_id"`
	EventName   string    `json:"event_name"`
	FilterType  string    `json:"filter_type"` // "exact", "prefix", "regex"
	FilterValue string    `json:"filter_value"`
	WebhookURL  *string   `json:"webhook_url,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateSubscriptionRequest represents subscription creation request
type CreateSubscriptionRequest struct {
	ChannelID   string  `json:"channel_id" validate:"required"`
	ChaincodeID string  `json:"chaincode_id" validate:"required"`
	EventName   string  `json:"event_name" validate:"required"`
	FilterType  string  `json:"filter_type"`
	FilterValue string  `json:"filter_value"`
	WebhookURL  *string `json:"webhook_url,omitempty"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID              `json:"id"`
	SubscriptionID uuid.UUID              `json:"subscription_id"`
	EventID        string                 `json:"event_id"`
	EventName      string                 `json:"event_name"`
	Payload        map[string]interface{} `json:"payload"`
	WebhookURL     string                 `json:"webhook_url"`
	Status         string                 `json:"status"`
	AttemptCount   int                    `json:"attempt_count"`
	ResponseCode   *int                   `json:"response_code,omitempty"`
	ResponseBody   *string                `json:"response_body,omitempty"`
	ErrorMessage   *string                `json:"error_message,omitempty"`
	NextRetryAt    *time.Time             `json:"next_retry_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
}

// WebsocketConnection represents an active WebSocket connection
type WebsocketConnection struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
	ConnectedAt time.Time  `json:"connected_at"`
	LastPingAt  *time.Time `json:"last_ping_at,omitempty"`
}

// Event represents a blockchain event
type Event struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channel_id"`
	ChaincodeID string                 `json:"chaincode_id"`
	EventName   string                 `json:"event_name"`
	TxID        string                 `json:"tx_id"`
	BlockNumber uint64                 `json:"block_number"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Webhook delivery status constants
const (
	DeliveryStatusPending   = "pending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"
	DeliveryStatusRetrying  = "retrying"
)

// Filter types
const (
	FilterTypeExact  = "exact"
	FilterTypePrefix = "prefix"
	FilterTypeRegex  = "regex"
)

