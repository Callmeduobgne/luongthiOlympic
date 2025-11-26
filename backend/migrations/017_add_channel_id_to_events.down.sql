-- Rollback: Remove channel_id, chaincode_id, filter_type, and filter_value columns
-- Migration: 017_add_channel_id_to_events

-- Drop indexes
DROP INDEX IF EXISTS idx_event_subscriptions_chaincode_id;
DROP INDEX IF EXISTS idx_event_subscriptions_channel_id;

-- Drop columns
ALTER TABLE events.event_subscriptions 
DROP COLUMN IF EXISTS filter_value;

ALTER TABLE events.event_subscriptions 
DROP COLUMN IF EXISTS filter_type;

ALTER TABLE events.event_subscriptions 
DROP COLUMN IF EXISTS chaincode_id;

ALTER TABLE events.event_subscriptions 
DROP COLUMN IF EXISTS channel_id;

