/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Frontend → Backend (port 9090) → Gateway (port 8080/8085) → Fabric
// Frontend KHÔNG gọi Gateway trực tiếp, chỉ gọi Backend
// Backend sẽ tự gọi Gateway khi cần blockchain operations
const getBaseURL = () => {
  const envUrl = import.meta.env.VITE_API_BASE_URL
  // Nếu có env variable và là backend URL, dùng nó
  if (envUrl && envUrl.includes('localhost:9090')) {
    return envUrl // Trong dev, có thể dùng trực tiếp backend URL
  }
  // Nếu có env variable và là container name (Docker), dùng relative URL
  if (envUrl && (envUrl.includes('api-gateway-nginx') || envUrl.includes('backend'))) {
    return '' // Dùng relative URL, Vite proxy sẽ forward
  }
  // Nếu có env variable khác, dùng nó
  if (envUrl && !envUrl.includes('localhost')) {
    return envUrl
  }
  // Production: use relative URL (nginx proxy)
  if (import.meta.env.PROD) {
    return ''
  }
  // Development: ALWAYS use relative URL để Vite proxy xử lý
  // Vite proxy đã được cấu hình trỏ tới backend (localhost:9090)
  return '' // Dùng Vite proxy → Backend (9090)
}

export const API_CONFIG = {
  BASE_URL: getBaseURL(),
  API_VERSION: '/api/v1',
} as const

