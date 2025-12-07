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
git clone https://github.com/Callmeduobgne/luongthiOlympic.git && cd luongthiOlympic

# Nếu CẦN tạo crypto materials từ đầu (chỉ khi chưa có thư mục organizations/)
sudo ./scripts/generate-crypto.sh

# Khởi động network (Docker Compose sẽ tự tạo network)
docker compose up -d
sleep 60

# Init Database
sudo ./scripts/init-database.sh

# Join orderers vào channel
sudo ./scripts/join-orderers.sh

# Join peers vào channel
sudo ./scripts/join-peers.sh

# Deploy chaincode
sudo ./scripts/deploy-chaincode-ccaas.sh

# Test chaincode
sudo ./scripts/test-chaincode.sh
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

> ⚠️ **QUAN TRỌNG**: Bước này chỉ cần thực hiện **MỘT LẦN** khi setup dự án mới.

Bạn có thể sử dụng script tự động để tiết kiệm thời gian hoặc thực hiện các bước thủ công để hiểu rõ quy trình.

### Cách 1: Sử dụng Script Tự Động (Khuyến nghị cho người mới)

```bash
sudo ./scripts/generate-crypto.sh
```

### Cách 2: Thực Hiện Thủ Công (Khuyến nghị để hiểu sâu)

Nếu bạn muốn kiểm soát từng bước, hãy làm theo hướng dẫn sau:

### 4.1. Pull Fabric Tools Image

```bash
docker pull hyperledger/fabric-tools:2.5.9
```

### 4.2. Tạo Crypto Materials (Certificates)

```bash
# Xóa crypto cũ nếu có (CHỈ KHI MUỐN TẠO LẠI TỪ ĐẦU)
sudo rm -rf core/organizations/* core/channel-artifacts/* core/system-genesis-block/*

# Tạo crypto materials bằng cryptogen
docker run --rm -v $(pwd)/core:/fabric -w /fabric \
  hyperledger/fabric-tools:2.5.9 \
  cryptogen generate --config=/fabric/crypto-config.yaml --output=/fabric/organizations
```

### 4.3. Tạo config.yaml và Genesis Block

Bạn cần tạo `config.yaml` cho các MSP (NodeOUs enabled) và `genesis.block`.

*(Chi tiết xem thêm trong script `scripts/generate-crypto.sh` hoặc tài liệu Fabric)*

```bash
# Tạo system genesis block
mkdir -p core/system-genesis-block
docker run --rm -v $(pwd)/core:/fabric -w /fabric/configtx \
  hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile RaftOrdererGenesis -channelID system-channel -outputBlock /fabric/system-genesis-block/genesis.block
```

### 4.4. Tạo Channel Artifacts

```bash
mkdir -p core/channel-artifacts

# 1. Channel Transaction
docker run --rm -v $(pwd)/core:/fabric -w /fabric/configtx hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile ThreePeersChannel -channelID ibnchannel -outputCreateChannelTx /fabric/channel-artifacts/ibnchannel.tx

# 2. Anchor Peers Update
docker run --rm -v $(pwd)/core:/fabric -w /fabric/configtx hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile ThreePeersChannel -channelID ibnchannel \
  -outputAnchorPeersUpdate /fabric/channel-artifacts/Org1MSPanchors.tx -asOrg Org1MSP

# 3. Channel Genesis Block (cho orderer join)
docker run --rm -v $(pwd)/core:/fabric -w /fabric/configtx hyperledger/fabric-tools:2.5.9 \
  configtxgen -profile ThreePeersChannel -channelID ibnchannel -outputBlock /fabric/channel-artifacts/ibnchannel.block
```

### 4.5. Phân quyền
### 5.1. Khởi động tất cả services

**Lưu ý**: Không cần tạo network thủ công (`docker network create`), Docker Compose sẽ tự động xử lý.

```bash
# Khởi động toàn bộ stack
docker compose up -d

# Kiểm tra trạng thái
docker compose ps
```

### 5.2. Đợi services healthy

