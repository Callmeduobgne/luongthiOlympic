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
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ibn-network/api-gateway/internal/models"
	"github.com/ibn-network/api-gateway/internal/repository/db"
	"go.uber.org/zap"
)

// SubscriptionService handles event subscription operations
type SubscriptionService struct {
	db      *pgxpool.Pool
	queries *db.Queries
	logger  *zap.Logger
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(dbPool *pgxpool.Pool, logger *zap.Logger) *SubscriptionService {
	return &SubscriptionService{
		db:      dbPool,
		queries: db.New(),
		logger:  logger,
	}
}

// CreateSubscription creates a new event subscription
func (s *SubscriptionService) CreateSubscription(
	ctx context.Context,
	req *models.CreateSubscriptionRequest,
	userID string,
	apiKeyID string,
) (*models.EventSubscription, error) {
	// Validate webhook URL for webhook type
	if req.Type == models.SubscriptionTypeWebhook && req.WebhookURL == "" {
		return nil, fmt.Errorf("webhook URL is required for webhook subscriptions")
	}

	// Convert user ID and API key ID to UUID
	var userUUID, apiKeyUUID pgtype.UUID
	if userID != "" {
		if err := userUUID.Scan(userID); err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
	}
	if apiKeyID != "" {
		if err := apiKeyUUID.Scan(apiKeyID); err != nil {
			return nil, fmt.Errorf("invalid API key ID: %w", err)
		}
	}

	// Convert filters to []byte
	var filtersBytes []byte
	if req.Filters != nil {
		var err error
		filtersBytes, err = json.Marshal(req.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filters: %w", err)
		}
	}

	// Create subscription in database
	params := db.CreateEventSubscriptionParams{
		UserID:       userUUID,
		ApiKeyID:     apiKeyUUID,
		Name:         req.Name,
		Type:         string(req.Type),
		ChannelName:  req.ChannelName,
		ChaincodeName: pgtype.Text{String: req.ChaincodeName, Valid: req.ChaincodeName != ""},
		EventName:    pgtype.Text{String: req.EventName, Valid: req.EventName != ""},
		WebhookUrl:   pgtype.Text{String: req.WebhookURL, Valid: req.WebhookURL != ""},
		WebhookSecret: pgtype.Text{String: req.WebhookSecret, Valid: req.WebhookSecret != ""},
		Filters:      filtersBytes,
		Active:       true,
	}

	dbSub, err := s.queries.CreateEventSubscription(ctx, s.db, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return s.dbSubscriptionToModel(&dbSub), nil
}

// GetSubscription retrieves a subscription by ID
func (s *SubscriptionService) GetSubscription(ctx context.Context, subscriptionID string) (*models.EventSubscription, error) {
	subUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription ID: %w", err)
	}

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(subUUID.String()); err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %w", err)
	}

	dbSub, err := s.queries.GetEventSubscriptionByID(ctx, s.db, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	return s.dbSubscriptionToModel(&dbSub), nil
}

// ListSubscriptions lists subscriptions with filters
func (s *SubscriptionService) ListSubscriptions(
	ctx context.Context,
	query *models.SubscriptionListQuery,
	userID string,
	apiKeyID string,
) (*models.SubscriptionListResponse, error) {
	// Convert user ID and API key ID to UUID
	var userUUID, apiKeyUUID pgtype.UUID
	if userID != "" {
		if err := userUUID.Scan(userID); err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
	}
	if apiKeyID != "" {
		if err := apiKeyUUID.Scan(apiKeyID); err != nil {
			return nil, fmt.Errorf("invalid API key ID: %w", err)
		}
	}

	// Set defaults
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	// Build query params
	var activeBool bool
	if query.Active != nil {
		activeBool = *query.Active
	}

	params := db.ListEventSubscriptionsParams{
		Column1: userUUID,
		Column2: apiKeyUUID,
		Column3: query.ChannelName,
		Column4: query.ChaincodeName,
		Column5: string(query.Type),
		Column6: activeBool,
		Limit:   int32(limit),
		Offset:  int32(offset),
	}

	// Get subscriptions
	dbSubs, err := s.queries.ListEventSubscriptions(ctx, s.db, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// Get total count
	countParams := db.CountEventSubscriptionsParams{
		Column1: userUUID,
		Column2: apiKeyUUID,
		Column3: query.ChannelName,
		Column4: query.ChaincodeName,
		Column5: string(query.Type),
		Column6: activeBool,
	}

	total, err := s.queries.CountEventSubscriptions(ctx, s.db, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// Convert to models
	subscriptions := make([]*models.EventSubscription, len(dbSubs))
	for i, dbSub := range dbSubs {
		subscriptions[i] = s.dbSubscriptionToModel(&dbSub)
	}

	return &models.SubscriptionListResponse{
		Subscriptions: subscriptions,
		Total:         total,
		Limit:         limit,
		Offset:        offset,
	}, nil
}

// UpdateSubscription updates a subscription
func (s *SubscriptionService) UpdateSubscription(
	ctx context.Context,
	subscriptionID string,
	req *models.UpdateSubscriptionRequest,
) (*models.EventSubscription, error) {
	subUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription ID: %w", err)
	}

	// Get existing subscription
	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(subUUID.String()); err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %w", err)
	}

	existing, err := s.queries.GetEventSubscriptionByID(ctx, s.db, pgUUID)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}

	// Build update params - use existing values if not provided
	name := existing.Name
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}

	active := existing.Active
	if req.Active != nil {
		active = *req.Active
	}

	webhookUrl := existing.WebhookUrl
	if req.WebhookURL != nil && *req.WebhookURL != "" {
		webhookUrl = pgtype.Text{String: *req.WebhookURL, Valid: true}
	}

	webhookSecret := existing.WebhookSecret
	if req.WebhookSecret != nil && *req.WebhookSecret != "" {
		webhookSecret = pgtype.Text{String: *req.WebhookSecret, Valid: true}
	}

	filters := existing.Filters
	if req.Filters != nil {
		var err error
		filters, err = json.Marshal(req.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filters: %w", err)
		}
	}

	params := db.UpdateEventSubscriptionParams{
		ID:            pgUUID,
		Name:          name,
		Active:        active,
		WebhookUrl:    webhookUrl,
		WebhookSecret: webhookSecret,
		Filters:       filters,
	}

	// Update subscription
	dbSub, err := s.queries.UpdateEventSubscription(ctx, s.db, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return s.dbSubscriptionToModel(&dbSub), nil
}

