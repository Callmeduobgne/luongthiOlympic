# IBN Network - Hướng Dẫn Cài Đặt Từ A-Z

Tài liệu này hướng dẫn chi tiết cách cài đặt và chạy dự án IBN Network (ICTU Blockchain Network) **từ đầu đến cuối**, bao gồm cả việc tạo crypto materials và genesis block.

## Mục Lục

1. [Yêu Cầu Hệ Thống](#1-yêu-cầu-hệ-thống)
2. [Cài Đặt Môi Trường](#2-cài-đặt-môi-trường)
3. [Clone Dự Án](#3-clone-dự-án)
4. [Tạo Crypto Materials & Genesis Block](#4-tạo-crypto-materials--genesis-block)
5. [Khởi Động Network](#5-khởi-động-network)
6. [Deploy Chaincode](#6-deploy-chaincode)
7. [Kiểm Tra Hệ Thống](#7-kiểm-tra-hệ-thống)
8. [Xử Lý Lỗi Thường Gặp](#8-xử-lý-lỗi-thường-gặp)

---

## Quick Start (TL;DR)

Nếu bạn đã quen với Hyperledger Fabric và chỉ cần các lệnh nhanh:

```bash
# Clone dự án
git clone <repository-url> ibn-docker-packaging && cd ibn-docker-packaging

# Nếu CẦN tạo crypto materials từ đầu (chỉ khi chưa có thư mục organizations/)
./scripts/generate-crypto.sh

# Khởi động network
docker network create ibn-network 2>/dev/null || true
docker compose up -d
sleep 60

# Join orderers vào channel
./scripts/join-orderers.sh

# Deploy chaincode
./scripts/deploy-chaincode-ccaas.sh

# Test chaincode
./scripts/test-chaincode.sh
```

---

## 1. Yêu Cầu Hệ Thống

### Phần cứng tối thiểu
- **CPU**: 4 cores
- **RAM**: 16GB (khuyến nghị 32GB)
- **Disk**: 50GB SSD

### Phần mềm cần thiết
| Phần mềm | Phiên bản | Kiểm tra |
|----------|-----------|----------|
| Docker | 24.0+ | `docker --version` |
| Docker Compose | 2.20+ | `docker compose version` |
| Git | 2.40+ | `git --version` |
| curl | any | `curl --version` |

### Hệ điều hành hỗ trợ
- ✅ Ubuntu 22.04/24.04
- ✅ Windows 10/11 với WSL2
- ✅ macOS 12+

---

## 2. Cài Đặt Môi Trường

### Ubuntu/WSL2

```bash
# Cập nhật hệ thống
sudo apt update && sudo apt upgrade -y

# Cài đặt Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Đăng xuất và đăng nhập lại để áp dụng group docker
# Hoặc chạy: newgrp docker

# Kiểm tra Docker
docker run hello-world
```

### Windows (WSL2)

1. Cài đặt [Docker Desktop](https://www.docker.com/products/docker-desktop/)
2. Bật WSL2 integration trong Docker Desktop Settings
3. Mở WSL terminal (Ubuntu)

### macOS

```bash
# Cài đặt Docker Desktop
brew install --cask docker

# Hoặc tải từ https://www.docker.com/products/docker-desktop/
```

---

## 3. Clone Dự Án

```bash
# Clone repository
git clone <repository-url> ibn-docker-packaging
cd ibn-docker-packaging

# Kiểm tra cấu trúc thư mục
ls -la
```

**Cấu trúc thư mục quan trọng:**
```
ibn-docker-packaging/
├── docker-compose.yml       # File compose chính
├── core/                    # Cấu hình Hyperledger Fabric
│   ├── config/              # core.yaml, orderer.yaml
│   ├── configtx/            # configtx.yaml cho genesis block
│   ├── crypto-config.yaml   # Định nghĩa orgs cho cryptogen
│   ├── organizations/       # MSP certificates (sẽ được tạo)
│   ├── channel-artifacts/   # Channel block (sẽ được tạo)
│   └── system-genesis-block/# Genesis block (sẽ được tạo)
├── teaTraceCC-go/           # Go chaincode
├── backend/                 # Backend API
├── frontend/                # Frontend React
└── scripts/                 # Utility scripts
```

---

## 4. Tạo Crypto Materials & Genesis Block

> ⚠️ **QUAN TRỌNG**: Bước này chỉ cần thực hiện **MỘT LẦN** khi setup dự án mới. Nếu repository đã có sẵn thư mục `organizations/`, `channel-artifacts/`, và `system-genesis-block/` thì **BỎ QUA** bước này.

### Sử dụng Script Tự Động (Khuyến nghị)

```bash
# Cấp quyền thực thi
chmod +x scripts/generate-crypto.sh

# Chạy script
./scripts/generate-crypto.sh
```

Script sẽ tự động:
1. Tạo crypto materials (certificates cho peers, orderers, users)
2. Tạo config.yaml cho MSP (NodeOUs)
3. Tạo genesis block
4. Tạo channel artifacts (ibnchannel.block, ibnchannel.tx, Org1MSPanchors.tx)
5. Set permissions
6. Verify tất cả files

### Hoặc Thực Hiện Thủ Công:

### 4.1. Pull Fabric Tools Image

```bash
# Pull image fabric-tools (chứa cryptogen và configtxgen)
docker pull hyperledger/fabric-tools:2.5.9
```

### 4.2. Tạo Crypto Materials (Certificates)

```bash
cd ibn-docker-packaging

# Xóa crypto cũ nếu có (CHỈ KHI MUỐN TẠO LẠI TỪ ĐẦU)
# rm -rf core/organizations/*

# Tạo crypto materials bằng cryptogen
docker run --rm \
  -v $(pwd)/core:/fabric \
  -w /fabric \
  hyperledger/fabric-tools:2.5.9 \
  cryptogen generate --config=/fabric/crypto-config.yaml --output=/fabric/organizations

# Kiểm tra kết quả
ls -la core/organizations/
```

**Kết quả mong đợi:**
```
core/organizations/
├── ordererOrganizations/
│   └── ibn.vn/
│       ├── ca/
│       ├── msp/
│       ├── orderers/
│       │   ├── orderer.ibn.vn/
│       │   ├── orderer1.ibn.vn/
│       │   └── orderer2.ibn.vn/
│       ├── tlsca/
│       └── users/
└── peerOrganizations/
    └── org1.ibn.vn/
        ├── ca/
        ├── msp/
        ├── peers/
        │   ├── peer0.org1.ibn.vn/
        │   ├── peer1.org1.ibn.vn/
        │   └── peer2.org1.ibn.vn/
        ├── tlsca/
        └── users/
```

### 4.3. Tạo config.yaml cho MSP

```bash
# Tạo config.yaml cho Org1MSP (bật NodeOUs)
cat > core/organizations/peerOrganizations/org1.ibn.vn/msp/config.yaml << 'EOF'
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

# Copy config.yaml cho từng Admin user
cp core/organizations/peerOrganizations/org1.ibn.vn/msp/config.yaml \
   core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/

# Tạo config.yaml cho OrdererMSP
cat > core/organizations/ordererOrganizations/ibn.vn/msp/config.yaml << 'EOF'
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

### 4.4. Tạo Genesis Block

```bash
# Tạo thư mục output
mkdir -p core/system-genesis-block

# Tạo genesis block cho orderer
docker run --rm \
  -v $(pwd)/core:/fabric \
  -w /fabric/configtx \
  hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile RaftOrdererGenesis \
    -channelID system-channel \
    -outputBlock /fabric/system-genesis-block/genesis.block

# Kiểm tra
ls -la core/system-genesis-block/
```

### 4.5. Tạo Channel Artifacts

```bash
# Tạo thư mục channel-artifacts
mkdir -p core/channel-artifacts

# Tạo channel transaction
docker run --rm \
  -v $(pwd)/core:/fabric \
  -w /fabric/configtx \
  hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile ThreePeersChannel \
    -channelID ibnchannel \
    -outputCreateChannelTx /fabric/channel-artifacts/ibnchannel.tx

# Tạo anchor peer update
docker run --rm \
  -v $(pwd)/core:/fabric \
  -w /fabric/configtx \
  hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile ThreePeersChannel \
    -channelID ibnchannel \
    -outputAnchorPeersUpdate /fabric/channel-artifacts/Org1MSPanchors.tx \
    -asOrg Org1MSP

# Tạo channel block (cho việc join orderers)
docker run --rm \
  -v $(pwd)/core:/fabric \
  -w /fabric/configtx \
  hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile ThreePeersChannel \
    -channelID ibnchannel \
    -outputBlock /fabric/channel-artifacts/ibnchannel.block

# Kiểm tra
ls -la core/channel-artifacts/
```

**Kết quả mong đợi:**
```
core/channel-artifacts/
├── ibnchannel.block      # Channel genesis block
├── ibnchannel.tx         # Channel transaction
└── Org1MSPanchors.tx     # Anchor peer update
```

### 4.6. Cấp quyền cho các files

```bash
# Cấp quyền đọc cho tất cả certificates
chmod -R 755 core/organizations/
chmod -R 755 core/channel-artifacts/
chmod -R 755 core/system-genesis-block/
```

### 4.7. Kiểm tra cuối cùng

```bash
# Kiểm tra các file quan trọng đã tồn tại
echo "=== Checking required files ==="

# Genesis block
[ -f "core/system-genesis-block/genesis.block" ] && echo "✅ genesis.block" || echo "❌ genesis.block MISSING"

# Channel block
[ -f "core/channel-artifacts/ibnchannel.block" ] && echo "✅ ibnchannel.block" || echo "❌ ibnchannel.block MISSING"

# Orderer TLS certs
[ -f "core/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/server.crt" ] && echo "✅ orderer TLS cert" || echo "❌ orderer TLS cert MISSING"

# Peer TLS certs
[ -f "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/server.crt" ] && echo "✅ peer0 TLS cert" || echo "❌ peer0 TLS cert MISSING"

# Admin MSP
[ -d "core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp" ] && echo "✅ Admin MSP" || echo "❌ Admin MSP MISSING"

echo "=== Done ==="
```

---

## 5. Khởi Động Network

### 5.1. Tạo Docker Network

```bash
docker network create ibn-network 2>/dev/null || true
```

### 5.2. Khởi động tất cả services

```bash
# Khởi động toàn bộ stack
docker compose up -d

# Kiểm tra trạng thái
docker compose ps
```

### 5.3. Đợi services healthy

```bash
# Đợi tất cả containers healthy (khoảng 1-2 phút)
echo "Đợi services khởi động..."
sleep 60

# Kiểm tra trạng thái
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "orderer|peer|couchdb|backend|postgres"
```

**Kết quả mong đợi:**
```
orderer.ibn.vn      Up X minutes (healthy)
orderer1.ibn.vn     Up X minutes (healthy)
orderer2.ibn.vn     Up X minutes (healthy)
peer0.org1.ibn.vn   Up X minutes (healthy)
peer1.org1.ibn.vn   Up X minutes (healthy)
peer2.org1.ibn.vn   Up X minutes (healthy)
couchdb0            Up X minutes (healthy)
couchdb1            Up X minutes (healthy)
couchdb2            Up X minutes (healthy)
ibn-backend         Up X minutes (healthy)
ibn-postgres        Up X minutes (healthy)
```

### 5.4. Khởi tạo Database và Admin User

> ⚠️ **QUAN TRỌNG**: Backend cần database được khởi tạo đúng để hoạt động!

```bash
# Chạy script khởi tạo database
./scripts/init-database.sh
```

Script sẽ:
1. Tạo các bảng cần thiết trong PostgreSQL
2. Chạy migrations
3. Tạo admin user mặc định

**Admin mặc định:**
- Email: `admin@ibn.vn`
- Password: `Admin123!@#`

**Kiểm tra đăng nhập:**
```bash
curl -X POST http://localhost:9900/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@ibn.vn", "password": "Admin123!@#"}'
```

### 5.5. Join Orderers vào Channel

⚠️ **QUAN TRỌNG**: Bước này cần thực hiện mỗi khi orderer volume bị reset hoặc lần đầu khởi động.

```bash
# Sử dụng script tự động (khuyến nghị)
./scripts/join-orderers.sh

# Hoặc thực hiện thủ công:
```

**Thủ công - Join từng orderer:**

```bash
# Join orderer chính (port 9443)
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  hyperledger/fabric-tools:2.5.9 \
  osnadmin channel join \
    --channelID ibnchannel \
    --config-block /fabric/channel-artifacts/ibnchannel.block \
    -o orderer.ibn.vn:9443 \
    --ca-file /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
    --client-cert /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/server.crt \
    --client-key /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/server.key

# Join orderer1 (port 10443)
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  hyperledger/fabric-tools:2.5.9 \
  osnadmin channel join \
    --channelID ibnchannel \
    --config-block /fabric/channel-artifacts/ibnchannel.block \
    -o orderer1.ibn.vn:10443 \
    --ca-file /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer1.ibn.vn/tls/ca.crt \
    --client-cert /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer1.ibn.vn/tls/server.crt \
    --client-key /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer1.ibn.vn/tls/server.key

# Join orderer2 (port 11443)
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  hyperledger/fabric-tools:2.5.9 \
  osnadmin channel join \
    --channelID ibnchannel \
    --config-block /fabric/channel-artifacts/ibnchannel.block \
    -o orderer2.ibn.vn:11443 \
    --ca-file /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer2.ibn.vn/tls/ca.crt \
    --client-cert /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer2.ibn.vn/tls/server.crt \
    --client-key /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer2.ibn.vn/tls/server.key
```

**Kết quả mong đợi cho mỗi orderer:**
```json
Status: 201
{
    "name": "ibnchannel",
    "url": "/participation/v1/channels/ibnchannel",
    "consensusRelation": "consenter",
    "status": "active",
    "height": 1
}
```

### 5.6. Join Peers vào Channel (Chỉ lần đầu)

> Nếu peer đã join channel trước đó và dữ liệu còn trong volume, bước này sẽ báo lỗi "already joined" - có thể bỏ qua.

```bash
# Join peer0 vào channel
docker exec peer0.org1.ibn.vn peer channel join -b /var/hyperledger/fabric/channel-artifacts/ibnchannel.block

# Join peer1 vào channel  
docker exec peer1.org1.ibn.vn peer channel join -b /var/hyperledger/fabric/channel-artifacts/ibnchannel.block

# Join peer2 vào channel
docker exec peer2.org1.ibn.vn peer channel join -b /var/hyperledger/fabric/channel-artifacts/ibnchannel.block
```

### 5.7. Kiểm tra Peer đã join channel

```bash
docker exec peer0.org1.ibn.vn peer channel list
```

**Kết quả:**
```
Channels peers has joined: 
ibnchannel
```

---

## 6. Deploy Chaincode

> 💡 **Gợi ý**: Sử dụng script tự động `./scripts/deploy-chaincode-ccaas.sh` để deploy nhanh. Các bước bên dưới là hướng dẫn thủ công.

### 6.1. Build Go Chaincode Image

```bash
cd teaTraceCC-go

# Build Docker image
docker build -t teatracecc-go:1.2.0 .

# Kiểm tra image
docker images | grep teatracecc-go
```

### 6.2. Tạo CCaaS Package

```bash
# Tạo thư mục package
mkdir -p /tmp/ccaas-package
cd /tmp/ccaas-package

# Tạo connection.json
cat > connection.json << 'EOF'
{
  "address": "teatracecc-go:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF

# Tạo metadata.json
cat > metadata.json << 'EOF'
{
  "type": "ccaas",
  "label": "teaTraceCC_1.2.0"
}
EOF

# Tạo package
tar czf code.tar.gz connection.json
tar czf teaTraceCC-go.tar.gz code.tar.gz metadata.json

# Quay lại thư mục gốc
cd -
```

### 6.3. Install Chaincode

```bash
# Copy package vào admin-service
docker cp /tmp/ccaas-package/teaTraceCC-go.tar.gz admin-service:/tmp/

# Install chaincode qua API
curl -s -X POST "http://localhost:9902/api/v1/chaincode/install" \
  -H "X-API-Key: admin-service-secret-key-change-in-production-min-32-chars" \
  -H "Content-Type: application/json" \
  -d '{"packagePath": "/tmp/teaTraceCC-go.tar.gz", "targets": ["peer0.org1.ibn.vn"]}'
```

**Lưu lại PACKAGE_ID từ kết quả:**
```json
{"data":{"packageId":"teaTraceCC_1.2.0:abc123..."},"success":true}
```

### 6.4. Approve Chaincode

```bash
# Thay PACKAGE_ID bằng giá trị thực tế
PACKAGE_ID="teaTraceCC_1.2.0:YOUR_PACKAGE_ID_HERE"

docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
  hyperledger/fabric-tools:2.5.9 \
  peer lifecycle chaincode approveformyorg \
    --channelID ibnchannel \
    --name teaTraceCC \
    --version 1.2.0 \
    --package-id "$PACKAGE_ID" \
    --sequence 1 \
    --tls \
    --cafile /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
    --orderer orderer.ibn.vn:7050 \
    --ordererTLSHostnameOverride orderer.ibn.vn
```

### 6.5. Commit Chaincode

```bash
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
  hyperledger/fabric-tools:2.5.9 \
  peer lifecycle chaincode commit \
    --channelID ibnchannel \
    --name teaTraceCC \
    --version 1.2.0 \
    --sequence 1 \
    --tls \
    --cafile /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
    --orderer orderer.ibn.vn:7050 \
    --ordererTLSHostnameOverride orderer.ibn.vn \
    --peerAddresses peer0.org1.ibn.vn:7051 \
    --tlsRootCertFiles /fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
```

### 6.6. Start Chaincode Container

```bash
# Thay PACKAGE_ID bằng giá trị thực tế
CHAINCODE_ID="teaTraceCC_1.2.0:YOUR_PACKAGE_ID_HERE"

docker run -d \
  --name teatracecc-go \
  --network ibn-network \
  --restart unless-stopped \
  -e CHAINCODE_ID="$CHAINCODE_ID" \
  -e CORE_CHAINCODE_ID_NAME="$CHAINCODE_ID" \
  -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 \
  teatracecc-go:1.2.0 \
  ./teaTraceCC

# Kiểm tra logs
docker logs teatracecc-go
```

**Kết quả mong đợi:**
```
2025/12/05 14:36:33 Starting chaincode as external service at 0.0.0.0:9999 with ID teaTraceCC_1.2.0:...
```

---

## 7. Kiểm Tra Hệ Thống

### 7.1. Kiểm tra chaincode đã committed

```bash
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
  hyperledger/fabric-tools:2.5.9 \
  peer lifecycle chaincode querycommitted --channelID ibnchannel --name teaTraceCC
```

### 7.2. Test CreateBatch

```bash
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
  hyperledger/fabric-tools:2.5.9 \
  peer chaincode invoke \
    --channelID ibnchannel \
    --name teaTraceCC \
    --tls \
    --cafile /fabric/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/tls/ca.crt \
    --orderer orderer.ibn.vn:7050 \
    --ordererTLSHostnameOverride orderer.ibn.vn \
    --peerAddresses peer0.org1.ibn.vn:7051 \
    --tlsRootCertFiles /fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
    -c '{"function":"CreateBatch","Args":["BATCH-001","Thai Nguyen Farm","2025-12-05","Dried and packaged","CERT-12345"]}'
```

### 7.3. Test Query

```bash
docker run --rm \
  --network ibn-network \
  -v $(pwd)/core:/fabric \
  -e CORE_PEER_LOCALMSPID=Org1MSP \
  -e CORE_PEER_MSPCONFIGPATH=/fabric/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE=/fabric/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt \
  -e CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051 \
  hyperledger/fabric-tools:2.5.9 \
  peer chaincode query \
    --channelID ibnchannel \
    --name teaTraceCC \
    -c '{"function":"GetBatchInfo","Args":["BATCH-001"]}'
```

### 7.4. Truy cập các services

| Service | URL | Mô tả |
|---------|-----|-------|
| Frontend | http://localhost:3000 | Web application |
| Backend API | http://localhost:9900 | REST API |
| Admin Service | http://localhost:9902 | Admin operations |
| Grafana | http://localhost:9300 | Monitoring dashboard |
| Prometheus | http://localhost:9901 | Metrics |
| CouchDB | http://localhost:5984 | State database |

---

## 8. Xử Lý Lỗi Thường Gặp

### 8.1. Lỗi "channel creation request not allowed"

**Nguyên nhân**: Orderer chưa join channel.

**Giải pháp**: Thực hiện lại bước 5.4 (Join Orderers vào Channel).

### 8.2. Lỗi "chaincode already successfully installed"

**Nguyên nhân**: Package đã được install trước đó.

**Giải pháp**: Đây không phải lỗi, chaincode đã được install. Tiếp tục approve và commit.

### 8.3. Lỗi "endorsement failure during invoke"

**Nguyên nhân**: Chaincode container chưa chạy hoặc không kết nối được.

**Giải pháp**:
```bash
# Kiểm tra chaincode container
docker ps | grep teatracecc

# Nếu không có, start lại
docker start teatracecc-go

# Kiểm tra logs
docker logs teatracecc-go
```

### 8.4. Lỗi Docker network

**Nguyên nhân**: Container không thể giao tiếp với nhau.

**Giải pháp**:
```bash
# Kiểm tra network
docker network inspect ibn-network

# Tạo lại network nếu cần
docker network rm ibn-network
docker network create ibn-network
docker compose down
docker compose up -d
```

### 8.5. Lỗi permission trên WSL

**Nguyên nhân**: File permissions không đúng.

**Giải pháp**:
```bash
# Cấp quyền cho thư mục organizations
chmod -R 755 core/organizations/
```

### 8.6. Upgrade chaincode

Khi cần upgrade chaincode:

```bash
# 1. Stop chaincode container cũ
docker stop teatracecc-go && docker rm teatracecc-go

# 2. Build image mới
docker build -t teatracecc-go:1.3.0 ./teaTraceCC-go/

# 3. Tạo package mới với label mới
# (Lặp lại bước 6.2 với version mới)

# 4. Install package mới
# (Lặp lại bước 6.3)

# 5. Approve với sequence tăng lên
# (Lặp lại bước 6.4 với --sequence 2)

# 6. Commit với sequence mới
# (Lặp lại bước 6.5 với --sequence 2)

# 7. Start chaincode container mới
# (Lặp lại bước 6.6 với CHAINCODE_ID mới)
```

---

## Script Tự Động

Để tiện lợi, bạn có thể sử dụng các script sau:

| Script | Mô tả | Cách dùng |
|--------|-------|-----------|
| `scripts/generate-crypto.sh` | Tạo crypto materials và genesis block | `./scripts/generate-crypto.sh` |
| `scripts/init-database.sh` | Khởi tạo database và admin user | `./scripts/init-database.sh` |
| `scripts/join-orderers.sh` | Join orderers vào channel | `./scripts/join-orderers.sh` |
| `scripts/deploy-chaincode-ccaas.sh` | Deploy chaincode tự động | `./scripts/deploy-chaincode-ccaas.sh` |
| `scripts/test-chaincode.sh` | Test các functions của chaincode | `./scripts/test-chaincode.sh` |

```bash
# Cấp quyền thực thi cho tất cả scripts
chmod +x scripts/*.sh

# Chạy script deploy
./scripts/deploy-chaincode-ccaas.sh
```

---

## Tóm Tắt Workflow

### Lần đầu setup (máy mới, chưa có gì):
```bash
# 1. Clone và vào thư mục dự án
git clone <repo-url> ibn-docker-packaging && cd ibn-docker-packaging

# 2. Cấp quyền cho scripts
chmod +x scripts/*.sh

# 3. Tạo crypto materials và channel artifacts (nếu chưa có)
./scripts/generate-crypto.sh

# 4. Khởi động network
docker network create ibn-network 2>/dev/null || true
docker compose up -d

# 5. Đợi healthy rồi init database và join orderers
sleep 60
./scripts/init-database.sh
./scripts/join-orderers.sh

# 6. Deploy chaincode
./scripts/deploy-chaincode-ccaas.sh

# 7. Test
./scripts/test-chaincode.sh
```

### Restart network (đã có crypto và đã deploy chaincode):
```bash
# 1. Khởi động network
docker compose up -d

# 2. Đợi healthy rồi init database và join orderers
sleep 60
./scripts/init-database.sh
./scripts/join-orderers.sh

# 3. Start chaincode container (nếu chưa có)
CHAINCODE_ID=$(cat teaTraceCC-go/PACKAGE_ID.txt)
docker start teatracecc-go 2>/dev/null || docker run -d --name teatracecc-go --network ibn-network \
  -e CHAINCODE_ID="$CHAINCODE_ID" -e CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999 \
  teatracecc-go:1.2.0 ./teaTraceCC
```

---

## Liên Hệ & Hỗ Trợ

- **Issues**: Tạo issue trên GitHub repository
- **Documentation**: Xem thêm trong thư mục `docs/`

---

© 2025 IBN Network (ICTU Blockchain Network). All rights reserved.
