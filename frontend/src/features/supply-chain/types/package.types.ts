/**
 * Package status types
 */
export type PackageStatus = 'CREATED' | 'VERIFIED' | 'SOLD' | 'EXPIRED'

/**
 * Tea Package types
 */
export interface TeaPackage {
    packageId: string
    batchId: string
    blockHash: string
    hashVersion: 'v1' | 'v2'
    txId: string
    weight: number
    productionDate: string
    expiryDate?: string
    qrCode?: string
    status: PackageStatus
    owner: string
    timestamp: string
}

/**
 * Create package request
 */
export interface CreatePackageRequest {
    package_id: string
    batch_id: string
    weight: number
    production_date: string
    expiry_date?: string
    qr_code?: string
}

/**
 * Update package status request
 */
export interface UpdatePackageStatusRequest {
    packageId: string
    status: PackageStatus
}

/**
 * Verify package request
 */
export interface VerifyPackageRequest {
    packageId: string
    blockHash?: string
}

/**
 * Verify package response
 */
export interface VerifyPackageResponse {
    isValid: boolean
    package: TeaPackage
}

/**
 * Get packages by batch request
 */
export interface GetPackagesByBatchRequest {
    batchId: string
    limit?: number
    offset?: number
}

/**
 * Get packages by batch response
 */
export interface GetPackagesByBatchResponse {
    packages: TeaPackage[]
    total: number
}
