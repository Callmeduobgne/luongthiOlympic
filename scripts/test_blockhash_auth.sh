#!/bin/bash

# Comprehensive Test Suite for Package BlockHash Authentication
# Tests all components: chaincode, backend API, QR code, verification

set -e

BASE_URL="http://localhost:9090"
BACKEND_HEALTHY=false
TOKEN=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

echo "🧪 Package BlockHash Authentication - Comprehensive Test Suite"
echo "=============================================================="
echo ""

# Helper function to print test result
test_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✅ PASS${NC} - $test_name"
        [ -n "$message" ] && echo "   $message"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}❌ FAIL${NC} - $test_name"
        [ -n "$message" ] && echo "   $message"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test 1: Check if backend is running
echo -e "${BLUE}📡 Test 1: Backend Health Check${NC}"
if curl -s -f "$BASE_URL/health" > /dev/null 2>&1; then
    test_result "Backend is running" "PASS" "Backend at $BASE_URL is healthy"
    BACKEND_HEALTHY=true
else
    test_result "Backend is running" "FAIL" "Backend at $BASE_URL is not accessible"
    echo ""
    echo "⚠️  Backend is not running. Please start it with:"
    echo "   docker-compose up -d ibn-backend"
    exit 1
fi
echo ""

# Test 2: Login and get token
echo -e "${BLUE}🔐 Test 2: Authentication${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@ibn.vn","password":"Admin123!"}' 2>/dev/null)

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token // .data.token // empty' 2>/dev/null)

if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    test_result "Login successful" "PASS" "Token obtained: ${TOKEN:0:20}..."
else
    test_result "Login successful" "FAIL" "Failed to obtain authentication token"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo ""

# Test 3: Create a test package
echo -e "${BLUE}📦 Test 3: Create Package${NC}"
PACKAGE_ID="PKG_TEST_$(date +%s)"
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/packages" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"package_id\": \"$PACKAGE_ID\",
        \"batch_id\": \"BATCH001\",
        \"weight\": 750,
        \"production_date\": \"2024-11-26\",
        \"expiry_date\": \"2025-11-26\"
    }" 2>/dev/null)

TX_ID=$(echo "$CREATE_RESPONSE" | jq -r '.tx_id // empty' 2>/dev/null)

if [ -n "$TX_ID" ] && [ "$TX_ID" != "null" ]; then
    test_result "Create package" "PASS" "Package $PACKAGE_ID created with tx_id: $TX_ID"
else
    test_result "Create package" "FAIL" "Failed to create package"
    echo "Response: $CREATE_RESPONSE"
fi
echo ""

# Wait for transaction to be committed
echo "⏳ Waiting 3 seconds for transaction to be committed..."
sleep 3
echo ""

# Test 4: Get Package (NEW ENDPOINT - Main test)
echo -e "${BLUE}🔍 Test 4: Get Package with BlockHash${NC}"
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/teatrace/packages/$PACKAGE_ID" \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null)

BLOCK_HASH=$(echo "$GET_RESPONSE" | jq -r '.blockHash // .block_hash // empty' 2>/dev/null)
PACKAGE_ID_RESP=$(echo "$GET_RESPONSE" | jq -r '.packageId // .package_id // empty' 2>/dev/null)

