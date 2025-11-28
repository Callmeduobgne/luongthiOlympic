#!/bin/bash

# Test login script
# Usage: ./test_login.sh <email> <password>

EMAIL="${1:-callmeduongne@ibn.vn}"
PASSWORD="${2:-callmenam1}"
API_URL="http://localhost:9900/api/v1/auth/login"

echo "========================================="
echo "Testing Login"
echo "========================================="
echo "Email: $EMAIL"
echo "Password: ****"
echo "API URL: $API_URL"
echo "========================================="

# Make login request
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

# Extract HTTP status code (last line)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

# Extract response body (all but last line)
BODY=$(echo "$RESPONSE" | sed '$d')

echo ""
echo "HTTP Status Code: $HTTP_CODE"
echo ""
echo "Response Body:"
echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
echo ""

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Login successful!"
    
    # Extract access token
    ACCESS_TOKEN=$(echo "$BODY" | jq -r '.access_token // .accessToken // .data.access_token // .data.accessToken' 2>/dev/null)
    
    if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
        echo ""
        echo "Access Token (first 50 chars): ${ACCESS_TOKEN:0:50}..."
        echo ""
        echo "Testing profile endpoint..."
        
        # Test profile endpoint
        PROFILE_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "http://localhost:9900/api/v1/profile" \
          -H "Authorization: Bearer $ACCESS_TOKEN")
        
        PROFILE_HTTP_CODE=$(echo "$PROFILE_RESPONSE" | tail -n1)
        PROFILE_BODY=$(echo "$PROFILE_RESPONSE" | sed '$d')
        
        echo "Profile HTTP Status: $PROFILE_HTTP_CODE"
        echo "Profile Response:"
        echo "$PROFILE_BODY" | jq '.' 2>/dev/null || echo "$PROFILE_BODY"
    fi
else
    echo "❌ Login failed!"
    echo ""
    echo "Checking if user exists in database..."
    docker exec ibn-postgres psql -U gateway -d ibn_gateway -c "SELECT email, username, role, is_active FROM users WHERE email = '$EMAIL';"
fi

echo ""
echo "========================================="
