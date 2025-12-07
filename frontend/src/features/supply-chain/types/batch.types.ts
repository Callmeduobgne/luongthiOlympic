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

