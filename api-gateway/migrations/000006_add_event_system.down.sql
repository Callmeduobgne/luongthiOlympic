-- Rollback event system tables
-- Migration: 000004_add_event_system (rollback)

DROP INDEX IF EXISTS idx_websocket_connections_connected_at;
DROP INDEX IF EXISTS idx_websocket_connections_connection_id;
DROP INDEX IF EXISTS idx_websocket_connections_user_id;
DROP INDEX IF EXISTS idx_websocket_connections_subscription_id;
DROP TABLE IF EXISTS websocket_connections;

DROP INDEX IF EXISTS idx_webhook_deliveries_created_at;
DROP INDEX IF EXISTS idx_webhook_deliveries_status;
DROP INDEX IF EXISTS idx_webhook_deliveries_subscription_id;
DROP TABLE IF EXISTS webhook_deliveries;

DROP INDEX IF EXISTS idx_event_subscriptions_type;
DROP INDEX IF EXISTS idx_event_subscriptions_active;
DROP INDEX IF EXISTS idx_event_subscriptions_chaincode;
DROP INDEX IF EXISTS idx_event_subscriptions_channel;
DROP INDEX IF EXISTS idx_event_subscriptions_api_key_id;
DROP INDEX IF EXISTS idx_event_subscriptions_user_id;
DROP TABLE IF EXISTS event_subscriptions;


