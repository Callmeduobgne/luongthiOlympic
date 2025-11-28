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
import { QrCode, Copy, Check, Package, Box, Hash, Search } from 'lucide-react'
import { QRCodeDisplay } from '@shared/components/ui/QRCodeDisplay'
import { Button } from '@shared/components/ui/Button'
import { Spinner } from '@shared/components/ui/Spinner'
import api from '@shared/utils/api'
import toast from 'react-hot-toast'

type QRCodeType = 'batch' | 'package' | 'transaction'

interface Transaction {
  tx_id: string
  function_name: string
  status: string
  block_number: number
  timestamp: string
  args: string[]
}

export const QRCodeGeneratorPage = () => {
  const [qrCodeType, setQrCodeType] = useState<QRCodeType>('batch')
  const [inputValue, setInputValue] = useState<string>('')
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // State for selected IDs
  const [selectedBatchId, setSelectedBatchId] = useState<string | null>(null)
  const [selectedPackageId, setSelectedPackageId] = useState<string | null>(null)
  const [selectedTxId, setSelectedTxId] = useState<string | null>(null)
  const [selectedPackageHash, setSelectedPackageHash] = useState<string | null>(null)

  // Fetch batches
  const { data: batches, isLoading: batchesLoading } = useQuery({
    queryKey: ['batches-list'],
    queryFn: async () => {
      const response = await api.get('/api/v1/teatrace/batches?limit=50')
      if (response.data?.batches) {
        return response.data.batches
      }
      if (response.data?.data?.batches) {
        return response.data.data.batches
      }
      if (response.data?.data && Array.isArray(response.data.data)) {
        return response.data.data
      }
      return []
    },
    enabled: qrCodeType === 'batch',
  })

  // Fetch packages (if API exists)
  const { data: packages, isLoading: packagesLoading } = useQuery({
    queryKey: ['packages-list'],
    queryFn: async () => {
      try {
        const response = await api.get('/api/v1/blockchain/transactions?limit=100')
        const transactions = response.data?.data || []

        const packageTransactions = transactions.filter((tx: Transaction) =>
          tx.function_name === 'CreatePackage' || tx.function_name === 'createPackage'
        )

        const packageList = packageTransactions.map((tx: Transaction) => ({
          package_id: tx.args?.[0] || tx.tx_id,
          id: tx.args?.[0] || tx.tx_id,
          tx_id: tx.tx_id,
          block_number: tx.block_number,
          timestamp: tx.timestamp,
          function_name: tx.function_name,
        }))

        return packageList
      } catch (error) {
        if (import.meta.env.DEV) {
          console.error('[DEV] Failed to fetch packages:', error)
        }
        return []
      }
    },
    enabled: qrCodeType === 'package',
  })

  // Fetch transactions
  const { data: transactions, isLoading: transactionsLoading } = useQuery<Transaction[]>({
    queryKey: ['transactions-list'],
    queryFn: async () => {
      const response = await api.get('/api/v1/blockchain/transactions?limit=50')
      return response.data.data || []
    },
    enabled: qrCodeType === 'transaction',
    refetchInterval: 10000,
  })

  const handleCopyId = (id: string) => {
    navigator.clipboard.writeText(id)
    setCopiedId(id)
    setTimeout(() => setCopiedId(null), 2000)
    toast.success('Đã copy ID')
  }

  const handleGenerateFromInput = () => {
    if (!inputValue.trim()) {
      toast.error('Vui lòng nhập ID')
      return
    }

    switch (qrCodeType) {
      case 'batch':
        setSelectedBatchId(inputValue.trim())
        setSelectedPackageId(null)
        setSelectedTxId(null)
        break
      case 'package':
        setSelectedPackageId(inputValue.trim())
        setSelectedBatchId(null)
        setSelectedTxId(null)
        break
      case 'transaction':
        setSelectedTxId(inputValue.trim())
        setSelectedBatchId(null)
        setSelectedPackageId(null)
        break
    }
    setInputValue('')
  }

  const handleSelectFromList = (id: string, hash?: string) => {
    switch (qrCodeType) {
      case 'batch':
        setSelectedBatchId(id)
        setSelectedPackageId(null)
        setSelectedTxId(null)
        setSelectedPackageHash(null)
        break
      case 'package':
        setSelectedPackageId(id)
        setSelectedPackageHash(hash || null)
        setSelectedBatchId(null)
        setSelectedTxId(null)
        break
      case 'transaction':
        setSelectedTxId(id)
        setSelectedBatchId(null)
        setSelectedPackageId(null)
        setSelectedPackageHash(null)
        break
    }
  }

  const getEntityId = (tx: Transaction) => {
    if (tx.args && tx.args.length > 0) {
      return tx.args[0]
    }
    return 'N/A'
  }

  const currentId = selectedBatchId || selectedPackageId || selectedTxId

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-green-900/20 to-gray-900">
      {/* Header */}
      <div className="border-b border-white/10 bg-black/40 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center space-x-4">
            <div>
              <h1 className="text-3xl font-bold text-white flex items-center gap-3">
                <QrCode className="w-8 h-8 text-green-400" />
                QR Code Generator
              </h1>
              <p className="text-gray-400 mt-1">
                Tạo QR code cho Batch, Package hoặc Transaction
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Left Panel: Input & Selection */}
          <div className="space-y-6">
            {/* Type Selection */}
            <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
              <h2 className="text-xl font-semibold text-white mb-4">
                Chọn loại QR Code
              </h2>
              <div className="grid grid-cols-3 gap-3">
                <button
                  onClick={() => {
                    setQrCodeType('batch')
                    setSelectedBatchId(null)
                    setSelectedPackageId(null)
                    setSelectedTxId(null)
                  }}
                  className={`p-4 rounded-xl border transition-all ${qrCodeType === 'batch'
                    ? 'bg-green-500/20 border-green-500/50 text-white'
                    : 'bg-white/5 border-white/10 text-gray-300 hover:bg-white/10'
                    }`}
                >
                  <Package className="w-6 h-6 mx-auto mb-2" />
                  <span className="text-sm font-medium">Batch</span>
                </button>
                <button
                  onClick={() => {
                    setQrCodeType('package')
                    setSelectedBatchId(null)
                    setSelectedPackageId(null)
                    setSelectedTxId(null)
                  }}
                  className={`p-4 rounded-xl border transition-all ${qrCodeType === 'package'
                    ? 'bg-green-500/20 border-green-500/50 text-white'
                    : 'bg-white/5 border-white/10 text-gray-300 hover:bg-white/10'
                    }`}
                >
                  <Box className="w-6 h-6 mx-auto mb-2" />
                  <span className="text-sm font-medium">Package</span>
                </button>
                <button
                  onClick={() => {
                    setQrCodeType('transaction')
                    setSelectedBatchId(null)
                    setSelectedPackageId(null)
                    setSelectedTxId(null)
                  }}
                  className={`p-4 rounded-xl border transition-all ${qrCodeType === 'transaction'
                    ? 'bg-green-500/20 border-green-500/50 text-white'
                    : 'bg-white/5 border-white/10 text-gray-300 hover:bg-white/10'
                    }`}
                >
                  <Hash className="w-6 h-6 mx-auto mb-2" />
                  <span className="text-sm font-medium">Transaction</span>
                </button>
              </div>
            </div>

            {/* Manual Input */}
            <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
              <h2 className="text-xl font-semibold text-white mb-4">
                Nhập ID trực tiếp
              </h2>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleGenerateFromInput()
                    }
                  }}
                  placeholder={`Nhập ${qrCodeType} ID...`}
                  className="flex-1 px-4 py-3 rounded-xl bg-white/5 border border-white/10 text-white placeholder-gray-500 focus:outline-none focus:border-green-500/50 focus:ring-2 focus:ring-green-500/20"
                />
                <Button
                  onClick={handleGenerateFromInput}
                  variant="primary"
                  className="px-6"
                >
                  <Search className="w-4 h-4 mr-2" />
                  Tạo
                </Button>
              </div>
            </div>

            {/* List Selection */}
            <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
              <h2 className="text-xl font-semibold text-white mb-4">
                Hoặc chọn từ danh sách
              </h2>

              {qrCodeType === 'batch' && (
                <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2">
                  {batchesLoading ? (
                    <div className="flex justify-center py-8">
                      <Spinner size="lg" />
                    </div>
                  ) : batches && batches.length > 0 ? (
                    batches.map((batch: any, index: number) => (
                      <div
                        key={batch.batch_id || batch.id || index}
                        onClick={() => handleSelectFromList(batch.batch_id || batch.id)}
                        className={`p-4 rounded-xl border transition-all cursor-pointer ${selectedBatchId === (batch.batch_id || batch.id)
                          ? 'bg-green-500/20 border-green-500/50'
                          : 'bg-white/5 border-white/10 hover:bg-white/10'
                          }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex-1">
                            <p className="text-sm text-white font-medium">
                              {batch.batch_id || batch.id}
                            </p>
                            {batch.name && (
                              <p className="text-xs text-gray-400 mt-1">{batch.name}</p>
                            )}
                          </div>
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleCopyId(batch.batch_id || batch.id)
                            }}
                            className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                          >
                            {copiedId === (batch.batch_id || batch.id) ? (
                              <Check className="w-4 h-4 text-green-400" />
                            ) : (
                              <Copy className="w-4 h-4 text-gray-400" />
                            )}
                          </button>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-center py-8 text-gray-400">
                      Không có batch nào
                    </div>
                  )}
                </div>
              )}

              {qrCodeType === 'package' && (
                <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2">
                  {packagesLoading ? (
                    <div className="flex justify-center py-8">
                      <Spinner size="lg" />
                    </div>
                  ) : packages && packages.length > 0 ? (
                    packages.map((pkg: any) => (
                      <div
                        key={pkg.tx_id}
                        onClick={() => handleSelectFromList(pkg.package_id || pkg.id, pkg.block_hash || pkg.blockHash)}
                        className={`p-4 rounded-xl border transition-all cursor-pointer ${selectedPackageId === (pkg.package_id || pkg.id)
                          ? 'bg-green-500/20 border-green-500/50'
                          : 'bg-white/5 border-white/10 hover:bg-white/10'
                          }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex-1">
                            <p className="text-sm text-white font-medium">
                              {pkg.package_id || pkg.id}
                            </p>
                            {pkg.name && (
                              <p className="text-xs text-gray-400 mt-1">{pkg.name}</p>
                            )}
                            {(pkg.block_hash || pkg.blockHash) && (
                              <div className="mt-2 flex items-center gap-2">
                                <Hash className="w-3 h-3 text-gray-500" />
                                <p className="text-xs text-gray-500 font-mono truncate">
                                  Hash: {(pkg.block_hash || pkg.blockHash).substring(0, 16)}...
                                </p>
                              </div>
                            )}
                          </div>
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleCopyId(pkg.package_id || pkg.id)
                            }}
                            className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                          >
                            {copiedId === (pkg.package_id || pkg.id) ? (
                              <Check className="w-4 h-4 text-green-400" />
                            ) : (
                              <Copy className="w-4 h-4 text-gray-400" />
                            )}
                          </button>
                        </div>
                      </div>
                    ))
                  ) : (
                    <div className="text-center py-8 text-gray-400">
                      Không có package nào
                    </div>
                  )}
                </div>
              )}

              {qrCodeType === 'transaction' && (
                <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2">
                  {transactionsLoading ? (
                    <div className="flex justify-center py-8">
                      <Spinner size="lg" />
                    </div>
                  ) : transactions && transactions.length > 0 ? (
                    transactions.map((tx) => (
                      <div
                        key={tx.tx_id}
                        onClick={() => handleSelectFromList(tx.tx_id)}
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
                              handleCopyId(tx.tx_id)
                            }}
                            className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                          >
                            {copiedId === tx.tx_id ? (
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
                    ))
                  ) : (
                    <div className="text-center py-8 text-gray-400">
                      Không có transaction nào
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>

          {/* Right Panel: QR Code Display */}
          <div className="space-y-4">
            <div className="bg-black/40 backdrop-blur-xl border border-white/10 rounded-2xl p-6">
              <h2 className="text-xl font-semibold text-white mb-4">
                QR Code Preview
              </h2>

              {!currentId ? (
                <div className="flex flex-col items-center justify-center py-24 text-center">
                  <QrCode className="w-16 h-16 text-gray-600 mb-4" />
                  <p className="text-gray-400">
                    Chọn hoặc nhập {qrCodeType} ID để tạo QR code
                  </p>
                </div>
              ) : (
                <div className="space-y-6">
                  <div className="flex justify-center">
                    <QRCodeDisplay
                      batchId={selectedBatchId || undefined}
                      packageId={selectedPackageId || undefined}
                      txId={selectedTxId || undefined}
                      size={320}
                      showDownload={true}
                    />
                  </div>

                  <div className="bg-white/5 rounded-xl p-4 space-y-3">
                    <div>
                      <label className="text-xs text-gray-400 uppercase tracking-wide">
                        {qrCodeType === 'batch' ? 'Batch ID' : qrCodeType === 'package' ? 'Package ID' : 'Transaction ID'}
                      </label>
                      <div className="flex items-center gap-2 mt-1">
                        <p className="text-sm text-white font-mono flex-1 truncate">
                          {currentId}
                        </p>
                        <button
                          onClick={() => handleCopyId(currentId)}
                          className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                        >
                          {copiedId === currentId ? (
                            <Check className="w-4 h-4 text-green-400" />
                          ) : (
                            <Copy className="w-4 h-4 text-gray-400" />
                          )}
                        </button>
                      </div>
                    </div>

                    {selectedPackageHash && (
                      <div>
                        <label className="text-xs text-gray-400 uppercase tracking-wide">
                          Block Hash
                        </label>
                        <div className="flex items-center gap-2 mt-1">
                          <p className="text-sm text-gray-300 font-mono flex-1 truncate">
                            {selectedPackageHash}
                          </p>
                          <button
                            onClick={() => handleCopyId(selectedPackageHash)}
                            className="p-2 rounded-lg bg-white/5 hover:bg-white/10 transition-colors"
                          >
                            {copiedId === selectedPackageHash ? (
                              <Check className="w-4 h-4 text-green-400" />
                            ) : (
                              <Copy className="w-4 h-4 text-gray-400" />
                            )}
                          </button>
                        </div>
                      </div>
                    )}

                    <div>
                      <label className="text-xs text-gray-400 uppercase tracking-wide">
                        Verification URL
                      </label>
                      <p className="text-sm text-green-400 mt-1 font-mono break-all">
                        {selectedBatchId
                          ? `${window.location.origin}/verify/batches/${selectedBatchId}`
                          : selectedPackageId
                            ? `${window.location.origin}/verify/packages/${selectedPackageId}`
                            : `${window.location.origin}/verify/hash?hash=${selectedTxId}`}
                      </p>
                    </div>
                  </div>

                  <Button
                    variant="primary"
                    className="w-full"
                    onClick={() => {
                      const url = selectedBatchId
                        ? `/verify/batches/${selectedBatchId}`
                        : selectedPackageId
                          ? `/verify/packages/${selectedPackageId}`
                          : `/verify/hash?hash=${selectedTxId}`
                      window.open(url, '_blank')
                    }}
                  >
                    Mở trang Verification
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
