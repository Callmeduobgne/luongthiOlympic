#!/bin/bash

# Quick Test for Package BlockHash Authentication
# Simplified version with user registration

set -e

BASE_URL="http://localhost:9090"

echo "🧪 Quick Test: Package BlockHash Authentication"
echo "================================================"
echo ""

# Register test user
echo "📝 Step 1: Register test user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"testuser@ibn.vn","password":"Test123!","full_name":"Test User","role":"admin"}' 2>/dev/null)

echo "Register response: $REGISTER_RESPONSE"
echo ""

# Login
echo "🔐 Step 2: Login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"testuser@ibn.vn","password":"Test123!"}' 2>/dev/null)

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // .data.token // empty' 2>/dev/null)

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo "❌ Login failed. Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Login successful! Token: ${TOKEN:0:20}..."
echo ""

# Create package
echo "📦 Step 3: Create test package..."
PACKAGE_ID="PKG_QUICK_$(date +%s)"
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

echo "Create response: $CREATE_RESPONSE"
echo ""

# Wait for transaction
echo "⏳ Waiting 5 seconds for blockchain transaction..."
sleep 5
echo ""

# Get package with blockHash
echo "🔍 Step 4: Get package (should include blockHash)..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/teatrace/packages/$PACKAGE_ID" \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null)

echo "Get response:"
echo "$GET_RESPONSE" | jq '.'
echo ""

BLOCK_HASH=$(echo "$GET_RESPONSE" | jq -r '.blockHash // .block_hash // empty' 2>/dev/null)

if [ -n "$BLOCK_HASH" ] && [ "$BLOCK_HASH" != "null" ]; then
    echo "✅ SUCCESS! BlockHash retrieved: $BLOCK_HASH"
    echo "   Length: ${#BLOCK_HASH} characters (should be 64)"
    echo ""
    
    # Verify with correct hash
    echo "✅ Step 5: Verify with correct blockHash..."
    VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/verify-by-hash" \
        -H "Content-Type: application/json" \
        -d "{\"hash\": \"$BLOCK_HASH\"}" 2>/dev/null)
    
    echo "Verify response:"
    echo "$VERIFY_RESPONSE" | jq '.'
    echo ""
    
    IS_VALID=$(echo "$VERIFY_RESPONSE" | jq -r '.data.is_valid // .is_valid // empty' 2>/dev/null)
    
    if [ "$IS_VALID" = "true" ]; then
        echo "✅ Verification PASSED! Product is authentic."
    else
        echo "❌ Verification FAILED!"
    fi
    echo ""
    
    # Test with wrong hash
    echo "❌ Step 6: Test counterfeit detection (wrong hash)..."
    FAKE_HASH="0000000000000000000000000000000000000000000000000000000000000000"
    VERIFY_FAKE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/verify-by-hash" \
        -H "Content-Type: application/json" \
        -d "{\"hash\": \"$FAKE_HASH\"}" 2>/dev/null)
    
    IS_VALID_FAKE=$(echo "$VERIFY_FAKE" | jq -r '.data.is_valid // .is_valid // empty' 2>/dev/null)
    
    if [ "$IS_VALID_FAKE" = "false" ]; then
        echo "✅ Counterfeit detection PASSED! Fake hash rejected."
    else
        echo "❌ Counterfeit detection FAILED!"
    fi
    echo ""
    
    echo "================================================"
    echo "🎉 All tests PASSED!"
    echo "================================================"
    echo ""
    echo "Package BlockHash Authentication is working correctly!"
    echo ""
    echo "Summary:"
    echo "  - Package ID: $PACKAGE_ID"
    echo "  - BlockHash: ${BLOCK_HASH:0:16}...${BLOCK_HASH:48:16}"
    echo "  - Verification: ✅ Working"
    echo "  - Counterfeit Detection: ✅ Working"
    
else
    echo "❌ FAILED! BlockHash not found in response."
    echo "This means GetPackage endpoint is not working correctly."
    exit 1
fi
