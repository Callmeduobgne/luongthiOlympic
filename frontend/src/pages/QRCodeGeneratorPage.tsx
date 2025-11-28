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
import { QrCode, Copy, Check, ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'
import { QRCodeDisplay } from '@shared/components/ui/QRCodeDisplay'
import { Button } from '@shared/components/ui/Button'
import { Spinner } from '@shared/components/ui/Spinner'
import api from '@shared/utils/api'

interface Transaction {
    tx_id: string
    function_name: string
    status: string
    block_number: number
    timestamp: string
    args: string[]
}

export const QRCodeGeneratorPage = () => {
    const [selectedTxId, setSelectedTxId] = useState<string | null>(null)
    const [copiedTxId, setCopiedTxId] = useState<string | null>(null)

    // Fetch transactions from API
    const { data: transactions, isLoading, error } = useQuery<Transaction[]>({
        queryKey: ['transactions'],
        queryFn: async () => {
            const response = await api.get('/api/v1/transactions?limit=50')
            return response.data.data || []
        },
        refetchInterval: 10000, // Refresh every 10 seconds
    })

    const handleCopyTxId = (txId: string) => {
        navigator.clipboard.writeText(txId)
        setCopiedTxId(txId)
        setTimeout(() => setCopiedTxId(null), 2000)
    }

    const getEntityId = (tx: Transaction) => {
        if (tx.args && tx.args.length > 0) {
            return tx.args[0]
        }
        return 'N/A'
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 via-green-900/20 to-gray-900">
            {/* Header */}
            <div className="border-b border-white/10 bg-black/40 backdrop-blur-xl">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                            <Link
                                to="/dashboard"
                                className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                            >
                                <ArrowLeft className="w-5 h-5 text-white" />
                            </Link>
                            <div>
                                <h1 className="text-3xl font-bold text-white flex items-center gap-3">
                                    <QrCode className="w-8 h-8 text-green-400" />
                                    QR Code Generator
                                </h1>
                                <p className="text-gray-400 mt-1">
                                    Generate QR codes for transaction verification
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                    {/* Transactions List */}
                    <div className="space-y-4">
                        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
                            <h2 className="text-xl font-semibold text-white mb-4">
                                Recent Transactions
                            </h2>

                            {isLoading && (
                                <div className="flex justify-center py-12">
                                    <Spinner size="lg" />
                                </div>
                            )}

                            {error && (
                                <div className="text-center py-12">
                                    <p className="text-red-400">Failed to load transactions</p>
                                </div>
                            )}

                            {transactions && transactions.length === 0 && (
                                <div className="text-center py-12">
                                    <QrCode className="w-12 h-12 text-gray-600 mx-auto mb-4" />
                                    <p className="text-gray-400">No transactions found</p>
                                </div>
                            )}

                            <div className="space-y-3 max-h-[600px] overflow-y-auto pr-2">
                                {transactions?.map((tx) => (
                                    <div
                                        key={tx.tx_id}
                                        onClick={() => setSelectedTxId(tx.tx_id)}
                                        className={`p-4 rounded-xl border transition-all cursor-pointer ${selectedTxId === tx.tx_id
                                                ? 'bg-green-500/20 border-green-500/50'
                                                : 'bg-white/5 border-white/10 hover:bg-white/10'
                                            }`}
                                    >
                                        <div className="flex items-start justify-between mb-2">
                                            <div className="flex-1">
                                                <div className="flex items-center gap-2 mb-1">
                                                    <span className="text-xs font-medium text-green-400 bg-green-500/20 px-2 py-0.5 rounded">
                                                        {tx.function_name}
                                                    </span>
                                                    <span className="text-xs text-gray-400">
                                                        Block #{tx.block_number}
                                                    </span>
                                                </div>
                                                <p className="text-sm text-white font-medium">
                                                    {getEntityId(tx)}
                                                </p>
                                            </div>
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation()
                                                    handleCopyTxId(tx.tx_id)
                                                }}
                                                className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                                            >
                                                {copiedTxId === tx.tx_id ? (
                                                    <Check className="w-4 h-4 text-green-400" />
                                                ) : (
                                                    <Copy className="w-4 h-4 text-gray-400" />
                                                )}
                                            </button>
                                        </div>
                                        <p className="text-xs text-gray-500 font-mono truncate">
                                            {tx.tx_id}
                                        </p>
                                        <p className="text-xs text-gray-400 mt-1">
                                            {new Date(tx.timestamp).toLocaleString()}
                                        </p>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* QR Code Display */}
                    <div className="space-y-4">
                        <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
                            <h2 className="text-xl font-semibold text-white mb-4">
                                QR Code Preview
                            </h2>

                            {!selectedTxId ? (
                                <div className="flex flex-col items-center justify-center py-24 text-center">
                                    <QrCode className="w-16 h-16 text-gray-600 mb-4" />
                                    <p className="text-gray-400">
                                        Select a transaction to generate QR code
                                    </p>
                                </div>
                            ) : (
                                <div className="space-y-6">
                                    <div className="flex justify-center">
                                        <QRCodeDisplay
                                            txId={selectedTxId}
                                            size={320}
                                            showDownload={true}
                                        />
                                    </div>

                                    <div className="bg-white/5 rounded-xl p-4 space-y-3">
                                        <div>
                                            <label className="text-xs text-gray-400 uppercase tracking-wide">
                                                Transaction ID
                                            </label>
                                            <div className="flex items-center gap-2 mt-1">
                                                <p className="text-sm text-white font-mono flex-1 truncate">
                                                    {selectedTxId}
                                                </p>
                                                <button
                                                    onClick={() => handleCopyTxId(selectedTxId)}
                                                    className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                                                >
                                                    {copiedTxId === selectedTxId ? (
                                                        <Check className="w-4 h-4 text-green-400" />
                                                    ) : (
                                                        <Copy className="w-4 h-4 text-gray-400" />
                                                    )}
                                                </button>
                                            </div>
                                        </div>

                                        <div>
                                            <label className="text-xs text-gray-400 uppercase tracking-wide">
                                                Verification URL
                                            </label>
                                            <p className="text-sm text-green-400 mt-1 font-mono break-all">
                                                {window.location.origin}/verify/hash?hash={selectedTxId}
                                            </p>
                                        </div>
                                    </div>

                                    <Button
                                        variant="primary"
                                        className="w-full"
                                        onClick={() => {
                                            window.open(
                                                `/verify/hash?hash=${selectedTxId}`,
                                                '_blank'
                                            )
                                        }}
                                    >
                                        Test Verification Page
                                    </Button>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    )
}
