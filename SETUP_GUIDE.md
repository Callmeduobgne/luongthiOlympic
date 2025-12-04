# 🚀 IBN Network - Hướng Dẫn Cài Đặt Thủ Công (Manual Setup)

Tài liệu này hướng dẫn chi tiết cách xây dựng lại mạng lưới Hyperledger Fabric (IBN Network) **bằng các lệnh thủ công**, giúp bạn kiểm soát hoàn toàn quá trình.

## 📋 1. Chuẩn Bị Môi Trường

Đảm bảo bạn đang ở thư mục gốc của dự án:
```bash
cd /mnt/e/luongbeo
```

Kiểm tra các công cụ cần thiết:
```bash
docker --version
docker compose version
```

---

## 🧹 2. Dọn Dẹp Môi Trường Cũ (Cleanup)

Trước khi bắt đầu, hãy xóa sạch containers và dữ liệu cũ để tránh xung đột.

```bash
# 1. Dừng và xóa containers
docker compose down --volumes --remove-orphans

# 2. Xóa crypto material cũ
rm -rf core/organizations
rm -rf core/channel-artifacts
rm -rf core/system-genesis-block
```

---

## 🔑 3. Tạo Crypto Material (Certificates)

Bước này tạo ra các chứng chỉ số (MSP) cho Orderer và Peers.

**Quan trọng:** Kiểm tra file `core/crypto-config.yaml` phải có `EnableNodeOUs: true`.

Chạy lệnh sau để tạo crypto bằng Docker (sử dụng image `fabric-tools`):

```bash
docker run --rm -v "$(pwd)/core":/core -w /core \
  hyperledger/fabric-tools:2.5 \
  cryptogen generate --config=crypto-config.yaml --output=organizations
```

**Sau khi tạo xong, cần tạo file cấu hình NodeOUs cho MSP:**

```bash
# Tạo config.yaml cho Peer Org
cat > core/organizations/peerOrganizations/org1.ibn.vn/msp/config.yaml <<EOF
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: orderer
EOF

# Tạo config.yaml cho Orderer Org
cat > core/organizations/ordererOrganizations/ibn.vn/msp/config.yaml <<EOF
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/ca.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: orderer
EOF
```

---

## 🧱 4. Tạo Genesis Block

Tạo block khởi nguyên cho hệ thống (System Genesis Block).

```bash
# Tạo thư mục chứa artifacts và genesis block
mkdir -p core/channel-artifacts
mkdir -p core/system-genesis-block

# Generate Genesis Block
docker run --rm -v "$(pwd)/core":/core -w /core \
  -e FABRIC_CFG_PATH=/core/configtx \
  hyperledger/fabric-tools:2.5 \
  configtxgen -profile RaftOrdererGenesis -channelID system-channel -outputBlock ./system-genesis-block/genesis.block
```

---

## 🐳 5. Khởi Động Network

Bây giờ chúng ta sẽ khởi động các containers (Peers, Orderer, CA...).

```bash
# Khởi động network (chạy ngầm)
docker compose up -d

# Kiểm tra trạng thái
docker compose ps
```
*Đợi khoảng 10-20 giây để các services khởi động hoàn toàn.*

---

## 🌐 6. Tạo Channel

Chúng ta sẽ tạo channel tên là `ibnchannel`.

### 6.1. Tạo Channel Genesis Block
Thay vì tạo transaction file, Fabric 2.5 khuyến nghị tạo genesis block cho application channel.

```bash
docker run --rm -v "$(pwd)/core":/core -w /core \
  -e FABRIC_CFG_PATH=/core/configtx \
  hyperledger/fabric-tools:2.5 \
  configtxgen -profile ThreePeersChannel -outputBlock ./channel-artifacts/ibnchannel.block -channelID ibnchannel
```

### 6.2. Join Orderer vào Channel
Copy block vào container Orderer và join.

