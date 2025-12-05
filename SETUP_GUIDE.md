# 🚀 IBN Network - Hướng Dẫn Cài Đặt Thủ Công (Manual Setup)

Tài liệu này hướng dẫn chi tiết cách xây dựng lại mạng lưới Hyperledger Fabric (IBN Network) **bằng các lệnh thủ công**, giúp bạn kiểm soát hoàn toàn quá trình.

## 🚨 Vấn Đề Thường Gặp & Giải Pháp Nhanh

### ❌ Script không hoạt động / Lỗi "broken pipe" khi deploy chaincode

**Nguyên nhân phổ biến:**
- **WSL2 Docker socket limitation** (nếu đang dùng WSL2)
- Docker daemon quá tải hoặc timeout
- Chaincode package thiếu file

**Giải pháp nhanh (theo thứ tự):**

1. **Pre-pull builder images** (5 phút):
   ```bash
   docker pull hyperledger/fabric-nodeenv:2.5.9
   docker pull hyperledger/fabric-ccenv:2.5.9
   # Sau đó chạy lại script hoặc bước 9.3
   ```

2. **Restart Docker Desktop** (Windows):
   - Mở Docker Desktop → Settings → Apply & Restart

3. **Tăng Docker resources** (Windows):
   - Docker Desktop → Settings → Resources → Advanced
   - Memory: ít nhất 4GB (khuyến nghị 8GB)
   - CPUs: ít nhất 2 (khuyến nghị 4)

4. **Skip chaincode tạm thời** (nếu chỉ cần test hệ thống):
   ```bash
   # Hệ thống vẫn hoạt động mà không có chaincode
   docker compose up -d
   bash scripts/setup.sh --create-admin
   # Vào http://localhost:9999 để test
   ```

5. **Xem chi tiết troubleshooting:** Phần 9.8 bên dưới

### ✅ Checklist Trước Khi Bắt Đầu

- [ ] Docker và Docker Compose đã cài đặt
- [ ] User đã được thêm vào `docker` group (không cần sudo)
- [ ] Đã kiểm tra `docker ps` chạy được
- [ ] (WSL2) Đã đọc phần cảnh báo về WSL2 ở trên

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

> 🔐 **Quyền hạn (rất quan trọng)**
>
> - Tài liệu này được thiết kế để **có thể dùng như “backup script” khi bạn KHÔNG có quyền sudo trong WSL**.
> - Điều kiện bắt buộc khi bạn không có sudo:
>   - Docker + Docker Compose đã được **admin cài sẵn** và chạy nền.
>   - User của bạn đã được admin thêm vào `docker` group (để chạy được `docker ...` không cần sudo).
> - Mọi lệnh trong guide **không dùng `sudo`**, trừ một vài dòng trong phần *Troubleshooting* được đánh dấu rõ là **(Admin only)** – nếu bạn không có quyền, hãy nhờ admin chạy giúp những lệnh đó.

### ✅ Checklist nhanh cho WSL **không có quyền admin**

- **Bước 1**: Mở WSL, chạy:
  ```bash
  docker ps
  ```
  - Nếu chạy được và hiện danh sách containers (hoặc rỗng) → **OK, tiếp tục các bước phía dưới**.
  - Nếu báo lỗi kiểu `permission denied` hoặc `Cannot connect to the Docker daemon`:
    - Bạn **không tự fix được nếu không có sudo**.
    - Gửi cho admin (hoặc người cài máy) đoạn sau và nhờ họ chạy trên Linux host (Admin only):
      ```bash
      # Thêm user vào docker group (Admin only)
      sudo usermod -aG docker <username>
      sudo systemctl restart docker
      ```
      Sau đó bạn **logout / login lại WSL** rồi thử lại `docker ps`.

- **Bước 2**: Sau khi `docker ps` chạy OK, bạn có thể **làm toàn bộ các bước còn lại trong file này mà không cần sudo**.

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

> ⚠️ **CẢNH BÁO QUAN TRỌNG CHO WSL2 USERS:**
> 
> Nếu bạn đang dùng **WSL2 (Windows Subsystem for Linux)**, có thể gặp lỗi `write unix @->/run/docker.sock: write: broken pipe` khi install chaincode. Đây là limitation đã biết của WSL2 với Docker-in-Docker.
> 
> **Giải pháp nhanh:**
> 1. **Pre-pull builder images** trước khi install (xem bước 9.3.1)
> 2. **Restart Docker Desktop** và tăng resources
> 3. Nếu vẫn lỗi, xem phần **9.8 - Troubleshooting** để có các workaround chi tiết
> 4. **Khuyến nghị:** Deploy trên Linux native để tránh hoàn toàn vấn đề này

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

