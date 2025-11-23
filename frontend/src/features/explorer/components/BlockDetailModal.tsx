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

import { useQuery } from '@tanstack/react-query'
import { X, Hash, Clock, FileText, Copy, Check } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Badge } from '@shared/components/ui/Badge'
import { explorerService } from '../services/explorerService'
import type { Block } from '@shared/types/blockchain.types'
import { useState } from 'react'

interface BlockDetailModalProps {
  block: Block
  onClose: () => void
}

export const BlockDetailModal = ({ block, onClose }: BlockDetailModalProps) => {
  const [copiedHash, setCopiedHash] = useState<string | null>(null)

  const { data: transactions, isLoading } = useQuery({
    queryKey: ['block-transactions', block.number],
    queryFn: () => explorerService.getBlockTransactions('ibnchannel', block.number),
    enabled: !!block,
  })

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp)
      return date.toLocaleString('vi-VN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    } catch {
      return timestamp
    }
  }

  const copyToClipboard = (text: string, hashType: string) => {
    navigator.clipboard.writeText(text)
    setCopiedHash(hashType)
    setTimeout(() => setCopiedHash(null), 2000)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <Card className="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
              Block #{block.number}
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              Chi tiết block và transactions
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* Block Info */}
        <div className="p-6 space-y-4">
          {/* Block Number */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Block Number
              </label>
              <div className="mt-1 flex items-center gap-2">
                <Hash className="w-4 h-4 text-gray-400" />
                <span className="text-lg font-semibold text-gray-900 dark:text-white">
                  {block.number}
                </span>
              </div>
            </div>

            {/* Timestamp */}
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Timestamp
              </label>
              <div className="mt-1 flex items-center gap-2">
                <Clock className="w-4 h-4 text-gray-400" />
                <span className="text-sm text-gray-900 dark:text-white">
                  {formatTimestamp(block.timestamp)}
                </span>
              </div>
            </div>
          </div>

          {/* Hash */}
          <div>
            <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Block Hash
            </label>
            <div className="mt-1 flex items-center gap-2">
              <code className="flex-1 text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded break-all">
                {block.hash || 'N/A'}
              </code>
              {block.hash && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(block.hash, 'hash')}
                >
                  {copiedHash === 'hash' ? (
                    <Check className="w-4 h-4 text-green-500" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              )}
            </div>
            {!block.hash && (
              <p className="text-xs text-yellow-500 dark:text-yellow-400 mt-1">
                Block hash chưa được lưu trong database. Có thể block này chưa có transactions hoặc indexer chưa chạy.
              </p>
            )}
          </div>

          {/* Previous Hash */}
          {block.previousHash && (
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Previous Hash
              </label>
              <div className="mt-1 flex items-center gap-2">
                <code className="flex-1 text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded break-all">
                  {block.previousHash}
                </code>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(block.previousHash, 'prevHash')}
                >
                  {copiedHash === 'prevHash' ? (
                    <Check className="w-4 h-4 text-green-500" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Data Hash */}
          {block.dataHash && (
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Data Hash
              </label>
              <div className="mt-1 flex items-center gap-2">
                <code className="flex-1 text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded break-all">
                  {block.dataHash}
                </code>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => copyToClipboard(block.dataHash, 'dataHash')}
                >
                  {copiedHash === 'dataHash' ? (
                    <Check className="w-4 h-4 text-green-500" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                Hash của dữ liệu transactions trong block
              </p>
            </div>
          )}

          {/* Transaction Count */}
          <div>
            <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Transactions
            </label>
            <div className="mt-1">
              <Badge variant="default" className="text-lg px-3 py-1">
                <FileText className="w-4 h-4 mr-2" />
                {block.transactionCount} transaction{block.transactionCount !== 1 ? 's' : ''}
              </Badge>
            </div>
          </div>
        </div>

        {/* Transactions List */}
        {isLoading && (
          <div className="p-6 text-center text-gray-600 dark:text-gray-400">
            Đang tải transactions...
          </div>
        )}

        {transactions && transactions.length > 0 && (
          <div className="p-6 border-t border-gray-200 dark:border-gray-700">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Transactions
            </h3>
            <div className="space-y-2">
              {transactions.map((tx, index) => (
                <div
                  key={tx.txId || index}
                  className="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <code className="text-xs font-mono text-gray-900 dark:text-white">
                          {tx.txId ? `${tx.txId.substring(0, 16)}...` : `TX-${index + 1}`}
                        </code>
                        {tx.isValid !== undefined && (
                          <Badge variant={tx.isValid ? 'success' : 'danger'}>
                            {tx.isValid ? 'Valid' : 'Invalid'}
                          </Badge>
                        )}
                      </div>
                      {tx.chaincodeName && (
                        <div className="text-sm text-gray-600 dark:text-gray-400">
                          Chaincode: <span className="font-medium">{tx.chaincodeName}</span>
                          {tx.functionName && (
                            <> • Function: <span className="font-medium">{tx.functionName}</span></>
                          )}
                        </div>
                      )}
                      {tx.timestamp && (
                        <div className="text-xs text-gray-500 dark:text-gray-500 mt-1">
                          {formatTimestamp(tx.timestamp)}
                        </div>
                      )}
                    </div>
                    {tx.txId && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(tx.txId, `tx-${index}`)}
                      >
                        {copiedHash === `tx-${index}` ? (
                          <Check className="w-4 h-4 text-green-500" />
                        ) : (
                          <Copy className="w-4 h-4" />
                        )}
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {transactions && transactions.length === 0 && !isLoading && (
          <div className="p-6 text-center border-t border-gray-200 dark:border-gray-700">
            <p className="text-gray-600 dark:text-gray-400 mb-2">
              Không có transactions trong block này
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-500">
              Block có thể là genesis block hoặc block config không chứa transactions
            </p>
          </div>
        )}

        {/* Footer */}
        <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-end">
          <Button onClick={onClose}>Đóng</Button>
        </div>
      </Card>
    </div>
  )
}

