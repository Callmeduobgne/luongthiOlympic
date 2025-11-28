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

import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

export type TransactionStatus = 'SUBMITTED' | 'VALID' | 'INVALID' | 'FAILED'

export interface Transaction {
  id: string
  txId: string
  channelName: string
  chaincodeName: string
  functionName: string
  args?: string[]
  transientData?: Record<string, unknown>
  userId?: string
  apiKeyId?: string
  status: TransactionStatus
  blockNumber?: number
  blockHash?: string
  timestamp: string
  errorMessage?: string
  endorsingOrgs?: string[]
  createdAt: string
  updatedAt: string
}

export interface TransactionReceipt {
  txId: string
  status: TransactionStatus
  blockNumber: number
  blockHash?: string
  timestamp: string
  channelName: string
  chaincodeName: string
  functionName: string
  result?: unknown
  errorMessage?: string
}

export interface TransactionListQuery {
  channel?: string
  chaincode?: string
  status?: TransactionStatus
  userId?: string
  limit?: number
  offset?: number
  startTime?: string // RFC3339
  endTime?: string // RFC3339
}

export interface TransactionListResponse {
  transactions: Transaction[]
  total: number
  limit: number
  offset: number
}

export const transactionExplorerService = {
  /**
   * List transactions with filters
   */
  async listTransactions(query: TransactionListQuery = {}): Promise<TransactionListResponse> {
    try {
      const params = new URLSearchParams()
      
      if (query.channel) params.append('channel', query.channel)
      if (query.chaincode) params.append('chaincode', query.chaincode)
      if (query.status) params.append('status', query.status)
      if (query.userId) params.append('userId', query.userId)
      if (query.limit) params.append('limit', query.limit.toString())
      if (query.offset) params.append('offset', query.offset.toString())
      if (query.startTime) params.append('startTime', query.startTime)
      if (query.endTime) params.append('endTime', query.endTime)

      const response = await api.get<any>(
        `${API_ENDPOINTS.TRANSACTIONS.LIST}?${params.toString()}`
      )
      
      // Backend trả về format: { success: true, data: [...], meta: { total, limit, offset } }
      // Hoặc: { data: [...], meta: { total, limit, offset } }
      const responseData = response.data
      const transactions = responseData?.data || []
      const meta = responseData?.meta || {}
      
      // Map backend transaction format sang frontend Transaction format
      const mappedTransactions: Transaction[] = transactions.map((tx: any) => ({
        id: tx.id || tx.txId || '',
        txId: tx.txId || tx.tx_id || '',
        channelName: tx.channelName || tx.channel_name || '',
        chaincodeName: tx.chaincodeName || tx.chaincode_name || '',
        functionName: tx.functionName || tx.function_name || '',
        args: tx.args || [],
        transientData: tx.transientData || tx.transient_data || {},
        userId: tx.userId || tx.user_id || undefined,
        apiKeyId: tx.apiKeyId || tx.api_key_id || undefined,
        status: this.mapStatus(tx.status),
        blockNumber: tx.blockNumber || tx.block_number || undefined,
        blockHash: tx.blockHash || tx.block_hash || undefined,
        timestamp: tx.timestamp || tx.completed_at || tx.committed_at || tx.submitted_at || new Date().toISOString(),
        errorMessage: tx.errorMessage || tx.error_message || undefined,
        endorsingOrgs: tx.endorsingOrgs || tx.endorsing_orgs || undefined,
        createdAt: tx.createdAt || tx.created_at || tx.timestamp || new Date().toISOString(),
        updatedAt: tx.updatedAt || tx.updated_at || tx.timestamp || new Date().toISOString(),
      }))
      
      return {
        transactions: mappedTransactions,
        total: meta.total || transactions.length,
        limit: meta.limit || query.limit || 20,
        offset: meta.offset || query.offset || 0,
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to list transactions:', error)
      }
      return {
        transactions: [],
        total: 0,
        limit: query.limit || 20,
        offset: query.offset || 0,
      }
    }
  },

  /**
   * Map backend status to frontend TransactionStatus
   */
  mapStatus(status: string): TransactionStatus {
    const upperStatus = status?.toUpperCase() || ''
    if (upperStatus === 'VALID' || upperStatus === 'COMMITTED' || upperStatus === 'SUCCESS') {
      return 'VALID'
    }
    if (upperStatus === 'INVALID' || upperStatus === 'FAILED' || upperStatus === 'ERROR') {
      return 'INVALID'
    }
    if (upperStatus === 'SUBMITTED' || upperStatus === 'PENDING') {
      return 'SUBMITTED'
    }
    return 'SUBMITTED' // Default
  },

  /**
   * Get transaction by ID or TxID
   */
  async getTransaction(idOrTxID: string): Promise<Transaction> {
    const response = await api.get<{ success: boolean; data: Transaction }>(
      API_ENDPOINTS.TRANSACTIONS.GET(idOrTxID)
    )
    return response.data.data
  },

  /**
   * Get transaction status
   */
  async getTransactionStatus(idOrTxID: string): Promise<{ status: TransactionStatus }> {
    const response = await api.get<{ success: boolean; data: { status: TransactionStatus } }>(
      API_ENDPOINTS.TRANSACTIONS.STATUS(idOrTxID)
    )
    return response.data.data
  },

  /**
   * Get transaction receipt
   */
  async getTransactionReceipt(idOrTxID: string): Promise<TransactionReceipt> {
    const response = await api.get<{ success: boolean; data: TransactionReceipt }>(
      API_ENDPOINTS.TRANSACTIONS.RECEIPT(idOrTxID)
    )
    return response.data.data
  },
}


