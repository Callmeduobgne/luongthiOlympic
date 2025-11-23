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

export interface InstalledChaincode {
  packageId: string
  label: string
  chaincode: {
    name: string
    version: string
    path: string
  }
}

export interface CommittedChaincode {
  name: string
  version: string
  sequence: number
  endorsementPlugin: string
  validationPlugin: string
  initRequired: boolean
  collections?: string[]
  approvedOrganizations?: string[]
}

export interface InstallChaincodeRequest {
  packagePath: string
  label?: string
}

export interface UploadPackageResponse {
  filePath: string
  filename: string
  size: string
}

export interface ApproveChaincodeRequest {
  channelName: string
  name: string
  version: string
  sequence: number
  packageId?: string
  initRequired?: boolean
  endorsementPlugin?: string
  validationPlugin?: string
  collections?: string[]
}

export interface CommitChaincodeRequest {
  channelName: string
  name: string
  version: string
  sequence: number
  initRequired?: boolean
  endorsementPlugin?: string
  validationPlugin?: string
  collections?: string[]
}

export interface InvokeChaincodeRequest {
  function: string
  args: string[]
  transient?: Record<string, string>
}

export interface QueryChaincodeRequest {
  function: string
  args: string[]
}

export const chaincodeService = {
  /**
   * List installed chaincodes
   */
  async listInstalled(peer?: string): Promise<InstalledChaincode[]> {
    try {
      const params = peer ? `?peer=${peer}` : ''
      const response = await api.get<{ success: boolean; data: InstalledChaincode[] }>(
        `${API_ENDPOINTS.CHAINCODE.INSTALLED}${params}`
      )
       
      // Ensure we always return an array
      if (Array.isArray(response.data?.data)) {
        return response.data.data
      }
      
      // Log warning in development only
      if (import.meta.env.DEV) {
        console.warn('[chaincodeService] Invalid response format for installed chaincodes:', response.data)
      }
      return []
    } catch (error: any) {
      if (import.meta.env.DEV) {
        console.error('[chaincodeService] Error loading installed chaincodes:', error)
        // Log detailed error for debugging
        if (error?.response?.data) {
          console.error('[chaincodeService] Error details:', error.response.data)
        }
      }
      // Re-throw error so UI can display it
      throw error
    }
  },

  /**
   * List committed chaincodes on a channel
   */
  async listCommitted(channel?: string): Promise<CommittedChaincode[]> {
    try {
      const params = channel ? `?channel=${channel}` : ''
      const response = await api.get<{ success: boolean; data: CommittedChaincode[] }>(
        `${API_ENDPOINTS.CHAINCODE.COMMITTED}${params}`
      )
      // Ensure we always return an array and handle approvedOrganizations safely
      if (Array.isArray(response.data?.data)) {
        return response.data.data.map(cc => ({
          ...cc,
          approvedOrganizations: Array.isArray(cc.approvedOrganizations) ? cc.approvedOrganizations : [],
        }))
      }
      return []
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('Error loading committed chaincodes:', error)
      }
      return [] // Return empty array on error
    }
  },

  /**
   * Get committed chaincode info
   */
  async getCommittedInfo(name: string, channel?: string): Promise<CommittedChaincode> {
    try {
      const params = channel ? `?channel=${channel}` : ''
      const response = await api.get<{ success: boolean; data: CommittedChaincode }>(
        `${API_ENDPOINTS.CHAINCODE.COMMITTED_INFO(name)}${params}`
      )
      return response.data.data
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[chaincodeService] Error getting committed chaincode info:', error)
      }
      throw error
    }
  },

  /**
   * Upload chaincode package file
   */
  async uploadPackage(file: File): Promise<UploadPackageResponse> {
    const formData = new FormData()
    formData.append('package', file)

    const response = await api.post<{ success: boolean; data: UploadPackageResponse }>(
      API_ENDPOINTS.CHAINCODE.UPLOAD,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    )
    return response.data.data
  },

  /**
   * Install chaincode
   */
  async install(request: InstallChaincodeRequest): Promise<{ packageId: string }> {
    const response = await api.post<{ success: boolean; data: { packageId: string } }>(
      API_ENDPOINTS.CHAINCODE.INSTALL,
      request
    )
    return response.data.data
  },

  /**
   * Approve chaincode definition
   */
  async approve(request: ApproveChaincodeRequest): Promise<void> {
    await api.post(API_ENDPOINTS.CHAINCODE.APPROVE, request)
  },

  /**
   * Commit chaincode definition
   */
  async commit(request: CommitChaincodeRequest): Promise<void> {
    await api.post(API_ENDPOINTS.CHAINCODE.COMMIT, request)
  },

  /**
   * Get latest version ID for a chaincode (for Phase 4 features)
   */
  async getLatestVersionId(chaincodeName: string, channel: string): Promise<string | null> {
    try {
      const params = `?channel=${channel}`
      const response = await api.get<{ success: boolean; data: { version_id: string } }>(
        `${API_ENDPOINTS.CHAINCODE.VERSION.GET_LATEST(chaincodeName)}${params}`
      )
      return response.data.data.version_id || null
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('Failed to get latest version ID:', error)
      }
      return null
    }
  },

  /**
   * Invoke chaincode function (submit transaction)
   * User và admin đều có thể invoke
   */
  async invoke(
    channel: string,
    chaincodeName: string,
    request: InvokeChaincodeRequest
  ): Promise<any> {
    const response = await api.post<{ success: boolean; data: any }>(
      API_ENDPOINTS.CHAINCODE.INVOKE(channel, chaincodeName),
      request
    )
    return response.data.data
  },

  /**
   * Query chaincode function (read-only)
   * User và admin đều có thể query
   */
  async query(
    channel: string,
    chaincodeName: string,
    request: QueryChaincodeRequest
  ): Promise<any> {
    const response = await api.post<{ success: boolean; data: any }>(
      API_ENDPOINTS.CHAINCODE.QUERY(channel, chaincodeName),
      request
    )
    return response.data.data
  },
}

