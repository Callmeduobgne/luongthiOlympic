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
 * 1. Gọi GET /api/v1/blockchain/verify-transaction/{txid} với transaction ID (hash)
 * 2. Backend sẽ query TRỰC TIẾP từ blockchain network qua Gateway (KHÔNG từ database)
 * 3. Trả về kết quả: isValid, message, transactionId
 * 
 * NOTE: Endpoint này query từ blockchain network, không từ database
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

    // Call backend endpoint to verify from blockchain network (not from database)
    const response = await api.get<{ 
      success: boolean
      data: {
        is_valid: boolean
        message: string
        tx_id: string
        verified_from?: string
      }
    }>(
      `/api/v1/blockchain/verify-transaction/${encodeURIComponent(hash)}`
    )
    
    if (response.data.success && response.data.data) {
      return {
        isValid: response.data.data.is_valid,
        message: response.data.data.message,
        transactionId: response.data.data.tx_id,
      }
    }

    // Fallback if response format is unexpected
    return {
      isValid: false,
      message: 'Không thể xác thực sản phẩm. Vui lòng thử lại sau.',
    }
  } catch (error: any) {
    console.error('Failed to verify product by blockhash:', error)
    
    // Handle 404 - transaction not found in blockchain network
    if (error.response?.status === 404 || error.response?.status === 200) {
      // Backend returns 200 with is_valid: false if transaction not found
      if (error.response?.data?.data?.is_valid === false) {
        return {
          isValid: false,
          message: error.response.data.data.message || 'Transaction không tồn tại trong blockchain network',
        }
      }
    }
    
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

