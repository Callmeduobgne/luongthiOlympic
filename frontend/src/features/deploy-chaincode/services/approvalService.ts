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

export interface ApprovalRequest {
  id: string
  chaincode_version_id: string
  operation: 'install' | 'approve' | 'commit'
  status: 'pending' | 'approved' | 'rejected' | 'cancelled'
  requested_by: string
  requested_at: string
  approved_at?: string
  rejected_at?: string
  expires_at?: string
  reason?: string
  metadata?: Record<string, any>
}

export interface ApprovalVote {
  id: string
  request_id: string
  voter_id: string
  vote_type: 'approve' | 'reject'
  voted_at: string
  reason?: string
}

export interface CreateApprovalRequestRequest {
  chaincode_version_id: string
  operation: 'install' | 'approve' | 'commit'
  reason?: string
  metadata?: Record<string, any>
}

export interface VoteRequest {
  approval_request_id: string
  vote: 'approve' | 'reject'
  comment?: string
}

export const approvalService = {
  /**
   * Create an approval request
   */
  async createRequest(request: CreateApprovalRequestRequest): Promise<ApprovalRequest> {
    const response = await api.post<{ success: boolean; data: ApprovalRequest }>(
      API_ENDPOINTS.CHAINCODE.APPROVAL.CREATE_REQUEST,
      request
    )
    return response.data.data
  },

  /**
   * Vote on an approval request
   */
  async vote(request: VoteRequest): Promise<void> {
    await api.post(API_ENDPOINTS.CHAINCODE.APPROVAL.VOTE, request)
  },

  /**
   * Get approval request by ID
   */
  async getRequest(id: string): Promise<ApprovalRequest> {
    const response = await api.get<{ success: boolean; data: ApprovalRequest }>(
      API_ENDPOINTS.CHAINCODE.APPROVAL.GET_REQUEST(id)
    )
    return response.data.data
  },

  /**
   * List approval requests
   */
  async listRequests(filters?: {
    status?: string
    operation?: string
    chaincode_version_id?: string
  }): Promise<ApprovalRequest[]> {
    const params = new URLSearchParams()
    if (filters?.status) params.append('status', filters.status)
    if (filters?.operation) params.append('operation', filters.operation)
    if (filters?.chaincode_version_id) params.append('chaincode_version_id', filters.chaincode_version_id)

    const queryString = params.toString()
    const url = queryString
      ? `${API_ENDPOINTS.CHAINCODE.APPROVAL.LIST_REQUESTS}?${queryString}`
      : API_ENDPOINTS.CHAINCODE.APPROVAL.LIST_REQUESTS

    const response = await api.get<{ success: boolean; data: ApprovalRequest[] }>(url)
    return response.data.data || []
  },
}

