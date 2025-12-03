#!/bin/bash

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

# Setup Fabric wallet for API Gateway

set -e

echo "Setting up Fabric wallet for API Gateway..."

# Check if running from correct directory
if [ ! -d "../core/organizations" ]; then
    echo "Error: Must run from api-gateway directory"
    echo "Usage: cd api-gateway && bash scripts/setup-wallet.sh"
    exit 1
fi

# Create wallet directory
mkdir -p wallet

# Check if certificates exist
CERT_PATH="../core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/signcerts"
KEY_PATH="../core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/keystore"

if [ ! -d "$CERT_PATH" ] || [ ! -d "$KEY_PATH" ]; then
    echo "Error: Certificates not found"
    echo "Please ensure Fabric network is set up and certificates are generated"
    exit 1
fi

echo "✓ Certificates found"
echo "✓ Wallet setup complete"
echo ""
echo "Wallet location: ./wallet"
echo "User: Admin@org1.ibn.vn"
echo "MSP ID: Org1MSP"
echo ""
echo "Next steps:"
echo "1. Set FABRIC_USER_CERT_PATH in .env"
echo "2. Set FABRIC_USER_KEY_PATH in .env"
echo "3. Run 'make test-connection' to verify"

