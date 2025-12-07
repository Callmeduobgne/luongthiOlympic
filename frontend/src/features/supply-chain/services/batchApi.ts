import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import type {
  TeaBatch,
  CreateBatchRequest,
  UpdateBatchStatusRequest,
  VerifyBatchRequest,
  VerifyBatchResponse,
} from '../types/batch.types'

type ApiResponse<T> = {
  success: boolean
  data: T
  message?: string
}

export const batchApi = {
  /**
   * Get batch information
   * Uses REST API endpoint (public, no auth required)
   */
  getBatchInfo: async (batchId: string): Promise<TeaBatch> => {
    const response = await api.get<ApiResponse<{ result: TeaBatch }>>(
      API_ENDPOINTS.BATCHES.GET(batchId)
    )
    return response.data.data.result
  },

  /**
   * Create new batch
   * Uses REST API endpoint (auth required)
   */
  createBatch: async (data: CreateBatchRequest): Promise<TeaBatch> => {
    const response = await api.post<ApiResponse<TeaBatch>>(
      API_ENDPOINTS.BATCHES.CREATE,
      data
    )
    return response.data.data
  },

  /**
   * Update batch status
   * Uses REST API endpoint (auth required)
   */
  updateBatchStatus: async (
    data: UpdateBatchStatusRequest
  ): Promise<TeaBatch> => {
    const response = await api.patch<ApiResponse<TeaBatch>>(
      API_ENDPOINTS.BATCHES.UPDATE_STATUS(data.batchId),
      { status: data.status }
    )
    return response.data.data
  },

  /**
   * Verify batch hash
   * Uses REST API endpoint (auth required)
   */
  verifyBatch: async (
    data: VerifyBatchRequest
  ): Promise<VerifyBatchResponse> => {
    const response = await api.post<ApiResponse<VerifyBatchResponse>>(
      API_ENDPOINTS.BATCHES.VERIFY(data.batchId),
      { hashInput: data.hashInput }
    )
    return response.data.data
  },

  /**
   * Get all batches
   * Uses REST API endpoint (auth required)
   */
  getAllBatches: async (): Promise<TeaBatch[]> => {
    const response = await api.get<ApiResponse<{ batches: TeaBatch[] }>>(
      API_ENDPOINTS.BATCHES.LIST
    )
    return response.data.data?.batches || []
  },
}

