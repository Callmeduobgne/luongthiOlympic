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
import { useQuery } from '@tanstack/react-query'
import { Radio, Search, Check, Smartphone, Link as LinkIcon, AlertCircle } from 'lucide-react'
import { Button } from '@shared/components/ui/Button'
import { Spinner } from '@shared/components/ui/Spinner'
import api from '@shared/utils/api'
import toast from 'react-hot-toast'
import { API_CONFIG } from '@shared/config/api.config'

interface Transaction {
    tx_id: string
    function_name: string
    status: string
    block_number: number
    timestamp: string
    args: string[]
    nfc_tag_id?: string
}

export const NFCManagerPage = () => {
    const [selectedTxId, setSelectedTxId] = useState<string | null>(null)
    const [isScanning, setIsScanning] = useState(false)
    const [searchTerm, setSearchTerm] = useState('')

    // Fetch transactions
    const { data: transactions, isLoading, refetch } = useQuery<Transaction[]>({
        queryKey: ['transactions-list-nfc'],
        queryFn: async () => {
            try {
                const response = await api.get('/api/v1/blockchain/transactions?limit=50')
                const responseData = response.data?.data
                const data = Array.isArray(responseData) ? responseData : (responseData?.transactions || [])

                return data.map((tx: any) => ({
                    tx_id: tx.txId || tx.tx_id || '',
                    function_name: tx.functionName || tx.function_name || '',
                    status: tx.status || '',
                    block_number: tx.blockNumber || tx.block_number || 0,
                    timestamp: tx.timestamp || '',
                    args: tx.args || [],
                    nfc_tag_id: tx.nfcTagId || tx.nfc_tag_id || '',
                })).filter((tx: Transaction) => tx.tx_id)
            } catch (error) {
                console.error('Failed to fetch transactions:', error)
                return []
            }
        },
        refetchInterval: 10000,
    })

    // Filter transactions
    const filteredTransactions = transactions?.filter(tx =>
        tx.tx_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
        (tx.args && tx.args.some(arg => arg.toLowerCase().includes(searchTerm.toLowerCase())))
    )

    const handleAssignNFC = async () => {
        if (!selectedTxId) return

        if (!('NDEFReader' in window)) {
            // @ts-ignore
            toast.error('Trình duyệt không hỗ trợ NFC. Vui lòng dùng Chrome trên Android.')
            return
        }

        try {
            setIsScanning(true)
            // @ts-ignore
            const ndef = new window.NDEFReader()
            await ndef.scan()
            // @ts-ignore
            toast.loading('Đang chờ thẻ NFC... Vui lòng chạm thẻ vào mặt sau điện thoại.', { id: 'nfc-scan' })

            ndef.onreading = async (event: any) => {
                const serialNumber = event.serialNumber
                if (!serialNumber) {
                    toast.error('Không đọc được mã thẻ (UID)', { id: 'nfc-scan' })
                    return
                }

                // Step 1: Write URL to Tag
                try {
                    const verificationUrl = `${API_CONFIG.FRONTEND_URL}/verify/nfc?tag=${serialNumber}`
                    await ndef.write({
                        records: [{ recordType: "url", data: verificationUrl }]
                    })
                    toast.success('Đã ghi URL vào thẻ thành công!', { id: 'nfc-scan' })
                } catch (writeError) {
                    console.error('Write Error:', writeError)
                    toast.error('Không thể ghi vào thẻ (Thẻ bị khóa hoặc lỗi). Vẫn tiếp tục liên kết Database...', { id: 'nfc-scan' })
                }

                // Step 2: Link to Database
                try {
                    await api.post(`/api/v1/blockchain/transactions/${selectedTxId}/nfc`, {
                        nfcId: serialNumber
                    })
                    toast.success(`Đã liên kết thẻ ${serialNumber} với giao dịch!`)
                    setIsScanning(false)
                    refetch() // Refresh list
                } catch (apiError) {
                    console.error('API Error:', apiError)
                    toast.error('Lỗi khi lưu vào Database')
                }
            }
        } catch (error) {
            console.error('NFC Error:', error)
            setIsScanning(false)
            toast.error('Không thể kích hoạt NFC', { id: 'nfc-scan' })
        }
    }

    const getEntityId = (tx: Transaction) => {
        if (tx.args && tx.args.length > 0) {
            return tx.args[0]
        }
        return 'N/A'
    }

    return (
        <div className="min-h-screen bg-gray-900">
            {/* Header */}
            <div className="border-b border-white/10 bg-black/40 backdrop-blur-xl">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                    <div className="flex items-center space-x-4">
                        <div>
                            <h1 className="text-3xl font-bold text-white flex items-center gap-3">
                                <Radio className="w-8 h-8 text-green-400" />
                                Quản Lý NFC (Hybrid Pro)
                            </h1>
                            <p className="text-gray-400 mt-1">
                                Gán thẻ NFC cho sản phẩm: Ghi URL (UX) + Lưu Database (Bảo mật)
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                    {/* Left Panel: List (7 cols) */}
                    <div className="lg:col-span-7 space-y-4">
                        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-xl p-4">
                            <div className="flex justify-between items-center mb-4">
                                <h2 className="text-lg font-semibold text-white">
                                    Danh sách Giao dịch
                                </h2>
                                <div className="relative w-64">
                                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-3.5 h-3.5" />
                                    <input
                                        type="text"
                                        placeholder="Tìm kiếm..."
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="w-full pl-9 pr-4 py-1.5 bg-white/5 border border-white/10 rounded-lg text-sm text-white focus:outline-none focus:border-green-500 transition-colors"
                                    />
                                </div>
                            </div>

                            <div className="space-y-2 max-h-[calc(100vh-240px)] overflow-y-auto pr-2 custom-scrollbar">
                                {isLoading ? (
                                    <div className="flex justify-center py-8">
                                        <Spinner size="lg" />
                                    </div>
                                ) : filteredTransactions && filteredTransactions.length > 0 ? (
                                    filteredTransactions.map((tx) => (
                                        <div
                                            key={tx.tx_id}
                                            onClick={() => setSelectedTxId(tx.tx_id)}
                                            className={`p-3 rounded-lg border transition-all cursor-pointer group ${selectedTxId === tx.tx_id
                                                ? 'bg-green-500/20 border-green-500/50'
                                                : 'bg-white/5 border-white/10 hover:bg-white/10 hover:border-white/20'
                                                }`}
                                        >
                                            <div className="flex items-center justify-between gap-3">
                                                <div className="flex-1 min-w-0">
                                                    <div className="flex items-center gap-2 mb-1">
                                                        <span className="text-[10px] font-bold uppercase tracking-wider text-green-400 bg-green-500/10 px-1.5 py-0.5 rounded border border-green-500/20">
                                                            {tx.function_name || 'N/A'}
                                                        </span>
                                                        {tx.nfc_tag_id && (
                                                            <span className="text-[10px] font-medium text-blue-400 bg-blue-500/10 px-1.5 py-0.5 rounded border border-blue-500/20 flex items-center gap-1">
                                                                <LinkIcon className="w-2.5 h-2.5" />
                                                                Linked
                                                            </span>
                                                        )}
                                                        <span className="text-xs text-gray-500 ml-auto font-mono">
                                                            {new Date(tx.timestamp).toLocaleTimeString()}
                                                        </span>
                                                    </div>
                                                    <div className="flex items-center justify-between">
                                                        <p className="text-sm text-white font-medium truncate pr-2">
                                                            {getEntityId(tx)}
                                                        </p>
                                                    </div>
                                                    <p className="text-[11px] text-gray-500 font-mono truncate opacity-60 group-hover:opacity-100 transition-opacity">
                                                        {tx.tx_id}
                                                    </p>
                                                </div>
                                                {selectedTxId === tx.tx_id && (
                                                    <div className="w-8 h-8 rounded-full bg-green-500/20 flex items-center justify-center border border-green-500/30">
                                                        <Check className="w-4 h-4 text-green-400" />
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    ))
                                ) : (
                                    <div className="text-center py-12 text-gray-500">
                                        <p>Không tìm thấy giao dịch nào</p>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    {/* Right Panel: Action (5 cols) */}
                    <div className="lg:col-span-5 space-y-6">
                        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-xl p-5 sticky top-6">
                            <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                                <Smartphone className="w-5 h-5 text-green-400" />
                                Thao tác Gán Thẻ
                            </h2>

                            {!selectedTxId ? (
                                <div className="flex flex-col items-center justify-center py-10 text-center border-2 border-dashed border-white/10 rounded-xl bg-white/5">
                                    <div className="w-12 h-12 rounded-full bg-white/5 flex items-center justify-center mb-3">
                                        <Radio className="w-6 h-6 text-gray-500" />
                                    </div>
                                    <p className="text-sm text-gray-400 font-medium">Chọn một giao dịch từ danh sách</p>
                                    <p className="text-xs text-gray-600 mt-1">để bắt đầu quy trình gán thẻ NFC</p>
                                </div>
                            ) : (
                                <div className="space-y-5">
                                    <div className="bg-green-500/5 border border-green-500/20 rounded-lg p-3">
                                        <div className="flex justify-between items-start mb-1">
                                            <h3 className="text-xs font-bold text-green-400 uppercase tracking-wide">Giao dịch đang chọn</h3>
                                            <button
                                                onClick={() => setSelectedTxId(null)}
                                                className="text-[10px] text-gray-400 hover:text-white transition-colors"
                                            >
                                                Hủy chọn
                                            </button>
                                        </div>
                                        <p className="text-white font-medium text-sm truncate">{getEntityId(transactions?.find(t => t.tx_id === selectedTxId) as Transaction)}</p>
                                        <p className="text-gray-500 font-mono text-[10px] break-all mt-1 leading-tight opacity-70">{selectedTxId}</p>
                                    </div>

                                    <div className="space-y-2">
                                        <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors">
                                            <div className="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center flex-shrink-0 text-xs font-bold text-white">1</div>
                                            <p className="text-sm text-gray-300">Chuẩn bị thẻ NFC trắng</p>
                                        </div>
                                        <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors">
                                            <div className="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center flex-shrink-0 text-xs font-bold text-white">2</div>
                                            <p className="text-sm text-gray-300">Bấm nút "Bắt đầu Gán" bên dưới</p>
                                        </div>
                                        <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-white/5 transition-colors">
                                            <div className="w-6 h-6 rounded-full bg-white/10 flex items-center justify-center flex-shrink-0 text-xs font-bold text-white">3</div>
                                            <p className="text-sm text-gray-300">Chạm thẻ vào mặt lưng điện thoại</p>
                                        </div>
                                    </div>

                                    <Button
                                        variant="primary"
                                        className="w-full py-3 text-base font-bold bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-500 hover:to-emerald-500 shadow-lg shadow-green-900/20"
                                        onClick={handleAssignNFC}
                                        disabled={isScanning}
                                    >
                                        {isScanning ? (
                                            <>
                                                <Spinner size="sm" className="mr-2" />
                                                Đang chờ thẻ...
                                            </>
                                        ) : (
                                            <>
                                                <Radio className="w-4 h-4 mr-2" />
                                                Bắt đầu Gán (Hybrid)
                                            </>
                                        )}
                                    </Button>

                                    <div className="flex gap-2 p-3 bg-yellow-500/5 border border-yellow-500/10 rounded-lg">
                                        <AlertCircle className="w-4 h-4 text-yellow-500/70 flex-shrink-0 mt-0.5" />
                                        <p className="text-[11px] text-yellow-200/60 leading-relaxed">
                                            Yêu cầu trình duyệt Chrome trên Android. Đảm bảo NFC đã được bật trong cài đặt điện thoại.
                                        </p>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
