# Hướng Dẫn Deploy Chaincode Từ Đầu

## ⚠️ Lưu Ý
Do vấn đề với MSP identity trong peer container, **khuyến nghị sử dụng Frontend UI** hoặc **Admin Service API** để deploy.

---

## Cách 1: Sử Dụng Frontend UI (KHUYẾN NGHỊ)

1. **Mở trình duyệt**: http://localhost:3000
2. **Navigate**: Deploy Chaincode page
3. **Upload package**: Chọn file `teaTraceCC.tar.gz` (đã được package sẵn)
4. **Install**: Click "Install" button
5. **Approve**: Click "Approve" button  
6. **Commit**: Click "Commit" button

---

## Cách 2: Sử Dụng Admin Service API

### Bước 1: Package Chaincode
```bash
cd /home/exp2/ibn/teaTraceCC
npm run build
cp msp-config.json dist/
# Package sẽ được tạo qua API upload
```

### Bước 2: Upload Package
```bash
curl -X POST http://localhost:8090/api/v1/chaincode/upload \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -F "package=@teaTraceCC.tar.gz"
```

### Bước 3: Install
```bash
curl -X POST http://localhost:8090/api/v1/chaincode/install \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "packagePath": "/tmp/chaincode-uploads/teaTraceCC.tar.gz",
    "label": "teaTraceCC_1.1"
  }'
```

### Bước 4: Approve & Commit
Sử dụng API tương ứng hoặc Frontend UI.

---

## Cách 3: Package Chaincode Trước

Nếu bạn muốn package chaincode trước:

```bash
cd /home/exp2/ibn/teaTraceCC

# Build
npm run build
cp msp-config.json dist/

# Package (cần peer CLI trên host hoặc trong container)
# Option 1: Sử dụng admin-service
docker exec admin-service mkdir -p /app/chaincode/teaTraceCC/dist
docker cp dist/. admin-service:/app/chaincode/teaTraceCC/dist/
docker exec -w /app admin-service peer lifecycle chaincode package /tmp/teaTraceCC.tar.gz \
  --path /app/chaincode/teaTraceCC/dist \
  --lang node \
  --label teaTraceCC_1.1

# Copy package về host
docker cp admin-service:/tmp/teaTraceCC.tar.gz ./teaTraceCC.tar.gz
```

Sau đó upload file `teaTraceCC.tar.gz` qua Frontend UI hoặc API.

---

## Xóa Chaincode Cũ

Để xóa chaincode cũ (nếu cần):

```bash
# Stop chaincode containers
docker ps --filter "name=dev-peer" --filter "name=teaTraceCC" -q | xargs -r docker stop
docker ps --filter "name=dev-peer" --filter "name=teaTraceCC" -q | xargs -r docker rm

# Note: Không thể "uninstall" chaincode đã commit, chỉ có thể deploy version mới
```

---

## Verify Deployment

Sau khi deploy, verify:

```bash
# Query committed chaincode
curl http://localhost:8090/api/v1/chaincode/committed?channel=ibnchannel

# Hoặc qua peer container (nếu có quyền)
docker exec peer0.org1.ibn.vn peer lifecycle chaincode querycommitted \
  --channelID ibnchannel --name teaTraceCC
```

---

## Troubleshooting

### Lỗi: "identity is not an admin"
- **Giải pháp**: Sử dụng Frontend UI hoặc Admin Service API (đã có admin identity)

### Lỗi: "chaincode already installed"
- **Giải pháp**: Đây là idempotent operation, có thể bỏ qua hoặc dùng package ID hiện có

### Lỗi: "sequence mismatch"
- **Giải pháp**: Tăng sequence number (1 → 2 → 3...)

---

**Khuyến nghị**: Sử dụng **Frontend UI** để deploy vì đơn giản và đã được cấu hình đúng với admin identity.
