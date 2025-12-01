# Tea Traceability Chaincode (teaTraceCC)

> Chaincode truy xuất nguồn gốc trà trên Hyperledger Fabric với hash verification và MSP authorization.

## 🚀 Thông tin Chaincode

| Thông tin | Giá trị |
|-----------|---------|
| **Name** | teaTraceCC |
| **Version** | 1.0.0 |
| **Language** | Node.js (TypeScript) |
| **Network** | Hyperledger Fabric 2.x |

## ⚡ Quick Start

### Tạo lô trà mới
```bash
peer chaincode invoke -C mychannel -n teaTraceCC \
  -c '{"Args":["createBatch","BATCH001","Moc Chau","2024-11-08","Organic","VN-ORG-2024"]}'
```

### Query thông tin lô trà
```bash
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

## 📋 Tính năng chính

✅ **Tạo lô trà** - Farmer tạo lô trà mới với thông tin đầy đủ  
✅ **Xác minh lô trà** - Verifier xác minh hash để chống giả mạo  
✅ **Cập nhật trạng thái** - Admin quản lý lifecycle của lô trà  
✅ **Hash verification** - SHA-256 đảm bảo tính toàn vẹn  
✅ **MSP portable** - Config linh hoạt, chạy trên mọi network  
✅ **Query functions** - Query batches theo nhiều tiêu chí (status, owner, all)  
✅ **History tracking** - Theo dõi lịch sử thay đổi của batch  
✅ **Input validation** - Validate tất cả inputs để đảm bảo data integrity  
✅ **Pagination support** - Hỗ trợ pagination cho các query functions

## 📋 Yêu cầu hệ thống

- ✅ **Hyperledger Fabric** 2.x
- ✅ **Node.js** >= 16.0.0
- ✅ **Docker** & Docker Compose
- ✅ **Fabric network** đã chạy (peer + orderer)

## 📦 Cài đặt & Deploy

### 1. Build chaincode
```bash
npm install
npm run build
```

### 2. Tùy chỉnh MSP (Optional)
Chỉnh `msp-config.json` nếu network dùng tên MSP khác:
```json
{
  "mspRoles": {
    "farmer": {"mspId": "YourOrgMSP"},
    "verifier": {"mspId": "YourOrg2MSP"},
    "admin": {"mspId": "YourOrg3MSP"}
  }
}
```

### 3. Package & Install
```bash
# Package
cp msp-config.json dist/
peer lifecycle chaincode package teaTraceCC.tar.gz \
  --path ./dist --lang node --label teaTraceCC_1.0

# Install
peer lifecycle chaincode install teaTraceCC.tar.gz
# Lưu lại PACKAGE_ID
```

### 4. Approve & Commit
```bash
# Approve (với PACKAGE_ID từ bước trên)
peer lifecycle chaincode approveformyorg \
  -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
  --channelID mychannel --name teaTraceCC --version 1.0 \
  --package-id <PACKAGE_ID> --sequence 1 --tls \
  --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

# Commit (sau khi đủ orgs approve)
peer lifecycle chaincode commit \
  -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com \
  --channelID mychannel --name teaTraceCC --version 1.0 \
  --sequence 1 --tls \
  --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
  --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt

