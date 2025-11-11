# Trạng Thái Deploy Chaincode teaTraceCC

**Ngày kiểm tra**: 2025-11-11  
**Chaincode**: teaTraceCC v1.0.0

## ❌ KẾT QUẢ KIỂM TRA

### Trạng thái hiện tại: **CHƯA ĐƯỢC DEPLOY**

## 📋 Chi tiết kiểm tra

### 1. Channels
- ❌ **Không có channel nào được join**
- Thư mục chains trống: `/var/hyperledger/production/ledgersData/chains/chains/`
- Cần tạo và join channel trước khi deploy chaincode

### 2. Chaincode Installation
- ❌ **Không có chaincode nào được install**
- Query `peer lifecycle chaincode queryinstalled` không trả về kết quả
- Chaincode chưa được package và install

### 3. Chaincode Containers
- ❌ **Không có chaincode container nào đang chạy**
- Không tìm thấy container `dev-peer*.org1.ibn.vn-teaTraceCC-*`

### 4. Chaincode Committed
- ❌ **Không có chaincode nào được commit lên channel**
- Query `peer lifecycle chaincode querycommitted` không tìm thấy chaincode

## 📝 Các bước cần thực hiện để deploy

### Bước 1: Tạo Channel (nếu chưa có)
```bash
cd ~/ibn/core
export FABRIC_CFG_PATH=./configtx
export PATH=./bin:$PATH

# Tạo channel transaction (nếu chưa có)
./bin/configtxgen -profile ThreePeersChannel \
  -channelID ibnchannel \
  -outputCreateChannelTx ./channel-artifacts/ibnchannel.tx \
  -configPath ./configtx

# Tạo genesis block cho channel
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=./organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp
export CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051
export ORDERER_CA=./organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem

./bin/peer channel create \
  -o orderer.ibn.vn:7050 \
  -c ibnchannel \
  -f ./channel-artifacts/ibnchannel.tx \
  --tls \
  --cafile $ORDERER_CA
```

### Bước 2: Join Peers vào Channel
```bash
# Join peer0
export CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051
./bin/peer channel join -b ./channel-artifacts/ibnchannel.block

# Join peer1
export CORE_PEER_ADDRESS=peer1.org1.ibn.vn:8051
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt
./bin/peer channel join -b ./channel-artifacts/ibnchannel.block

# Join peer2
export CORE_PEER_ADDRESS=peer2.org1.ibn.vn:9051
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt
./bin/peer channel join -b ./channel-artifacts/ibnchannel.block
```

### Bước 3: Package Chaincode
```bash
cd ~/ibn/teaTraceCC

# Build chaincode
npm install
npm run build

# Copy msp-config.json vào dist
cp msp-config.json dist/

# Package chaincode
cd ~/ibn/core
export PATH=./bin:$PATH
export FABRIC_CFG_PATH=./config

./bin/peer lifecycle chaincode package teaTraceCC.tar.gz \
  --path ../teaTraceCC/dist \
  --lang node \
  --label teaTraceCC_1.0
```

### Bước 4: Install Chaincode trên tất cả Peers
```bash
# Install trên peer0
export CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
./bin/peer lifecycle chaincode install teaTraceCC.tar.gz

# Install trên peer1
export CORE_PEER_ADDRESS=peer1.org1.ibn.vn:8051
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt
./bin/peer lifecycle chaincode install teaTraceCC.tar.gz

# Install trên peer2
export CORE_PEER_ADDRESS=peer2.org1.ibn.vn:9051
export CORE_PEER_TLS_ROOTCERT_FILE=./organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt
./bin/peer lifecycle chaincode install teaTraceCC.tar.gz

# Lưu lại PACKAGE_ID từ output
```

### Bước 5: Approve Chaincode
```bash
# Approve trên peer0 (với PACKAGE_ID từ bước trên)
export CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051
export ORDERER_CA=./organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem

./bin/peer lifecycle chaincode approveformyorg \
  -o orderer.ibn.vn:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --channelID ibnchannel \
  --name teaTraceCC \
  --version 1.0 \
  --package-id <PACKAGE_ID> \
  --sequence 1 \
  --tls \
  --cafile $ORDERER_CA
```

### Bước 6: Commit Chaincode
```bash
./bin/peer lifecycle chaincode commit \
  -o orderer.ibn.vn:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --channelID ibnchannel \
  --name teaTraceCC \
  --version 1.0 \
  --sequence 1 \
  --tls \
  --cafile $ORDERER_CA \
  --peerAddresses peer0.org1.ibn.vn:7051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  --peerAddresses peer1.org1.ibn.vn:8051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt \
  --peerAddresses peer2.org1.ibn.vn:9051 \
  --tlsRootCertFiles ./organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt
```

### Bước 7: Verify Deployment
```bash
# Query committed chaincodes
./bin/peer lifecycle chaincode querycommitted --channelID ibnchannel

# Test chaincode
./bin/peer chaincode query \
  -C ibnchannel \
  -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

## ⚠️ Lưu ý

1. **Báo cáo cũ**: Báo cáo `BAO_CAO_TANG_CORE.md` có thể đề cập đến deployment cũ hoặc từ network khác
2. **Network mới**: Network hiện tại đã được reset, cần deploy lại từ đầu
3. **Channel name**: Cần xác định channel name (ibnchannel hoặc teachannel) trước khi deploy

## ✅ Checklist Deploy

- [ ] Channel đã được tạo
- [ ] Tất cả peers đã join channel
- [ ] Chaincode đã được package
- [ ] Chaincode đã được install trên tất cả peers
- [ ] Chaincode đã được approve
- [ ] Chaincode đã được commit
- [ ] Chaincode container đang chạy
- [ ] Test chaincode thành công

---

**Trạng thái**: ❌ **CHƯA DEPLOY**  
**Cần thực hiện**: Tất cả các bước từ 1-7

