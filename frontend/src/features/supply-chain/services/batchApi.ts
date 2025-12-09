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
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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

