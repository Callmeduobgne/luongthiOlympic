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
    const params = channel ? `?channel=${channel}` : ''
    const response = await api.get<{ success: boolean; data: MetricsSummary }>(
      `${API_ENDPOINTS.METRICS.SUMMARY}${params}`
    )
    return response.data.data
  },

  async getLatestBlocks(_channel: string, _limit: number = 10): Promise<Block[]> {
    // Backend doesn't have list blocks endpoint, only get by number
    // Use channel info to get latest block number, then fetch recent blocks
    try {
      await api.get(API_ENDPOINTS.BLOCKS.CHANNEL_INFO)
      // Channel info contains blockchain info, but we need to parse it
      // For now, return empty array to avoid errors
      // TODO: Implement logic to fetch blocks by number from latest to latest-limit
      if (import.meta.env.DEV) {
        console.warn('[DEV] getLatestBlocks: Backend only supports GetBlockByNumber, not ListBlocks')
      }
      return []
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to get channel info for blocks:', error)
      }
      return []
    }
  },

  async getNetworkInfo(): Promise<NetworkInfo> {
    // Backend doesn't have /api/v1/network/info
    // Use blockchain channel info instead
    const response = await api.get<any>(
      API_ENDPOINTS.BLOCKS.CHANNEL_INFO
    )
    const data = response.data.data
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