```bash
# Copy block vào orderer
docker cp core/channel-artifacts/ibnchannel.block orderer.ibn.vn:/var/hyperledger/orderer/ibnchannel.block

# Join Orderer (thực hiện qua osnadmin - nếu có cấu hình, hoặc đơn giản là orderer tự join qua system channel cũ - nhưng ở đây ta dùng cách thủ công join peer trước)
# Lưu ý: Với cấu hình hiện tại, Orderer thường đã sẵn sàng. Ta tập trung join Peer.
```

---

## 🔗 7. Join Peers vào Channel

Thực hiện join lần lượt cho từng Peer (`peer0`, `peer1`, `peer2`).

### Peer 0
```bash
# 1. Copy block và Admin MSP vào container
docker cp core/channel-artifacts/ibnchannel.block peer0.org1.ibn.vn:/root/ibnchannel.block
docker cp core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp peer0.org1.ibn.vn:/tmp/admin_msp

# 2. Join Channel (Sử dụng Admin MSP)
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer channel join -b /root/ibnchannel.block
```

### Peer 1
```bash
# 1. Copy block và Admin MSP
docker cp core/channel-artifacts/ibnchannel.block peer1.org1.ibn.vn:/root/ibnchannel.block
docker cp core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp peer1.org1.ibn.vn:/tmp/admin_msp

# 2. Join Channel
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
  -e CORE_PEER_ADDRESS="peer1.org1.ibn.vn:8051" \
  peer1.org1.ibn.vn \
  peer channel join -b /root/ibnchannel.block
```

### Peer 2
```bash
# 1. Copy block và Admin MSP
docker cp core/channel-artifacts/ibnchannel.block peer2.org1.ibn.vn:/root/ibnchannel.block
docker cp core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp peer2.org1.ibn.vn:/tmp/admin_msp

# 2. Join Channel
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
  -e CORE_PEER_ADDRESS="peer2.org1.ibn.vn:9051" \
  peer2.org1.ibn.vn \
  peer channel join -b /root/ibnchannel.block
```

---

## ✅ 8. Kiểm Tra Kết Quả

Kiểm tra xem Peer 0 đã join thành công chưa:

```bash
docker exec peer0.org1.ibn.vn peer channel list
```

Nếu thấy output:
```
Channels peers has joined: 
ibnchannel
```
🎉 **Chúc mừng! Bạn đã setup thành công mạng lưới thủ công.**

---

## 📜 9. Deploy Chaincode teaTraceCC

Hướng dẫn deploy chaincode **teaTraceCC** (Node.js/TypeScript) lên network IBN.

### 9.1. Build Chaincode

Trước tiên, cần build chaincode từ source code:

```bash
# Di chuyển vào thư mục chaincode
cd teaTraceCC

# Cài đặt dependencies
npm install

# Build TypeScript sang JavaScript
npm run build

# Copy file cấu hình MSP vào thư mục dist
cp msp-config.json dist/
```

### 9.2. Package Chaincode

Tạo package file cho chaincode:

```bash
# Quay về thư mục gốc
cd ..

# Copy thư mục dist vào container peer0 để package
docker cp teaTraceCC/dist peer0.org1.ibn.vn:/opt/chaincode/teaTraceCC

# Package chaincode
docker exec peer0.org1.ibn.vn peer lifecycle chaincode package teaTraceCC.tar.gz \
  --path /opt/chaincode/teaTraceCC \
  --lang node \
  --label teaTraceCC_1.1.0

# Copy package file ra ngoài để dùng cho các peer khác
docker cp peer0.org1.ibn.vn:/opt/chaincode/teaTraceCC.tar.gz ./teaTraceCC.tar.gz
```

### 9.3. Install Chaincode trên các Peers

Cài đặt chaincode trên tất cả các peers:

```bash
# Install trên Peer0
docker exec peer0.org1.ibn.vn peer lifecycle chaincode install teaTraceCC.tar.gz

# Install trên Peer1
docker cp teaTraceCC.tar.gz peer1.org1.ibn.vn:/opt/chaincode/
docker exec peer1.org1.ibn.vn peer lifecycle chaincode install /opt/chaincode/teaTraceCC.tar.gz

# Install trên Peer2
docker cp teaTraceCC.tar.gz peer2.org1.ibn.vn:/opt/chaincode/
docker exec peer2.org1.ibn.vn peer lifecycle chaincode install /opt/chaincode/teaTraceCC.tar.gz
```