# Copy package.json vào dist/ (QUAN TRỌNG: Node.js chaincode cần package.json)
cp package.json dist/

# Regenerate package-lock.json trong dist/ để đồng bộ với package.json
# (QUAN TRỌNG: npm ci yêu cầu lock file đồng bộ với package.json)
# Sử dụng --omit=dev để chỉ có production dependencies (giống như Fabric build)
cd dist/
rm -f package-lock.json
npm install --omit=dev --package-lock-only
cd ..

# Kiểm tra cấu trúc chaincode (đảm bảo có đủ file)
ls -la dist/
# Phải có: index.js, package.json, package-lock.json, msp-config.json, và các file models/, utils/
```

### 9.2. Package Chaincode

Tạo package file cho chaincode:

```bash
# Quay về thư mục gốc
cd ..

# Kiểm tra Docker daemon đang chạy (QUAN TRỌNG để tránh lỗi "broken pipe")
docker ps > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "❌ Docker daemon không chạy. Vui lòng khởi động Docker trước."
    exit 1
fi

# Tạo thư mục chaincode trong container peer0 (nếu chưa tồn tại)
docker exec peer0.org1.ibn.vn mkdir -p /opt/chaincode

# Xóa thư mục cũ nếu có (tránh conflict)
docker exec peer0.org1.ibn.vn rm -rf /opt/chaincode/teaTraceCC

# Copy thư mục dist vào container peer0 để package
docker cp teaTraceCC/dist peer0.org1.ibn.vn:/opt/chaincode/teaTraceCC

# Kiểm tra file đã được copy đúng (đảm bảo có package.json)
docker exec peer0.org1.ibn.vn ls -la /opt/chaincode/teaTraceCC/ | grep package.json
if [ $? -ne 0 ]; then
    echo "❌ Lỗi: package.json không có trong dist/. Vui lòng kiểm tra lại bước 9.1."
    exit 1
fi

# Package chaincode (với đầy đủ environment variables)
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
  -w /opt/chaincode \
  peer0.org1.ibn.vn \
  peer lifecycle chaincode package teaTraceCC.tar.gz \
  --path /opt/chaincode/teaTraceCC \
  --lang node \
  --label teaTraceCC_1.1.0

# Kiểm tra package file đã được tạo
if [ $? -ne 0 ]; then
    echo "❌ Lỗi: Không thể tạo package file. Kiểm tra lại cấu trúc chaincode."
    exit 1
fi

# Copy package file ra ngoài để dùng cho các peer khác
docker cp peer0.org1.ibn.vn:/opt/chaincode/teaTraceCC.tar.gz ./teaTraceCC.tar.gz

# Kiểm tra file đã được copy ra ngoài
if [ ! -f ./teaTraceCC.tar.gz ]; then
    echo "❌ Lỗi: Không thể copy package file ra ngoài."
    exit 1
fi

echo "✅ Package file đã được tạo: ./teaTraceCC.tar.gz"
ls -lh ./teaTraceCC.tar.gz
```

### 9.3. Install Chaincode trên các Peers

Cài đặt chaincode trên tất cả các peers:

**⚠️ LƯU Ý QUAN TRỌNG:** 
- Quá trình install có thể mất **2-5 phút** cho mỗi peer (do build Docker image)
- Nếu gặp lỗi "broken pipe", đây thường là lỗi timeout hoặc **WSL2 Docker socket limitation** - cần retry hoặc xem phần 9.8
- Đảm bảo Docker daemon đang chạy và có đủ resources

#### 9.3.1. Pre-pull Builder Images (QUAN TRỌNG cho WSL2)

**Bước này giúp tránh lỗi "broken pipe" trong WSL2** bằng cách pull builder images về host trước:

```bash
# Pull chaincode builder images về host (tránh peer phải download khi build)
echo "📥 Pulling chaincode builder images..."
docker pull hyperledger/fabric-nodeenv:2.5.9
docker pull hyperledger/fabric-ccenv:2.5.9

