# Core Blockchain Layer - Tổng Hợp

**Ngày tạo:** 2025-11-12  
**Version:** 1.1.0  
**Last Updated:** 2025-11-24  
**Layer:** Core/Blockchain (Hyperledger Fabric Network)

---

## 📋 Tổng Quan

Tài liệu này tổng hợp tất cả thông tin về **Core Blockchain Layer** của hệ thống IBN Network, bao gồm:
- Kiến trúc network
- Chaincode operations
- Network configuration
- Commands và utilities

---

## 📚 Tài Liệu Tham Khảo

### 1. Network Architecture
**File:** `network-architecture-analysis.md` (609 dòng)

**Nội dung:**
- Kiến trúc Hyperledger Fabric Network
- Topology và các thành phần
- Orderer cluster (Raft consensus)
- Peer nodes và CouchDB
- Security & Certificates
- Monitoring & Logging
- Kết nối API Gateway

**Thông số kỹ thuật:**
- Hyperledger Fabric Version: 2.5.9
- Consensus Algorithm: Raft (etcdraft)
- Orderer Nodes: 3 nodes (High Availability)
- Peer Nodes: 3 nodes (Org1)
- State Database: CouchDB (3 instances)
- Channel: ibnchannel
- Chaincode: teaTraceCC v1.0.0
- Domain: `.ibn.vn`

### 2. Chaincode Commands
**File:** `chaincode-commands.md` (125 dòng)

**Nội dung:**
- Commands để query/invoke chaincode
- Helper scripts
- Prerequisites và setup
- Troubleshooting

**Các commands chính:**
- Query batch
- Create batch
- Verify batch
- Update status
- Health check

### 3. Chaincode Documentation
**File:** `tea_1.0.md` (588 dòng)

**Nội dung:**
- Hướng dẫn sử dụng chaincode teaTraceCC
- Cấu trúc dữ liệu TeaBatch
- Query operations
- Invoke transactions
- Ví dụ sử dụng
- Troubleshooting

**Chaincode Info:**
- Name: teaTraceCC
- Version: 1.0.0
- Sequence: 2
- Channel: ibnchannel
- Language: Node.js (TypeScript)

---

## 🏗️ Kiến Trúc Network

### Topology

```
┌─────────────────────────────────────────────────────────┐
│              IBN BLOCKCHAIN NETWORK                     │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │         ORDERER CLUSTER (Raft)                   │   │
│  │  orderer:7050  orderer1:8050  orderer2:9050      │   │
│  └──────────────────────────────────────────────────┘   │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │         PEER NODES (Org1)                        │   │
│  │  peer0:7051  peer1:8051  peer2:9051              │   │
│  │  + CouchDB instances                             │   │
│  └──────────────────────────────────────────────────┘   │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │         FABRIC CA                                │   │
│  │  ca.org1.ibn.vn:7054                             │   │
│  └──────────────────────────────────────────────────┘   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```
### Các Thành Phần

1. **Orderer Cluster**
   - 3 orderer nodes với Raft consensus
   - High availability
   - TLS enabled

2. **Peer Nodes**
   - 3 peer nodes (Org1MSP)
   - Mỗi peer có CouchDB riêng
   - Endorsement và commit

3. **Fabric CA**
   - Certificate Authority
   - User enrollment/registration
   - Certificate management

4. **Channel**
   - ibnchannel
   - Chaincode: teaTraceCC v1.0.0

---

## 🔧 Chaincode Operations

### Query Operations

```bash
# Query batch by ID
peer chaincode query -C ibnchannel -n teaTraceCC \
  -c '{"Args":["getBatchInfo","BATCH001"]}'
```

### Invoke Operations

```bash
# Create batch
peer chaincode invoke -C ibnchannel -n teaTraceCC \
  -c '{"Args":["createBatch","BATCH001","Moc Chau, Son La","2024-11-12","Organic processing","VN-ORG-2024"]}'

# Verify batch
peer chaincode invoke -C ibnchannel -n teaTraceCC \
  -c '{"Args":["verifyBatch","BATCH001","hash_input_string"]}'

# Update status
peer chaincode invoke -C ibnchannel -n teaTraceCC \
  -c '{"Args":["updateBatchStatus","BATCH001","VERIFIED"]}'
```

### Helper Scripts