if [ -n "$BLOCK_HASH" ] && [ "$BLOCK_HASH" != "null" ] && [ ${#BLOCK_HASH} -eq 64 ]; then
    test_result "GetPackage endpoint" "PASS" "BlockHash retrieved: ${BLOCK_HASH:0:16}... (64 chars)"
    test_result "BlockHash format" "PASS" "BlockHash is valid SHA-256 hex (64 characters)"
else
    test_result "GetPackage endpoint" "FAIL" "BlockHash not found or invalid format"
    echo "Response: $GET_RESPONSE"
    echo "BlockHash: $BLOCK_HASH"
fi

if [ "$PACKAGE_ID_RESP" = "$PACKAGE_ID" ]; then
    test_result "Package data integrity" "PASS" "Package ID matches: $PACKAGE_ID"
else
    test_result "Package data integrity" "FAIL" "Package ID mismatch"
fi
echo ""

# Test 5: Verify Package with Correct BlockHash
echo -e "${BLUE}✅ Test 5: Verify with Correct BlockHash${NC}"
if [ -n "$BLOCK_HASH" ] && [ "$BLOCK_HASH" != "null" ]; then
    VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/verify-by-hash" \
        -H "Content-Type: application/json" \
        -d "{\"hash\": \"$BLOCK_HASH\"}" 2>/dev/null)
    
    IS_VALID=$(echo "$VERIFY_RESPONSE" | jq -r '.data.is_valid // .is_valid // empty' 2>/dev/null)
    
    if [ "$IS_VALID" = "true" ]; then
        test_result "Verify correct hash" "PASS" "Product verified as authentic"
    else
        test_result "Verify correct hash" "FAIL" "Verification failed for correct hash"
        echo "Response: $VERIFY_RESPONSE"
    fi
else
    test_result "Verify correct hash" "FAIL" "No blockHash available for verification"
fi
echo ""

# Test 6: Verify with Wrong BlockHash (Counterfeit Detection)
echo -e "${BLUE}❌ Test 6: Verify with Wrong BlockHash (Counterfeit Detection)${NC}"
FAKE_HASH="0000000000000000000000000000000000000000000000000000000000000000"
VERIFY_FAKE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teatrace/verify-by-hash" \
    -H "Content-Type: application/json" \
    -d "{\"hash\": \"$FAKE_HASH\"}" 2>/dev/null)

IS_VALID_FAKE=$(echo "$VERIFY_FAKE_RESPONSE" | jq -r '.data.is_valid // .is_valid // empty' 2>/dev/null)

if [ "$IS_VALID_FAKE" = "false" ]; then
    test_result "Detect counterfeit" "PASS" "Fake hash correctly rejected"
else
    test_result "Detect counterfeit" "FAIL" "Fake hash was not rejected"
    echo "Response: $VERIFY_FAKE_RESPONSE"
fi
echo ""

# Test 7: QR Code Generation
echo -e "${BLUE}📱 Test 7: QR Code Generation${NC}"
QR_RESPONSE=$(curl -s -w "%{http_code}" -o /tmp/qr_test.png \
    -X GET "$BASE_URL/api/v1/qrcode/packages/$PACKAGE_ID" \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null)

if [ "$QR_RESPONSE" = "200" ] && [ -f /tmp/qr_test.png ]; then
    FILE_TYPE=$(file /tmp/qr_test.png | grep -o "PNG image data")
    if [ -n "$FILE_TYPE" ]; then
        FILE_SIZE=$(stat -f%z /tmp/qr_test.png 2>/dev/null || stat -c%s /tmp/qr_test.png 2>/dev/null)
        test_result "QR code generation" "PASS" "QR code generated (${FILE_SIZE} bytes)"
    else
        test_result "QR code generation" "FAIL" "File is not a valid PNG"
    fi
else
    test_result "QR code generation" "FAIL" "Failed to generate QR code (HTTP $QR_RESPONSE)"
fi
echo ""

# Test 8: Package Data Completeness
echo -e "${BLUE}📋 Test 8: Package Data Completeness${NC}"
BATCH_ID=$(echo "$GET_RESPONSE" | jq -r '.batchId // .batch_id // empty' 2>/dev/null)
WEIGHT=$(echo "$GET_RESPONSE" | jq -r '.weight // empty' 2>/dev/null)
PROD_DATE=$(echo "$GET_RESPONSE" | jq -r '.productionDate // .production_date // empty' 2>/dev/null)
STATUS=$(echo "$GET_RESPONSE" | jq -r '.status // empty' 2>/dev/null)

FIELDS_OK=true
[ -z "$BATCH_ID" ] || [ "$BATCH_ID" = "null" ] && FIELDS_OK=false
[ -z "$WEIGHT" ] || [ "$WEIGHT" = "null" ] && FIELDS_OK=false
[ -z "$PROD_DATE" ] || [ "$PROD_DATE" = "null" ] && FIELDS_OK=false
[ -z "$STATUS" ] || [ "$STATUS" = "null" ] && FIELDS_OK=false

if [ "$FIELDS_OK" = true ]; then
    test_result "Package data fields" "PASS" "All required fields present"
    echo "   Batch: $BATCH_ID, Weight: $WEIGHT, Date: $PROD_DATE, Status: $STATUS"
else
    test_result "Package data fields" "FAIL" "Missing required fields"
fi
echo ""

# Test 9: Get Non-Existent Package (Error Handling)
echo -e "${BLUE}🚫 Test 9: Error Handling (Non-Existent Package)${NC}"
ERROR_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null \
    -X GET "$BASE_URL/api/v1/teatrace/packages/PKG_NONEXISTENT" \
    -H "Authorization: Bearer $TOKEN" 2>/dev/null)

if [ "$ERROR_RESPONSE" = "404" ]; then
    test_result "404 for non-existent package" "PASS" "Correctly returns 404 Not Found"
else
    test_result "404 for non-existent package" "FAIL" "Expected 404, got $ERROR_RESPONSE"
fi
echo ""

# Summary
echo "=============================================================="
echo -e "${BLUE}📊 Test Summary${NC}"
echo "=============================================================="
echo -e "Total Tests:  $TESTS_TOTAL"
echo -e "${GREEN}Passed:       $TESTS_PASSED${NC}"
echo -e "${RED}Failed:       $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed! Package BlockHash authentication is working correctly.${NC}"
    echo ""
    echo "✅ System is ready for production use!"
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed. Please review the errors above.${NC}"
    exit 1
fi
