import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { packageApi } from '../services/packageApi'
import type {
    CreatePackageRequest,
    UpdatePackageStatusRequest,
    VerifyPackageRequest,
    GetPackagesByBatchRequest,
} from '../types/package.types'
import toast from 'react-hot-toast'
import type { AxiosError } from 'axios'

/**
 * Query keys
 */
export const packageKeys = {
    all: ['packages'] as const,
    detail: (id: string) => [...packageKeys.all, 'detail', id] as const,
    byBatch: (batchId: string) => [...packageKeys.all, 'byBatch', batchId] as const,
}

/**
 * Get package information
 */
export function usePackage(packageId: string) {
    return useQuery({
        queryKey: packageKeys.detail(packageId),
        queryFn: () => packageApi.getPackageInfo(packageId),
        enabled: !!packageId,
        retry: 1,
    })
}

/**
 * Get packages by batch ID
 */
export function usePackagesByBatch(params: GetPackagesByBatchRequest) {
    return useQuery({
        queryKey: packageKeys.byBatch(params.batchId),
        queryFn: () => packageApi.getPackagesByBatch(params),
        enabled: !!params.batchId,
        retry: 1,
    })
}

/**
 * Create package mutation
 */
export function useCreatePackage() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (data: CreatePackageRequest) => packageApi.createPackage(data),
        onSuccess: (data) => {
            // Invalidate all packages and packages by batch
            queryClient.invalidateQueries({ queryKey: packageKeys.all })
            queryClient.invalidateQueries({ queryKey: packageKeys.byBatch(data.batchId) })
            toast.success(`Package ${data.packageId} created successfully`)
        },
        onError: (error: AxiosError<{ message?: string }>) => {
            toast.error(
                error.response?.data?.message || 'Failed to create package'
            )
        },
    })
}

/**
 * Update package status mutation
 */
export function useUpdatePackageStatus() {
    const queryClient = useQueryClient()

    return useMutation({
        mutationFn: (data: UpdatePackageStatusRequest) =>
            packageApi.updatePackageStatus(data),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: packageKeys.detail(data.packageId) })
            queryClient.invalidateQueries({ queryKey: packageKeys.all })
            queryClient.invalidateQueries({ queryKey: packageKeys.byBatch(data.batchId) })
            toast.success(`Package status updated to ${data.status}`)
        },
        onError: (error: AxiosError<{ message?: string }>) => {
            toast.error(
                error.response?.data?.message || 'Failed to update package status'
            )
        },
    })
}

/**
 * Verify package mutation
 */
export function useVerifyPackage() {
    return useMutation({
        mutationFn: (data: VerifyPackageRequest) => packageApi.verifyPackage(data),
        onSuccess: (data) => {
            if (data.isValid) {
                toast.success('Package verification successful')
            } else {
                toast.error('Package verification failed')
            }
        },
        onError: (error: AxiosError<{ message?: string }>) => {
            toast.error(
                error.response?.data?.message || 'Failed to verify package'
            )
        },
    })
}

/**
 * Get all packages
 */
export function useAllPackages() {
    return useQuery({
        queryKey: packageKeys.all,
        queryFn: () => packageApi.getAllPackages(),
        retry: 1,
    })
}
