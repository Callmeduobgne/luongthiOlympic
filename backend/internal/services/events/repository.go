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
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles event data access
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new event repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateSubscription creates a new event subscription
func (r *Repository) CreateSubscription(ctx context.Context, sub *EventSubscription) error {
	query := `
		INSERT INTO events.event_subscriptions 
		(id, user_id, channel_id, chaincode_id, event_name, filter_type, filter_value, webhook_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		sub.ID, sub.UserID, sub.ChannelID, sub.ChaincodeID, sub.EventName,
		sub.FilterType, sub.FilterValue, sub.WebhookURL, sub.IsActive,
	).Scan(&sub.CreatedAt, &sub.UpdatedAt)
}

// GetSubscription retrieves a subscription by ID
func (r *Repository) GetSubscription(ctx context.Context, id uuid.UUID) (*EventSubscription, error) {
	sub := &EventSubscription{}
	query := `
		SELECT id, user_id, channel_id, chaincode_id, event_name, filter_type, 
		       filter_value, webhook_url, is_active, created_at, updated_at
		FROM events.event_subscriptions 
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.UserID, &sub.ChannelID, &sub.ChaincodeID, &sub.EventName,
		&sub.FilterType, &sub.FilterValue, &sub.WebhookURL, &sub.IsActive,
		&sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (r *Repository) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]*EventSubscription, error) {
	query := `
		SELECT id, user_id, channel_id, chaincode_id, event_name, filter_type, 
		       filter_value, webhook_url, is_active, created_at, updated_at
		FROM events.event_subscriptions 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*EventSubscription
	for rows.Next() {
		sub := &EventSubscription{}
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ChannelID, &sub.ChaincodeID, &sub.EventName,
			&sub.FilterType, &sub.FilterValue, &sub.WebhookURL, &sub.IsActive,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// GetActiveSubscriptions retrieves all active subscriptions
func (r *Repository) GetActiveSubscriptions(ctx context.Context) ([]*EventSubscription, error) {
	query := `
		SELECT id, user_id, channel_id, chaincode_id, event_name, filter_type, 
		       filter_value, webhook_url, is_active, created_at, updated_at
		FROM events.event_subscriptions 
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []*EventSubscription
	for rows.Next() {
		sub := &EventSubscription{}
		err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ChannelID, &sub.ChaincodeID, &sub.EventName,
			&sub.FilterType, &sub.FilterValue, &sub.WebhookURL, &sub.IsActive,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

// UpdateSubscription updates a subscription
func (r *Repository) UpdateSubscription(ctx context.Context, sub *EventSubscription) error {
	query := `
		UPDATE events.event_subscriptions 
		SET is_active = $2, webhook_url = $3, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, sub.ID, sub.IsActive, sub.WebhookURL)
	return err
}

// DeleteSubscription deletes a subscription
func (r *Repository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events.event_subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CreateWebhookDelivery creates a webhook delivery record
func (r *Repository) CreateWebhookDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	query := `
		INSERT INTO events.webhook_deliveries 
		(id, subscription_id, event_id, event_name, payload, webhook_url, status, attempt_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`

	return r.db.QueryRow(ctx, query,
		delivery.ID, delivery.SubscriptionID, delivery.EventID, delivery.EventName,
		delivery.Payload, delivery.WebhookURL, delivery.Status, delivery.AttemptCount,
	).Scan(&delivery.CreatedAt)
}

// UpdateWebhookDelivery updates a webhook delivery record
func (r *Repository) UpdateWebhookDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	query := `
		UPDATE events.webhook_deliveries 
		SET status = $2, attempt_count = $3, response_code = $4, 
		    response_body = $5, error_message = $6, next_retry_at = $7, 
		    completed_at = $8, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		delivery.ID, delivery.Status, delivery.AttemptCount, delivery.ResponseCode,
		delivery.ResponseBody, delivery.ErrorMessage, delivery.NextRetryAt, delivery.CompletedAt,
	)

	return err
}

// GetPendingWebhookDeliveries retrieves webhook deliveries that need retry
func (r *Repository) GetPendingWebhookDeliveries(ctx context.Context, limit int) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, subscription_id, event_id, event_name, payload, webhook_url, 
		       status, attempt_count, response_code, response_body, error_message, 
		       next_retry_at, created_at, completed_at
		FROM events.webhook_deliveries 
		WHERE status IN ($1, $2) 
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		  AND attempt_count < 5
		ORDER BY created_at ASC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, DeliveryStatusPending, DeliveryStatusRetrying, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		d := &WebhookDelivery{}
		err := rows.Scan(
			&d.ID, &d.SubscriptionID, &d.EventID, &d.EventName, &d.Payload,
			&d.WebhookURL, &d.Status, &d.AttemptCount, &d.ResponseCode,
			&d.ResponseBody, &d.ErrorMessage, &d.NextRetryAt,
			&d.CreatedAt, &d.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan delivery: %w", err)
		}
		deliveries = append(deliveries, d)
	}

	return deliveries, nil
}

// CreateWebsocketConnection creates a websocket connection record
func (r *Repository) CreateWebsocketConnection(ctx context.Context, conn *WebsocketConnection) error {
	query := `
		INSERT INTO events.websocket_connections 
		(id, user_id, ip_address, user_agent, connected_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		conn.ID, conn.UserID, conn.IPAddress, conn.UserAgent, conn.ConnectedAt,
	)

	return err
}

// UpdateWebsocketPing updates the last ping time
func (r *Repository) UpdateWebsocketPing(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE events.websocket_connections SET last_ping_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeleteWebsocketConnection deletes a websocket connection
func (r *Repository) DeleteWebsocketConnection(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events.websocket_connections WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// GetActiveWebsocketConnections retrieves active websocket connections
func (r *Repository) GetActiveWebsocketConnections(ctx context.Context) ([]*WebsocketConnection, error) {
	query := `
		SELECT id, user_id, ip_address, user_agent, connected_at, last_ping_at
		FROM events.websocket_connections 
		WHERE last_ping_at IS NULL OR last_ping_at > NOW() - INTERVAL '2 minutes'
		ORDER BY connected_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}
	defer rows.Close()

	var connections []*WebsocketConnection
	for rows.Next() {
		conn := &WebsocketConnection{}
		err := rows.Scan(
			&conn.ID, &conn.UserID, &conn.IPAddress, &conn.UserAgent,
			&conn.ConnectedAt, &conn.LastPingAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

