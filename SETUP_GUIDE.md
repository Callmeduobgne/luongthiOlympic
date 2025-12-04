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

## 📜 9. Deploy Chaincode (Tùy chọn)

Nếu muốn deploy chaincode (ví dụ `basic`):

```bash
# Package chaincode
docker exec peer0.org1.ibn.vn peer lifecycle chaincode package basic.tar.gz --path /opt/gopath/src/github.com/chaincode/basic --lang golang --label basic_1.0

# Install trên Peer0
docker exec peer0.org1.ibn.vn peer lifecycle chaincode install basic.tar.gz

# (Lặp lại install cho peer1, peer2 nếu cần)

# Approve & Commit (Cần thêm các bước approve và commit tùy theo policy)
```
