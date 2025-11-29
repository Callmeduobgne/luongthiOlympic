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

import type { BatchStatus } from '@shared/utils/constants'

/**
 * Tea Batch types
 */
export interface TeaBatch {
  batchId: string
  farmLocation: string
  harvestDate: string
  processingInfo: string
  qualityCert: string
  hashValue: string
  owner: string
  timestamp: string
  status: BatchStatus
}

export interface CreateBatchRequest {
  batch_id: string
  farm_name: string
  harvest_date: string
  certification: string
  certificate_id: string
}

export interface UpdateBatchStatusRequest {
  batchId: string
  status: BatchStatus
}

export interface VerifyBatchRequest {
  batchId: string
  hashInput: string
}

export interface VerifyBatchResponse {
  isValid: boolean
  batch: TeaBatch
}

export interface BatchTimelineEvent {
  timestamp: string
  status: BatchStatus
  description: string
  actor?: string
}

