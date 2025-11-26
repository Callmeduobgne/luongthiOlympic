-- Add channel_id, chaincode_id, filter_type, and filter_value columns to event_subscriptions table
-- Migration: 017_add_channel_id_to_events

-- Add channel_id column (copy from channel_name)
ALTER TABLE events.event_subscriptions 
ADD COLUMN IF NOT EXISTS channel_id VARCHAR(100);

-- Add chaincode_id column (copy from chaincode_name)
ALTER TABLE events.event_subscriptions 
ADD COLUMN IF NOT EXISTS chaincode_id VARCHAR(100);

-- Add filter_type column (for event filtering)
ALTER TABLE events.event_subscriptions 
ADD COLUMN IF NOT EXISTS filter_type VARCHAR(50);

-- Add filter_value column (for event filtering)
ALTER TABLE events.event_subscriptions 
ADD COLUMN IF NOT EXISTS filter_value TEXT;

-- Copy data from channel_name to channel_id
UPDATE events.event_subscriptions 
SET channel_id = channel_name 
WHERE channel_id IS NULL AND channel_name IS NOT NULL;

-- Copy data from chaincode_name to chaincode_id
UPDATE events.event_subscriptions 
SET chaincode_id = chaincode_name 
WHERE chaincode_id IS NULL AND chaincode_name IS NOT NULL;

-- Add index for channel_id
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_channel_id 
ON events.event_subscriptions(channel_id) 
WHERE is_active = TRUE AND deleted_at IS NULL;

-- Add index for chaincode_id
CREATE INDEX IF NOT EXISTS idx_event_subscriptions_chaincode_id 
ON events.event_subscriptions(chaincode_id) 
WHERE is_active = TRUE AND deleted_at IS NULL;

