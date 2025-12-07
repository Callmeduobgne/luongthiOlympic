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
