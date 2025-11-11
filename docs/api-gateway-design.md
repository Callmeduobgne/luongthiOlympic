# API Gateway Design - Tổng Hợp Dữ Liệu Thiết Kế

> Tài liệu tổng hợp đầy đủ thông tin cần thiết để thiết kế API Gateway cho Hyperledger Fabric Network IBN

**Ngày tạo:** 2025-11-11  
**Network:** IBN (ibn.vn)  
**Chaincode:** teaTraceCC v1.0  
**Fabric Version:** 2.5

---

## 📋 Mục Lục

1. [Network Topology](#network-topology)
2. [Connection Profile](#connection-profile)
3. [Chaincode Functions](#chaincode-functions)
4. [Data Models](#data-models)
5. [MSP & Authorization](#msp--authorization)
6. [Certificate Paths](#certificate-paths)
7. [Docker Compose Configuration](#docker-compose-configuration)
8. [API Endpoints Design](#api-endpoints-design)
9. [Error Handling](#error-handling)
10. [Security Configuration](#security-configuration)

---

## Network Topology

### Peers

| Peer Name | Container | Host Port | Internal Port | Chaincode Port | MSP ID |
|-----------|-----------|-----------|---------------|----------------|--------|
| peer0.org1.ibn.vn | peer0.org1.ibn.vn | 7051 | 7051 | 7052 | Org1MSP |
| peer1.org1.ibn.vn | peer1.org1.ibn.vn | 8051 | 8051 | 8052 | Org1MSP |
| peer2.org1.ibn.vn | peer2.org1.ibn.vn | 9051 | 9051 | 9052 | Org1MSP |

### Orderers

| Orderer Name | Container | Host Port | Internal Port | Admin Port | MSP ID |
|--------------|-----------|-----------|---------------|------------|--------|
| orderer.ibn.vn | orderer.ibn.vn | 7050 | 7050 | 9443 | OrdererMSP |
| orderer1.ibn.vn | orderer1.ibn.vn | 8050 | 8050 | 9447 | OrdererMSP |
| orderer2.ibn.vn | orderer2.ibn.vn | 9050 | 9050 | 9448 | OrdererMSP |

### CouchDB Instances

| CouchDB | Container | Port | Username | Password |
|---------|-----------|------|----------|----------|
| couchdb0 | couchdb0 | 5984 | admin | adminpw |
| couchdb1 | couchdb1 | 5984 | admin | adminpw |
| couchdb2 | couchdb2 | 5984 | admin | adminpw |

### Network Information

- **Network Name:** `fabric-network`
- **Network Driver:** `bridge`
- **Channel Name:** `ibnchannel`
- **Chaincode Name:** `teaTraceCC`
- **Chaincode Version:** `1.0`
- **Chaincode Sequence:** `2`
- **Package ID:** `teaTraceCC_1.0:98cfde5435a0f97398b9a8e1fecc4c1374106133bcefba1f5122a20de6efae60`

---

## Connection Profile

### Gateway Connection Configuration

```json
{
  "name": "ibn-network",
  "version": "1.0.0",
  "client": {
    "organization": "Org1",
    "connection": {
      "timeout": {
        "peer": {
          "endorser": "300"
        }
      }
    }
  },
  "channels": {
    "ibnchannel": {
      "orderers": [
        "orderer.ibn.vn"
      ],
      "peers": {
        "peer0.org1.ibn.vn": {
          "endorsingPeer": true,
          "chaincodeQuery": true,
          "ledgerQuery": true,
          "eventSource": true
        },
        "peer1.org1.ibn.vn": {
          "endorsingPeer": true,
          "chaincodeQuery": true,
          "ledgerQuery": true,
          "eventSource": true
        },
        "peer2.org1.ibn.vn": {
          "endorsingPeer": true,
          "chaincodeQuery": true,
          "ledgerQuery": true,
          "eventSource": true
        }
      }
    }
  },
  "organizations": {
    "Org1": {
      "mspid": "Org1MSP",
      "peers": [
        "peer0.org1.ibn.vn",
        "peer1.org1.ibn.vn",
        "peer2.org1.ibn.vn"
      ],
      "certificateAuthorities": [
        "ca.org1.ibn.vn"
      ]
    }
  },
  "orderers": {
    "orderer.ibn.vn": {
      "url": "grpcs://localhost:7050",
      "tlsCACerts": {
        "path": "core/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem"
      },
      "grpcOptions": {
        "ssl-target-name-override": "orderer.ibn.vn",
        "hostnameOverride": "orderer.ibn.vn"
      }
    },
    "orderer1.ibn.vn": {
      "url": "grpcs://localhost:8050",
      "tlsCACerts": {
        "path": "core/organizations/ordererOrganizations/ibn.vn/orderers/orderer1.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem"
      },
      "grpcOptions": {
        "ssl-target-name-override": "orderer1.ibn.vn",
        "hostnameOverride": "orderer1.ibn.vn"
      }
    },
    "orderer2.ibn.vn": {
      "url": "grpcs://localhost:9050",
      "tlsCACerts": {
        "path": "core/organizations/ordererOrganizations/ibn.vn/orderers/orderer2.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem"
      },
      "grpcOptions": {
        "ssl-target-name-override": "orderer2.ibn.vn",
        "hostnameOverride": "orderer2.ibn.vn"
      }
    }
  },
  "peers": {
    "peer0.org1.ibn.vn": {
      "url": "grpcs://localhost:7051",
      "tlsCACerts": {
        "path": "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt"
      },
      "grpcOptions": {
        "ssl-target-name-override": "peer0.org1.ibn.vn",
        "hostnameOverride": "peer0.org1.ibn.vn"
      }
    },
    "peer1.org1.ibn.vn": {
      "url": "grpcs://localhost:8051",
      "tlsCACerts": {
        "path": "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt"
      },
      "grpcOptions": {
        "ssl-target-name-override": "peer1.org1.ibn.vn",
        "hostnameOverride": "peer1.org1.ibn.vn"
      }
    },
    "peer2.org1.ibn.vn": {
      "url": "grpcs://localhost:9051",
      "tlsCACerts": {
        "path": "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt"
      },
      "grpcOptions": {
        "ssl-target-name-override": "peer2.org1.ibn.vn",
        "hostnameOverride": "peer2.org1.ibn.vn"
      }
    }
  }
}
```

### Gateway Client Configuration

```javascript
// Gateway connection settings
const gatewayConfig = {
  channel: "ibnchannel",
  chaincode: "teaTraceCC",
  mspId: "Org1MSP",
  userCertPath: "core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/signcerts/cert.pem",
  userKeyPath: "core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/keystore",
  peerEndpoints: [
    {
      name: "peer0.org1.ibn.vn",
      url: "grpcs://localhost:7051",
      tlsCACert: "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt"
    },
    {
      name: "peer1.org1.ibn.vn",
      url: "grpcs://localhost:8051",
      tlsCACert: "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt"
    },
    {
      name: "peer2.org1.ibn.vn",
      url: "grpcs://localhost:9051",
      tlsCACert: "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt"
    }
  ],
  ordererEndpoints: [
    {
      name: "orderer.ibn.vn",
      url: "grpcs://localhost:7050",
      tlsCACert: "core/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem"
    }
  ]
};
```

---

## Chaincode Functions

### 1. createBatch

**Function Signature:**
```typescript
public async createBatch(
  ctx: Context,
  batchId: string,
  farmLocation: string,
  harvestDate: string,
  processingInfo: string,
  qualityCert: string
): Promise<TeaBatch>
```

**Authorization:** Farmer (Org1MSP)

**Parameters:**
- `batchId` (string): ID duy nhất của lô trà
- `farmLocation` (string): Vị trí nông trại
- `harvestDate` (string): Ngày thu hoạch (YYYY-MM-DD)
- `processingInfo` (string): Thông tin xử lý
- `qualityCert` (string): Chứng chỉ chất lượng

**Returns:** `TeaBatch` object

**Behavior:**
- Tạo lô trà mới với status "CREATED"
- Tự động generate hashValue (SHA-256)
- Set owner từ MSP ID của caller
- Set timestamp từ transaction timestamp

**Error Cases:**
- Batch ID đã tồn tại
- MSP không có quyền (không phải Org1MSP)

---

### 2. verifyBatch

**Function Signature:**
```typescript
public async verifyBatch(
  ctx: Context,
  batchId: string,
  hashInput: string
): Promise<{ isValid: boolean; batch: TeaBatch }>
```

**Authorization:** Farmer, Verifier, Admin (Org1MSP, Org2MSP, Org3MSP)

**Parameters:**
- `batchId` (string): ID của lô trà cần xác minh
- `hashInput` (string): Chuỗi input để verify hash

**Returns:** 
```typescript
{
  isValid: boolean;
  batch: TeaBatch;
}
```

**Behavior:**
- Verify hash của batch với hashInput
- Nếu valid và status chưa là "VERIFIED", tự động update status thành "VERIFIED"
- Return kết quả verification và batch object

**Error Cases:**
- Batch ID không tồn tại
- MSP không có quyền

---

### 3. getBatchInfo

**Function Signature:**
```typescript
public async getBatchInfo(
  ctx: Context,
  batchId: string
): Promise<TeaBatch>
```

**Authorization:** Public (không cần kiểm tra MSP)

**Parameters:**
- `batchId` (string): ID của lô trà

**Returns:** `TeaBatch` object

**Behavior:**
- Query thông tin lô trà từ ledger
- Không thay đổi state (read-only)

**Error Cases:**
- Batch ID không tồn tại

---

### 4. updateBatchStatus

**Function Signature:**
```typescript
public async updateBatchStatus(
  ctx: Context,
  batchId: string,
  status: string
): Promise<TeaBatch>
```

**Authorization:** Farmer, Admin (Org1MSP, Org3MSP)

**Parameters:**
- `batchId` (string): ID của lô trà
- `status` (string): Trạng thái mới (CREATED, VERIFIED, EXPIRED)

**Returns:** `TeaBatch` object

**Behavior:**
- Update status của batch
- Update timestamp
- Normalize status to uppercase

**Error Cases:**
- Batch ID không tồn tại
- Status không hợp lệ (không phải CREATED, VERIFIED, EXPIRED)
- MSP không có quyền

---

## Data Models

### TeaBatch Interface

```typescript
export interface TeaBatch {
  batchId: string;
  farmLocation: string;
  harvestDate: string;        // Format: YYYY-MM-DD
  processingInfo: string;
  qualityCert: string;
  hashValue: string;          // SHA-256 hash
  owner: string;              // MSP ID
  timestamp: string;           // ISO 8601 format
  status: TeaBatchStatus;     // "CREATED" | "VERIFIED" | "EXPIRED"
}

export type TeaBatchStatus = "CREATED" | "VERIFIED" | "EXPIRED";
```

### CreateTeaBatchInput Interface

```typescript
export interface CreateTeaBatchInput {
  batchId: string;
  farmLocation: string;
  harvestDate: string;
  processingInfo: string;
  qualityCert: string;
}
```

### VerifyBatchResponse Interface

```typescript
export interface VerifyBatchResponse {
  isValid: boolean;
  batch: TeaBatch;
}
```

### Hash Generation

**Hash Payload Format:**
```
batchId|farmLocation|harvestDate|processingInfo|qualityCert
```

**Hash Algorithm:** SHA-256

**Example:**
```
Input: "BATCH001|Moc Chau|2024-11-08|Organic|VN-ORG-2024"
Hash: SHA-256("BATCH001|Moc Chau|2024-11-08|Organic|VN-ORG-2024")
```

---

## MSP & Authorization

### MSP Configuration

**File:** `teaTraceCC/msp-config.json`

```json
{
  "mspRoles": {
    "farmer": {
      "mspId": "Org1MSP",
      "description": "Farmer - create tea batches"
    },
    "verifier": {
      "mspId": "Org2MSP",
      "description": "Verifier - verify tea batches"
    },
    "admin": {
      "mspId": "Org3MSP",
      "description": "Admin - manage batch status"
    }
  }
}
```

### Function Authorization Matrix

| Function | Farmer (Org1MSP) | Verifier (Org2MSP) | Admin (Org3MSP) | Public |
|----------|------------------|-------------------|-----------------|--------|
| createBatch | ✅ | ❌ | ❌ | ❌ |
| verifyBatch | ✅ | ✅ | ✅ | ❌ |
| getBatchInfo | ✅ | ✅ | ✅ | ✅ |
| updateBatchStatus | ✅ | ❌ | ✅ | ❌ |

### Authorization Implementation

Gateway cần kiểm tra MSP ID của user trước khi gọi chaincode:

```javascript
function checkAuthorization(mspId, functionName) {
  const authMatrix = {
    'createBatch': ['Org1MSP'],
    'verifyBatch': ['Org1MSP', 'Org2MSP', 'Org3MSP'],
    'getBatchInfo': ['*'], // Public
    'updateBatchStatus': ['Org1MSP', 'Org3MSP']
  };
  
  const allowedMsps = authMatrix[functionName];
  if (allowedMsps.includes('*')) return true;
  return allowedMsps.includes(mspId);
}
```

---

## Certificate Paths

### User Identity (Admin@org1.ibn.vn)

**Base Path:** `core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/`

**Files:**
- **Certificate:** `signcerts/cert.pem`
- **Private Key:** `keystore/` (file name pattern: `*_sk`)
- **CA Certificate:** `cacerts/ca.org1.ibn.vn-cert.pem`

### Peer TLS Certificates

**peer0.org1.ibn.vn:**
- **TLS CA:** `core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt`
- **TLS Server Cert:** `core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/server.crt`
- **TLS Server Key:** `core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/server.key`

**peer1.org1.ibn.vn:**
- **TLS CA:** `core/organizations/peerOrganizations/org1.ibn.vn/peers/peer1.org1.ibn.vn/tls/ca.crt`

**peer2.org1.ibn.vn:**
- **TLS CA:** `core/organizations/peerOrganizations/org1.ibn.vn/peers/peer2.org1.ibn.vn/tls/ca.crt`

### Orderer TLS Certificates

**orderer.ibn.vn:**
- **TLS CA:** `core/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem`

**orderer1.ibn.vn:**
- **TLS CA:** `core/organizations/ordererOrganizations/ibn.vn/orderers/orderer1.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem`

**orderer2.ibn.vn:**
- **TLS CA:** `core/organizations/ordererOrganizations/ibn.vn/orderers/orderer2.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem`

### Certificate Loading Example

```javascript
const fs = require('fs');
const path = require('path');

// Load user certificate
const userCertPath = path.join(
  __dirname,
  '../core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/signcerts/cert.pem'
);
const userCert = fs.readFileSync(userCertPath).toString();

// Load user private key
const keystorePath = path.join(
  __dirname,
  '../core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/keystore'
);
const keyFiles = fs.readdirSync(keystorePath).filter(f => f.endsWith('_sk'));
const userKeyPath = path.join(keystorePath, keyFiles[0]);
const userKey = fs.readFileSync(userKeyPath).toString();

// Load peer TLS CA
const peerTlsCaPath = path.join(
  __dirname,
  '../core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt'
);
const peerTlsCa = fs.readFileSync(peerTlsCaPath).toString();
```

---

## Docker Compose Configuration

### Network Configuration

```yaml
networks:
  fabric-network:
    driver: bridge
```

### Peer Service Configuration

#### peer0.org1.ibn.vn

```yaml
peer0.org1.ibn.vn:
  container_name: peer0.org1.ibn.vn
  image: hyperledger/fabric-peer:2.5
  environment:
    - CORE_PEER_ID=peer0.org1.ibn.vn
    - CORE_PEER_ADDRESS=peer0.org1.ibn.vn:7051
    - CORE_PEER_LISTENADDRESS=0.0.0.0:7051
    - CORE_PEER_CHAINCODEADDRESS=peer0.org1.ibn.vn:7052
    - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:7052
    - CORE_PEER_LOCALMSPID=Org1MSP
    - CORE_PEER_TLS_ENABLED=true
    - CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt
    - CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key
    - CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt
    - CORE_LEDGER_STATE_STATEDATABASE=CouchDB
    - CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=couchdb0:5984
    - CORE_LEDGER_STATE_COUCHDBCONFIG_USERNAME=admin
    - CORE_LEDGER_STATE_COUCHDBCONFIG_PASSWORD=adminpw
  ports:
    - "7051:7051"   # Peer endpoint
    - "7052:7052"   # Chaincode endpoint
    - "9446:9443"   # Operations endpoint
  volumes:
    - ../organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/msp:/etc/hyperledger/fabric/msp
    - ../organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/:/etc/hyperledger/fabric/tls
  networks:
    - fabric-network
```

#### peer1.org1.ibn.vn

```yaml
peer1.org1.ibn.vn:
  container_name: peer1.org1.ibn.vn
  image: hyperledger/fabric-peer:2.5
  environment:
    - CORE_PEER_ID=peer1.org1.ibn.vn
    - CORE_PEER_ADDRESS=peer1.org1.ibn.vn:8051
    - CORE_PEER_LISTENADDRESS=0.0.0.0:8051
    - CORE_PEER_CHAINCODEADDRESS=peer1.org1.ibn.vn:8052
    - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:8052
    - CORE_PEER_LOCALMSPID=Org1MSP
    - CORE_PEER_TLS_ENABLED=true
  ports:
    - "8051:8051"
    - "8052:8052"
    - "9444:9444"
  networks:
    - fabric-network
```

#### peer2.org1.ibn.vn

```yaml
peer2.org1.ibn.vn:
  container_name: peer2.org1.ibn.vn
  image: hyperledger/fabric-peer:2.5
  environment:
    - CORE_PEER_ID=peer2.org1.ibn.vn
    - CORE_PEER_ADDRESS=peer2.org1.ibn.vn:9051
    - CORE_PEER_LISTENADDRESS=0.0.0.0:9051
    - CORE_PEER_CHAINCODEADDRESS=peer2.org1.ibn.vn:9052
    - CORE_PEER_CHAINCODELISTENADDRESS=0.0.0.0:9052
    - CORE_PEER_LOCALMSPID=Org1MSP
    - CORE_PEER_TLS_ENABLED=true
  ports:
    - "9051:9051"
    - "9052:9052"
    - "9445:9445"
  networks:
    - fabric-network
```

### Orderer Service Configuration

#### orderer.ibn.vn

```yaml
orderer.ibn.vn:
  container_name: orderer.ibn.vn
  image: hyperledger/fabric-orderer:2.5
  environment:
    - ORDERER_GENERAL_LISTENADDRESS=0.0.0.0
    - ORDERER_GENERAL_LISTENPORT=7050
    - ORDERER_GENERAL_LOCALMSPID=OrdererMSP
    - ORDERER_GENERAL_TLS_ENABLED=true
    - ORDERER_CHANNELPARTICIPATION_ENABLED=true
    - ORDERER_ADMIN_LISTENADDRESS=0.0.0.0:9443
  ports:
    - "7050:7050"   # Orderer endpoint
    - "9443:9443"   # Admin endpoint
  networks:
    - fabric-network
```

### CouchDB Service Configuration

```yaml
couchdb0:
  container_name: couchdb0
  image: couchdb:3.3
  environment:
    - COUCHDB_USER=admin
    - COUCHDB_PASSWORD=adminpw
  ports:
    - "5984:5984"
  networks:
    - fabric-network
```

---

## API Endpoints Design

### Base URL

```
http://localhost:3000/api/v1
```

### Endpoints

#### 1. Create Tea Batch

**POST** `/batches`

**Authorization:** Required (Org1MSP)

**Request Body:**
```json
{
  "batchId": "BATCH001",
  "farmLocation": "Moc Chau, Son La",
  "harvestDate": "2024-11-08",
  "processingInfo": "Organic processing, no pesticides",
  "qualityCert": "VN-ORG-2024"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "batchId": "BATCH001",
    "farmLocation": "Moc Chau, Son La",
    "harvestDate": "2024-11-08",
    "processingInfo": "Organic processing, no pesticides",
    "qualityCert": "VN-ORG-2024",
    "hashValue": "a1b2c3d4e5f6...",
    "owner": "Org1MSP",
    "timestamp": "2024-11-08T10:00:00.000Z",
    "status": "CREATED"
  },
  "transactionId": "abc123..."
}
```

**Error Response (400 Bad Request):**
```json
{
  "success": false,
  "error": {
    "code": "BATCH_EXISTS",
    "message": "Batch with id 'BATCH001' already exists."
  }
}
```

---

#### 2. Get Batch Info

**GET** `/batches/:batchId`

**Authorization:** Optional (Public endpoint)

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "batchId": "BATCH001",
    "farmLocation": "Moc Chau, Son La",
    "harvestDate": "2024-11-08",
    "processingInfo": "Organic processing, no pesticides",
    "qualityCert": "VN-ORG-2024",
    "hashValue": "a1b2c3d4e5f6...",
    "owner": "Org1MSP",
    "timestamp": "2024-11-08T10:00:00.000Z",
    "status": "CREATED"
  }
}
```

**Error Response (404 Not Found):**
```json
{
  "success": false,
  "error": {
    "code": "BATCH_NOT_FOUND",
    "message": "Batch with id 'BATCH001' does not exist."
  }
}
```

---

#### 3. Verify Batch

**POST** `/batches/:batchId/verify`

**Authorization:** Required (Org1MSP, Org2MSP, Org3MSP)

**Request Body:**
```json
{
  "hashInput": "BATCH001|Moc Chau, Son La|2024-11-08|Organic processing|VN-ORG-2024"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "isValid": true,
    "batch": {
      "batchId": "BATCH001",
      "status": "VERIFIED",
      ...
    }
  },
  "transactionId": "def456..."
}
```

**Error Response (400 Bad Request):**
```json
{
  "success": false,
  "error": {
    "code": "VERIFICATION_FAILED",
    "message": "Hash verification failed."
  }
}
```

---

#### 4. Update Batch Status

**PATCH** `/batches/:batchId/status`

**Authorization:** Required (Org1MSP, Org3MSP)

**Request Body:**
```json
{
  "status": "VERIFIED"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "batchId": "BATCH001",
    "status": "VERIFIED",
    "timestamp": "2024-11-08T12:00:00.000Z",
    ...
  },
  "transactionId": "ghi789..."
}
```

**Error Response (400 Bad Request):**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_STATUS",
    "message": "Invalid status 'INVALID'. Allowed values: CREATED, VERIFIED, EXPIRED."
  }
}
```

---

### API Gateway Implementation Example

```javascript
const express = require('express');
const { Gateway, Wallets } = require('fabric-network');
const fs = require('fs');
const path = require('path');

const app = express();
app.use(express.json());

// Load connection profile
const ccpPath = path.resolve(__dirname, 'connection-profile.json');
const ccp = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

// Initialize Gateway
const gateway = new Gateway();

// Connect to network
async function connectGateway() {
  const wallet = await Wallets.newFileSystemWallet('./wallet');
  
  const identity = await wallet.get('Admin@org1.ibn.vn');
  if (!identity) {
    throw new Error('Identity not found in wallet');
  }

  await gateway.connect(ccp, {
    wallet,
    identity: 'Admin@org1.ibn.vn',
    discovery: { enabled: true, asLocalhost: true }
  });
}

// API Routes
app.post('/api/v1/batches', async (req, res) => {
  try {
    const network = await gateway.getNetwork('ibnchannel');
    const contract = network.getContract('teaTraceCC');
    
    const { batchId, farmLocation, harvestDate, processingInfo, qualityCert } = req.body;
    
    const result = await contract.submitTransaction(
      'createBatch',
      batchId,
      farmLocation,
      harvestDate,
      processingInfo,
      qualityCert
    );
    
    const batch = JSON.parse(result.toString());
    res.status(201).json({ success: true, data: batch });
  } catch (error) {
    res.status(400).json({ success: false, error: { message: error.message } });
  }
});

app.get('/api/v1/batches/:batchId', async (req, res) => {
  try {
    const network = await gateway.getNetwork('ibnchannel');
    const contract = network.getContract('teaTraceCC');
    
    const result = await contract.evaluateTransaction('getBatchInfo', req.params.batchId);
    const batch = JSON.parse(result.toString());
    
    res.json({ success: true, data: batch });
  } catch (error) {
    res.status(404).json({ success: false, error: { message: error.message } });
  }
});

// Start server
const PORT = process.env.PORT || 3000;
connectGateway().then(() => {
  app.listen(PORT, () => {
    console.log(`API Gateway running on port ${PORT}`);
  });
});
```

---

## Error Handling

### Chaincode Error Codes

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| BATCH_EXISTS | 400 | Batch ID đã tồn tại |
| BATCH_NOT_FOUND | 404 | Batch ID không tồn tại |
| INVALID_STATUS | 400 | Status không hợp lệ |
| UNAUTHORIZED | 403 | MSP không có quyền |
| VERIFICATION_FAILED | 400 | Hash verification thất bại |
| NETWORK_ERROR | 500 | Lỗi kết nối network |
| TRANSACTION_FAILED | 500 | Transaction thất bại |

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {}
  },
  "timestamp": "2024-11-08T10:00:00.000Z"
}
```

---

## Security Configuration

### TLS Configuration

- **TLS Enabled:** Yes
- **Protocol:** gRPC over TLS (grpcs://)
- **Certificate Validation:** Required
- **Hostname Override:** Enabled (for localhost development)

### Authentication

- **Method:** X.509 Certificates
- **User Identity:** Admin@org1.ibn.vn
- **MSP ID:** Org1MSP

### Authorization

- **Level 1:** API Gateway authentication (JWT/OAuth)
- **Level 2:** Fabric MSP authorization (chaincode level)

### Environment Variables

```bash
# Network Configuration
CHANNEL_NAME=ibnchannel
CHAINCODE_NAME=teaTraceCC
MSP_ID=Org1MSP

# Peer Endpoints
PEER0_URL=grpcs://localhost:7051
PEER1_URL=grpcs://localhost:8051
PEER2_URL=grpcs://localhost:9051

# Orderer Endpoints
ORDERER_URL=grpcs://localhost:7050

# Certificate Paths
USER_CERT_PATH=core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/signcerts/cert.pem
USER_KEY_PATH=core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp/keystore
PEER_TLS_CA_PATH=core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/tls/ca.crt
ORDERER_TLS_CA_PATH=core/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp/tlscacerts/tlsca.ibn.vn-cert.pem
```

---

## Quick Reference

### Network Endpoints Summary

```
Peers:
  - peer0.org1.ibn.vn: grpcs://localhost:7051
  - peer1.org1.ibn.vn: grpcs://localhost:8051
  - peer2.org1.ibn.vn: grpcs://localhost:9051

Orderers:
  - orderer.ibn.vn: grpcs://localhost:7050
  - orderer1.ibn.vn: grpcs://localhost:8050
  - orderer2.ibn.vn: grpcs://localhost:9050

Channel: ibnchannel
Chaincode: teaTraceCC v1.0
MSP: Org1MSP
```

### Chaincode Functions Summary

| Function | Type | Authorization | Parameters |
|----------|------|---------------|------------|
| createBatch | Submit | Org1MSP | 5 params |
| verifyBatch | Submit | Org1MSP, Org2MSP, Org3MSP | 2 params |
| getBatchInfo | Evaluate | Public | 1 param |
| updateBatchStatus | Submit | Org1MSP, Org3MSP | 2 params |

---

## Tài Liệu Tham Khảo

- [Hyperledger Fabric Gateway SDK](https://hyperledger.github.io/fabric-gateway/)
- [Fabric Contract API](https://hyperledger.github.io/fabric-chaincode-node/release-2.5/api/)
- [Fabric Network Configuration](https://hyperledger-fabric.readthedocs.io/en/release-2.5/network/network.html)

---

**Cập nhật lần cuối:** 2025-11-11  
**Version:** 1.0  
**Maintainer:** IBN Network Team

