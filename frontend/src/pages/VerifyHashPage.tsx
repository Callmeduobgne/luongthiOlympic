
import { useSearchParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, XCircle, Shield, Clock, Hash, Layers, ArrowLeft, Package, Calendar, Scale, Copy, Sprout, Factory, Award } from 'lucide-react'
import { verifyByHash } from '@/services/verifyService'
import type { VerifyResult } from '@/services/verifyService'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'

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

    const handleCopy = (text: string, label: string) => {
        navigator.clipboard.writeText(text)
        toast.success(`Đã sao chép ${label}`)
    }

    if (!hash) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
                <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="max-w-md w-full bg-white rounded-2xl shadow-xl p-8 text-center"
                >
                    <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
                    <h2 className="text-2xl font-bold text-gray-900 mb-2">Thiếu Thông Tin</h2>
                    <p className="text-gray-600 mb-6">Vui lòng cung cấp hash hoặc transaction ID để xác thực.</p>
                    <Link to="/" className="inline-flex items-center text-green-600 hover:text-green-700 font-medium">
                        <ArrowLeft className="w-4 h-4 mr-2" />
                        Quay về trang chủ
                    </Link>
                </motion.div>
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
                <motion.div
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="max-w-md w-full bg-white rounded-2xl shadow-xl p-8"
                >
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
                </motion.div>
            </div>
        )
    }

    const isValid = verifyResult?.is_valid
    const productDetails = verifyResult?.product_details

    return (
        <div className={`min-h-screen flex items-center justify-center p-4 bg-gradient-to-br ${isValid ? 'from-green-50 to-emerald-100' : 'from-red-50 to-pink-100'
            }`}>
            <motion.div
                initial={{ opacity: 0, y: 30 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5 }}
                className="max-w-2xl w-full bg-white rounded-2xl shadow-2xl overflow-hidden"
            >
                {/* Header */}
                <div className={`p-8 ${isValid ? 'bg-gradient-to-r from-green-500 to-emerald-600' : 'bg-gradient-to-r from-red-500 to-pink-600'}`}>
                    <div className="relative">
                        <Link to="/" className="absolute top-0 left-0 text-white/80 hover:text-white transition-colors">
                            <ArrowLeft className="w-6 h-6" />
                        </Link>
                        <div className="text-center text-white">
                            {isValid ? (
                                <motion.div
                                    initial={{ scale: 0 }}
                                    animate={{ scale: 1 }}
                                    transition={{ type: "spring", stiffness: 260, damping: 20, delay: 0.2 }}
                                >
                                    <CheckCircle2 className="w-20 h-20 mx-auto mb-4" />
                                </motion.div>
                            ) : (
                                <motion.div
                                    initial={{ scale: 0 }}
                                    animate={{ scale: 1 }}
                                    transition={{ type: "spring", stiffness: 260, damping: 20, delay: 0.2 }}
                                >
                                    <XCircle className="w-20 h-20 mx-auto mb-4" />
                                </motion.div>
                            )}
                            <motion.h1
                                initial={{ opacity: 0, y: 10 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: 0.3 }}
                                className="text-3xl font-bold mb-2"
                            >
                                {isValid ? 'Xác Thực Thành Công' : 'Xác Thực Thất Bại'}
                            </motion.h1>
                            <motion.p
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                transition={{ delay: 0.4 }}
                                className="text-white/90"
                            >
                                {verifyResult?.message}
                            </motion.p>
                        </div>
                    </div>
                </div>

                {/* Content */}
                <div className="p-8 space-y-6">
                    {/* Verification Info */}
                    <div className="space-y-3">
                        <div className="flex items-center justify-between py-3 border-b border-gray-200 group">
                            <span className="text-gray-600 flex items-center gap-2">
                                <Hash className="w-4 h-4" />
                                Transaction ID
                            </span>
                            <div className="flex items-center gap-2">
                                <span className="font-mono text-xs text-gray-700 truncate max-w-xs" title={verifyResult?.transaction_id}>
                                    {verifyResult?.transaction_id}
                                </span>
                                <button
                                    onClick={() => handleCopy(verifyResult?.transaction_id || '', 'Transaction ID')}
                                    className="text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity"
                                    title="Sao chép"
                                >
                                    <Copy className="w-3.5 h-3.5" />
                                </button>
                            </div>
                        </div>

                        {verifyResult?.batch_id && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200 group">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Layers className="w-4 h-4" />
                                    Batch ID
                                </span>
                                <div className="flex items-center gap-2">
                                    <span className="font-mono font-semibold text-gray-900">
                                        {verifyResult.batch_id}
                                    </span>
                                    <button
                                        onClick={() => handleCopy(verifyResult.batch_id || '', 'Batch ID')}
                                        className="text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity"
                                        title="Sao chép"
                                    >
                                        <Copy className="w-3.5 h-3.5" />
                                    </button>
                                </div>
                            </div>
                        )}

                        {verifyResult?.package_id && (
                            <div className="flex items-center justify-between py-3 border-b border-gray-200 group">
                                <span className="text-gray-600 flex items-center gap-2">
                                    <Package className="w-4 h-4" />
                                    Package ID
                                </span>
                                <div className="flex items-center gap-2">
                                    <span className="font-mono font-semibold text-gray-900">
                                        {verifyResult.package_id}
                                    </span>
                                    <button
                                        onClick={() => handleCopy(verifyResult.package_id || '', 'Package ID')}
                                        className="text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity"
                                        title="Sao chép"
                                    >
                                        <Copy className="w-3.5 h-3.5" />
                                    </button>
                                </div>
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
                        <motion.div
                            initial={{ opacity: 0, y: 20 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.5 }}
                            className="bg-gray-50 rounded-xl p-6 border border-gray-200"
                        >
                            <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                                <Package className="w-5 h-5 text-green-600" />
                                Thông Tin Sản Phẩm
                            </h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                {productDetails.harvest_date && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Sprout className="w-3 h-3" /> Ngày thu hoạch
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.harvest_date}</div>
                                    </div>
                                )}
                                {productDetails.production_date && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Calendar className="w-3 h-3" /> Ngày sản xuất
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.production_date}</div>
                                    </div>
                                )}
                                {productDetails.expiry_date && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Calendar className="w-3 h-3" /> Hạn sử dụng
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.expiry_date}</div>
                                    </div>
                                )}
                                {productDetails.weight && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Scale className="w-3 h-3" /> Khối lượng
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.weight}g</div>
                                    </div>
                                )}
                                {productDetails.processing_info && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Factory className="w-3 h-3" /> Quy trình
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.processing_info}</div>
                                    </div>
                                )}
                                {productDetails.quality_cert && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1 flex items-center gap-1">
                                            <Award className="w-3 h-3" /> Chứng chỉ
                                        </div>
                                        <div className="font-medium text-gray-900">{productDetails.quality_cert}</div>
                                    </div>
                                )}
                                {productDetails.farm_location && (
                                    <div className="bg-white p-3 rounded-lg border border-gray-100 col-span-full hover:shadow-md transition-shadow">
                                        <div className="text-xs text-gray-500 mb-1">Nông trại</div>
                                        <div className="font-medium text-gray-900">{productDetails.farm_location}</div>
                                    </div>
                                )}
                            </div>
                        </motion.div>
                    )}

                    {/* Success Message */}
                    {isValid && (
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            transition={{ delay: 0.6 }}
                            className="bg-green-50 border border-green-200 rounded-lg p-4"
                        >
                            <p className="text-sm text-green-800">
                                ✅ Thông tin này được truy xuất trực tiếp từ <strong>IBN Network Blockchain</strong>.
                                Dữ liệu là bất biến và không thể làm giả.
                            </p>
                        </motion.div>
                    )}
                </div>
            </motion.div>
        </div>
    )
}

