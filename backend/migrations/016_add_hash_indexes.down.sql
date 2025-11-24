-- Remove indexes added for hash verification

DROP INDEX IF EXISTS auth.idx_transactions_teatrace_tx_id;
-- Note: idx_transactions_tx_id and idx_transactions_chaincode already exist from api-gateway migrations
-- We don't drop them as they're used by other queries. This migration only adds the composite index.

