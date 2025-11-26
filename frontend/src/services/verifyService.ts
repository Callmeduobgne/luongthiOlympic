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

import { API_ENDPOINTS } from '@/shared/config/api.config'

// Merkle proof step structure
export interface ProofStep {
    hash: string
    position: 'left' | 'right'
}

// QR Code data structure with Merkle proof
export interface QRCodeData {
    packageId?: string
    batchId?: string
    blockHash?: string
    verificationHash?: string
    txId: string
    blockNumber?: number
    merkleProof?: ProofStep[]
    merkleRoot?: string
    verifyUrl: string
    timestamp?: number
}

// Verification result from backend
export interface VerifyResult {
    is_valid: boolean
    message: string
    transaction_id?: string
    batch_id?: string
    package_id?: string
    block_number?: number
    entity_type?: 'transaction' | 'batch' | 'package'
    verified_at: string
    verification_method?: 'merkle_proof' | 'blockchain_query'
}

// Verify product using Merkle proof (fast path)
export const verifyWithMerkleProof = async (
    qrData: QRCodeData
): Promise<VerifyResult> => {
    const response = await fetch(API_ENDPOINTS.BATCHES.VERIFY_BY_HASH || '/api/v1/teatrace/verify-by-hash', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            hash: qrData.txId || qrData.verificationHash || qrData.blockHash,
            merkleProof: qrData.merkleProof,
            merkleRoot: qrData.merkleRoot,
            blockNumber: qrData.blockNumber,
        }),
    })

    if (!response.ok) {
        throw new Error(`Verification failed: ${response.statusText}`)
    }

    const data = await response.json()
    return data.data
}

// Verify product by hash only (legacy path)
export const verifyByHash = async (hash: string): Promise<VerifyResult> => {
    const response = await fetch(API_ENDPOINTS.BATCHES.VERIFY_BY_HASH || '/api/v1/teatrace/verify-by-hash', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            hash,
        }),
    })

    if (!response.ok) {
        throw new Error(`Verification failed: ${response.statusText}`)
    }

    const data = await response.json()
    return data.data
}

// Get QR code data for a package
export const getPackageQRData = async (packageId: string): Promise<QRCodeData> => {
    const response = await fetch(`/api/v1/qrcode/packages/${packageId}/data`)

    if (!response.ok) {
        throw new Error(`Failed to get QR data: ${response.statusText}`)
    }

    return response.json()
}

// Get QR code data for a batch
export const getBatchQRData = async (batchId: string): Promise<QRCodeData> => {
    const response = await fetch(`/api/v1/qrcode/batches/${batchId}/data`)

    if (!response.ok) {
        throw new Error(`Failed to get QR data: ${response.statusText}`)
    }

    return response.json()
}
