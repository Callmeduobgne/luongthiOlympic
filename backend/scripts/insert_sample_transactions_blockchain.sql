-- Insert sample transactions for teaTraceCC chaincode into blockchain.transactions
-- This script creates sample transaction data for testing

-- First, get a user_id (use the first user if exists)
DO $$
DECLARE
    v_user_id UUID;
    v_tx_id VARCHAR(255);
    v_timestamp TIMESTAMPTZ;
BEGIN
    -- Get first user
    SELECT id INTO v_user_id FROM auth.users LIMIT 1;
    
    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'No users found. Please create a user first.';
    END IF;

    -- Insert sample transactions for teaTraceCC
    v_timestamp := NOW() - INTERVAL '1 day';
    
    -- Transaction 1: createBatch
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'createBatch',
        '["BATCH001", "Moc Chau", "2024-11-25", "Organic Processing", "VN-ORG-2024"]'::jsonb,
        '{"source": "farmer"}'::jsonb,
        'committed',
        1001,
        v_timestamp,
        v_timestamp + INTERVAL '5 seconds'
    );

    -- Transaction 2: createPackage
    v_timestamp := NOW() - INTERVAL '12 hours';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'createPackage',
        '["PKG001", "BATCH001", "500", "2024-11-25", "2025-11-25"]'::jsonb,
        '{"weight": 500, "unit": "gram"}'::jsonb,
        'committed',
        1002,
        v_timestamp,
        v_timestamp + INTERVAL '5 seconds'
    );

    -- Transaction 3: verifyBatch
    v_timestamp := NOW() - INTERVAL '6 hours';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'verifyBatch',
        '["BATCH001", "abc123def456..."]'::jsonb,
        NULL::jsonb,
        'committed',
        1003,
        v_timestamp,
        v_timestamp + INTERVAL '5 seconds'
    );

    -- Transaction 4: getAllBatches
    v_timestamp := NOW() - INTERVAL '3 hours';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'getAllBatches',
        '["100", "0"]'::jsonb,
        NULL::jsonb,
        'committed',
        1004,
        v_timestamp,
        v_timestamp + INTERVAL '2 seconds'
    );

    -- Transaction 5: createPackage (another one)
    v_timestamp := NOW() - INTERVAL '1 hour';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'createPackage',
        '["PKG002", "BATCH001", "1000", "2024-11-26", "2025-11-26"]'::jsonb,
        '{"weight": 1000, "unit": "gram"}'::jsonb,
        'committed',
        1005,
        v_timestamp,
        v_timestamp + INTERVAL '5 seconds'
    );

    -- Transaction 6: getPackagesByBatch
    v_timestamp := NOW() - INTERVAL '30 minutes';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'getPackagesByBatch',
        '["BATCH001", "50", "0"]'::jsonb,
        NULL::jsonb,
        'committed',
        1006,
        v_timestamp,
        v_timestamp + INTERVAL '2 seconds'
    );

    -- Transaction 7: verifyPackage
    v_timestamp := NOW() - INTERVAL '15 minutes';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'verifyPackage',
        '["PKG001", "hash123..."]'::jsonb,
        NULL::jsonb,
        'committed',
        1007,
        v_timestamp,
        v_timestamp + INTERVAL '3 seconds'
    );

    -- Transaction 8: updateBatchStatus
    v_timestamp := NOW() - INTERVAL '10 minutes';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'updateBatchStatus',
        '["BATCH001", "VERIFIED"]'::jsonb,
        NULL::jsonb,
        'committed',
        1008,
        v_timestamp,
        v_timestamp + INTERVAL '4 seconds'
    );

    -- Transaction 9: getBatchInfo
    v_timestamp := NOW() - INTERVAL '5 minutes';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'getBatchInfo',
        '["BATCH001"]'::jsonb,
        NULL::jsonb,
        'committed',
        1009,
        v_timestamp,
        v_timestamp + INTERVAL '1 second'
    );

    -- Transaction 10: createPackage (latest)
    v_timestamp := NOW() - INTERVAL '2 minutes';
    v_tx_id := 'tx_' || substr(md5(random()::text), 1, 16) || '_' || extract(epoch from now())::bigint;
    INSERT INTO blockchain.transactions (
        id, tx_id, user_id, channel_name, chaincode_name, function_name,
        args, transient_data, status, block_number, submitted_at, committed_at
    ) VALUES (
        gen_random_uuid(),
        v_tx_id,
        v_user_id,
        'ibnchannel',
        'teaTraceCC',
        'createPackage',
        '["PKG003", "BATCH001", "750", "2024-11-26", "2025-11-26"]'::jsonb,
        '{"weight": 750, "unit": "gram"}'::jsonb,
        'committed',
        1010,
        v_timestamp,
        v_timestamp + INTERVAL '5 seconds'
    );

    RAISE NOTICE 'Inserted 10 sample transactions for teaTraceCC chaincode into blockchain.transactions';
    RAISE NOTICE 'User ID: %', v_user_id;
END $$;

-- Verify inserted data
SELECT 
    tx_id,
    chaincode_name,
    function_name,
    status,
    block_number,
    submitted_at
FROM blockchain.transactions
WHERE chaincode_name = 'teaTraceCC'
ORDER BY COALESCE(submitted_at, created_at) DESC
LIMIT 10;