# Kiểm tra images đã có
echo "✅ Builder images ready:"
docker images | grep -E "(fabric-nodeenv|fabric-ccenv)"
echo ""
```

**Sau khi pull xong, tiếp tục với các bước bên dưới.**

#### 9.3.2. Install Chaincode

```bash
# Kiểm tra package file tồn tại
if [ ! -f ./teaTraceCC.tar.gz ]; then
    echo "❌ Lỗi: File teaTraceCC.tar.gz không tồn tại. Vui lòng chạy lại bước 9.2."
    exit 1
fi

# Kiểm tra Docker daemon và peer containers
echo "🔍 Kiểm tra Docker daemon và peer containers..."
if ! docker ps > /dev/null 2>&1; then
    echo "❌ Docker daemon không chạy. Vui lòng khởi động Docker trước."
    exit 1
fi

if ! docker ps | grep -q "peer0.org1.ibn.vn"; then
    echo "❌ Peer containers không chạy. Vui lòng khởi động network trước (docker compose up -d)."
    exit 1
fi

echo "✅ Docker daemon và peer containers đang chạy"
echo ""

# Tạo thư mục chaincode cho các peers (nếu chưa tồn tại)
docker exec peer0.org1.ibn.vn mkdir -p /opt/chaincode
docker exec peer1.org1.ibn.vn mkdir -p /opt/chaincode
docker exec peer2.org1.ibn.vn mkdir -p /opt/chaincode

# Copy Admin MSP vào các peers (cần Admin MSP để có quyền install chaincode)
docker cp core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp peer0.org1.ibn.vn:/tmp/admin_msp
docker cp core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp peer1.org1.ibn.vn:/tmp/admin_msp
docker cp core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp peer2.org1.ibn.vn:/tmp/admin_msp

# Hàm install với retry mechanism
install_chaincode() {
    local peer_name=$1
    local peer_port=$2
    local package_path=$3
    local max_retries=3
    local retry_count=0
    
    while [ $retry_count -lt $max_retries ]; do
        echo "📦 Installing chaincode on ${peer_name} (attempt $((retry_count + 1))/${max_retries})..."
        
        # Install với timeout 5 phút
        if timeout 300 docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
          -e CORE_PEER_TLS_ENABLED=true \
          -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
          -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
          -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
          -w /opt/chaincode \
          ${peer_name} \
          peer lifecycle chaincode install ${package_path} 2>&1; then
            echo "✅ Installed on ${peer_name}"
            return 0
        else
            retry_count=$((retry_count + 1))
            if [ $retry_count -lt $max_retries ]; then
                echo "⚠️  Install failed, retrying in 10 seconds..."
                sleep 10
            else
                echo "❌ Failed to install on ${peer_name} after ${max_retries} attempts"
                return 1
            fi
        fi
    done
}

# Install trên Peer0
install_chaincode "peer0.org1.ibn.vn" "7051" "teaTraceCC.tar.gz"
INSTALL_PEER0=$?
echo ""

# Install trên Peer1
docker cp teaTraceCC.tar.gz peer1.org1.ibn.vn:/opt/chaincode/
install_chaincode "peer1.org1.ibn.vn" "8051" "/opt/chaincode/teaTraceCC.tar.gz"
INSTALL_PEER1=$?
echo ""

# Install trên Peer2
docker cp teaTraceCC.tar.gz peer2.org1.ibn.vn:/opt/chaincode/
install_chaincode "peer2.org1.ibn.vn" "9051" "/opt/chaincode/teaTraceCC.tar.gz"
INSTALL_PEER2=$?
echo ""

# Tóm tắt kết quả
echo "📊 Tóm tắt kết quả install:"
if [ $INSTALL_PEER0 -eq 0 ]; then echo "  ✅ peer0.org1.ibn.vn"; else echo "  ❌ peer0.org1.ibn.vn"; fi
if [ $INSTALL_PEER1 -eq 0 ]; then echo "  ✅ peer1.org1.ibn.vn"; else echo "  ❌ peer1.org1.ibn.vn"; fi
if [ $INSTALL_PEER2 -eq 0 ]; then echo "  ✅ peer2.org1.ibn.vn"; else echo "  ❌ peer2.org1.ibn.vn"; fi

