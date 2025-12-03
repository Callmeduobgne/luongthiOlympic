#!/bin/bash

<<<<<<< HEAD
# Copyright 2025 IBN Network (ICTU Blockchain Network)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

=======
>>>>>>> a8e22501ad4e0eabb60ad50615b06815c01724dc
# Configuration
API_URL="http://localhost:9900"
EMAIL="admin@ibn.vn"
PASSWORD="Admin123!"

echo "🚀 Starting Live API Test..."

# 0. Register
echo -e "\n0️⃣  Registering new user..."
EMAIL="test_admin_$(date +%s)@ibn.vn"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"role\": \"admin\",
    \"msp_id\": \"Org1MSP\"
  }")

if [[ $REGISTER_RESPONSE == *"error"* ]]; then
  echo "❌ Registration failed!"
  echo "Response: $REGISTER_RESPONSE"
  exit 1
fi
echo "✅ Registration successful for $EMAIL"

# 1. Login
echo -e "\n1️⃣  Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Login failed!"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi
echo "✅ Login successful! Token acquired."

# 2. Create Batch
BATCH_ID="BATCH_TEST_$(date +%s)"
echo -e "\n2️⃣  Creating Batch: $BATCH_ID..."
CREATE_BATCH_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/teatrace/batches" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"batch_id\": \"$BATCH_ID\",
    \"farm_name\": \"Thai Nguyen Farm\",
    \"harvest_date\": \"2025-12-01\",
    \"certification\": \"VietGAP\",
    \"certificate_id\": \"CERT_12345\"
  }")

if [[ $CREATE_BATCH_RESPONSE == *"error"* ]]; then
  echo "❌ Create Batch failed!"
  echo "Response: $CREATE_BATCH_RESPONSE"
  exit 1
fi
echo "✅ Batch created successfully!"
echo "Response: $CREATE_BATCH_RESPONSE"

echo "Sleeping 2s to ensure batch is committed..."
sleep 2

# 3. Create Package
PACKAGE_ID="PKG_TEST_$(date +%s)"
echo -e "\n3️⃣  Creating Package: $PACKAGE_ID..."
CREATE_PKG_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/teatrace/packages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"package_id\": \"$PACKAGE_ID\",
    \"batch_id\": \"$BATCH_ID\",
    \"weight\": 500,
    \"production_date\": \"2025-12-02\",
    \"expiry_date\": \"2026-12-02\"
  }")

if [[ $CREATE_PKG_RESPONSE == *"error"* ]]; then
  echo "❌ Create Package failed!"
  echo "Response: $CREATE_PKG_RESPONSE"
  exit 1
fi
echo "✅ Package created successfully!"
echo "Response: $CREATE_PKG_RESPONSE"

# 4. Verify Package (Get Info)
echo -e "\n4️⃣  Verifying Package Info..."
GET_PKG_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/teatrace/packages/$PACKAGE_ID" \
  -H "Authorization: Bearer $TOKEN")

if [[ $GET_PKG_RESPONSE == *"error"* ]]; then
  echo "❌ Get Package failed!"
  echo "Response: $GET_PKG_RESPONSE"
  exit 1
fi
echo "✅ Package info retrieved successfully!"
echo "Response: $GET_PKG_RESPONSE"

echo -e "\n🎉 Test Completed Successfully!"