```bash
# Đợi tất cả containers healthy (khoảng 1-2 phút)
echo "Đợi services khởi động..."
sleep 60

# Kiểm tra trạng thái
docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "orderer|peer|couchdb|backend|postgres"
```

**Kết quả mong đợi:** Các services đều ở trạng thái `(healthy)`.

### 5.3. Khởi tạo Database và Admin User

> ⚠️ **QUAN TRỌNG**: Backend cần database được khởi tạo đúng để hoạt động!

```bash
# Chạy script khởi tạo database
./scripts/init-database.sh
```

Script sẽ:
1. Chờ database sẵn sàng
2. Chạy migrations (tạo bảng)
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

### 5.4. Join Orderers vào Channel

⚠️ **QUAN TRỌNG**: Bước này cần thực hiện mỗi khi orderer volume bị reset hoặc lần đầu khởi động.

```bash
# Sử dụng script tự động
./scripts/join-orderers.sh
```

### 5.5. Join Peers vào Channel (Chỉ lần đầu)

> Nếu peer đã join channel trước đó và dữ liệu còn trong volume, bước này có thể bỏ qua.

```bash
# Script tự động join peers (nếu chưa có trong scripts, dùng lệnh docker exec)
# Join peer0
docker exec peer0.org1.ibn.vn peer channel join -b /var/hyperledger/fabric/channel-artifacts/ibnchannel.block
# Join peer1
docker exec peer1.org1.ibn.vn peer channel join -b /var/hyperledger/fabric/channel-artifacts/ibnchannel.block
# Join peer2
docker exec peer2.org1.ibn.vn peer channel join -b /var/hyperledger/fabric/channel-artifacts/ibnchannel.block
```

### 5.6. Kiểm tra Peer đã join channel

```bash
docker exec peer0.org1.ibn.vn peer channel list
```

---

## 6. Deploy Chaincode

> 💡 **Gợi ý**: Sử dụng script tự động `./scripts/deploy-chaincode-ccaas.sh` để deploy nhanh.

```bash
./scripts/deploy-chaincode-ccaas.sh
```

---

## 7. Kiểm Tra Hệ Thống

### 7.1. Truy cập các services

| Service | URL | Mô tả |
|---------|-----|-------|
| Frontend | http://localhost:9999 | Web application (User) |
| Backend API | http://localhost:9900 | REST API |
| Admin Service | http://localhost:9902 | Admin operations |
| Chaincode | External Service | Port 9999 |
| Grafana | http://localhost:9300 | Monitoring dashboard |
| Prometheus | http://localhost:9901 | Metrics |

---

## 8. Xử Lý Lỗi Thường Gặp

### 8.1. Lỗi "network ibn-network was found but has incorrect label"

**Nguyên nhân**: Bạn đã tạo network thủ công bằng `docker network create`.

**Giải pháp**:
```bash
docker network rm ibn-network
docker compose up -d
```

### 8.2. Lỗi Database "relation does not exist"

**Nguyên nhân**: Database chưa được khởi tạo hoặc Volume mount bị sai.

**Giải pháp**:
1. Đảm bảo file `docker-compose.yml` **KHÔNG** mount thư mục migrations vào `/docker-entrypoint-initdb.d`.
2. Chạy `./scripts/init-database.sh` sau khi DB container healthy.

### 8.3. Lỗi "channel creation request not allowed"

**Nguyên nhân**: Orderer chưa join channel.

**Giải pháp**: Chạy `./scripts/join-orderers.sh`.


### 8.4. Upgrade chaincode

Khi cần upgrade chaincode:

```bash
# 1. Stop chaincode container cũ
docker stop teatracecc-go && docker rm teatracecc-go

# 2. Build image mới
docker build -t teatracecc-go:1.3.0 ./teaTraceCC-go/

# 3. Tạo package mới với label mới
# (Lặp lại bước 6.2 với version mới, ví dụ 1.3.0)

# 4. Install package mới
# (Lặp lại bước 6.3 với package mới)

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
| `scripts/clean-generated-files.sh` | Xóa toàn bộ crypto & artifacts cũ | `./scripts/clean-generated-files.sh` |
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
# Lưu ý: Docker Compose tự tạo network "ibn-network"
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
