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

// In production (Docker), use relative URL (nginx will proxy /api to api-gateway)
// In development, use Vite proxy (relative URL) hoặc service name
const getBaseURL = () => {
  const envUrl = import.meta.env.VITE_API_BASE_URL
  // Nếu có env variable và là container name, dùng relative URL để Vite proxy xử lý
  if (envUrl && envUrl.includes('api-gateway-nginx')) {
    return '' // Dùng relative URL, Vite proxy sẽ forward đến api-gateway-nginx
  }
  // Nếu env variable là localhost, cũng dùng relative URL để Vite proxy xử lý
  if (envUrl && envUrl.includes('localhost')) {
    return '' // Dùng relative URL, Vite proxy sẽ forward
  }
  if (envUrl && !envUrl.includes('localhost')) {
    return envUrl // Chỉ dùng env URL nếu không phải localhost
  }
  // Production: use relative URL (nginx proxy)
  if (import.meta.env.PROD) {
    return ''
  }
  // Development: ALWAYS use relative URL để Vite proxy xử lý (target: api-gateway-nginx:80)
  // Vite proxy sẽ forward /api requests đến api-gateway-nginx:80
  return '' // Dùng Vite proxy
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
    SUMMARY: '/api/v1/metrics/summary',
    TRANSACTIONS: '/api/v1/metrics/transactions',
    BLOCKS: '/api/v1/metrics/blocks',
    PERFORMANCE: '/api/v1/metrics/performance',
    PEERS: '/api/v1/metrics/peers',
  },
  BLOCKS: {
    LIST: (channel: string) => `/api/v1/blocks/${channel}`,
    LATEST: (channel: string) => `/api/v1/blocks/${channel}/latest`,
    GET: (channel: string, number: number) => `/api/v1/blocks/${channel}/${number}`,
  },
  NETWORK: {
    INFO: '/api/v1/network/info',
    CHANNELS: '/api/v1/network/channels',
    CHANNEL_INFO: (name: string) => `/api/v1/network/channels/${name}`,
    PEERS: '/api/v1/network/peers',
    ORDERERS: '/api/v1/network/orderers',
    LOGS: '/api/v1/network/logs',
  },
  BATCHES: {
    GET: (id: string) => `/api/v1/batches/${id}`,
    CREATE: '/api/v1/batches',
    VERIFY: (id: string) => `/api/v1/batches/${id}/verify`,
    UPDATE_STATUS: (id: string) => `/api/v1/batches/${id}/status`,
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
    LIST: '/api/v1/transactions',
    GET: (id: string) => `/api/v1/transactions/${id}`,
    STATUS: (id: string) => `/api/v1/transactions/${id}/status`,
    RECEIPT: (id: string) => `/api/v1/transactions/${id}/receipt`,
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