Sử dụng helper script để dễ dàng hơn:
```bash
./scripts/chaincode-helper.sh query getBatchInfo BATCH001
./scripts/chaincode-helper.sh create TEST002 "Moc Chau" "2024-11-12" "Organic processing" "VN-ORG-2024"
./scripts/chaincode-helper.sh verify TEST001 "hash_input_string"
./scripts/chaincode-helper.sh status TEST001 VERIFIED
```

---

## 📦 Chaincode Functions

### teaTraceCC v1.0.0

**Query Functions:**
- `getBatchInfo(batchId)` - Get batch information by ID (Public access)

**Invoke Functions:**
- `createBatch(batchId, farmLocation, harvestDate, processingInfo, qualityCert)` - Create new batch (Farmer role required)
- `verifyBatch(batchId, hashInput)` - Verify batch hash (Farmer, Verifier, Admin roles)
- `updateBatchStatus(batchId, status)` - Update batch status (Farmer, Admin roles)

**Status Values:**
- `CREATED` - Batch mới được tạo
- `VERIFIED` - Batch đã được xác minh hash
- `EXPIRED` - Batch đã hết hạn

### Data Model: TeaBatch

```json
{
  "batchId": "BATCH001",
  "farmLocation": "Moc Chau, Son La",
  "harvestDate": "2024-11-12",
  "processingInfo": "Organic processing, no pesticides",
  "qualityCert": "VN-ORGANIC-2024",
  "hashValue": "abc123...",
  "owner": "Org1MSP",
  "timestamp": "2024-11-12T10:00:00.000Z",
  "status": "CREATED"
}
```

**Field Descriptions:**
- `batchId` - Unique identifier cho batch
- `farmLocation` - Vị trí nông trại (thay vì `farmName`)
- `harvestDate` - Ngày thu hoạch (YYYY-MM-DD)
- `processingInfo` - Thông tin xử lý (thay vì `certification`)
- `qualityCert` - Chứng chỉ chất lượng (thay vì `certificateId`)
- `hashValue` - SHA-256 hash để verify integrity (thay vì `verificationHash`)
- `owner` - MSP ID của owner
- `timestamp` - ISO 8601 timestamp (thay vì `createdAt`/`updatedAt`)
- `status` - Trạng thái: CREATED, VERIFIED, EXPIRED

---

## 🔐 Security & Certificates

### Certificate Structure

```
core/organizations/
├── peerOrganizations/
│   └── org1.ibn.vn/
│       ├── msp/
│       └── users/
│           └── Admin@org1.ibn.vn/
│               └── msp/
└── ordererOrganizations/
    └── ibn.vn/
        └── msp/
```

### MSP Configuration

- **Org1MSP** - Organization 1
- **OrdererMSP** - Orderer organization
- TLS certificates cho tất cả components

---

## 📊 Monitoring & Logging

### Health Checks

- Orderer health: `http://localhost:7050/healthz`
- Peer health: `http://localhost:7051/healthz`
- CouchDB health: `http://localhost:5984/_up`

### Logs

```bash
# Orderer logs
docker logs orderer.ibn.vn

# Peer logs
docker logs peer0.org1.ibn.vn

# CouchDB logs
docker logs couchdb0
```

---

## 🔗 Kết Nối API Gateway

API Gateway kết nối với Core Blockchain Layer thông qua:
- **Fabric Gateway SDK** - Go client
- **Connection Profile** - Network configuration
- **Certificates** - TLS và MSP certificates
- **Channel** - ibnchannel
- **Chaincode** - teaTraceCC

---

## 📝 Tóm Tắt

### Thành Phần Chính
- ✅ 3 Orderer nodes (Raft)
- ✅ 3 Peer nodes (Org1)
- ✅ 3 CouchDB instances
- ✅ 1 Fabric CA
- ✅ 1 Channel (ibnchannel)
- ✅ 1 Chaincode (teaTraceCC v1.0.0)

### Tính Năng
- ✅ High availability
- ✅ TLS encryption
- ✅ Certificate management
- ✅ Health monitoring
- ✅ Production-ready

### Tài Liệu
- `network-architecture-analysis.md` - Kiến trúc chi tiết
- `chaincode-commands.md` - Commands reference
- `tea_1.0.md` - Chaincode documentation

---

**Last Updated:** 2025-01-27

