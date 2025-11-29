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
   * Backend doesn't have list blocks endpoint, so we:
   * 1. Get channel height to know total blocks
   * 2. Get transactions to populate blocks with transaction counts
   * 3. Create all blocks from 0 to height
   */
  async listBlocks(channel: string, limit: number = 50, offset: number = 0): Promise<BlockListResponse> {
    try {
      // 1. Lấy channel info để biết height (tổng số blocks)
      const channelInfoResponse = await api.get<any>(API_ENDPOINTS.BLOCKS.CHANNEL_INFO)
      const channelInfo = channelInfoResponse.data?.data || channelInfoResponse.data
      const height = channelInfo?.height || 0

      if (height === 0) {
        // Nếu không có height, fallback sang lấy từ transactions
        return await this.listBlocksFromTransactions(channel, limit, offset)
      }

      // 2. Lấy transactions để tính transactionCount cho mỗi block
      const transactionsResponse = await api.get<any>(
        '/api/v1/blockchain/transactions',
        {
          params: { limit: 1000 }, // Lấy nhiều để có đủ data
        }
      )

      const responseData = transactionsResponse.data?.data
      const transactions = Array.isArray(responseData) ? responseData : (responseData?.transactions || [])

      // 3. Group transactions theo block_number để tính transactionCount
      const blockTxCountMap = new Map<number, number>()
      const blockTimestampMap = new Map<number, string>()
      const blockHashMap = new Map<number, string>()

      transactions.forEach((tx: any) => {
        const blockNumber = tx.block_number || tx.blockNumber
        if (!blockNumber) return

        // Count transactions per block
        blockTxCountMap.set(blockNumber, (blockTxCountMap.get(blockNumber) || 0) + 1)

        // Store timestamp (lấy timestamp mới nhất)
        const txTimestamp = tx.timestamp || tx.completed_at || tx.committed_at || tx.submitted_at
        if (txTimestamp) {
          const existing = blockTimestampMap.get(blockNumber)
          if (!existing || txTimestamp > existing) {
            blockTimestampMap.set(blockNumber, txTimestamp)
          }
        }

        // Store hash (lấy hash đầu tiên tìm thấy)
        const txHash = tx.block_hash || tx.blockHash
        if (txHash && !blockHashMap.has(blockNumber)) {
          blockHashMap.set(blockNumber, txHash)
        }
      })

      // 4. Tạo tất cả blocks từ 0 đến height-1 (hoặc 1 đến height tùy blockchain)
      // Fabric thường bắt đầu từ block 0 (genesis block)
      const allBlocks: Block[] = []
      for (let i = height - 1; i >= 0; i--) {
        allBlocks.push({
          number: i,
          hash: blockHashMap.get(i) || '',
          previousHash: '', // Không có trong data hiện tại
          timestamp: blockTimestampMap.get(i) || new Date().toISOString(),
          transactionCount: blockTxCountMap.get(i) || 0,
          channel: channel,
          dataHash: '', // Placeholder
        })
      }

      // 5. Paginate
      const total = allBlocks.length
      const paginatedBlocks = allBlocks.slice(offset, offset + limit)

      return {
        blocks: paginatedBlocks,
        total,
        limit,
        offset,
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DEV] Failed to list blocks:', error)
      }
      // Fallback: lấy từ transactions nếu có lỗi
      return await this.listBlocksFromTransactions(channel, limit, offset)
    }
  },

  /**
   * Fallback: List blocks from transactions only (old logic)
   */
  async listBlocksFromTransactions(channel: string, limit: number = 50, offset: number = 0): Promise<BlockListResponse> {
    try {
      const response = await api.get<any>(
        '/api/v1/blockchain/transactions',
        {
          params: { limit: limit * 10 },
        }
      )

      const responseData = response.data?.data
      const transactions = Array.isArray(responseData) ? responseData : (responseData?.transactions || [])

      if (!Array.isArray(transactions) || transactions.length === 0) {
        return {
          blocks: [],
          total: 0,
          limit,
          offset,
        }
      }

      const blockMap = new Map<number, Block>()

      transactions.forEach((tx: any) => {
        const blockNumber = tx.block_number || tx.blockNumber
        if (!blockNumber) return

        if (!blockMap.has(blockNumber)) {
          blockMap.set(blockNumber, {
            number: blockNumber,
            hash: tx.block_hash || tx.blockHash || '',
            previousHash: '',
            timestamp: tx.timestamp || tx.completed_at || tx.committed_at || tx.submitted_at || new Date().toISOString(),
            transactionCount: 0,
            channel: tx.channel_name || tx.channelName || channel,
            dataHash: '', // Placeholder
          })
        }

        const block = blockMap.get(blockNumber)!
        block.transactionCount = (block.transactionCount || 0) + 1
      })

      const allBlocks = Array.from(blockMap.values())
        .sort((a, b) => b.number - a.number)

      return {
        blocks: allBlocks.slice(offset, offset + limit),
        total: allBlocks.length,
        limit,
        offset,
      }
    } catch (error) {
      return {
        blocks: [],
        total: 0,
        limit,
        offset,
      }
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


