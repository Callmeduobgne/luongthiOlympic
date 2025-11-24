-- Add indexes for hash verification queries
-- These indexes optimize queries for verify-by-hash endpoint
-- Note: transactions table is in auth schema (shared with api-gateway)

-- Composite index for teaTraceCC transactions (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_transactions_teatrace_tx_id 
ON auth.transactions(chaincode_name, tx_id) 
WHERE chaincode_name = 'teaTraceCC';

-- Note: blockHash and verificationHash are stored in chaincode state (not in transactions table)
-- These indexes optimize the most common verification queries:
-- 1. Find transaction by tx_id (for transaction ID verification) - uses existing idx_transactions_tx_id
-- 2. Filter transactions by chaincode_name (to check if belongs to teaTraceCC) - uses existing idx_transactions_chaincode
-- 3. Composite index for teaTraceCC + tx_id lookup (fastest path) - new index above
