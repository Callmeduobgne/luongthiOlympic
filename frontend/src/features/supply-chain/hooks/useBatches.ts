import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { batchApi } from '../services/batchApi'
import type {
  CreateBatchRequest,
  UpdateBatchStatusRequest,
  VerifyBatchRequest,
} from '../types/batch.types'
import toast from 'react-hot-toast'
import type { AxiosError } from 'axios'

/**
 * Query keys
 */
export const batchKeys = {
  all: ['batches'] as const,
  detail: (id: string) => [...batchKeys.all, 'detail', id] as const,
}

/**
 * Get batch information
 */
export function useBatch(batchId: string) {
  return useQuery({
    queryKey: batchKeys.detail(batchId),
    queryFn: () => batchApi.getBatchInfo(batchId),
    enabled: !!batchId,
    retry: 1,
  })
}

/**
 * Create batch mutation
 */
export function useCreateBatch() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateBatchRequest) => batchApi.createBatch(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: batchKeys.all })
      toast.success(`Batch ${data.batchId} created successfully`)
    },
    onError: (error: AxiosError<{ message?: string }>) => {
      toast.error(
        error.response?.data?.message || 'Failed to create batch'
      )
    },
  })
}

/**
 * Update batch status mutation
 */
export function useUpdateBatchStatus() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: UpdateBatchStatusRequest) =>
      batchApi.updateBatchStatus(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: batchKeys.detail(data.batchId) })
      queryClient.invalidateQueries({ queryKey: batchKeys.all })
      toast.success(`Batch status updated to ${data.status}`)
    },
    onError: (error: AxiosError<{ message?: string }>) => {
      toast.error(
        error.response?.data?.message || 'Failed to update batch status'
      )
    },
  })
}

/**
 * Verify batch mutation
 */
export function useVerifyBatch() {
  return useMutation({
    mutationFn: (data: VerifyBatchRequest) => batchApi.verifyBatch(data),
    onSuccess: (data) => {
      if (data.isValid) {
        toast.success('Batch verification successful')
      } else {
        toast.error('Batch verification failed')
      }
    },
    onError: (error: AxiosError<{ message?: string }>) => {
      toast.error(
        error.response?.data?.message || 'Failed to verify batch'
      )
    },
  })
}

/**
 * Get all batches
 */
export function useAllBatches() {
  return useQuery({
    queryKey: batchKeys.all,
    queryFn: () => batchApi.getAllBatches(),
    retry: 1,
  })
}
