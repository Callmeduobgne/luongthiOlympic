/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

export interface QRCodeData {
  batchId?: string
  packageId?: string
  verificationHash?: string
  blockHash?: string
  verifyUrl?: string
  txId?: string
}

export interface QRCodeBase64Response {
  success: boolean
  data: {
    dataUri: string
    batchId?: string
    packageId?: string
  }
}

export interface QRCodeDataResponse {
  success: boolean
  data: QRCodeData
}

export const qrcodeApi = {
  /**
   * Get QR code as base64 data URI for batch
   * Returns data URI that can be used directly in <img src={dataUri} />
   */
  getBatchQRCodeBase64: async (batchId: string): Promise<string> => {
    const response = await api.get<QRCodeBase64Response>(
      API_ENDPOINTS.QRCODE.BATCH_BASE64(batchId)
    )
    return response.data.data.dataUri
  },

  /**
   * Get QR code data structure for batch (JSON)
   */
  getBatchQRCodeData: async (batchId: string): Promise<QRCodeData> => {
    const response = await api.get<QRCodeDataResponse>(
      API_ENDPOINTS.QRCODE.BATCH_DATA(batchId)
    )
    return response.data.data
  },

  /**
   * Get QR code PNG URL for batch
   * Returns URL that can be used in <img src={url} />
   */
  getBatchQRCodeUrl: (batchId: string): string => {
    return `${API_ENDPOINTS.QRCODE.BATCH_PNG(batchId)}`
  },

  /**
   * Get QR code as base64 data URI for package
   */
  getPackageQRCodeBase64: async (packageId: string): Promise<string> => {
    const response = await api.get<QRCodeBase64Response>(
      API_ENDPOINTS.QRCODE.PACKAGE_BASE64(packageId)
    )
    return response.data.data.dataUri
  },

  /**
   * Get QR code data structure for package (JSON)
   */
  getPackageQRCodeData: async (packageId: string): Promise<QRCodeData> => {
    const response = await api.get<QRCodeDataResponse>(
      API_ENDPOINTS.QRCODE.PACKAGE_DATA(packageId)
    )
    return response.data.data
  },

  /**
   * Get QR code PNG URL for package
   */
  getPackageQRCodeUrl: (packageId: string): string => {
    return `${API_ENDPOINTS.QRCODE.PACKAGE_PNG(packageId)}`
  },

  /**
   * Get QR code from transaction ID (auto-detect batch or package)
   * Returns PNG image URL
   */
  getTransactionQRCodeUrl: (txId: string): string => {
    return `${API_ENDPOINTS.QRCODE.TRANSACTION(txId)}`
  },
}

 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

export interface QRCodeData {
  batchId?: string
  packageId?: string
  verificationHash?: string
  blockHash?: string
  verifyUrl?: string
  txId?: string
}

export interface QRCodeBase64Response {
  success: boolean
  data: {
    dataUri: string
    batchId?: string
    packageId?: string
  }
}

export interface QRCodeDataResponse {
  success: boolean
  data: QRCodeData
}

export const qrcodeApi = {
  /**
   * Get QR code as base64 data URI for batch
   * Returns data URI that can be used directly in <img src={dataUri} />
   */
  getBatchQRCodeBase64: async (batchId: string): Promise<string> => {
    const response = await api.get<QRCodeBase64Response>(
      API_ENDPOINTS.QRCODE.BATCH_BASE64(batchId)
    )
    return response.data.data.dataUri
  },

  /**
   * Get QR code data structure for batch (JSON)
   */
  getBatchQRCodeData: async (batchId: string): Promise<QRCodeData> => {
    const response = await api.get<QRCodeDataResponse>(
      API_ENDPOINTS.QRCODE.BATCH_DATA(batchId)
    )
    return response.data.data
  },

  /**
   * Get QR code PNG URL for batch
   * Returns URL that can be used in <img src={url} />
   */
  getBatchQRCodeUrl: (batchId: string): string => {
    return `${API_ENDPOINTS.QRCODE.BATCH_PNG(batchId)}`
  },

  /**
   * Get QR code as base64 data URI for package
   */
  getPackageQRCodeBase64: async (packageId: string): Promise<string> => {
    const response = await api.get<QRCodeBase64Response>(
      API_ENDPOINTS.QRCODE.PACKAGE_BASE64(packageId)
    )
    return response.data.data.dataUri
  },

  /**
   * Get QR code data structure for package (JSON)
   */
  getPackageQRCodeData: async (packageId: string): Promise<QRCodeData> => {
    const response = await api.get<QRCodeDataResponse>(
      API_ENDPOINTS.QRCODE.PACKAGE_DATA(packageId)
    )
    return response.data.data
  },

  /**
   * Get QR code PNG URL for package
   */
  getPackageQRCodeUrl: (packageId: string): string => {
    return `${API_ENDPOINTS.QRCODE.PACKAGE_PNG(packageId)}`
  },

  /**
   * Get QR code from transaction ID (auto-detect batch or package)
   * Returns PNG image URL
   */
  getTransactionQRCodeUrl: (txId: string): string => {
    return `${API_ENDPOINTS.QRCODE.TRANSACTION(txId)}`
  },
}


