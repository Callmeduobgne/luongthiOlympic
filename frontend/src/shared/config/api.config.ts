/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

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

export const API_ENDPOINTS = {
  AUTH: {
    LOGIN: '/api/v1/auth/login',
    REGISTER: '/api/v1/auth/register',
    LOGOUT: '/api/v1/auth/logout',
    REFRESH: '/api/v1/auth/refresh',
    PROFILE: '/api/v1/auth/profile',
    UPLOAD_AVATAR: '/api/v1/auth/profile/avatar',
  },
  BATCHES: {
    LIST: '/api/v1/batches',
    CREATE: '/api/v1/batches',
    GET: (id: string) => `/api/v1/batches/${id}`,
    UPDATE: (id: string) => `/api/v1/batches/${id}`,
    DELETE: (id: string) => `/api/v1/batches/${id}`,
    UPDATE_STATUS: (id: string) => `/api/v1/batches/${id}/status`,
    VERIFY: (id: string) => `/api/v1/batches/${id}/verify`,
    VERIFY_BY_HASH: '/api/v1/batches/verify/hash',
  },
  PACKAGES: {
    LIST: '/api/v1/packages',
    CREATE: '/api/v1/packages',
    GET: (id: string) => `/api/v1/packages/${id}`,
    UPDATE: (id: string) => `/api/v1/packages/${id}`,
    DELETE: (id: string) => `/api/v1/packages/${id}`,
    UPDATE_STATUS: (id: string) => `/api/v1/packages/${id}/status`,
    VERIFY: (id: string) => `/api/v1/packages/${id}/verify`,
    BY_BATCH: (batchId: string) => `/api/v1/packages/batch/${batchId}`,
  },
  QRCODE: {
    BATCH_BASE64: (batchId: string) => `/api/v1/qrcode/batches/${batchId}/base64`,
    PACKAGE_BASE64: (packageId: string) => `/api/v1/qrcode/packages/${packageId}/base64`,
    TRANSACTION: (txId: string) => `/api/v1/qrcode/transactions/${txId}`,
    BATCH_DATA: (batchId: string) => `/api/v1/qrcode/batches/${batchId}/data`,
    BATCH_PNG: (batchId: string) => `/api/v1/qrcode/batches/${batchId}/png`,
    PACKAGE_DATA: (packageId: string) => `/api/v1/qrcode/packages/${packageId}/data`,
    PACKAGE_PNG: (packageId: string) => `/api/v1/qrcode/packages/${packageId}/png`,
  },
  BLOCKS: {
    LIST: (channel: string) => `/api/v1/blocks/${channel}`,
    GET: (channel: string, blockNumber: number) => `/api/v1/blocks/${channel}/${blockNumber}`,
    CHANNEL_INFO: (channel: string) => `/api/v1/blocks/${channel}/info`,
  },
  TRANSACTIONS: {
    LIST: '/api/v1/transactions',
    GET: (txId: string) => `/api/v1/transactions/${txId}`,
    STATUS: (txId: string) => `/api/v1/transactions/${txId}/status`,
    RECEIPT: (txId: string) => `/api/v1/transactions/${txId}/receipt`,
  },
  CHAINCODE: {
    LIST: '/api/v1/chaincode',
    DEPLOY: '/api/v1/chaincode/deploy',
    APPROVE: '/api/v1/chaincode/approve',
    COMMIT: '/api/v1/chaincode/commit',
    TEST: '/api/v1/chaincode/test',
    TESTING: {
      RUN: '/api/v1/chaincode/testing/run',
      GET_SUITE: (id: string) => `/api/v1/chaincode/testing/suites/${id}`,
      LIST_SUITES: '/api/v1/chaincode/testing/suites',
      GET_CASES: (suiteId: string) => `/api/v1/chaincode/testing/suites/${suiteId}/cases`,
    },
    ROLLBACK: {
      CREATE: '/api/v1/chaincode/rollback',
      EXECUTE: (id: string) => `/api/v1/chaincode/rollback/${id}/execute`,
      GET: (id: string) => `/api/v1/chaincode/rollback/${id}`,
      LIST: '/api/v1/chaincode/rollback',
      HISTORY: (id: string) => `/api/v1/chaincode/rollback/${id}/history`,
      CANCEL: (id: string) => `/api/v1/chaincode/rollback/${id}/cancel`,
    },
    APPROVAL: {
      CREATE_REQUEST: '/api/v1/chaincode/approval',
      VOTE: '/api/v1/chaincode/approval/vote',
      GET_REQUEST: (id: string) => `/api/v1/chaincode/approval/${id}`,
      LIST_REQUESTS: '/api/v1/chaincode/approval',
    },
    INSTALLED: '/api/v1/chaincode/installed',
    COMMITTED: '/api/v1/chaincode/committed',
    COMMITTED_INFO: (name: string) => `/api/v1/chaincode/committed/${name}/info`,
    UPLOAD: '/api/v1/chaincode/upload',
    INSTALL: '/api/v1/chaincode/install',
    VERSION: {
      GET_LATEST: (chaincodeName: string) => `/api/v1/chaincode/${chaincodeName}/version/latest`,
    },
    INVOKE: (channel: string, chaincodeName: string) => `/api/v1/chaincode/${channel}/${chaincodeName}/invoke`,
    QUERY: (channel: string, chaincodeName: string) => `/api/v1/chaincode/${channel}/${chaincodeName}/query`,
  },
  NETWORK: {
    OVERVIEW: '/api/v1/network/overview',
    CHANNELS: '/api/v1/network/channels',
    PEERS: '/api/v1/network/peers',
    ORDERERS: '/api/v1/network/orderers',
    LOGS: '/api/v1/network/logs',
    INFO: '/api/v1/network/info',
    CHANNEL_INFO: (channel: string) => `/api/v1/network/channels/${channel}/info`,
  },
  EXPLORER: {
    BLOCKS: (channel: string) => `/api/v1/explorer/blocks/${channel}`,
    TRANSACTIONS: '/api/v1/explorer/transactions',
  },
  DASHBOARD: {
    WS: (channel: string) => `/api/v1/dashboard/ws/${channel}`,
    METRICS: '/api/v1/dashboard/metrics',
  },
  METRICS: {
    SUMMARY: '/api/v1/metrics/summary',
  },
} as const

export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || '',
  TIMEOUT: 30000,
  FRONTEND_URL: typeof window !== 'undefined'
    ? (import.meta.env.VITE_FRONTEND_URL || window.location.origin)
    : import.meta.env.VITE_FRONTEND_URL || '',
} as const

