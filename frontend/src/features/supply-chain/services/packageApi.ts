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

import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import type {
    TeaPackage,
    CreatePackageRequest,
    UpdatePackageStatusRequest,
    VerifyPackageRequest,
    VerifyPackageResponse,
    GetPackagesByBatchRequest,
    GetPackagesByBatchResponse,
} from '../types/package.types'

type ApiResponse<T> = {
    success: boolean
    data: T
    message?: string
}

export const packageApi = {
    /**
     * Get package information
     * Uses REST API endpoint (public, no auth required)
     */
    getPackageInfo: async (packageId: string): Promise<TeaPackage> => {
        const response = await api.get<ApiResponse<{ result: TeaPackage }>>(
            API_ENDPOINTS.PACKAGES.GET(packageId)
        )
        return response.data.data.result
    },

    /**
     * Create new package
     * Uses REST API endpoint (auth required)
     */
    createPackage: async (data: CreatePackageRequest): Promise<TeaPackage> => {
        const response = await api.post<ApiResponse<TeaPackage>>(
            API_ENDPOINTS.PACKAGES.CREATE,
            data
        )
        return response.data.data
    },

    /**
     * Update package status
     * Uses REST API endpoint (auth required)
     */
    updatePackageStatus: async (
        data: UpdatePackageStatusRequest
    ): Promise<TeaPackage> => {
        const response = await api.patch<ApiResponse<TeaPackage>>(
            API_ENDPOINTS.PACKAGES.UPDATE_STATUS(data.packageId),
            { status: data.status }
        )
        return response.data.data
    },

    /**
     * Verify package hash
     * Uses REST API endpoint (auth required)
     */
    verifyPackage: async (
        data: VerifyPackageRequest
    ): Promise<VerifyPackageResponse> => {
        const response = await api.post<ApiResponse<VerifyPackageResponse>>(
            API_ENDPOINTS.PACKAGES.VERIFY(data.packageId),
            { blockHash: data.blockHash }
        )
        return response.data.data
    },

    /**
     * Get all packages
     * Uses REST API endpoint (auth required)
     */
    getAllPackages: async (): Promise<TeaPackage[]> => {
        const response = await api.get<ApiResponse<{ packages: TeaPackage[] }>>(
            API_ENDPOINTS.PACKAGES.LIST
        )
        return response.data.data?.packages || []
    },

    /**
     * Get packages by batch ID
     * Uses REST API endpoint (auth required)
     */
    getPackagesByBatch: async (
        params: GetPackagesByBatchRequest
    ): Promise<GetPackagesByBatchResponse> => {
        const { batchId, limit = 100, offset = 0 } = params
        const response = await api.get<ApiResponse<GetPackagesByBatchResponse>>(
            API_ENDPOINTS.PACKAGES.BY_BATCH(batchId),
            { params: { limit, offset } }
        )
        return response.data.data
    },
}
