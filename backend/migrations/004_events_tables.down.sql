-- Drop events tables
DROP TRIGGER IF EXISTS update_event_subscriptions_updated_at ON events.event_subscriptions;

DROP TABLE IF EXISTS events.websocket_connections;
DROP TABLE IF EXISTS events.webhook_deliveries;
DROP TABLE IF EXISTS events.event_subscriptions;