// DeleteSubscription deletes a subscription
func (s *SubscriptionService) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	subUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		return fmt.Errorf("invalid subscription ID: %w", err)
	}

	var pgUUID pgtype.UUID
	if err := pgUUID.Scan(subUUID.String()); err != nil {
		return fmt.Errorf("failed to convert UUID: %w", err)
	}

	err = s.queries.DeleteEventSubscription(ctx, s.db, pgUUID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	return nil
}

// GetActiveSubscriptionsByChannelAndChaincode gets active subscriptions matching channel and chaincode
func (s *SubscriptionService) GetActiveSubscriptionsByChannelAndChaincode(
	ctx context.Context,
	channelName string,
	chaincodeName string,
	eventName string,
) ([]*models.EventSubscription, error) {
	params := db.GetActiveSubscriptionsByChannelAndChaincodeParams{
		ChannelName: channelName,
		Column2:     chaincodeName,
		Column3:     eventName,
	}

	dbSubs, err := s.queries.GetActiveSubscriptionsByChannelAndChaincode(ctx, s.db, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	subscriptions := make([]*models.EventSubscription, len(dbSubs))
	for i, dbSub := range dbSubs {
		subscriptions[i] = s.dbSubscriptionToModel(&dbSub)
	}

	return subscriptions, nil
}

// dbSubscriptionToModel converts database subscription to model
func (s *SubscriptionService) dbSubscriptionToModel(dbSub *db.EventSubscription) *models.EventSubscription {
	sub := &models.EventSubscription{
		ID:          uuidToString(dbSub.ID),
		Name:        dbSub.Name,
		Type:        models.SubscriptionType(dbSub.Type),
		ChannelName: dbSub.ChannelName,
		Active:      dbSub.Active,
	}

	if dbSub.UserID.Valid {
		sub.UserID = uuidToString(dbSub.UserID)
	}
	if dbSub.ApiKeyID.Valid {
		sub.ApiKeyID = uuidToString(dbSub.ApiKeyID)
	}
	if dbSub.ChaincodeName.Valid {
		sub.ChaincodeName = dbSub.ChaincodeName.String
	}
	if dbSub.EventName.Valid {
		sub.EventName = dbSub.EventName.String
	}
	if dbSub.WebhookUrl.Valid {
		sub.WebhookURL = dbSub.WebhookUrl.String
	}
	if dbSub.WebhookSecret.Valid {
		sub.WebhookSecret = dbSub.WebhookSecret.String
	}
	if len(dbSub.Filters) > 0 {
		var filters map[string]interface{}
		if err := json.Unmarshal(dbSub.Filters, &filters); err == nil {
			sub.Filters = filters
		}
	}

	if dbSub.CreatedAt.Valid {
		sub.CreatedAt = dbSub.CreatedAt.Time
	}
	if dbSub.UpdatedAt.Valid {
		sub.UpdatedAt = dbSub.UpdatedAt.Time
	}

	return sub
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

