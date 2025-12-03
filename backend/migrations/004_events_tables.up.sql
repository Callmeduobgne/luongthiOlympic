-- Events schema tables

-- Event subscriptions table
CREATE TABLE events.event_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    channel_name VARCHAR(100) NOT NULL,
    chaincode_name VARCHAR(100),
    event_name VARCHAR(255),
    filters JSONB, -- Event filtering rules
    delivery_type VARCHAR(50) NOT NULL, -- webhook, websocket
    webhook_url TEXT,
    webhook_secret VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    retry_count INT DEFAULT 3,
    timeout_seconds INT DEFAULT 30,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Webhook deliveries table (for tracking)
CREATE TABLE events.webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES events.event_subscriptions(id) ON DELETE CASCADE,
    event_name VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, sent, failed
    http_status INT,
    response_body TEXT,
    error_message TEXT,
    attempt_count INT DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- WebSocket connections table
CREATE TABLE events.websocket_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    connection_id VARCHAR(255) UNIQUE NOT NULL,
    subscriptions UUID[] DEFAULT '{}', -- Array of subscription IDs
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    connected_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_ping_at TIMESTAMPTZ,
    disconnected_at TIMESTAMPTZ
);

-- Indexes for events schema
CREATE INDEX idx_event_subscriptions_user_id ON events.event_subscriptions(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_event_subscriptions_channel ON events.event_subscriptions(channel_name) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_event_subscriptions_chaincode ON events.event_subscriptions(chaincode_name) WHERE is_active = TRUE AND deleted_at IS NULL;
CREATE INDEX idx_event_subscriptions_is_active ON events.event_subscriptions(is_active) WHERE deleted_at IS NULL;

CREATE INDEX idx_webhook_deliveries_subscription_id ON events.webhook_deliveries(subscription_id);
CREATE INDEX idx_webhook_deliveries_status ON events.webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_next_retry ON events.webhook_deliveries(next_retry_at) WHERE status = 'failed' AND attempt_count < 3;
CREATE INDEX idx_webhook_deliveries_created_at ON events.webhook_deliveries(created_at DESC);

CREATE INDEX idx_websocket_connections_user_id ON events.websocket_connections(user_id);
CREATE INDEX idx_websocket_connections_connection_id ON events.websocket_connections(connection_id);
CREATE INDEX idx_websocket_connections_is_active ON events.websocket_connections(is_active);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_event_subscriptions_updated_at BEFORE UPDATE ON events.event_subscriptions
    FOR EACH ROW EXECUTE FUNCTION auth.update_updated_at_column();