# Kiểm tra ít nhất 1 peer đã install thành công
if [ $INSTALL_PEER0 -ne 0 ] && [ $INSTALL_PEER1 -ne 0 ] && [ $INSTALL_PEER2 -ne 0 ]; then
    echo ""
    echo "❌ Không thể install chaincode trên bất kỳ peer nào."
    echo "💡 Xem phần Troubleshooting (mục 9.8) để xử lý lỗi."
    exit 1
fi

# Query package ID từ peer đã install thành công
echo ""
echo "🔍 Querying installed chaincode để lấy PACKAGE_ID..."
if [ $INSTALL_PEER0 -eq 0 ]; then
    docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
      -e CORE_PEER_TLS_ENABLED=true \
      -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
      -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
      -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
      peer0.org1.ibn.vn \
      peer lifecycle chaincode queryinstalled | grep "teaTraceCC_1.1.0"
fi
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
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
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
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
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
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
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
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
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
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
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
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
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
- File `msp-config.json` và `package.json` phải được copy vào thư mục `dist/` trước khi package
- PACKAGE_ID sẽ khác nhau mỗi lần install, cần lưu lại để dùng cho approve
- Channel name: `ibnchannel`
- MSP ID: `Org1MSP`

### 9.8. Troubleshooting - Lỗi "broken pipe"

Nếu gặp lỗi `write unix @->/run/docker.sock: write: broken pipe` khi install chaincode:

**Nguyên nhân chính:**
- **WSL2 Docker socket limitation** (phổ biến nhất trong WSL2)
- **Timeout trong quá trình build Docker image**
- Docker daemon bị disconnect hoặc quá tải
- Chaincode package thiếu file quan trọng (package.json, package-lock.json)
- Peer container không có đủ resources để build image

**⚠️ QUAN TRỌNG: Nếu bạn đang dùng WSL2, hãy đọc phần "Giải pháp đặc biệt cho WSL2" trước!**

**Giải pháp theo thứ tự ưu tiên:**

#### 0. **🔴 GIẢI PHÁP ĐẶC BIỆT CHO WSL2 (ĐỌC TRƯỚC)**

Nếu bạn đang dùng **WSL2 (Windows Subsystem for Linux)**, lỗi "broken pipe" thường do **WSL2 Docker socket limitation**. Đây là vấn đề đã biết của WSL2 với Docker-in-Docker.

**Workaround 1: Pre-pull Builder Images (Thử ngay)**

Pull các builder images về host trước khi install để peer không phải download:

```bash
# Pull chaincode builder images về host
docker pull hyperledger/fabric-nodeenv:2.5.9
docker pull hyperledger/fabric-ccenv:2.5.9

# Kiểm tra images đã có
docker images | grep -E "(fabric-nodeenv|fabric-ccenv)"

# Sau đó thử install lại chaincode
# (Chạy lại bước 9.3)
```

**Workaround 2: Restart Docker Desktop (Windows)**

Nếu đang dùng Docker Desktop trên Windows:

```bash
# 1. Mở Docker Desktop
# 2. Click Settings → General
# 3. Bật "Use the WSL 2 based engine" (nếu chưa bật)
# 4. Click "Apply & Restart"
# 5. Đợi Docker Desktop khởi động lại xong
# 6. Thử install chaincode lại
```

**Workaround 3: Tăng Docker Resources (Windows)**

Trong Docker Desktop:
1. Settings → Resources → Advanced
2. Tăng **Memory** lên ít nhất **4GB** (khuyến nghị 8GB)
3. Tăng **CPUs** lên ít nhất **2** (khuyến nghị 4)
4. Click "Apply & Restart"
5. Thử install lại

**Workaround 4: Skip Chaincode Tạm Thời (Nếu cần test hệ thống)**

Nếu chaincode không deploy được nhưng bạn cần test các phần khác của hệ thống:

```bash
# Hệ thống vẫn hoạt động được mà không có chaincode:
# ✅ Backend API: Hoạt động
# ✅ Frontend Dashboard: Hoạt động  
# ✅ Authentication: Hoạt động
# ✅ Database: Hoạt động
# ❌ Blockchain transactions: Không hoạt động (cần chaincode)

# Để test admin login và dashboard:
docker compose up -d
bash scripts/setup.sh --create-admin

# Sau đó vào http://localhost:9999 để test
```

**Workaround 5: Deploy trên Linux Native (Khuyến nghị cho Production)**

