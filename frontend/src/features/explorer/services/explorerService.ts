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
import type { Block, Transaction } from '@shared/types/blockchain.types'

export interface BlockListResponse {
  blocks: Block[]
  total: number
  limit: number
  offset: number
}

export const explorerService = {
  /**
   * List blocks with pagination
   * Note: Backend doesn't have list endpoint, only GetBlockByNumber
   * This is a placeholder - implement pagination by fetching blocks by number
   */
  async listBlocks(_channel: string, limit: number = 50, offset: number = 0): Promise<BlockListResponse> {
    // Backend only supports GetBlockByNumber, not ListBlocks
    // Return empty for now - need to implement logic to fetch blocks by number range
    if (import.meta.env.DEV) {
      console.warn('[DEV] listBlocks: Backend only supports GetBlockByNumber, not ListBlocks')
    }
    return {
      blocks: [],
      total: 0,
      limit,
      offset,
    }
  },

  /**
   * Get latest block
   * Uses channel info to get latest block number, then fetches that block
   */
  async getLatestBlock(_channel: string): Promise<Block> {
    try {
      // Get channel info to find latest block number
      await api.get(API_ENDPOINTS.BLOCKS.CHANNEL_INFO)
      // Parse channel info to get latest block number
      // For now, return error - need to parse hex-encoded blockchain info
      throw new Error('getLatestBlock: Need to parse channel info to get latest block number')
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to get latest block:', error)
      }
      throw error
    }
  },

  /**
   * Get block by number
   * Backend endpoint: /api/v1/blockchain/info/blocks/{number}
   */
  async getBlock(channel: string, blockNumber: number): Promise<Block> {
    const response = await api.get<Block>(
      API_ENDPOINTS.BLOCKS.GET(channel, blockNumber)
    )
    // Backend returns BlockInfo with RawBlock (hex-encoded), need to parse
    return response.data as Block
  },

  /**
   * Get transactions in a block
   * Note: Backend doesn't have this endpoint directly
   * Need to parse block data to extract transactions
   */
  async getBlockTransactions(_channel: string, _blockNumber: number): Promise<Transaction[]> {
    // Backend returns raw block (hex-encoded), need to parse to extract transactions
    // For now, return empty array
    if (import.meta.env.DEV) {
      console.warn('[DEV] getBlockTransactions: Need to parse raw block data to extract transactions')
    }
    return []
  },
}