**Lưu ý:** Sau mỗi lệnh install, lưu lại **PACKAGE_ID** từ output. Ví dụ:
```
2024-12-04 10:00:00.000 UTC 0001 INFO [cli.lifecycle.chaincode] submitInstallProposal -> Installed remotely: response:<status:200 payload:"\nEteaTraceCC_1.1.0:abc123def456..." >
```

### 9.4. Approve Chaincode Definition

Approve chaincode definition cho Org1:

```bash
# Thay <PACKAGE_ID> bằng PACKAGE_ID thực tế từ bước trên
PACKAGE_ID="teaTraceCC_1.1.0:abc123def456..."  # Thay bằng PACKAGE_ID thực tế

# Approve trên Peer0 (với Admin MSP)
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp/users/Admin@org1.ibn.vn/msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer lifecycle chaincode approveformyorg \
  -o orderer.ibn.vn:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --channelID ibnchannel \
  --name teaTraceCC \
  --version 1.1.0 \
  --package-id ${PACKAGE_ID} \
  --sequence 1 \
  --tls \
  --cafile /etc/hyperledger/fabric/orderer/tls/ca.crt
```

### 9.5. Commit Chaincode Definition

Commit chaincode definition lên channel:

```bash
# Commit chaincode (cần ít nhất 1 peer từ mỗi org)
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp/users/Admin@org1.ibn.vn/msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer lifecycle chaincode commit \
  -o orderer.ibn.vn:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  --channelID ibnchannel \
  --name teaTraceCC \
  --version 1.1.0 \
  --sequence 1 \
  --tls \
  --cafile /etc/hyperledger/fabric/orderer/tls/ca.crt \
  --peerAddresses peer0.org1.ibn.vn:7051 \
  --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
  --peerAddresses peer1.org1.ibn.vn:8051 \
  --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
  --peerAddresses peer2.org1.ibn.vn:9051 \
  --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt
```

### 9.6. Kiểm Tra Chaincode đã Deploy

Kiểm tra chaincode đã được commit thành công:

```bash
# Query committed chaincode
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp/users/Admin@org1.ibn.vn/msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer lifecycle chaincode querycommitted --channelID ibnchannel
```

Nếu thấy output có `teaTraceCC` với version `1.1.0`, chaincode đã được deploy thành công! 🎉

### 9.7. Ví dụ Sử dụng Chaincode

#### Tạo lô trà mới
```bash
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp/users/Admin@org1.ibn.vn/msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer chaincode invoke \
  -o orderer.ibn.vn:7050 \
  --ordererTLSHostnameOverride orderer.ibn.vn \
  -C ibnchannel \
  -n teaTraceCC \
  --tls \
  --cafile /etc/hyperledger/fabric/orderer/tls/ca.crt \
  --peerAddresses peer0.org1.ibn.vn:7051 \
  --tlsRootCertFiles /etc/hyperledger/fabric/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH001","Moc Chau, Son La","2024-12-04","Organic processing","VN-ORG-2024"]}'
```

#### Query thông tin lô trà
```bash
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp/users/Admin@org1.ibn.vn/msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer chaincode query \
  -C ibnchannel \
  -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

#### Query tất cả batches
```bash
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp/users/Admin@org1.ibn.vn/msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  peer0.org1.ibn.vn \
  peer chaincode query \
  -C ibnchannel \
  -n teaTraceCC \
  -c '{"Args":["getAllBatches","50","0"]}'
```

---

**Lưu ý:**
- Chaincode teaTraceCC sử dụng **Node.js**, không phải Golang
- Cần build TypeScript trước khi package
- File `msp-config.json` phải được copy vào thư mục `dist/` trước khi package
- PACKAGE_ID sẽ khác nhau mỗi lần install, cần lưu lại để dùng cho approve
- Channel name: `ibnchannel`
- MSP ID: `Org1MSP`
