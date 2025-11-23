-- Add event system tables for event subscriptions, webhooks, and WebSocket connections
-- Migration: 000004_add_event_system

-- Event Subscriptions Table
CREATE TABLE IF NOT EXISTS event_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    api_key_id UUID REFERENCES api_keys(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('websocket', 'sse', 'webhook')),
    channel_name VARCHAR(255) NOT NULL,
    chaincode_name VARCHAR(255),
    event_name VARCHAR(255), -- NULL = all events
    webhook_url TEXT, -- For webhook type
    webhook_secret VARCHAR(255), -- For webhook signature
    filters JSONB, -- Additional filters
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Webhook Deliveries Table (Audit Trail)
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID REFERENCES event_subscriptions(id) ON DELETE CASCADE,
    event_id VARCHAR(255) NOT NULL,
    webhook_url TEXT NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'success', 'failed')),
    status_code INT,
    response_body TEXT,
    error_message TEXT,
    attempts INT NOT NULL DEFAULT 0,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- WebSocket Connections Table
CREATE TABLE IF NOT EXISTS websocket_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID REFERENCES event_subscriptions(id) ON DELETE CASCADE,
    connection_id VARCHAR(255) UNIQUE NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    connected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    disconnected_at TIMESTAMP,
    last_ping_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_user_id ON event_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_api_key_id ON event_subscriptions(api_key_id);
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_channel ON event_subscriptions(channel_name);
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_chaincode ON event_subscriptions(chaincode_name);
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_active ON event_subscriptions(active);
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_type ON event_subscriptions(type);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_subscription_id ON webhook_deliveries(subscription_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created_at ON webhook_deliveries(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_websocket_connections_subscription_id ON websocket_connections(subscription_id);
CREATE INDEX IF NOT EXISTS idx_websocket_connections_user_id ON websocket_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_websocket_connections_connection_id ON websocket_connections(connection_id);
CREATE INDEX IF NOT EXISTS idx_websocket_connections_connected_at ON websocket_connections(connected_at DESC);


