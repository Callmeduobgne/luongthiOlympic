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

export interface RollbackOperation {
  id: string
  chaincode_name: string
  channel_name: string
  from_version_id: string
  to_version_id: string
  from_version: string
  to_version: string
  from_sequence: number
  to_sequence: number
  status: 'pending' | 'in_progress' | 'completed' | 'failed' | 'cancelled'
  reason?: string
  rollback_type: 'version' | 'sequence'
  started_at?: string
  completed_at?: string
  duration_ms?: number
  requested_by: string
  executed_by?: string
  error_message?: string
  error_code?: string
  metadata?: Record<string, any>
  created_at: string
  updated_at: string
}

export interface RollbackHistory {
  id: string
  rollback_operation_id: string
  chaincode_version_id: string
  operation: string
  previous_status?: string
  new_status?: string
  details?: Record<string, any>
  created_at: string
}

export interface CreateRollbackRequest {
  chaincode_name: string
  channel_name: string
  to_version_id?: string
  reason?: string
  metadata?: Record<string, any>
}

export const rollbackService = {
  /**
   * Create a rollback operation
   */
  async createRollback(request: CreateRollbackRequest): Promise<RollbackOperation> {
    const response = await api.post<{ success: boolean; data: RollbackOperation }>(
      API_ENDPOINTS.CHAINCODE.ROLLBACK.CREATE,
      request
    )
    return response.data.data
  },

  /**
   * Execute a rollback operation
   */
  async executeRollback(id: string): Promise<void> {
    await api.post(API_ENDPOINTS.CHAINCODE.ROLLBACK.EXECUTE(id))
  },

  /**
   * Get rollback operation by ID
   */
  async getRollback(id: string): Promise<RollbackOperation> {
    const response = await api.get<{ success: boolean; data: RollbackOperation }>(
      API_ENDPOINTS.CHAINCODE.ROLLBACK.GET(id)
    )
    return response.data.data
  },

  /**
   * List rollback operations
   */
  async listRollbacks(filters?: {
    chaincode_name?: string
    channel_name?: string
    status?: string
  }): Promise<RollbackOperation[]> {
    try {
      const params = new URLSearchParams()
      if (filters?.chaincode_name) params.append('chaincode_name', filters.chaincode_name)
      if (filters?.channel_name) params.append('channel_name', filters.channel_name)
      if (filters?.status) params.append('status', filters.status)

      const queryString = params.toString()
      const url = queryString
        ? `${API_ENDPOINTS.CHAINCODE.ROLLBACK.LIST}?${queryString}`
        : API_ENDPOINTS.CHAINCODE.ROLLBACK.LIST

      const response = await api.get<{ success: boolean; data: { operations: RollbackOperation[]; count: number } }>(url)
      // Ensure we always return an array
      if (Array.isArray(response.data?.data?.operations)) {
        return response.data.data.operations
      }
      return []
    } catch (error) {
      console.error('Error loading rollback operations:', error)
      return [] // Return empty array on error
    }
  },

  /**
   * Get rollback history
   */
  async getRollbackHistory(id: string): Promise<RollbackHistory[]> {
    const response = await api.get<{ success: boolean; data: { history: RollbackHistory[]; count: number } }>(
      API_ENDPOINTS.CHAINCODE.ROLLBACK.HISTORY(id)
    )
    return response.data.data.history || []
  },

  /**
   * Cancel a rollback operation
   */
  async cancelRollback(id: string): Promise<void> {
    await api.delete(API_ENDPOINTS.CHAINCODE.ROLLBACK.CANCEL(id))
  },
}

