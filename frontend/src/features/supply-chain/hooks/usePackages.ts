/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

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