# Verify
peer lifecycle chaincode querycommitted --channelID mychannel
```

## 🔐 Phân quyền MSP

| MSP Role | Vai trò | Quyền hạn |
|----------|---------|-----------|
| **Farmer** (Org1MSP) | Nông dân | Tạo lô trà, cập nhật status |
| **Verifier** (Org1MSP) | Kiểm định | Xác minh lô trà |
| **Admin** (Org1MSP) | Quản trị | Cập nhật status, xác minh |

*Lưu ý: Có thể thay đổi trong msp-config.json*

## 🔄 Workflow

```
CREATED → VERIFIED → EXPIRED
```

- **CREATED**: Lô trà mới được tạo
- **VERIFIED**: Đã xác minh hash, đảm bảo chính hãng
- **EXPIRED**: Hết hạn sử dụng

## 📊 Mô hình dữ liệu

```typescript
{
  batchId: "BATCH001",
  farmLocation: "Moc Chau, Son La",
  harvestDate: "2024-11-08",
  processingInfo: "Organic processing, no pesticides",
  qualityCert: "VN-ORGANIC-2024",
  hashValue: "abc123...",
  owner: "Org1MSP",
  timestamp: "2024-11-08T10:00:00.000Z",
  status: "VERIFIED"
}
```

## 📝 API Reference

### createBatch(batchId, farmLocation, harvestDate, processingInfo, qualityCert)
- **Quyền**: Farmer
- **Mô tả**: Tạo lô trà mới
- **Parameters**:
  - `batchId`: ID lô trà (unique, alphanumeric + dash/underscore, max 100 chars)
  - `farmLocation`: Vị trí nông trại (max 200 chars)
  - `harvestDate`: Ngày thu hoạch (YYYY-MM-DD format)
  - `processingInfo`: Thông tin xử lý (max 1000 chars)
  - `qualityCert`: Chứng chỉ chất lượng (max 100 chars)
- **Validation**: Tất cả parameters được validate trước khi xử lý

### verifyBatch(batchId, hashInput)
- **Quyền**: Farmer, Verifier, Admin (tất cả đều Org1MSP)
- **Mô tả**: Xác minh hash của lô trà
- **Parameters**:
  - `batchId`: ID lô trà
  - `hashInput`: Raw string để verify (format: `batchId|farmLocation|harvestDate|processingInfo|qualityCert`)
- **Returns**: `{isValid: boolean, batch: TeaBatch}`
- **Note**: `hashInput` sẽ được hash bằng SHA-256 trước khi so sánh với hash đã lưu

### getBatchInfo(batchId)
- **Quyền**: Public
- **Mô tả**: Xem thông tin lô trà
- **Parameters**:
  - `batchId`: ID lô trà
- **Returns**: `TeaBatch | null` (null nếu không tồn tại)

### updateBatchStatus(batchId, status)
- **Quyền**: Farmer, Admin
- **Mô tả**: Cập nhật trạng thái lô trà
- **Parameters**:
  - `batchId`: ID lô trà
  - `status`: Trạng thái mới (CREATED, VERIFIED, EXPIRED)
- **Returns**: `TeaBatch` với status đã cập nhật

### getAllBatches(limit?, offset?)
- **Quyền**: Public
- **Mô tả**: Query tất cả batches với pagination
- **Parameters**:
  - `limit`: Số lượng batches trả về (default: 100, max: 1000)
  - `offset`: Số batches bỏ qua (default: 0)
- **Returns**: `{batches: TeaBatch[], total: number}`

### getBatchesByStatus(status, limit?, offset?)
- **Quyền**: Public
- **Mô tả**: Query batches theo trạng thái
- **Parameters**:
  - `status`: Trạng thái (CREATED, VERIFIED, EXPIRED)
  - `limit`: Số lượng batches trả về (default: 100, max: 1000)
  - `offset`: Số batches bỏ qua (default: 0)
- **Returns**: `{batches: TeaBatch[], total: number}`

### getBatchesByOwner(owner, limit?, offset?)
- **Quyền**: Public
- **Mô tả**: Query batches theo owner (MSP ID)
- **Parameters**:
  - `owner`: MSP ID của owner
  - `limit`: Số lượng batches trả về (default: 100, max: 1000)
  - `offset`: Số batches bỏ qua (default: 0)
- **Returns**: `{batches: TeaBatch[], total: number}`

### getBatchHistory(batchId)
- **Quyền**: Public
- **Mô tả**: Lấy lịch sử thay đổi của batch (tất cả transactions)
- **Parameters**:
  - `batchId`: ID lô trà
- **Returns**: `TeaBatch[]` (array các version của batch, oldest first)

## 🔧 Ví dụ Sử dụng

### Tạo lô trà
```bash
peer chaincode invoke -C mychannel -n teaTraceCC \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
  -c '{"Args":["createBatch","BATCH001","Moc Chau, Son La","2024-11-08","Organic processing","VN-ORG-2024"]}'
```

### Xác minh lô trà
```bash
# hashInput format: batchId|farmLocation|harvestDate|processingInfo|qualityCert
# Chaincode sẽ hash input này và so sánh với hash đã lưu
peer chaincode invoke -C mychannel -n teaTraceCC \
  -c '{"Args":["verifyBatch","BATCH001","BATCH001|Moc Chau, Son La|2024-11-08|Organic processing|VN-ORG-2024"]}'
```

### Cập nhật trạng thái
```bash
peer chaincode invoke -C mychannel -n teaTraceCC \
  -c '{"Args":["updateBatchStatus","BATCH001","EXPIRED"]}'
```

### Query thông tin
```bash
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

### Query tất cả batches
```bash
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getAllBatches","50","0"]}'
```

### Query batches theo status
```bash
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getBatchesByStatus","VERIFIED","50","0"]}'
```

### Query batches theo owner
```bash
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getBatchesByOwner","Org1MSP","50","0"]}'
```

### Lấy lịch sử batch
```bash
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getBatchHistory","BATCH001"]}'
```

## ⚠️ Lưu ý

- **Hash verification** - Sử dụng SHA-256 để verify integrity
- **MSP authorization** - Kiểm tra quyền nghiêm ngặt theo role
- **Portable** - Dễ dàng deploy trên nhiều network khác nhau

## 🐛 Troubleshooting

### Lỗi: "MSP không có quyền"
```bash
# Giải pháp: Kiểm tra CORE_PEER_LOCALMSPID
export CORE_PEER_LOCALMSPID=Org1MSP
```

### Lỗi: "Batch already exists"
```bash
# Giải pháp: Dùng batchId khác hoặc query batch hiện tại
peer chaincode query -C mychannel -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

## 📦 Release Package (Sẵn sàng gửi đi)

### Files để gửi:
- `teaTraceCC-release.tar.gz` (52KB) - Full source code
- `teaTraceCC-release.tar.gz.sha256` - Checksum để verify

### Nội dung package:
- ✅ Source code đầy đủ (TypeScript)
- ✅ README.md (hướng dẫn đầy đủ 211 dòng)
- ✅ Config files (package.json, msp-config.json, tsconfig.json, .gitignore)
- ✅ Chaincode đã build sẵn (teaTraceCC.tar.gz - version 2.0)
- ❌ Không có node_modules (chạy `npm install` để cài)

### Người nhận sử dụng:
```bash
# 1. Verify checksum (optional)
sha256sum -c teaTraceCC-release.tar.gz.sha256

# 2. Giải nén
tar -xzf teaTraceCC-release.tar.gz
cd teaTraceCC

# 3. Đọc README.md

# 4. Chọn: Dùng teaTraceCC.tar.gz có sẵn HOẶC build lại
npm install
npm run build

# 5. Deploy theo hướng dẫn trong README
```

## 📄 License

Apache-2.0 | ICTU - Đại học Công nghệ thông tin và Truyền thông