Nếu có Ubuntu VM hoặc máy Linux thật, deploy trên đó sẽ không gặp vấn đề WSL2:

```bash
# Clone project sang Linux native
git clone <repo> && cd <project>
bash scripts/setup.sh --fresh
```

**Workaround 6: Sử dụng External Chaincode (Nâng cao)**

Nếu các workaround trên không work, có thể chuyển sang External Chaincode mode (chaincode chạy như service riêng, không cần Docker build trong peer). Xem phần 9.9 bên dưới.

---

#### 1. **Retry với timeout dài hơn (Thử ngay)**
Hướng dẫn đã có retry mechanism tự động (3 lần), nhưng nếu vẫn fail:

```bash
# Thử install lại với timeout 5 phút
timeout 300 docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
  -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
  -w /opt/chaincode \
  peer0.org1.ibn.vn \
  peer lifecycle chaincode install teaTraceCC.tar.gz
```

#### 2. **Kiểm tra và cleanup Docker resources**
```bash
# Kiểm tra Docker system resources
docker system df

# Cleanup unused Docker resources (cẩn thận - sẽ xóa unused images/containers)
docker system prune -f

# Kiểm tra Docker daemon
docker ps
# Nếu lỗi, khởi động lại Docker (Admin only): 
#   - Linux: sudo systemctl restart docker
#   - Windows: Restart Docker Desktop (Docker Desktop UI)
```

#### 3. **Kiểm tra package file và cấu trúc**
```bash
# Kiểm tra file tồn tại và không bị corrupt
ls -lh ./teaTraceCC.tar.gz
file ./teaTraceCC.tar.gz
# Phải là: gzip compressed data

# Kiểm tra cấu trúc bên trong package
tar -tzf teaTraceCC.tar.gz | head -5
# Phải có: metadata.json, code.tar.gz

# Extract và kiểm tra code.tar.gz
mkdir -p /tmp/check_package
cd /tmp/check_package
tar -xzf /mnt/e/luongbeo/teaTraceCC.tar.gz
tar -tzf code.tar.gz | grep -E "(package\.json|index\.js|package-lock\.json)" | head -5
# Phải có: package.json, package-lock.json, index.js (hoặc src/index.js)
```

#### 4. **Kiểm tra peer containers và logs**
```bash
# Kiểm tra peer containers đang chạy và healthy
docker ps | grep peer
# Phải thấy peer0, peer1, peer2 với status "Up" và "(healthy)"

# Kiểm tra logs của peer container để tìm lỗi chi tiết
docker logs peer0.org1.ibn.vn --tail 100 | grep -E "(chaincode|error|failed|broken|npm|docker)" -i

# Kiểm tra logs của Docker build process
docker logs peer0.org1.ibn.vn --tail 200 | grep -A 20 "buildImage"
```

#### 5. **Restart peer container (nếu cần)**
```bash
# Restart peer container
docker restart peer0.org1.ibn.vn

# Đợi container khởi động lại (30 giây)
echo "Đợi peer container khởi động lại..."
sleep 30

# Kiểm tra container đã sẵn sàng
docker exec peer0.org1.ibn.vn peer version

# Thử install lại
```

#### 6. **Kiểm tra Docker socket permissions (Linux only)**
```bash
# Kiểm tra quyền truy cập Docker socket
ls -la /var/run/docker.sock
# Phải có quyền: srw-rw----

# Nếu không có quyền, thêm user vào docker group (Admin only):
#   sudo usermod -aG docker $USER
# Sau đó logout/login lại để áp dụng thay đổi
```

#### 7. **Kiểm tra package-lock.json đã sync**
```bash
# Vào thư mục dist và kiểm tra
cd teaTraceCC/dist

# Kiểm tra package.json và package-lock.json có sync không
npm ls --depth=0 2>&1 | head -10
# Nếu có lỗi về version mismatch, regenerate package-lock.json:
rm -f package-lock.json
npm install --omit=dev --package-lock-only

# Quay lại và package lại
cd ../..
# Chạy lại bước 9.2
```

#### 8. **Kiểm tra Docker build images đang chạy**
```bash
# Xem có Docker build process nào đang chạy không
docker ps -a | grep -E "(build|chaincode)"

# Xem Docker images liên quan đến chaincode
docker images | grep -E "(dev-peer|chaincode|teaTraceCC)"
```

