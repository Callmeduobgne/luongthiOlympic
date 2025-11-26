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

import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, XCircle, Shield, Clock, Hash, Layers } from 'lucide-react'
import { getPackageQRData, verifyWithMerkleProof } from '@/services/verifyService'
import type { QRCodeData, VerifyResult } from '@/services/verifyService'

export default function VerifyPackage() {
    const { packageId } = useParams<{ packageId: string }>()
    const [showProofDetails, setShowProofDetails] = useState(false)

    // Get QR data with Merkle proof
    const { data: qrData, isLoading: qrLoading, error: qrError } = useQuery<QRCodeData>({
        queryKey: ['qr-data', packageId],
        queryFn: () => getPackageQRData(packageId!),
        enabled: !!packageId,
    })

    // Verify with Merkle proof
    const { data: verifyResult, isLoading: verifyLoading, error: verifyError } = useQuery<VerifyResult>({
        queryKey: ['verify', qrData?.txId],
        queryFn: () => verifyWithMerkleProof(qrData!),
        enabled: !!qrData,
    })

    if (qrLoading || verifyLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mb-4"></div>
                    <p className="text-gray-600">Đang xác thực sản phẩm...</p>
                </div>
            </div>
        )
    }

    if (qrError || verifyError) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-red-50 to-pink-100 p-4">
                <div className="max-w-md w-full bg-white rounded-2xl shadow-xl p-8">
                    <div className="text-center">
                        <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
                        <h2 className="text-2xl font-bold text-gray-900 mb-2">Lỗi Xác Thực</h2>
                        <p className="text-gray-600">
                            {(qrError as Error)?.message || (verifyError as Error)?.message || 'Không thể xác thực sản phẩm'}
                        </p>
                    </div>
                </div>
            </div>
        )
    }

    const isValid = verifyResult?.is_valid
    const isMerkleProof = verifyResult?.verification_method === 'merkle_proof'

    return (
        <div className={`min-h-screen flex items-center justify-center p-4 bg-gradient-to-br ${isValid ? 'from-green-50 to-emerald-100' : 'from-red-50 to-pink-100'
            }`}>
            <div className="max-w-2xl w-full bg-white rounded-2xl shadow-2xl overflow-hidden">
                {/* Header */}
                <div className={`p-8 ${isValid ? 'bg-gradient-to-r from-green-500 to-emerald-600' : 'bg-gradient-to-r from-red-500 to-pink-600'}`}>
                    <div className="text-center text-white">
                        {isValid ? (
                            <CheckCircle2 className="w-20 h-20 mx-auto mb-4" />
                        ) : (
                            <XCircle className="w-20 h-20 mx-auto mb-4" />
                        )}
                        <h1 className="text-3xl font-bold mb-2">
                            {isValid ? 'Sản Phẩm Chính Hãng' : 'Sản Phẩm Không Hợp Lệ'}
                        </h1>
                        <p className="text-white/90">{verifyResult?.message}</p>
                    </div>
                </div>

                {/* Content */}
                <div className="p-8 space-y-6">
                    {/* Product Info */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between py-3 border-b border-gray-200">
                            <span className="text-gray-600 flex items-center gap-2">
                                <Hash className="w-4 h-4" />
                                Package ID
                            </span>
                            <span className="font-mono font-semibold text-gray-900">{packageId}</span>
                        </div>

                        {qrData?.txId && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Layers className="w-4 h-4" />
                                    Transaction
                                </span>
                                <span className="font-mono text-xs text-gray-700 truncate max-w-xs">
                                    {qrData.txId}
                                </span>
                            </div>
                        )}

                        {qrData?.blockNumber && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Layers className="w-4 h-4" />
                                    Block Number
                                </span>
                                <span className="font-mono font-semibold text-gray-900">
                                    {qrData.blockNumber}
                                </span>
                            </div>
                        )}

                        {verifyResult?.verified_at && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Clock className="w-4 h-4" />
                                    Verified At
                                </span>
                                <span className="text-gray-900">
                                    {new Date(verifyResult.verified_at).toLocaleString('vi-VN')}
                                </span>
                            </div>
                        )}

                        <div className="flex items-center justify-between py-3">
                            <span className="text-gray-600 flex items-center gap-2">
                                <Shield className="w-4 h-4" />
                                Verification Method
                            </span>
                            <span className={`px-3 py-1 rounded-full text-sm font-semibold ${isMerkleProof
                                    ? 'bg-purple-100 text-purple-800'
                                    : 'bg-blue-100 text-blue-800'
                                }`}>
                                {isMerkleProof ? '🔐 Merkle Proof (Cryptographic)' : '⛓️ Blockchain Query'}
                            </span>
                        </div>
                    </div>

                    {/* Merkle Proof Details */}
                    {qrData?.merkleProof && qrData.merkleProof.length > 0 && (
                        <div className="bg-gray-50 rounded-lg p-4">
                            <button
                                onClick={() => setShowProofDetails(!showProofDetails)}
                                className="w-full flex items-center justify-between text-left"
                            >
                                <span className="font-semibold text-gray-900">
                                    Merkle Proof Details
                                </span>
                                <span className="text-gray-500">
                                    {showProofDetails ? '▼' : '▶'}
                                </span>
                            </button>

                            {showProofDetails && (
                                <div className="mt-4 space-y-3">
                                    <div>
                                        <div className="text-xs font-semibold text-gray-600 mb-1">Merkle Root:</div>
                                        <div className="font-mono text-xs bg-white p-2 rounded border border-gray-200 break-all">
                                            {qrData.merkleRoot}
                                        </div>
                                    </div>

                                    <div>
                                        <div className="text-xs font-semibold text-gray-600 mb-1">
                                            Proof Path ({qrData.merkleProof.length} steps):
                                        </div>
                                        <div className="space-y-2">
                                            {qrData.merkleProof.map((step, i) => (
                                                <div key={i} className="bg-white p-2 rounded border border-gray-200">
                                                    <div className="flex items-center gap-2 mb-1">
                                                        <span className="text-xs font-semibold text-gray-600">
                                                            Step {i + 1}:
                                                        </span>
                                                        <span className={`text-xs px-2 py-0.5 rounded ${step.position === 'left'
                                                                ? 'bg-blue-100 text-blue-800'
                                                                : 'bg-green-100 text-green-800'
                                                            }`}>
                                                            {step.position}
                                                        </span>
                                                    </div>
                                                    <div className="font-mono text-xs text-gray-700 break-all">
                                                        {step.hash}
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}

                    {/* Info Box */}
                    {isValid && (
                        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                            <p className="text-sm text-green-800">
                                ✅ Sản phẩm này đã được xác thực thành công từ blockchain.
                                {isMerkleProof && ' Xác thực sử dụng Merkle proof - đảm bảo tính toàn vẹn dữ liệu.'}
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
