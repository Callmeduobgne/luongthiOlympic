# Phân Tích Kiến Trúc Network - IBN Blockchain Network

## 📋 Mục Lục

1. [Tổng Quan](#tổng-quan)
2. [Kiến Trúc Network](#kiến-trúc-network)
3. [Các Thành Phần Đã Triển Khai](#các-thành-phần-đã-triển-khai)
4. [Cấu Hình Chi Tiết](#cấu-hình-chi-tiết)
5. [Monitoring & Logging](#monitoring--logging)
6. [Security & Certificates](#security--certificates)
7. [Network Topology](#network-topology)
8. [Kết Nối API Gateway](#kết-nối-api-gateway)
9. [Tóm Tắt & Đánh Giá](#tóm-tắt--đánh-giá)

---

## 🎯 Tổng Quan

IBN Blockchain Network là một hệ thống Hyperledger Fabric 2.5.9 được thiết kế cho ứng dụng truy xuất nguồn gốc trà (Tea Traceability). Network được xây dựng với kiến trúc production-ready, tập trung vào tính khả dụng cao, khả năng mở rộng và bảo mật.

### Thông Số Kỹ Thuật

- **Hyperledger Fabric Version**: 2.5.9
- **Consensus Algorithm**: Raft (etcdraft)
- **Orderer Nodes**: 3 nodes (High Availability)
- **Peer Nodes**: 3 nodes (Org1)
- **State Database**: CouchDB (3 instances)
- **Channel**: ibnchannel
- **Chaincode**: teaTraceCC v1.0
- **Domain**: `.ibn.vn`

---

## 🏗️ Kiến Trúc Network

### Sơ Đồ Tổng Quan

```
┌─────────────────────────────────────────────────────────────────┐
│                    IBN BLOCKCHAIN NETWORK                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              ORDERER CLUSTER (Raft)                      │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │   │
│  │  │ orderer    │  │ orderer1   │  │ orderer2   │         │   │
│  │  │ :7050      │  │ :8050      │  │ :9050      │         │   │
│  │  │ :9443      │  │ :9447      │  │ :9448      │         │   │
│  │  └────────────┘  └────────────┘  └────────────┘         │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           │                                       │
│                           │ Consensus & Ordering                 │
│                           ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              PEER ORGANIZATION (Org1MSP)                  │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │   │
│  │  │ peer0      │  │ peer1      │  │ peer2      │         │   │
│  │  │ :7051      │  │ :8051      │  │ :9051      │         │   │
│  │  │ + CouchDB0 │  │ + CouchDB1 │  │ + CouchDB2 │         │   │
│  │  └────────────┘  └────────────┘  └────────────┘         │   │
│  └──────────────────────────────────────────────────────────┘   │
│                           │                                       │
│                           │ gRPC/TLS                              │
│                           ▼                                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              API GATEWAY LAYER                            │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │   │
│  │  │ Gateway 1  │  │ Gateway 2  │  │ Gateway 3  │         │   │
│  │  │ :8081      │  │ :8082      │  │ :8083      │         │   │
│  │  └────────────┘  └────────────┘  └────────────┘         │   │
│  │                           │                               │   │
│  │                           ▼                               │   │
│  │                    ┌────────────┐                         │   │
│  │                    │   Nginx    │                         │   │
│  │                    │  :8080    │                         │   │
│  │                    └────────────┘                         │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Network Layers

1. **Consensus Layer**: 3-node Raft cluster cho high availability
2. **Peer Layer**: 3 peers với CouchDB state database
3. **Application Layer**: API Gateway với load balancing
4. **Monitoring Layer**: Prometheus + Grafana + Loki

---

## 🔧 Các Thành Phần Đã Triển Khai

### 1. Orderer Cluster (Raft Consensus)

#### Cấu Hình

| Node | Container | Port | Admin Port | Status |
|------|-----------|------|------------|--------|
| Leader | orderer.ibn.vn | 7050 | 9443 | ✅ Running |
| Follower | orderer1.ibn.vn | 8050 | 9447 | ✅ Running |
| Follower | orderer2.ibn.vn | 9050 | 9448 | ✅ Running |

#### Đặc Điểm

- **Consensus**: Raft (etcdraft) - Byzantine Fault Tolerant
- **High Availability**: Có thể chịu được lỗi 1 node (3/2+1)
- **TLS Enabled**: Tất cả communication đều được mã hóa
- **Metrics**: Prometheus metrics endpoint tại `/metrics`
- **Bootstrap**: File-based genesis block

#### Cấu Hình Raft

```yaml
EtcdRaft:
  Consenters:
    - Host: orderer.ibn.vn
      Port: 7050
    - Host: orderer1.ibn.vn
      Port: 8050
    - Host: orderer2.ibn.vn
      Port: 9050
```

### 2. Peer Organization (Org1MSP)

#### Cấu Hình

| Peer | Container | Port | Operations | CouchDB | Status |
|------|-----------|------|------------|---------|--------|
| Anchor | peer0.org1.ibn.vn | 7051 | 9444 | couchdb0:5984 | ✅ Running |
| Peer1 | peer1.org1.ibn.vn | 8051 | 9445 | couchdb1:5984 | ✅ Running |
| Peer2 | peer2.org1.ibn.vn | 9051 | 9446 | couchdb2:5984 | ✅ Running |

#### Đặc Điểm

- **State Database**: CouchDB cho rich queries
- **TLS Enabled**: Tất cả peer-to-peer và peer-to-orderer communication
- **Metrics**: Prometheus metrics endpoint
- **Health Checks**: Tự động kiểm tra sức khỏe
- **Chaincode**: teaTraceCC v1.0 đã được deploy

### 3. Channel Configuration

#### Channel: `ibnchannel`

- **Type**: Application channel
- **Consortium**: SampleConsortium
- **Organizations**: Org1MSP
- **Capabilities**: V2_0 (Fabric 2.5)
- **Policies**: 
  - Readers: ANY Readers
  - Writers: ANY Writers
  - Admins: MAJORITY Admins
  - Endorsement: MAJORITY Endorsement

#### Anchor Peer

- **peer0.org1.ibn.vn:7051** - Anchor peer cho Org1MSP

### 4. Chaincode Deployment

#### Chaincode: `teaTraceCC`

- **Version**: 1.0
- **Language**: Node.js/TypeScript
- **Package ID**: `teaTraceCC_1.0:98cfde5435a0f97398b9a8e1fecc4c1374106133bcefba1f5122a20de6efae60`
- **Status**: ✅ Installed, Approved, Committed
- **Peers**: Deployed trên cả 3 peers

#### Functions

- `createBatch` - Tạo batch trà mới
- `getBatchInfo` - Lấy thông tin batch
- `verifyBatch` - Xác minh hash của batch
- `updateBatchStatus` - Cập nhật trạng thái batch

---

## ⚙️ Cấu Hình Chi Tiết

### Network Configuration

#### Docker Network

```yaml
networks:
  fabric-network:
    external: true  # External network để gateway có thể kết nối
```

**Lý do sử dụng external network**:
- Cho phép API Gateway containers kết nối trực tiếp với Fabric network
- Đảm bảo DNS resolution giữa gateway và peers/orderers
- Tách biệt network management giữa core và gateway

### Orderer Configuration

#### Raft Cluster Settings

```yaml
OrdererType: etcdraft
BatchTimeout: 2s
BatchSize:
  MaxMessageCount: 500
  AbsoluteMaxBytes: 10 MB
  PreferredMaxBytes: 2 MB
```

#### TLS Configuration

- **Client TLS**: Enabled
- **Server TLS**: Enabled
- **Cluster TLS**: Enabled (cho Raft communication)
- **Certificates**: Tự động generate bằng cryptogen

### Peer Configuration

#### CouchDB Integration

Mỗi peer có một CouchDB instance riêng:
- **couchdb0** → peer0.org1.ibn.vn
- **couchdb1** → peer1.org1.ibn.vn
- **couchdb2** → peer2.org1.ibn.vn

#### Port Mapping

| Service | Internal Port | External Port | Purpose |
|---------|--------------|---------------|---------|
| peer0 | 7051 | 7051 | gRPC endpoint |
| peer0 | 9444 | 9444 | Operations |
| peer1 | 8051 | 8051 | gRPC endpoint |
| peer1 | 9445 | 9445 | Operations |
| peer2 | 9051 | 9051 | gRPC endpoint |
| peer2 | 9446 | 9446 | Operations |

---

## 📊 Monitoring & Logging

### 1. Prometheus Monitoring

#### Configuration

- **Scrape Interval**: 15 seconds
- **Metrics Endpoints**:
  - Orderers: `/metrics` (Prometheus format)
  - Peers: `/metrics` (Prometheus format)

#### Scraped Targets

```yaml
- orderer.ibn.vn:9443/metrics
- orderer1.ibn.vn:9447/metrics
- orderer2.ibn.vn:9448/metrics
- peer0.org1.ibn.vn:9444/metrics
- peer1.org1.ibn.vn:9445/metrics
- peer2.org1.ibn.vn:9446/metrics
```

#### Metrics Collected

- **Orderer Metrics**:
  - Block processing time
  - Transaction throughput
  - Raft leader election
  - Cluster health

- **Peer Metrics**:
  - Endorsement latency
  - Commit latency
  - Chaincode execution time
  - CouchDB query performance

### 2. Grafana Dashboards

#### Pre-configured Dashboards

- **Fabric Network Overview**: Tổng quan network health
- **Orderer Cluster Status**: Raft cluster monitoring
- **Peer Performance**: Peer metrics và throughput
- **Channel Statistics**: Channel-level metrics

### 3. Loki Logging

#### Configuration

- **Log Aggregation**: Centralized logging từ tất cả containers
- **Storage**: Local filesystem
- **Retention**: Configurable

#### Log Sources

- Orderer logs
- Peer logs
- Chaincode logs
- CouchDB logs
- API Gateway logs

---

## 🔐 Security & Certificates

### Certificate Structure

```
organizations/
├── ordererOrganizations/
│   └── ibn.vn/
│       ├── msp/
│       └── orderers/
│           ├── orderer.ibn.vn/
│           ├── orderer1.ibn.vn/
│           └── orderer2.ibn.vn/
└── peerOrganizations/
    └── org1.ibn.vn/
        ├── msp/
        ├── users/
        │   └── Admin@org1.ibn.vn/
        └── peers/
            ├── peer0.org1.ibn.vn/
            ├── peer1.org1.ibn.vn/
            └── peer2.org1.ibn.vn/
```

### MSP (Membership Service Provider)

#### OrdererMSP
- **Domain**: ibn.vn
- **Role**: Ordering service
- **Certificates**: TLS và signing certificates

#### Org1MSP
- **Domain**: org1.ibn.vn
- **Role**: Peer organization
- **Users**: Admin@org1.ibn.vn (được dùng bởi API Gateway)
- **Node OUs**: Enabled (phân biệt peer, admin, client)

### TLS Configuration

- **TLS Enabled**: ✅ Tất cả communication
- **Certificate Validation**: Strict
- **mTLS**: Enabled cho peer-to-peer và peer-to-orderer

---

## 🌐 Network Topology

### Physical Topology

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Host                              │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │         fabric-network (External Bridge)             │  │
│  │                                                       │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │ Orderer  │  │ Orderer1 │  │ Orderer2 │          │  │
│  │  │ Cluster  │  │          │  │          │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  │       │              │              │                │  │
│  │       └──────────────┼──────────────┘                │  │
│  │                      │                                │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │  Peer0   │  │  Peer1   │  │  Peer2   │          │  │
│  │  │ +CouchDB0│  │ +CouchDB1│  │ +CouchDB2│          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  │       │              │              │                │  │
│  │       └──────────────┼──────────────┘                │  │
│  │                      │                                │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │  │
│  │  │ Gateway1 │  │ Gateway2 │  │ Gateway3 │          │  │
│  │  └──────────┘  └──────────┘  └──────────┘          │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Network Segmentation

- **fabric-network**: Core Fabric network (orderers, peers, CouchDB)
- **gateway-network**: API Gateway network (gateways, PostgreSQL, Redis, Nginx)

**Cross-network connectivity**: Gateway containers kết nối cả 2 networks để:
- Kết nối với peers (fabric-network)
- Kết nối với databases (gateway-network)

---

## 🔗 Kết Nối API Gateway

### Gateway-to-Fabric Integration

#### Connection Profile

```json
{
  "channel": "ibnchannel",
  "chaincode": "teaTraceCC",
  "msp": {
    "id": "Org1MSP",
    "userCert": "Admin@org1.ibn.vn-cert.pem",
    "userKey": "keystore/priv_sk"
  },
  "peers": [
    {
      "name": "peer0.org1.ibn.vn",
      "endpoint": "peer0.org1.ibn.vn:7051",
      "tlsCA": "peer0.org1.ibn.vn/tls/ca.crt"
    },
    {
      "name": "peer1.org1.ibn.vn",
      "endpoint": "peer1.org1.ibn.vn:8051",
      "tlsCA": "peer1.org1.ibn.vn/tls/ca.crt"
    },
    {
      "name": "peer2.org1.ibn.vn",
      "endpoint": "peer2.org1.ibn.vn:9051",
      "tlsCA": "peer2.org1.ibn.vn/tls/ca.crt"
    }
  ]
}
```

#### Gateway Instances

| Instance | Container | Port | Connected Peer | Status |
|----------|-----------|------|----------------|--------|
| Gateway 1 | api-gateway-1 | 8081 | peer0.org1.ibn.vn:7051 | ✅ Healthy |
| Gateway 2 | api-gateway-2 | 8082 | peer1.org1.ibn.vn:8051 | ✅ Healthy |
| Gateway 3 | api-gateway-3 | 8083 | peer2.org1.ibn.vn:9051 | ✅ Healthy |

#### Load Balancing

- **Nginx**: Round-robin load balancing
- **Port**: 8080 (external)
- **Backend**: 3 gateway instances
- **Health Checks**: Automatic failover

### Synchronization Status

✅ **Đã Đồng Bộ Hoàn Toàn**:
- Network connectivity: ✅
- DNS resolution: ✅
- Certificate paths: ✅
- Channel access: ✅
- Chaincode access: ✅
- Health checks: ✅

---

## 📈 Tóm Tắt & Đánh Giá

### Những Gì Đã Hoàn Thành

#### 1. Network Infrastructure ✅

- [x] Multi-orderer Raft cluster (3 nodes)
- [x] Multi-peer setup (3 peers)
- [x] CouchDB state database (3 instances)
- [x] External network configuration
- [x] TLS/SSL encryption
- [x] Health checks và monitoring

#### 2. Channel & Chaincode ✅

- [x] Channel `ibnchannel` created
- [x] All peers joined channel
- [x] Anchor peer configured
- [x] Chaincode `teaTraceCC` packaged
- [x] Chaincode installed on all peers
- [x] Chaincode approved và committed

#### 3. Monitoring & Observability ✅

- [x] Prometheus metrics collection
- [x] Grafana dashboards
- [x] Loki log aggregation
- [x] Promtail log collection
- [x] Health check endpoints

#### 4. API Gateway Integration ✅

- [x] Gateway-to-Fabric connectivity
- [x] Certificate mounting
- [x] Network synchronization
- [x] Load balancing
- [x] Health monitoring

### Điểm Mạnh

1. **High Availability**
   - Raft cluster có thể chịu được 1 node failure
   - 3 peers đảm bảo redundancy
   - Load balancing với 3 gateway instances

2. **Security**
   - TLS/SSL cho tất cả communication
   - Certificate-based authentication
   - MSP-based authorization

3. **Scalability**
   - Có thể thêm peers/organizations
   - Horizontal scaling với gateway instances
   - CouchDB cho rich queries

4. **Observability**
   - Comprehensive monitoring
   - Centralized logging
   - Health check automation

### Cải Tiến Đề Xuất

1. **Network Resilience**
   - Thêm orderer nodes (5 nodes cho better fault tolerance)
   - Implement network partitioning tests
   - Disaster recovery procedures

2. **Performance Optimization**
   - Connection pooling optimization
   - Caching strategies
   - Query optimization

3. **Security Hardening**
   - Certificate rotation automation
   - Audit logging
   - Access control policies

4. **Documentation**
   - Operational runbooks
   - Troubleshooting guides
   - Performance tuning guides

---

## 📝 File Cấu Hình Quan Trọng

### Core Network

- `core/docker/docker-compose.yml` - Docker Compose cho Fabric network
- `core/configtx/configtx.yaml` - Channel và organization configuration
- `core/config/orderer.yaml` - Orderer configuration
- `core/config/core.yaml` - Peer configuration
- `core/crypto-config.yaml` - Certificate generation config

### Monitoring

- `monitoring/prometheus.yml` - Prometheus scrape configuration
- `monitoring/docker-compose-monitoring.yml` - Monitoring stack
- `monitoring/grafana/provisioning/` - Grafana dashboards

### Logging

- `logging/docker-compose-logging.yml` - Logging stack
- `logging/loki-config.yml` - Loki configuration
- `logging/promtail-config.yml` - Promtail configuration

### API Gateway

- `api-gateway/docker/docker-compose.yml` - Gateway deployment
- `api-gateway/internal/config/` - Gateway configuration
- `api-gateway/internal/services/fabric/` - Fabric integration

---

## 🚀 Deployment Status

### Current Status

| Component | Status | Health |
|-----------|--------|--------|
| Orderer Cluster | ✅ Running | Healthy (3/3) |
| Peer Network | ✅ Running | Healthy (3/3) |
| CouchDB | ✅ Running | Healthy (3/3) |
| API Gateway | ✅ Running | Healthy (3/3) |
| Monitoring | ✅ Running | Active |
| Logging | ✅ Running | Active |

### Network Health

```bash
# Kiểm tra network health
docker ps --filter "network=fabric-network" --format "{{.Names}}: {{.Status}}"

# Kết quả:
# orderer.ibn.vn: Up X minutes (healthy)
# orderer1.ibn.vn: Up X minutes (healthy)
# orderer2.ibn.vn: Up X minutes (healthy)
# peer0.org1.ibn.vn: Up X minutes (healthy)
# peer1.org1.ibn.vn: Up X minutes (healthy)
# peer2.org1.ibn.vn: Up X minutes (healthy)
```

---

## 📚 Tài Liệu Tham Khảo

- [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)
- [Fabric Gateway SDK](https://github.com/hyperledger/fabric-gateway)
- [Raft Consensus](https://raft.github.io/)
- [Prometheus Monitoring](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/docs/)

---

**Tài liệu được tạo tự động từ cấu hình hiện tại của IBN Blockchain Network**

*Last Updated: 2025-11-11*

