-- Seed default approval policies
-- Run this once to initialize approval policies for production

-- Insert default policies
INSERT INTO blockchain.approval_policies (operation, required_approvals, expiration_hours, is_active, conditions)
VALUES 
    ('approve', 1, 24, true, '{"description": "Requires 1 admin approval for chaincode approve operation"}'),
    ('commit', 1, 24, true, '{"description": "Requires 1 admin approval for chaincode commit operation"}'),
    ('install', 1, 24, false, '{"description": "Optional approval for chaincode install operation"}')
ON CONFLICT (operation) DO UPDATE SET
    required_approvals = EXCLUDED.required_approvals,
    expiration_hours = EXCLUDED.expiration_hours,
    is_active = EXCLUDED.is_active,
    conditions = EXCLUDED.conditions,
    updated_at = CURRENT_TIMESTAMP;

-- For production strict mode (uncomment to require 2 admins)
-- UPDATE blockchain.approval_policies 
-- SET required_approvals = 2 
-- WHERE operation IN ('approve', 'commit');

-- Verify
SELECT operation, required_approvals, expiration_hours, is_active 
FROM blockchain.approval_policies 
ORDER BY operation;
