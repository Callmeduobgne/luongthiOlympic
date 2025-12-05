#!/bin/bash

# Copyright (c) 2025 IBN Network
# Script to join all orderers to the channel

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
echo "║         IBN Network - Join Orderers to Channel            ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if fabric-tools image exists
if ! docker image inspect hyperledger/fabric-tools:2.5.9 >/dev/null 2>&1; then
    log_info "Pulling fabric-tools image..."
    docker pull hyperledger/fabric-tools:2.5.9
fi

# Orderer configurations: name:admin_port
declare -a ORDERERS=(
    "orderer.ibn.vn:9443"
    "orderer1.ibn.vn:10443"
    "orderer2.ibn.vn:11443"
)

for ORDERER_CONFIG in "${ORDERERS[@]}"; do
    ORDERER_NAME="${ORDERER_CONFIG%%:*}"
    ADMIN_PORT="${ORDERER_CONFIG##*:}"
    
    log_info "Joining ${ORDERER_NAME} to channel ${CHANNEL_NAME}..."
    
    RESULT=$(docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        hyperledger/fabric-tools:2.5.9 \
        osnadmin channel join \
            --channelID "${CHANNEL_NAME}" \
            --config-block /fabric/channel-artifacts/${CHANNEL_NAME}.block \
            -o "${ORDERER_NAME}:${ADMIN_PORT}" \
            --ca-file "/fabric/organizations/ordererOrganizations/ibn.vn/orderers/${ORDERER_NAME}/tls/ca.crt" \
            --client-cert "/fabric/organizations/ordererOrganizations/ibn.vn/orderers/${ORDERER_NAME}/tls/server.crt" \
            --client-key "/fabric/organizations/ordererOrganizations/ibn.vn/orderers/${ORDERER_NAME}/tls/server.key" 2>&1) || true
    
    if echo "${RESULT}" | grep -q '"status": "active"'; then
        log_success "${ORDERER_NAME} joined successfully!"
    elif echo "${RESULT}" | grep -q "already exists"; then
        log_info "${ORDERER_NAME} already joined to channel"
    else
        log_error "Failed to join ${ORDERER_NAME}: ${RESULT}"
    fi
done

echo ""
log_info "Verifying channel membership..."

for ORDERER_CONFIG in "${ORDERERS[@]}"; do
    ORDERER_NAME="${ORDERER_CONFIG%%:*}"
    ADMIN_PORT="${ORDERER_CONFIG##*:}"
    
    docker run --rm \
        --network "${NETWORK_NAME}" \
        -v "${PROJECT_ROOT}/core:/fabric" \
        hyperledger/fabric-tools:2.5.9 \
        osnadmin channel list \
            -o "${ORDERER_NAME}:${ADMIN_PORT}" \
            --ca-file "/fabric/organizations/ordererOrganizations/ibn.vn/orderers/${ORDERER_NAME}/tls/ca.crt" \
            --client-cert "/fabric/organizations/ordererOrganizations/ibn.vn/orderers/${ORDERER_NAME}/tls/server.crt" \
            --client-key "/fabric/organizations/ordererOrganizations/ibn.vn/orderers/${ORDERER_NAME}/tls/server.key" 2>&1 | grep -A2 "${CHANNEL_NAME}" || true
done

echo ""
echo -e "${GREEN}All orderers have been processed.${NC}"