#### 9. **Giải pháp cuối cùng: Clean install**
Nếu tất cả các bước trên không giải quyết được:

```bash
# 1. Xóa package file cũ
rm -f ./teaTraceCC.tar.gz

# 2. Xóa thư mục dist và rebuild
cd teaTraceCC
rm -rf dist
npm run build
cp msp-config.json dist/
cp package.json dist/
cd dist/
rm -f package-lock.json
npm install --omit=dev --package-lock-only
cd ../..

# 3. Chạy lại từ bước 9.2 (Package Chaincode)
# 4. Chạy lại bước 9.3 (Install Chaincode) với retry mechanism
```

**Lưu ý quan trọng:**
- Lỗi "broken pipe" thường xảy ra do **WSL2 Docker socket limitation** hoặc **timeout** trong quá trình build Docker image
- Quá trình install có thể mất **2-5 phút** cho mỗi peer
- Hướng dẫn đã có **retry mechanism tự động** (3 lần với timeout 5 phút mỗi lần)
- Nếu vẫn fail sau 3 lần retry, thử các workaround cho WSL2 ở trên
- **Khuyến nghị:** Nếu có thể, deploy trên Linux native để tránh hoàn toàn vấn đề WSL2

---

### 9.9. External Chaincode (Workaround Nâng Cao)

Nếu tất cả các giải pháp trên không work và bạn vẫn cần chaincode hoạt động, có thể chuyển sang **External Chaincode** mode. External Chaincode chạy như một service riêng, không cần Docker build trong peer.

**⚠️ Lưu ý:** External Chaincode phức tạp hơn và cần sửa đổi chaincode code. Chỉ dùng khi thực sự cần thiết.

**Bước 1: Sửa chaincode để chạy như gRPC server**

Chaincode cần được sửa để expose gRPC endpoint. Xem tài liệu Hyperledger Fabric về External Chaincode.

**Bước 2: Tạo connection.json**

```bash
# Tạo file connection.json trong chaincode directory
cat > chaincode/teaTraceCC/connection.json <<EOF
{
  "address": "chaincode-teaTraceCC:9999",
  "dial_timeout": "10s",
  "tls_required": false
}
EOF
```

**Bước 3: Package với connection.json**

```bash
# Package chaincode với connection.json thay vì code
docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
  -e CORE_PEER_TLS_ENABLED=true \
  -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
  -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
  -w /opt/chaincode \
  peer0.org1.ibn.vn \
  peer lifecycle chaincode package teaTraceCC_external.tar.gz \
  --path /opt/chaincode/teaTraceCC \
  --lang external \
  --label teaTraceCC_1.0
```

**Bước 4: Deploy chaincode service**

Thêm service vào `docker-compose.yml`:

```yaml
chaincode-teaTraceCC:
  container_name: chaincode-teaTraceCC
  image: node:18-alpine
  working_dir: /app
  command: sh -c "npm install && node dist/index.js"
  environment:
    - CHAINCODE_ID=teaTraceCC_1.0:latest
    - CHAINCODE_SERVER_ADDRESS=0.0.0.0:9999
  volumes:
    - ./chaincode/teaTraceCC/dist:/app
  networks:
    - ibn-network
  restart: unless-stopped
```

**Bước 5: Install và approve như bình thường**

Sau khi chaincode service chạy, install và approve như các bước 9.3-9.5.

**⚠️ Lưu ý:** External Chaincode cần chaincode code được sửa để hỗ trợ gRPC server mode. Đây là giải pháp nâng cao, chỉ dùng khi thực sự cần thiết.

---

### 9.10. Tóm Tắt Các Giải Pháp

| Giải pháp | Độ khó | Thời gian | Khuyến nghị |
|-----------|--------|-----------|-------------|
| Pre-pull builder images | Dễ | 5 phút | ✅ Thử đầu tiên |
| Restart Docker Desktop | Dễ | 2 phút | ✅ Thử thứ hai |
| Tăng Docker resources | Dễ | 5 phút | ✅ Thử thứ ba |
| Skip chaincode tạm thời | Dễ | 0 phút | ✅ Nếu chỉ cần test hệ thống |
| Deploy trên Linux native | Trung bình | 30 phút | ✅✅ Khuyến nghị cho production |
| External Chaincode | Khó | 1-2 giờ | ⚠️ Chỉ khi thực sự cần |
