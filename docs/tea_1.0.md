# Hướng Dẫn Sử Dụng Chaincode teaTraceCC

> Hướng dẫn chi tiết về cách query, invoke và quản lý dữ liệu trên chaincode teaTraceCC

## 📋 Mục Lục

- [Thông Tin Chaincode](#thông-tin-chaincode)
- [Cấu Hình Môi Trường](#cấu-hình-môi-trường)
- [Query Dữ Liệu](#query-dữ-liệu)
- [Invoke Transactions](#invoke-transactions)
- [Ví Dụ Sử Dụng](#ví-dụ-sử-dụng)
- [Troubleshooting](#troubleshooting)

---

## Thông Tin Chaincode

| Thông tin | Giá trị |
|-----------|---------|
| **Name** | teaTraceCC |
| **Version** | 1.0 |
| **Sequence** | 2 |
| **Channel** | ibnchannel |
| **Language** | Node.js (TypeScript) |
| **Package ID** | teaTraceCC_1.0:98cfde5435a0f97398b9a8e1fecc4c1374106133bcefba1f5122a20de6efae60 |

### Cấu Trúc Dữ Liệu TeaBatch

```json
{
  "batchId": "string",
  "farmLocation": "string",
  "harvestDate": "string",
  "processingInfo": "string",
  "qualityCert": "string",
  "hashValue": "string",
  "owner": "string",
  "timestamp": "string",
  "status": "CREATED|VERIFIED|EXPIRED"
}
```

### Các Functions Có Sẵn

1. **createBatch** - Tạo lô trà mới (Quyền: Farmer/Org1MSP)
2. **verifyBatch** - Xác minh hash lô trà (Quyền: Farmer, Verifier, Admin)
3. **getBatchInfo** - Lấy thông tin lô trà (Quyền: Public)
4. **updateBatchStatus** - Cập nhật trạng thái (Quyền: Farmer, Admin)

---

## Cấu Hình Môi Trường

### Thiết Lập Biến Môi Trường

```bash
# Di chuyển đến thư mục core
cd ~/ibn/core

# Thiết lập PATH
export PATH=./bin:$PATH
unset FABRIC_CFG_PATH

# Cấu hình TLS và MSP
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_TLS_SERVERHOSTOVERRIDE=peer0.org1.ibn.vn

# Orderer CA
export ORDERER_CA=./organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem
```

### Cấu Hình Cho Các Peer Khác

**Peer1 (port 8051):**
```bash
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:8051
export CORE_PEER_TLS_SERVERHOSTOVERRIDE=peer1.org1.ibn.vn
```

**Peer2 (port 9051):**
```bash
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:9051
export CORE_PEER_TLS_SERVERHOSTOVERRIDE=peer2.org1.ibn.vn
```

---

## Query Dữ Liệu

### 1. Query Thông Tin Lô Trà (getBatchInfo)

Lấy thông tin chi tiết của một lô trà theo batchId.

```bash
./bin/peer chaincode query \
  -C ibnchannel \
  -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

**Kết quả mẫu:**
```json
{
  "batchId": "BATCH001",
  "farmLocation": "Moc Chau, Son La",
  "harvestDate": "2024-11-08",
  "processingInfo": "Organic processing, no pesticides",
  "qualityCert": "VN-ORG-2024",
  "hashValue": "a1b2c3d4e5f6...",
  "owner": "Org1MSP",
  "timestamp": "2024-11-08T10:00:00.000Z",
  "status": "CREATED"
}
```

### 2. Kiểm Tra Chaincode Đã Commit

```bash
./bin/peer lifecycle chaincode querycommitted \
  --channelID ibnchannel \
  --name teaTraceCC
```

**Kết quả:**
```
Committed chaincode definition for chaincode 'teaTraceCC' on channel 'ibnchannel':
Version: 1.0, Sequence: 2, Endorsement Plugin: escc, Validation Plugin: vscc, Approvals: [Org1MSP: true]
```

### 3. Kiểm Tra Chaincode Đã Install

```bash
./bin/peer lifecycle chaincode queryinstalled
```

### 4. Kiểm Tra Thông Tin Channel

```bash
./bin/peer channel getinfo -c ibnchannel
```

**Kết quả:**
```
Blockchain info: {
  "height": 6,
  "currentBlockHash": "...",
  "previousBlockHash": "..."
}
```

---

## Invoke Transactions

### 1. Tạo Lô Trà Mới (createBatch)

**Quyền:** Farmer (Org1MSP)

```bash
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  --peerAddresses localhost:8051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt \
  --peerAddresses localhost:9051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH001","Moc Chau, Son La","2024-11-08","Organic processing, no pesticides","VN-ORG-2024"]}'
```

**Tham số:**
- `batchId`: ID duy nhất của lô trà (ví dụ: "BATCH001")
- `farmLocation`: Vị trí nông trại (ví dụ: "Moc Chau, Son La")
- `harvestDate`: Ngày thu hoạch (format: YYYY-MM-DD)
- `processingInfo`: Thông tin xử lý (ví dụ: "Organic processing, no pesticides")
- `qualityCert`: Chứng chỉ chất lượng (ví dụ: "VN-ORG-2024")

**Kết quả:**
```
[chaincodeCmd] ClientWait -> txid [abc123...] committed with status (VALID)
```

### 2. Xác Minh Lô Trà (verifyBatch)

**Quyền:** Farmer, Verifier, Admin

```bash
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["verifyBatch","BATCH001","hashInputString"]}'
```

**Tham số:**
- `batchId`: ID của lô trà cần xác minh
- `hashInput`: Chuỗi input để verify hash

**Kết quả:**
```json
{
  "isValid": true,
  "batch": {
    "batchId": "BATCH001",
    "status": "VERIFIED",
    ...
  }
}
```

### 3. Cập Nhật Trạng Thái (updateBatchStatus)

**Quyền:** Farmer, Admin

**Trạng thái hợp lệ:** `CREATED`, `VERIFIED`, `EXPIRED`

```bash
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["updateBatchStatus","BATCH001","VERIFIED"]}'
```

**Tham số:**
- `batchId`: ID của lô trà
- `status`: Trạng thái mới (CREATED, VERIFIED, hoặc EXPIRED)

**Kết quả:**
```json
{
  "batchId": "BATCH001",
  "status": "VERIFIED",
  "timestamp": "2024-11-08T12:00:00.000Z",
  ...
}
```

---

## Ví Dụ Sử Dụng

### Workflow Hoàn Chỉnh

#### Bước 1: Tạo Lô Trà Mới

```bash
# Thiết lập môi trường
cd ~/ibn/core
export PATH=./bin:$PATH
unset FABRIC_CFG_PATH
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_TLS_SERVERHOSTOVERRIDE=peer0.org1.ibn.vn
export ORDERER_CA=./organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem

# Tạo lô trà
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH001","Moc Chau, Son La","2024-11-08","Organic processing","VN-ORG-2024"]}'
```

#### Bước 2: Query Thông Tin Lô Trà

```bash
./bin/peer chaincode query \
  -C ibnchannel \
  -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

#### Bước 3: Xác Minh Lô Trà

```bash
# Lấy hashInput từ batch đã tạo (sử dụng thông tin từ getBatchInfo)
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["verifyBatch","BATCH001","BATCH001Moc Chau, Son La2024-11-08Organic processingVN-ORG-2024"]}'
```

#### Bước 4: Cập Nhật Trạng Thái

```bash
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["updateBatchStatus","BATCH001","EXPIRED"]}'
```

### Ví Dụ Tạo Nhiều Lô Trà

```bash
# Lô trà 1
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH001","Moc Chau","2024-11-08","Organic","VN-ORG-2024"]}'

# Lô trà 2
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH002","Da Lat","2024-11-09","Premium","VN-PREMIUM-2024"]}'

# Lô trà 3
./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH003","Bao Loc","2024-11-10","Standard","VN-STD-2024"]}'
```

### Query Nhiều Lô Trà

```bash
# Query lô trà 1
./bin/peer chaincode query -C ibnchannel -n teaTraceCC -c '{"Args":["getBatchInfo","BATCH001"]}'

# Query lô trà 2
./bin/peer chaincode query -C ibnchannel -n teaTraceCC -c '{"Args":["getBatchInfo","BATCH002"]}'

# Query lô trà 3
./bin/peer chaincode query -C ibnchannel -n teaTraceCC -c '{"Args":["getBatchInfo","BATCH003"]}'
```

---

## Script Helper

### Script Query Đơn Giản

Tạo file `query-batch.sh`:

```bash
#!/bin/bash

BATCH_ID=$1

if [ -z "$BATCH_ID" ]; then
  echo "Usage: ./query-batch.sh <BATCH_ID>"
  exit 1
fi

cd ~/ibn/core
export PATH=./bin:$PATH
unset FABRIC_CFG_PATH
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_TLS_SERVERHOSTOVERRIDE=peer0.org1.ibn.vn

./bin/peer chaincode query \
  -C ibnchannel \
  -n teaTraceCC \
  -c "{\"Args\":[\"getBatchInfo\",\"$BATCH_ID\"]}"
```

**Sử dụng:**
```bash
chmod +x query-batch.sh
./query-batch.sh BATCH001
```

### Script Invoke Đơn Giản

Tạo file `create-batch.sh`:

```bash
#!/bin/bash

BATCH_ID=$1
FARM_LOCATION=$2
HARVEST_DATE=$3
PROCESSING_INFO=$4
QUALITY_CERT=$5

if [ -z "$BATCH_ID" ] || [ -z "$FARM_LOCATION" ] || [ -z "$HARVEST_DATE" ] || [ -z "$PROCESSING_INFO" ] || [ -z "$QUALITY_CERT" ]; then
  echo "Usage: ./create-batch.sh <BATCH_ID> <FARM_LOCATION> <HARVEST_DATE> <PROCESSING_INFO> <QUALITY_CERT>"
  exit 1
fi

cd ~/ibn/core
export PATH=./bin:$PATH
unset FABRIC_CFG_PATH
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
export CORE_PEER_ADDRESS=localhost:7051
export CORE_PEER_TLS_SERVERHOSTOVERRIDE=peer0.org1.ibn.vn
export ORDERER_CA=./organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem

./bin/peer chaincode invoke \
  -o localhost:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --tls \
  --cafile $ORDERER_CA \
  -C ibnchannel \
  -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -c "{\"Args\":[\"createBatch\",\"$BATCH_ID\",\"$FARM_LOCATION\",\"$HARVEST_DATE\",\"$PROCESSING_INFO\",\"$QUALITY_CERT\"]}"
```

**Sử dụng:**
```bash
chmod +x create-batch.sh
./create-batch.sh "BATCH001" "Moc Chau, Son La" "2024-11-08" "Organic processing" "VN-ORG-2024"
```

---

## Troubleshooting

### Lỗi: "endorsement failure"

**Nguyên nhân:** Chaincode container chưa khởi động hoặc gặp lỗi DNS.

**Giải pháp:**
1. Kiểm tra chaincode containers:
   ```bash
   docker ps -a | grep dev-peer
   ```

2. Xem logs của chaincode container:
   ```bash
   docker logs <container_name>
   ```

3. Kiểm tra network:
   ```bash
   docker network ls | grep fabric
   ```

### Lỗi: "MSP không có quyền"

**Nguyên nhân:** MSP hiện tại không có quyền thực hiện function.

**Giải pháp:**
- Kiểm tra MSP ID: `echo $CORE_PEER_LOCALMSPID`
- Xem phân quyền trong `teaTraceCC/msp-config.json`
- Đảm bảo đang sử dụng đúng MSP có quyền

### Lỗi: "Batch with id 'XXX' does not exist"

**Nguyên nhân:** Batch ID không tồn tại trong ledger.

**Giải pháp:**
- Kiểm tra lại batch ID
- Tạo batch mới trước khi query
- Sử dụng `createBatch` để tạo batch

### Lỗi: "Invalid status"

**Nguyên nhân:** Trạng thái không hợp lệ.

**Giải pháp:**
- Chỉ sử dụng: `CREATED`, `VERIFIED`, `EXPIRED`
- Kiểm tra chính tả và chữ hoa/thường

### Lỗi: "container exited with 0"

**Nguyên nhân:** Chaincode container không thể kết nối với peer.

**Giải pháp:**
1. Xóa và để peer tự tạo lại container:
   ```bash
   docker rm -f $(docker ps -a | grep dev-peer | awk '{print $1}')
   ```

2. Thử lại query/invoke để peer tự động khởi động container mới

### Kiểm Tra Trạng Thái Network

```bash
# Kiểm tra containers đang chạy
docker ps | grep -E "peer|orderer"

# Kiểm tra logs của peer
docker logs peer0.org1.ibn.vn --tail 50

# Kiểm tra channel info
cd ~/ibn/core
export PATH=./bin:$PATH
./bin/peer channel getinfo -c ibnchannel
```

---

## Phân Quyền MSP

| MSP Role | MSP ID | Quyền Hạn |
|----------|--------|-----------|
| **Farmer** | Org1MSP | createBatch, updateBatchStatus, verifyBatch |
| **Verifier** | Org2MSP | verifyBatch, getBatchInfo |
| **Admin** | Org3MSP | updateBatchStatus, verifyBatch, getBatchInfo |

**Lưu ý:** Hiện tại network chỉ có Org1MSP, nên chỉ có thể thực hiện các function của Farmer.

---

## Workflow Trạng Thái

```
CREATED → VERIFIED → EXPIRED
```

- **CREATED**: Lô trà mới được tạo
- **VERIFIED**: Đã xác minh hash, đảm bảo chính hãng
- **EXPIRED**: Hết hạn sử dụng

---

## Tài Liệu Tham Khảo

- [Hyperledger Fabric Chaincode Documentation](https://hyperledger-fabric.readthedocs.io/en/latest/chaincode4ade.html)
- [Fabric Contract API](https://hyperledger.github.io/fabric-chaincode-node/release-2.5/api/)
- [Peer CLI Commands](https://hyperledger-fabric.readthedocs.io/en/latest/commands/peerchaincode.html)

---

**Cập nhật lần cuối:** 2025-11-11  
**Chaincode Version:** 1.0  
**Network:** IBN (ibn.vn)

