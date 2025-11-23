#!/bin/bash

# Helper script for chaincode operations
# Usage: ./scripts/chaincode-helper.sh [query|invoke] [args...]

set -e

CHANNEL="ibnchannel"
CHAINCODE="teaTraceCC"
PEER="peer0.org1.ibn.vn"
ORDERER="orderer2.ibn.vn:9050"
ADMIN_MSP="/tmp/admin-msp"
TLS_CA="/tmp/orderer-tls-ca.pem"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

function query_batch() {
    local batch_id=$1
    echo -e "${YELLOW}Querying batch: ${batch_id}${NC}"
    docker exec -e CORE_PEER_MSPCONFIGPATH=${ADMIN_MSP} -e CORE_PEER_LOCALMSPID=Org1MSP \
        ${PEER} peer chaincode query -C ${CHANNEL} -n ${CHAINCODE} \
        -c "{\"Args\":[\"getBatchInfo\",\"${batch_id}\"]}" 2>&1 | jq '.' || cat
}

function create_batch() {
    local batch_id=$1
    local farm_location=$2
    local harvest_date=$3
    local processing_info=$4
    local quality_cert=$5
    
    echo -e "${YELLOW}Creating batch: ${batch_id}${NC}"
    docker exec -e CORE_PEER_MSPCONFIGPATH=${ADMIN_MSP} -e CORE_PEER_LOCALMSPID=Org1MSP \
        ${PEER} peer chaincode invoke -C ${CHANNEL} -n ${CHAINCODE} \
        --peerAddresses ${PEER}:7051 \
        --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
        -o ${ORDERER} --tls --cafile ${TLS_CA} \
        -c "{\"Args\":[\"createBatch\",\"${batch_id}\",\"${farm_location}\",\"${harvest_date}\",\"${processing_info}\",\"${quality_cert}\"]}" \
        --waitForEvent 2>&1 | tail -5
}

function verify_batch() {
    local batch_id=$1
    local hash_input=$2
    
    echo -e "${YELLOW}Verifying batch: ${batch_id}${NC}"
    docker exec -e CORE_PEER_MSPCONFIGPATH=${ADMIN_MSP} -e CORE_PEER_LOCALMSPID=Org1MSP \
        ${PEER} peer chaincode invoke -C ${CHANNEL} -n ${CHAINCODE} \
        --peerAddresses ${PEER}:7051 \
        --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
        -o ${ORDERER} --tls --cafile ${TLS_CA} \
        -c "{\"Args\":[\"verifyBatch\",\"${batch_id}\",\"${hash_input}\"]}" \
        --waitForEvent 2>&1 | tail -5
}

function update_status() {
    local batch_id=$1
    local status=$2
    
    echo -e "${YELLOW}Updating batch status: ${batch_id} -> ${status}${NC}"
    docker exec -e CORE_PEER_MSPCONFIGPATH=${ADMIN_MSP} -e CORE_PEER_LOCALMSPID=Org1MSP \
        ${PEER} peer chaincode invoke -C ${CHANNEL} -n ${CHAINCODE} \
        --peerAddresses ${PEER}:7051 \
        --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
        -o ${ORDERER} --tls --cafile ${TLS_CA} \
        -c "{\"Args\":[\"updateBatchStatus\",\"${batch_id}\",\"${status}\"]}" \
        --waitForEvent 2>&1 | tail -5
}

# Main
case "$1" in
    query)
        if [ -z "$2" ]; then
            echo "Usage: $0 query <batch_id>"
            exit 1
        fi
        query_batch "$2"
        ;;
    create)
        if [ -z "$6" ]; then
            echo "Usage: $0 create <batch_id> <farm_location> <harvest_date> <processing_info> <quality_cert>"
            exit 1
        fi
        create_batch "$2" "$3" "$4" "$5" "$6"
        ;;
    verify)
        if [ -z "$3" ]; then
            echo "Usage: $0 verify <batch_id> <hash_input>"
            exit 1
        fi
        verify_batch "$2" "$3"
        ;;
    status)
        if [ -z "$3" ]; then
            echo "Usage: $0 status <batch_id> <status>"
            echo "Status: CREATED, VERIFIED, EXPIRED"
            exit 1
        fi
        update_status "$2" "$3"
        ;;
    *)
        echo "Chaincode Helper Script"
        echo ""
        echo "Usage: $0 <command> [args...]"
        echo ""
        echo "Commands:"
        echo "  query <batch_id>                    - Query batch information"
        echo "  create <id> <location> <date> <info> <cert>  - Create new batch"
        echo "  verify <batch_id> <hash_input>      - Verify batch hash"
        echo "  status <batch_id> <status>          - Update batch status"
        echo ""
        echo "Examples:"
        echo "  $0 query health-check"
        echo "  $0 create TEST002 \"Farm A\" \"2024-11-12\" \"Organic\" \"CERT-001\""
        echo "  $0 verify TEST001 \"hash_input_string\""
        echo "  $0 status TEST001 VERIFIED"
        exit 1
        ;;
esac
