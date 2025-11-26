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

import { useSearchParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, XCircle, Shield, Clock, Hash, Layers, ArrowLeft, Package, Calendar, Scale } from 'lucide-react'
import { verifyByHash } from '@/services/verifyService'
import type { VerifyResult } from '@/services/verifyService'

// Extend VerifyResult to include product_details if not already present in the type definition
interface ExtendedVerifyResult extends VerifyResult {
    product_details?: {
        farm_location?: string
        harvest_date?: string
        production_date?: string
        expiry_date?: string
        weight?: number
        processing_info?: string
        quality_cert?: string
        status?: string
    }
}

export default function VerifyHashPage() {
    const [searchParams] = useSearchParams()
    const hash = searchParams.get('hash')

    // Verify by hash
    const { data: verifyResult, isLoading, error } = useQuery<ExtendedVerifyResult>({
        queryKey: ['verify-hash', hash],
        queryFn: () => verifyByHash(hash!),
        enabled: !!hash,
        retry: false
    })

    if (!hash) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
                <div className="max-w-md w-full bg-white rounded-2xl shadow-xl p-8 text-center">
                    <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
                    <h2 className="text-2xl font-bold text-gray-900 mb-2">Thiếu Thông Tin</h2>
                    <p className="text-gray-600 mb-6">Vui lòng cung cấp hash hoặc transaction ID để xác thực.</p>
                    <Link to="/" className="inline-flex items-center text-green-600 hover:text-green-700 font-medium">
                        <ArrowLeft className="w-4 h-4 mr-2" />
                        Quay về trang chủ
                    </Link>
                </div>
            </div>
        )
    }

    if (isLoading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
                <div className="text-center">
                    <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mb-4"></div>
                    <p className="text-gray-600">Đang truy xuất dữ liệu blockchain...</p>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-red-50 to-pink-100 p-4">
                <div className="max-w-md w-full bg-white rounded-2xl shadow-xl p-8">
                    <div className="text-center">
                        <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
                        <h2 className="text-2xl font-bold text-gray-900 mb-2">Lỗi Xác Thực</h2>
                        <p className="text-gray-600 mb-6">
                            {(error as Error)?.message || 'Không thể xác thực hash này. Vui lòng kiểm tra lại.'}
                        </p>
                        <Link to="/" className="inline-flex items-center text-gray-600 hover:text-gray-900 font-medium">
                            <ArrowLeft className="w-4 h-4 mr-2" />
                            Quay về trang chủ
                        </Link>
                    </div>
                </div>
            </div>
        )
    }

    const isValid = verifyResult?.is_valid
    const productDetails = verifyResult?.product_details

    return (
        <div className={`min-h-screen flex items-center justify-center p-4 bg-gradient-to-br ${isValid ? 'from-green-50 to-emerald-100' : 'from-red-50 to-pink-100'
            }`}>
            <div className="max-w-2xl w-full bg-white rounded-2xl shadow-2xl overflow-hidden">
                {/* Header */}
                <div className={`p-8 ${isValid ? 'bg-gradient-to-r from-green-500 to-emerald-600' : 'bg-gradient-to-r from-red-500 to-pink-600'}`}>
                    <div className="relative">
                        <Link to="/" className="absolute top-0 left-0 text-white/80 hover:text-white transition-colors">
                            <ArrowLeft className="w-6 h-6" />
                        </Link>
                        <div className="text-center text-white">
                            {isValid ? (
                                <CheckCircle2 className="w-20 h-20 mx-auto mb-4" />
                            ) : (
                                <XCircle className="w-20 h-20 mx-auto mb-4" />
                            )}
                            <h1 className="text-3xl font-bold mb-2">
                                {isValid ? 'Xác Thực Thành Công' : 'Xác Thực Thất Bại'}
                            </h1>
                            <p className="text-white/90">{verifyResult?.message}</p>
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div className="p-8 space-y-6">
                    {/* Verification Info */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between py-3 border-b border-gray-200">
                            <span className="text-gray-600 flex items-center gap-2">
                                <Hash className="w-4 h-4" />
                                Transaction ID
                            </span>
                            <span className="font-mono text-xs text-gray-700 truncate max-w-xs" title={verifyResult?.transaction_id}>
                                {verifyResult?.transaction_id}
                            </span>
                        </div>

                        {verifyResult?.batch_id && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Layers className="w-4 h-4" />
                                    Batch ID
                                </span>
                                <span className="font-mono font-semibold text-gray-900">
                                    {verifyResult.batch_id}
                                </span>
                            </div>
                        )}

                        {verifyResult?.package_id && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Package className="w-4 h-4" />
                                    Package ID
                                </span>
                                <span className="font-mono font-semibold text-gray-900">
                                    {verifyResult.package_id}
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
                                Method
                            </span>
                            <span className="px-3 py-1 rounded-full text-sm font-semibold bg-blue-100 text-blue-800">
                                ⛓️ Blockchain Query
                            </span>
                        </div>
                    </div>

                    {/* Product Details Section */}
                    {isValid && productDetails && (
                        <div className="bg-gray-50 rounded-xl p-6 border border-gray-200">
                            <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                                <Package className="w-5 h-5 text-green-600" />
                                Thông Tin Sản Phẩm
                            </h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {productDetails.production_date && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Calendar className="w-3 h-3" /> Ngày sản xuất
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.production_date}</div>
                                    </div>
                                )}
                                {productDetails.expiry_date && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Calendar className="w-3 h-3" /> Hạn sử dụng
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.expiry_date}</div>
                                    </div>
                                )}
                                {productDetails.weight && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Scale className="w-3 h-3" /> Khối lượng
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.weight}g</div>
                                    </div>
                                )}
                                {productDetails.farm_location && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 col-span-full">
                                        <div className="text-xs text-gray-500 mb-1">Nông trại</div>
                                        <div className="font-medium text-gray-900">{productDetails.farm_location}</div>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Success Message */}
                    {isValid && (
                        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                            <p className="text-sm text-green-800">
                                ✅ Thông tin này được truy xuất trực tiếp từ <strong>IBN Network Blockchain</strong>.
                                Dữ liệu là bất biến và không thể làm giả.
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    )
}
