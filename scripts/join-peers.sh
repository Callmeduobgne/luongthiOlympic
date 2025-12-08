#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to join peers to the channel

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

CHANNEL_NAME="ibnchannel"
NETWORK_NAME="ibn-network"

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo -e "${GREEN}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║            IBN Network - Join Peers to Channel            ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if fabric-tools image exists
if ! docker image inspect hyperledger/fabric-tools:2.5.9 >/dev/null 2>&1; then
    log_info "Pulling fabric-tools image..."
    docker pull hyperledger/fabric-tools:2.5.9
fi

# Peers to join
PEERS=("peer0.org1.ibn.vn" "peer1.org1.ibn.vn" "peer2.org1.ibn.vn")

# Cổng của peer
# peer0.org1.ibn.vn:7051
# peer1.org1.ibn.vn:8051
# peer2.org1.ibn.vn:9051
for PEER_HOST in "${PEERS[@]}"; do
    # Determine port based on peer host
    PORT=7051
    if [[ "$PEER_HOST" == "peer1.org1.ibn.vn" ]]; then
        PORT=8051
    elif [[ "$PEER_HOST" == "peer2.org1.ibn.vn" ]]; then
        PORT=9051
    fi

    log_info "Joining ${PEER_HOST} on port ${PORT} to channel ${CHANNEL_NAME}..."
    
    # Check if already joined by listing channels
    LIST_RESULT=$(docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/${PEER_HOST}/tls/ca.crt" \
        -e CORE_PEER_ADDRESS="${PEER_HOST}:${PORT}" \
        hyperledger/fabric-tools:2.5.9 \
        peer channel list 2>&1)
        
    if echo "${LIST_RESULT}" | grep -q "${CHANNEL_NAME}"; then
        log_info "${PEER_HOST} is already in channel ${CHANNEL_NAME}"
        continue
    fi
    
    # Join channel
    JOIN_RESULT=$(docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        -e CORE_PEER_LOCALMSPID=Org1MSP \
        -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/${PEER_HOST}/tls/ca.crt" \
        -e CORE_PEER_ADDRESS="${PEER_HOST}:${PORT}" \
        hyperledger/fabric-tools:2.5.9 \
        peer channel join -b "/fabric/channel-artifacts/${CHANNEL_NAME}.block" 2>&1)
        
    if echo "${JOIN_RESULT}" | grep -q "Successfully submitted proposal to join channel"; then
        log_success "${PEER_HOST} joined successfully"
    else
        log_error "Failed to join ${PEER_HOST}: ${JOIN_RESULT}"
        exit 1
    fi
done

echo ""
log_success "All peers have been processed."