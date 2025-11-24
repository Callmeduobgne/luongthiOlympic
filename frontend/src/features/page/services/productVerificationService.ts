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
 * distributed under the License is distributed on "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import api from '@shared/utils/api'

export interface VerifyProductByBlockhashRequest {
  blockhash: string
}

export interface VerifyProductByBlockhashResponse {
  isValid: boolean
  message: string
  transactionId?: string
  batchId?: string
  packageId?: string
}

/**
 * Verify product by blockhash/hash using backend endpoint
 * 
 * Flow xử lý khi người dùng nhập hash:
 * 1. Gọi POST /api/v1/teatrace/verify-by-hash với hash
 * 2. Backend sẽ:
 *    - Check cache (24h TTL)
 *    - Query database với index trên tx_id, chaincode_id
 *    - Query batches/packages nếu cần
 * 3. Trả về kết quả: isValid, message, transactionId, batchId, packageId
 */
export const verifyProductByBlockhash = async (
  blockhash: string
): Promise<VerifyProductByBlockhashResponse> => {
  try {
    const hash = blockhash.trim()
    
    if (!hash) {
      return {
        isValid: false,
        message: 'Vui lòng nhập hash hoặc transaction ID',
      }
    }

    // Call backend endpoint (with rate limiting: 10 req/min/IP)
    const response = await api.post<{ success: boolean; data: VerifyProductByBlockhashResponse }>(
      '/api/v1/teatrace/verify-by-hash',
      { hash }
    )
    
    if (response.data.success && response.data.data) {
      return response.data.data
    }

    // Fallback if response format is unexpected
    return {
      isValid: false,
      message: 'Không thể xác thực sản phẩm. Vui lòng thử lại sau.',
    }
  } catch (error: any) {
    console.error('Failed to verify product by blockhash:', error)
    
    // Handle rate limit error
    if (error.response?.status === 429) {
      return {
        isValid: false,
        message: 'Quá nhiều yêu cầu. Vui lòng thử lại sau 1 phút.',
      }
    }
    
    return {
      isValid: false,
      message: 'Không thể xác thực sản phẩm. Vui lòng thử lại sau.',
    }
  }
}

