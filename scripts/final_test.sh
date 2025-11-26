#!/bin/bash

# Final Test for Package BlockHash Authentication
# Uses newly created user

set -e

BASE_URL="http://localhost:9090"
EMAIL="blockhash_test@ibn.vn"
PASSWORD="BlockHash123!"

echo "🧪 Final Test: Package BlockHash Authentication"
echo "================================================"
echo ""

# Login
echo "🔐 Step 1: Login with test user..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" 2>/dev/null)

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // .token // .data.token // empty' 2>/dev/null)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ Login failed. Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Login successful!"
echo ""

# Create package
echo "📦 Step 2: Create test package..."
PACKAGE_ID="PKG_FINAL_$(date +%s)"
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/packages" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"package_id\": \"$PACKAGE_ID\",
        \"batch_id\": \"BATCH001\",
        \"weight\": 500,
        \"production_date\": \"2024-11-26\",
        \"expiry_date\": \"2025-11-26\"
    }" 2>/dev/null)

TX_ID=$(echo "$CREATE_RESPONSE" | jq -r '.tx_id // empty' 2>/dev/null)

if [ -n "$TX_ID" ] && [ "$TX_ID" != "null" ]; then
    echo "✅ Package created! TX ID: $TX_ID"
else
    echo "❌ Failed to create package"
    echo "Response: $CREATE_RESPONSE"
    exit 1
fi
echo ""

# Wait for transaction
echo "⏳ Waiting 5 seconds for blockchain transaction..."
sleep 5
echo ""

# Get package with blockHash
echo "🔍 Step 3: Get package with blockHash..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/teatrace/packages/$PACKAGE_ID" \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null)

echo "$GET_RESPONSE" | jq '.'
echo ""

BLOCK_HASH=$(echo "$GET_RESPONSE" | jq -r '.blockHash // .block_hash // empty' 2>/dev/null)

if [ -n "$BLOCK_HASH" ] && [ "$BLOCK_HASH" != "null" ] && [ ${#BLOCK_HASH} -eq 64 ]; then
    echo "✅ SUCCESS! BlockHash retrieved!"
    echo "   BlockHash: ${BLOCK_HASH:0:32}..."
    echo "   Length: ${#BLOCK_HASH} characters ✓"
    echo ""
else
    echo "❌ FAILED! BlockHash not found or invalid"
    exit 1
fi

# Verify with correct hash
echo "✅ Step 4: Verify with correct blockHash..."
VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/verify-by-hash" \
    -H "Content-Type: application/json" \
    -d "{\"hash\": \"$BLOCK_HASH\"}" 2>/dev/null)

echo "$VERIFY_RESPONSE" | jq '.'
echo ""

IS_VALID=$(echo "$VERIFY_RESPONSE" | jq -r '.data.is_valid // .is_valid // empty' 2>/dev/null)

if [ "$IS_VALID" = "true" ]; then
    echo "✅ Verification PASSED! Product is authentic."
else
    echo "❌ Verification FAILED!"
    exit 1
fi
echo ""

# Test with wrong hash
echo "❌ Step 5: Test counterfeit detection..."
FAKE_HASH="0000000000000000000000000000000000000000000000000000000000000000"
VERIFY_FAKE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/verify-by-hash" \
    -H "Content-Type: application/json" \
    -d "{\"hash\": \"$FAKE_HASH\"}" 2>/dev/null)

IS_VALID_FAKE=$(echo "$VERIFY_FAKE" | jq -r '.data.is_valid // .is_valid // empty' 2>/dev/null)

if [ "$IS_VALID_FAKE" = "false" ]; then
    echo "✅ Counterfeit detection PASSED!"
else
    echo "❌ Counterfeit detection FAILED!"
    exit 1
fi
echo ""

echo "================================================"
echo "🎉 ALL TESTS PASSED!"
echo "================================================"
echo ""
echo "✅ Package BlockHash Authentication is working perfectly!"
echo ""
echo "Test Results:"
echo "  ✓ Backend API responding"
echo "  ✓ Authentication working"
echo "  ✓ Package creation successful"
echo "  ✓ GetPackage endpoint working"
echo "  ✓ BlockHash present and valid (64 chars)"
echo "  ✓ Verification with correct hash: PASS"
echo "  ✓ Counterfeit detection: PASS"
echo ""
echo "Package Details:"
echo "  - Package ID: $PACKAGE_ID"
echo "  - Transaction ID: $TX_ID"
echo "  - BlockHash: ${BLOCK_HASH:0:16}...${BLOCK_HASH:48:16}"
echo ""
echo "System is READY for production! 🚀"
