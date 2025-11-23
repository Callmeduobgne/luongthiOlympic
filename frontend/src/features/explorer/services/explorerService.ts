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
   */
  async listBlocks(channel: string, limit: number = 50, offset: number = 0): Promise<BlockListResponse> {
    const response = await api.get<{ success: boolean; data: BlockListResponse }>(
      `${API_ENDPOINTS.BLOCKS.LIST(channel)}?limit=${limit}&offset=${offset}`
    )
    return response.data.data
  },

  /**
   * Get latest block
   */
  async getLatestBlock(channel: string): Promise<Block> {
    const response = await api.get<{ success: boolean; data: Block }>(
      API_ENDPOINTS.BLOCKS.LATEST(channel)
    )
    return response.data.data
  },

  /**
   * Get block by number
   */
  async getBlock(channel: string, blockNumber: number): Promise<Block> {
    const response = await api.get<{ success: boolean; data: Block }>(
      API_ENDPOINTS.BLOCKS.GET(channel, blockNumber)
    )
    return response.data.data
  },

  /**
   * Get transactions in a block
   */
  async getBlockTransactions(channel: string, blockNumber: number): Promise<Transaction[]> {
    const response = await api.get<{ success: boolean; data: Transaction[] }>(
      `${API_ENDPOINTS.BLOCKS.GET(channel, blockNumber)}/transactions`
    )
    return response.data.data || []
  },
}


