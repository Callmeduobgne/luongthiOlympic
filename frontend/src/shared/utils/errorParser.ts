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

/**
 * Utility functions for parsing API errors
 */

interface ApiErrorResponse {
  error?: {
    code?: string
    message?: string
    detail?: string
  }
  message?: string
}

interface AxiosError {
  response?: {
    data?: ApiErrorResponse
    status?: number
  }
  message?: string
}

/**
 * Extract user-friendly error message from API error
 */
export const extractErrorMessage = (error: unknown): string => {
  const axiosError = error as AxiosError

  // If it's an axios error with response
  if (axiosError.response?.data) {
    const data = axiosError.response.data

    // Try to get error from error object
    if (data.error) {
      // Prefer detail if available (more specific)
      if (data.error.detail) {
        // Check if it's a package format error
        if (data.error.detail.includes('could not parse') || 
            data.error.detail.includes('tar entry') ||
            data.error.detail.includes('Invalid chaincode package format')) {
          return extractPackageFormatError(data.error.detail)
        }
        // Check if it's a peer command error
        if (data.error.detail.includes('peer command failed')) {
          return extractPeerCommandError(data.error.detail)
        }
        return data.error.detail
      }
      // Fallback to message
      if (data.error.message) {
        return data.error.message
      }
    }

    // Fallback to top-level message
    if (data.message) {
      return data.message
    }
  }

  // If it's an Error object
  if (error instanceof Error) {
    // Check if message contains useful info
    const msg = error.message
    if (msg.includes('Invalid chaincode package format') || 
        msg.includes('could not parse') ||
        msg.includes('tar entry')) {
      return extractPackageFormatError(msg)
    }
    return msg
  }

  // Default fallback
  return 'Đã xảy ra lỗi. Vui lòng thử lại.'
}

/**
 * Extract package format error message
 */
const extractPackageFormatError = (errorDetail: string): string => {
  // Look for "could not parse" or "tar entry" errors
  if (errorDetail.includes('tar entry')) {
    const match = errorDetail.match(/tar entry ([^ ]+) is not a regular file/)
    if (match) {
      return `Package không đúng format: entry "${match[1]}" không phải là file hợp lệ. Vui lòng kiểm tra lại cấu trúc package.`
    }
    return 'Package không đúng format. Vui lòng đảm bảo package được tạo bằng lệnh `peer lifecycle chaincode package`.'
  }

  if (errorDetail.includes('could not parse')) {
    return 'Không thể parse chaincode package. Package phải chứa metadata.json và code.tar.gz.'
  }

  if (errorDetail.includes('Invalid chaincode package format')) {
    // Extract the actual error after the prefix
    const actualError = errorDetail.replace('Invalid chaincode package format: ', '')
    return `Package không đúng format: ${actualError}`
  }

  return errorDetail
}

/**
 * Extract peer command error message
 */
const extractPeerCommandError = (errorDetail: string): string => {
  // Look for "Error:" in peer output
  if (errorDetail.includes('Error:')) {
    const errorMatch = errorDetail.match(/Error: (.+?)(\n|$)/)
    if (errorMatch) {
      let errorMsg = errorMatch[1]
      // Remove redundant prefixes
      errorMsg = errorMsg.replace(/^chaincode install failed with status: \d+ - /, '')
      errorMsg = errorMsg.replace(/^failed to invoke backing implementation of 'InstallChaincode': /, '')
      return errorMsg
    }
  }
  return errorDetail
}

/**
 * Get error code from API error
 */
export const extractErrorCode = (error: unknown): string | undefined => {
  const axiosError = error as AxiosError
  if (axiosError.response?.data?.error?.code) {
    return axiosError.response.data.error.code
  }
  return undefined
}

