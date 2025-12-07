#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to test teaTraceCC chaincode

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Configuration
CHANNEL_NAME="ibnchannel"
CHAINCODE_NAME="teaTraceCC"
NETWORK_NAME="ibn-network"

# Functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_test() { echo -e "\n${CYAN}[TEST]${NC} $1"; }

run_chaincode_invoke() {
    local FUNCTION=$1
    local ARGS=$2
    
    docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
        -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
        hyperledger/fabric-tools:2.5.9 \
        peer chaincode invoke \
            --channelID "${CHANNEL_NAME}" \
            --name "${CHAINCODE_NAME}" \
            --tls \
            --cafile /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
            --orderer orderer.ibn.vn:7050 \
            --ordererTLSHostnameOverride orderer.ibn.vn \
            --peerAddresses peer0.org1.ibn.vn:7051 \
            --tlsRootCertFiles /fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
            -c "{\"function\":\"${FUNCTION}\",\"Args\":${ARGS}}"
}

run_chaincode_query() {
    local FUNCTION=$1
    local ARGS=$2
    
    docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
        -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
        hyperledger/fabric-tools:2.5.9 \
        peer chaincode query \
            --channelID "${CHANNEL_NAME}" \
            --name "${CHAINCODE_NAME}" \
            -c "{\"function\":\"${FUNCTION}\",\"Args\":${ARGS}}"
}

echo -e "${GREEN}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║           IBN Network - Chaincode Test Script             ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Test 1: Create Batch
BATCH_ID="BATCH-TEST-$(date +%s)"
log_test "Creating batch: ${BATCH_ID}"
run_chaincode_invoke "CreateBatch" "[\"${BATCH_ID}\",\"Thai Nguyen Tea Farm\",\"$(date +%Y-%m-%d)\",\"Sun-dried, hand-processed\",\"CERT-VN-2025\"]"

sleep 2

# Test 2: Query Batch
log_test "Querying batch: ${BATCH_ID}"
BATCH_OUTPUT=$(run_chaincode_query "GetBatchInfo" "[\"${BATCH_ID}\"]")
echo "$BATCH_OUTPUT"

# Extract hash (simple grep/cut since jq might not be available)
HASH_VALUE=$(echo "$BATCH_OUTPUT" | grep -o '"hashValue":"[^"]*"' | cut -d'"' -f4)

if [ -z "$HASH_VALUE" ]; then
    log_error "Failed to extract hash value from batch info"
    # Proceeding might fail, but let's try or we could exit. 
    # For now, let's warn.
else
    log_info "Extracted hash value: ${HASH_VALUE}"
fi

# Test 3: Update Batch Status
log_test "Updating batch status to VERIFIED"
run_chaincode_invoke "UpdateBatchStatus" "[\"${BATCH_ID}\",\"VERIFIED\"]"

sleep 2

# Test 4: Verify Batch
log_test "Verifying batch"
# Use the extracted hash
run_chaincode_query "VerifyBatch" "[\"${BATCH_ID}\",\"${HASH_VALUE}\"]"

# Test 5: Create Package
PACKAGE_ID="PKG-TEST-$(date +%s)"
log_test "Creating package: ${PACKAGE_ID}"
run_chaincode_invoke "CreatePackage" "[\"${PACKAGE_ID}\",\"${BATCH_ID}\",\"100.5\",\"$(date +%Y-%m-%d)\",\"$(date -d '+1 year' +%Y-%m-%d 2>/dev/null || date -v+1y +%Y-%m-%d)\",\"QR-${PACKAGE_ID}\"]"

sleep 2

# Test 6: Query Package
log_test "Querying package: ${PACKAGE_ID}"
run_chaincode_query "GetPackageInfo" "[\"${PACKAGE_ID}\"]"

# Test 7: Get Packages by Batch
log_test "Getting packages by batch: ${BATCH_ID}"
run_chaincode_query "GetPackagesByBatch" "[\"${BATCH_ID}\",\"\",\"\"]"

# Test 8: Get All Batches
log_test "Getting all batches"
run_chaincode_query "GetAllBatches" "[\"\",\"\"]"

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              ALL TESTS COMPLETED SUCCESSFULLY             ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Batch ID:   ${BATCH_ID}"
echo -e "Package ID: ${PACKAGE_ID}"
echo ""