export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/api/v1/auth/login',
    REFRESH: '/api/v1/auth/refresh',
    PROFILE: '/api/v1/auth/profile',
    UPLOAD_AVATAR: '/api/v1/auth/avatar',
  },
  METRICS: {
    // Backend endpoints: /api/v1/metrics/*
    SUMMARY: '/api/v1/metrics/snapshot', // Backend uses 'snapshot' not 'summary'
    ALL: '/api/v1/metrics', // Get all metrics
    AGGREGATIONS: '/api/v1/metrics/aggregations', // Get aggregations
    BY_NAME: '/api/v1/metrics/by-name', // Get metric by name (query param: ?name=...)
    // Note: Backend doesn't have separate endpoints for transactions/blocks/performance/peers
    // Use snapshot or aggregations instead
  },
  BLOCKS: {
    // Backend endpoints: /api/v1/blockchain/blocks/{number}
    // Note: Backend doesn't have list blocks endpoint, only get by number
    LIST: (_channel: string) => `/api/v1/blockchain/channel/info`, // Use channel info instead
    LATEST: (_channel: string) => `/api/v1/blockchain/channel/info`, // Use channel info instead
    GET: (_channel: string, number: number) => `/api/v1/blockchain/blocks/${number}`,
    GET_BY_TXID: (txid: string) => `/api/v1/blockchain/blocks/tx/${txid}`,
    CHANNEL_INFO: '/api/v1/blockchain/channel/info',
  },
  NETWORK: {
    // Backend only has /api/v1/network/logs
    // Other endpoints use blockchain/channel/info as fallback
    INFO: '/api/v1/blockchain/channel/info', // Use blockchain channel info
    CHANNELS: '/api/v1/blockchain/channel/info', // Use blockchain channel info
    CHANNEL_INFO: (_name: string) => `/api/v1/blockchain/channel/info`, // Use blockchain channel info
    PEERS: '/api/v1/blockchain/channel/info', // Placeholder - parse from channel info
    ORDERERS: '/api/v1/blockchain/channel/info', // Placeholder - parse from channel info
    LOGS: '/api/v1/network/logs', // Backend has this endpoint
  },
  BATCHES: {
    // Backend uses /api/v1/teatrace/batches
    GET: (id: string) => `/api/v1/teatrace/batches/${id}`,
    CREATE: '/api/v1/teatrace/batches',
    LIST: '/api/v1/teatrace/batches',
    VERIFY: (id: string) => `/api/v1/teatrace/batches/${id}/verify`,
    UPDATE_STATUS: (id: string) => `/api/v1/teatrace/batches/${id}/status`,
    HEALTH: '/api/v1/teatrace/health',
    VERIFY_BY_HASH: '/api/v1/teatrace/verify-by-hash', // NEW: Merkle proof verification
  },
  QRCODE: {
    // Batch QR codes
    BATCH_PNG: (batchId: string) => `/api/v1/qrcode/batches/${batchId}`,
    BATCH_BASE64: (batchId: string) => `/api/v1/qrcode/batches/${batchId}/base64`,
    BATCH_DATA: (batchId: string) => `/api/v1/qrcode/batches/${batchId}/data`,
    // Package QR codes
    PACKAGE_PNG: (packageId: string) => `/api/v1/qrcode/packages/${packageId}`,
    PACKAGE_BASE64: (packageId: string) => `/api/v1/qrcode/packages/${packageId}/base64`,
    PACKAGE_DATA: (packageId: string) => `/api/v1/qrcode/packages/${packageId}/data`,
    // Transaction QR code (auto-detect batch or package)
    TRANSACTION: (txId: string) => `/api/v1/qrcode/transactions/${txId}`,
  },
  CHAINCODE: {
    INSTALLED: '/api/v1/chaincode/installed',
    COMMITTED: '/api/v1/chaincode/committed',
    COMMITTED_INFO: (name: string) => `/api/v1/chaincode/committed/${name}`,
    UPLOAD: '/api/v1/chaincode/upload',
    INSTALL: '/api/v1/chaincode/install',
    APPROVE: '/api/v1/chaincode/approve',
    COMMIT: '/api/v1/chaincode/commit',
    QUERY: (channel: string, name: string) => `/api/v1/channels/${channel}/chaincodes/${name}/query`,
    INVOKE: (channel: string, name: string) => `/api/v1/channels/${channel}/chaincodes/${name}/invoke`,
    // Approval workflow
    APPROVAL: {
      CREATE_REQUEST: '/api/v1/chaincode/approval/request',
      VOTE: '/api/v1/chaincode/approval/vote',
      GET_REQUEST: (id: string) => `/api/v1/chaincode/approval/request/${id}`,
      LIST_REQUESTS: '/api/v1/chaincode/approval/requests',
    },
    // Rollback
    ROLLBACK: {
      CREATE: '/api/v1/chaincode/rollback',
      EXECUTE: (id: string) => `/api/v1/chaincode/rollback/${id}/execute`,
      GET: (id: string) => `/api/v1/chaincode/rollback/${id}`,
      LIST: '/api/v1/chaincode/rollback',
      HISTORY: (id: string) => `/api/v1/chaincode/rollback/${id}/history`,
      CANCEL: (id: string) => `/api/v1/chaincode/rollback/${id}`,
    },
    // Testing
    TESTING: {
      RUN: '/api/v1/chaincode/testing/run',
      LIST_SUITES: '/api/v1/chaincode/testing/suites',
      GET_SUITE: (id: string) => `/api/v1/chaincode/testing/suites/${id}`,
      GET_CASES: (id: string) => `/api/v1/chaincode/testing/suites/${id}/cases`,
    },
    // Version management
    VERSION: {
      CREATE_TAG: '/api/v1/chaincode/version/tags',
      GET_TAGS: (versionId: string) => `/api/v1/chaincode/version/versions/${versionId}/tags`,
      GET_BY_TAG: (chaincodeName: string, tagName: string) => `/api/v1/chaincode/version/chaincodes/${chaincodeName}/tags/${tagName}`,
      CREATE_DEPENDENCY: '/api/v1/chaincode/version/dependencies',
      GET_DEPENDENCIES: (versionId: string) => `/api/v1/chaincode/version/versions/${versionId}/dependencies`,
      CREATE_RELEASE_NOTE: '/api/v1/chaincode/version/release-notes',
      GET_RELEASE_NOTE: (versionId: string) => `/api/v1/chaincode/version/versions/${versionId}/release-notes`,
      COMPARE: '/api/v1/chaincode/version/compare',
      GET_LATEST: (chaincodeName: string) => `/api/v1/chaincode/version/chaincodes/${chaincodeName}/latest`,
      GET_HISTORY: (chaincodeName: string) => `/api/v1/chaincode/version/chaincodes/${chaincodeName}/history`,
      GET_COMPARISONS: (versionId: string) => `/api/v1/chaincode/version/versions/${versionId}/comparisons`,
    },
    // CI/CD
    CICD: {
      CREATE_PIPELINE: '/api/v1/chaincode/cicd/pipelines',
      LIST_PIPELINES: '/api/v1/chaincode/cicd/pipelines',
      GET_PIPELINE: (id: string) => `/api/v1/chaincode/cicd/pipelines/${id}`,
      TRIGGER_EXECUTION: '/api/v1/chaincode/cicd/executions',
      LIST_EXECUTIONS: '/api/v1/chaincode/cicd/executions',
      GET_EXECUTION: (id: string) => `/api/v1/chaincode/cicd/executions/${id}`,
      GET_ARTIFACTS: (id: string) => `/api/v1/chaincode/cicd/executions/${id}/artifacts`,
      PROCESS_WEBHOOK: (pipelineId: string) => `/api/v1/chaincode/cicd/webhooks/${pipelineId}`,
    },
  },
  TRANSACTIONS: {
    // Backend uses /api/v1/blockchain/transactions
    LIST: '/api/v1/blockchain/transactions',
    GET: (id: string) => `/api/v1/blockchain/transactions/${id}`,
    GET_BY_TXID: (txid: string) => `/api/v1/blockchain/info/transaction/${txid}`,
    HISTORY: (id: string) => `/api/v1/blockchain/transactions/${id}/history`,
    GET_BY_TXID_ALT: (txid: string) => `/api/v1/blockchain/txid/${txid}`,
    SUBMIT: '/api/v1/blockchain/transactions',
    QUERY: '/api/v1/blockchain/query',
    STATUS: (id: string) => `/api/v1/blockchain/transactions/${id}/status`,
    RECEIPT: (id: string) => `/api/v1/blockchain/transactions/${id}/receipt`,
  },
  USERS: {
    LIST: '/api/v1/users',
    GET: (id: string) => `/api/v1/users/${id}`,
    ENROLL: '/api/v1/users/enroll',
    REGISTER: '/api/v1/users/register',
    REENROLL: (id: string) => `/api/v1/users/${id}/reenroll`,
    REVOKE: (id: string) => `/api/v1/users/${id}/revoke`,
    CERTIFICATE: (id: string) => `/api/v1/users/${id}/certificate`,
  },
} as const
