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
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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
