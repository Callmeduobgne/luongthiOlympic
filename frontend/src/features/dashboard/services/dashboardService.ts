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

export interface MetricsSummary {
  transactions: {
    total: number
    valid: number
    invalid: number
    submitted: number
    successRate: number
    averageDuration: number
    last24Hours: number
    last7Days: number
    last30Days: number
  }
  blocks: {
    total: number
    last24Hours: number
    last7Days: number
    averageBlockTime: number
    largestBlock: number
  }
  performance: {
    averageResponseTime: number
    throughput: number
    errorRate: number
  }
}

export interface Block {
  number: number
  hash: string
  previousHash: string
  timestamp: string
  transactionCount: number
  channel: string
}

export interface NetworkInfo {
  name: string
  version: string
  channels: string[]
  chaincodes: string[]
  peers: number
  orderers: number
}

export const dashboardService = {
  async getMetricsSummary(channel?: string): Promise<MetricsSummary> {
    try {
      // Backend trả về MetricSnapshot { timestamp, metrics: map[string]float64 }
      // Cần parse thành MetricsSummary format
      const response = await api.get<any>(
        `${API_ENDPOINTS.METRICS.SUMMARY}${channel ? `?channel=${channel}` : ''}`
      )

      // Backend có thể trả về:
      // 1. { timestamp, metrics: {...} } (MetricSnapshot)
      // 2. { success: true, data: {...} }
      // 3. { data: {...} }

      const snapshot = response.data?.data || response.data
      const metricsMap = snapshot?.metrics || {}

      // Backend metrics service không có transaction metrics
      // Luôn lấy từ transactions thực tế để đảm bảo có dữ liệu
      // Chỉ dùng metrics map nếu có transaction metrics cụ thể
      const hasTransactionMetrics = metricsMap['transaction_total'] ||
        metricsMap['tx_total'] ||
        metricsMap['transaction_valid'] ||
        metricsMap['tx_valid']

      if (!hasTransactionMetrics) {
        // Luôn fallback sang transactions thực tế
        return await this.getMetricsFromTransactions()
      }

      // Parse metrics map thành MetricsSummary (nếu có)
      return {
        transactions: {
          total: metricsMap['transaction_total'] || metricsMap['tx_total'] || 0,
          valid: metricsMap['transaction_valid'] || metricsMap['tx_valid'] || 0,
          invalid: metricsMap['transaction_invalid'] || metricsMap['tx_invalid'] || 0,
          submitted: metricsMap['transaction_submitted'] || metricsMap['tx_submitted'] || 0,
          successRate: metricsMap['transaction_success_rate'] || metricsMap['tx_success_rate'] || 0,
          averageDuration: metricsMap['transaction_duration_avg'] || metricsMap['tx_duration_avg'] || 0,
          last24Hours: metricsMap['transaction_24h'] || metricsMap['tx_24h'] || 0,
          last7Days: metricsMap['transaction_7d'] || metricsMap['tx_7d'] || 0,
          last30Days: metricsMap['transaction_30d'] || metricsMap['tx_30d'] || 0,
        },
        blocks: {
          total: metricsMap['block_total'] || 0,
          last24Hours: metricsMap['block_24h'] || 0,
          last7Days: metricsMap['block_7d'] || 0,
          averageBlockTime: metricsMap['block_time_avg'] || 0,
          largestBlock: metricsMap['block_largest'] || 0,
        },
        performance: {
          averageResponseTime: metricsMap['response_time_avg'] || 0,
          throughput: metricsMap['throughput'] || 0,
          errorRate: metricsMap['error_rate'] || 0,
        },
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to get metrics summary:', error)
      }
      // Fallback: lấy từ transactions
      return await this.getMetricsFromTransactions()
    }
  },

  async getMetricsFromTransactions(): Promise<MetricsSummary> {
    try {
      // Lấy transactions từ database để tính metrics
      const response = await api.get<any>('/api/v1/blockchain/transactions', {
        params: { limit: 1000 }, // Lấy nhiều để tính toán chính xác
      })

      const responseData = response.data?.data
      const transactions = Array.isArray(responseData) ? responseData : (responseData?.transactions || [])

      if (!Array.isArray(transactions) || transactions.length === 0) {
        // Return empty metrics nếu không có transactions
        return {
          transactions: {
            total: 0,
            valid: 0,
            invalid: 0,
            submitted: 0,
            successRate: 0,
            averageDuration: 0,
            last24Hours: 0,
            last7Days: 0,
            last30Days: 0,
          },
          blocks: {
            total: 0,
            last24Hours: 0,
            last7Days: 0,
            averageBlockTime: 0,
            largestBlock: 0,
          },
          performance: {
            averageResponseTime: 0,
            throughput: 0,
            errorRate: 0,
          },
        }
      }

      const now = new Date()
      const last24h = new Date(now.getTime() - 24 * 60 * 60 * 1000)
      const last7d = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
      const last30d = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)

      const total = transactions.length
      // Backend trả về status: "VALID", "INVALID", "submitted", "pending", "committed", "failed", etc.
      const valid = transactions.filter((tx: any) =>
        tx.status === 'VALID' ||
        tx.status === 'committed' ||
        tx.status === 'success' ||
        (tx.status && tx.status.toUpperCase() === 'VALID')
      ).length
      const invalid = transactions.filter((tx: any) =>
        tx.status === 'INVALID' ||
        tx.status === 'failed' ||
        tx.status === 'error' ||
        (tx.status && tx.status.toUpperCase() === 'INVALID')
      ).length
      const submitted = transactions.filter((tx: any) =>
        tx.status === 'submitted' ||
        tx.status === 'pending' ||
        (tx.status && !['VALID', 'INVALID', 'committed', 'failed', 'success', 'error'].includes(tx.status))
      ).length

      const last24hTxs = transactions.filter((tx: any) => {
        const txTime = new Date(tx.timestamp || tx.completed_at || tx.committed_at || tx.submitted_at || 0)
        return txTime >= last24h && !isNaN(txTime.getTime())
      })

      const last7dTxs = transactions.filter((tx: any) => {
        const txTime = new Date(tx.timestamp || tx.completed_at || tx.committed_at || tx.submitted_at || 0)
        return txTime >= last7d && !isNaN(txTime.getTime())
      })

      const last30dTxs = transactions.filter((tx: any) => {
        const txTime = new Date(tx.timestamp || tx.completed_at || tx.committed_at || tx.submitted_at || 0)
        return txTime >= last30d && !isNaN(txTime.getTime())
      })

      // Tính average duration (nếu có)
      const durations = transactions
        .filter((tx: any) => tx.completed_at && tx.submitted_at)
        .map((tx: any) => {
          const completed = new Date(tx.completed_at).getTime()
          const submitted = new Date(tx.submitted_at).getTime()
          return completed - submitted
        })
      const avgDuration = durations.length > 0
        ? durations.reduce((a: number, b: number) => a + b, 0) / durations.length
        : 0

      // Group by block_number để tính blocks
      const blockNumbers = new Set<number>()
      transactions.forEach((tx: any) => {
        if (tx.block_number || tx.blockNumber) {
          blockNumbers.add(tx.block_number || tx.blockNumber)
        }
      })

      return {
        transactions: {
          total,
          valid,
          invalid,
          submitted,
          successRate: total > 0 ? (valid / total) * 100 : 0,
          averageDuration: avgDuration,
          last24Hours: last24hTxs.length,
          last7Days: last7dTxs.length,
          last30Days: last30dTxs.length,
        },
        blocks: {
          total: blockNumbers.size,
          last24Hours: 0, // Cần tính từ block timestamps
          last7Days: 0,
          averageBlockTime: 0,
          largestBlock: 0,
        },
        performance: {
          averageResponseTime: avgDuration,
          throughput: total > 0 ? total / 30 : 0, // Approximate
          errorRate: total > 0 ? (invalid / total) * 100 : 0,
        },
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to get metrics from transactions:', error)
      }
      // Return empty metrics nếu có lỗi
      return {
        transactions: {
          total: 0,
          valid: 0,
          invalid: 0,
          submitted: 0,
          successRate: 0,
          averageDuration: 0,
          last24Hours: 0,
          last7Days: 0,
          last30Days: 0,
        },
        blocks: {
          total: 0,
          last24Hours: 0,
          last7Days: 0,
          averageBlockTime: 0,
          largestBlock: 0,
        },
        performance: {
          averageResponseTime: 0,
          throughput: 0,
          errorRate: 0,
        },
      }
    }
  },

  async getLatestBlocks(channel: string, limit: number = 10): Promise<Block[]> {
    try {
      // Backend endpoint /api/v1/blocks/{channel} chỉ trả về [] (stub)
      // Thay vào đó, lấy từ transactions và group theo block_number
      const response = await api.get<any>(
        '/api/v1/blockchain/transactions',
        {
          params: { limit: limit * 10 }, // Lấy nhiều transactions để có đủ blocks
        }
      )

      const responseData = response.data?.data
      const transactions = Array.isArray(responseData) ? responseData : (responseData?.transactions || [])

      if (!Array.isArray(transactions) || transactions.length === 0) {
        return []
      }

      // Group transactions theo block_number và tạo blocks
      const blockMap = new Map<number, Block>()

      transactions.forEach((tx: any) => {
        const blockNumber = tx.block_number || tx.blockNumber
        if (!blockNumber) return

        if (!blockMap.has(blockNumber)) {
          blockMap.set(blockNumber, {
            number: blockNumber,
            hash: tx.block_hash || tx.blockHash || '',
            previousHash: '', // Không có trong transaction data
            timestamp: tx.completed_at || tx.committed_at || tx.submitted_at || new Date().toISOString(),
            transactionCount: 0,
            channel: tx.channel_name || tx.channelName || channel,
          })
        }

        // Tăng transaction count
        const block = blockMap.get(blockNumber)!
        block.transactionCount = (block.transactionCount || 0) + 1
      })

      // Convert map to array và sort theo block number (desc)
      const blocks = Array.from(blockMap.values())
        .sort((a, b) => b.number - a.number)
        .slice(0, limit)

      return blocks
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to get latest blocks:', error)
      }
      return []
    }
  },

  async getNetworkInfo(): Promise<NetworkInfo> {
    // Backend doesn't have /api/v1/network/info
    // Use blockchain channel info instead
    try {
      const response = await api.get<any>(
        API_ENDPOINTS.BLOCKS.CHANNEL_INFO
      )
      const data = response.data?.data || response.data
      // Ensure data is properly formatted
      if (data && typeof data === 'object') {
        // Extract channels (array of channel objects)
        const channels = Array.isArray(data.channels) ? data.channels : []
        const channelNames = channels.map((ch: any) => ch.name || ch).filter(Boolean)

        // Extract chaincodes from all channels (flatten and unique)
        const allChaincodes: string[] = channels
          .flatMap((ch: any) => Array.isArray(ch.chaincodes) ? ch.chaincodes : [])
          .filter((cc: any): cc is string => cc && typeof cc === 'string')
        const uniqueChaincodes = Array.from(new Set(allChaincodes))

        // Extract peers (array of peer objects)
        const peers = Array.isArray(data.peers) ? data.peers : []
        const peerCount = peers.length

        // Extract orderers (array of orderer objects)
        const orderers = Array.isArray(data.orderers) ? data.orderers : []
        const ordererCount = orderers.length

        return {
          name: data.name || 'IBN Network',
          version: data.version || '',
          channels: channelNames,
          chaincodes: uniqueChaincodes,
          peers: peerCount,
          orderers: ordererCount,
        }
      }
    } catch (error) {
      // If endpoint fails (e.g., 500 error), return fallback empty data
      // This prevents ERR_BAD_RESPONSE errors in frontend
      console.warn('Failed to fetch network info, using fallback:', error)
    }
    // Return fallback empty data if endpoint fails or data is invalid
    return {
      name: 'IBN Network',
      version: '',
      channels: [],
      chaincodes: [],
      peers: 0,
      orderers: 0,
    }
  },

  async getChannelInfo(channel: string) {
    const response = await api.get(
      API_ENDPOINTS.NETWORK.CHANNEL_INFO(channel)
    )
    return response.data.data
  },
}

