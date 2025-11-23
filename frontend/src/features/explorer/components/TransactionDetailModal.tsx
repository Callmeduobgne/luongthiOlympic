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
import { X, Hash, Clock, Copy, Check, AlertCircle, CheckCircle, XCircle, Loader, ExternalLink } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Badge } from '@shared/components/ui/Badge'
import { transactionExplorerService, type Transaction, type TransactionStatus } from '../services/transactionExplorerService'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

interface TransactionDetailModalProps {
  transaction: Transaction
  onClose: () => void
}

export const TransactionDetailModal = ({ transaction, onClose }: TransactionDetailModalProps) => {
  const [copiedHash, setCopiedHash] = useState<string | null>(null)
  const navigate = useNavigate()

  const { data: receipt, isLoading: receiptLoading } = useQuery({
    queryKey: ['transaction-receipt', transaction.txId],
    queryFn: () => transactionExplorerService.getTransactionReceipt(transaction.txId),
    enabled: !!transaction.txId,
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

  const getStatusBadge = (status: TransactionStatus) => {
    switch (status) {
      case 'VALID':
        return <Badge variant="success"><CheckCircle className="w-3 h-3 mr-1" />Valid</Badge>
      case 'INVALID':
        return <Badge variant="danger"><XCircle className="w-3 h-3 mr-1" />Invalid</Badge>
      case 'FAILED':
        return <Badge variant="danger"><XCircle className="w-3 h-3 mr-1" />Failed</Badge>
      case 'SUBMITTED':
        return <Badge variant="warning"><Loader className="w-3 h-3 mr-1 animate-spin" />Submitted</Badge>
      default:
        return <Badge variant="default">{status}</Badge>
    }
  }

  const handleViewBlock = () => {
    if (transaction.blockNumber !== undefined) {
      navigate(`/explorer?block=${transaction.blockNumber}`)
      onClose()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <Card className="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
              Transaction Details
            </h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
              {getStatusBadge(transaction.status)}
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <X className="w-5 h-5" />
          </Button>
        </div>

        {/* Transaction Info */}
        <div className="p-6 space-y-4">
          {/* TxID */}
          <div>
            <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Transaction ID (TxID)
            </label>
            <div className="mt-1 flex items-center gap-2">
              <code className="flex-1 text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded break-all">
                {transaction.txId}
              </code>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => copyToClipboard(transaction.txId, 'txid')}
              >
                {copiedHash === 'txid' ? (
                  <Check className="w-4 h-4 text-green-500" />
                ) : (
                  <Copy className="w-4 h-4" />
                )}
              </Button>
            </div>
          </div>

          {/* ID */}
          <div>
            <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
              Internal ID
            </label>
            <div className="mt-1 flex items-center gap-2">
              <code className="flex-1 text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded break-all">
                {transaction.id}
              </code>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => copyToClipboard(transaction.id, 'id')}
              >
                {copiedHash === 'id' ? (
                  <Check className="w-4 h-4 text-green-500" />
                ) : (
                  <Copy className="w-4 h-4" />
                )}
              </Button>
            </div>
          </div>

          {/* Grid: Channel, Chaincode, Function */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Channel
              </label>
              <div className="mt-1 text-sm text-gray-900 dark:text-white">
                {transaction.channelName}
              </div>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Chaincode
              </label>
              <div className="mt-1 text-sm text-gray-900 dark:text-white">
                {transaction.chaincodeName}
              </div>
            </div>
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Function
              </label>
              <div className="mt-1 text-sm text-gray-900 dark:text-white">
                {transaction.functionName}
              </div>
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
                {formatTimestamp(transaction.timestamp)}
              </span>
            </div>
          </div>

          {/* Block Info */}
          {transaction.blockNumber !== undefined && (
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Block Information
              </label>
              <div className="mt-1 flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <Hash className="w-4 h-4 text-gray-400" />
                  <span className="text-sm text-gray-900 dark:text-white">
                    Block #{transaction.blockNumber}
                  </span>
                </div>
                <Button variant="ghost" size="sm" onClick={handleViewBlock}>
                  <ExternalLink className="w-4 h-4 mr-1" />
                  Xem Block
                </Button>
              </div>
              {transaction.blockHash && (
                <div className="mt-2 flex items-center gap-2">
                  <code className="flex-1 text-xs font-mono text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 p-2 rounded break-all">
                    {transaction.blockHash}
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => copyToClipboard(transaction.blockHash!, 'blockHash')}
                  >
                    {copiedHash === 'blockHash' ? (
                      <Check className="w-4 h-4 text-green-500" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </Button>
                </div>
              )}
            </div>
          )}

          {/* Arguments */}
          {transaction.args && transaction.args.length > 0 && (
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Arguments
              </label>
              <div className="mt-1 space-y-1">
                {transaction.args.map((arg, index) => (
                  <div key={index} className="text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded">
                    [{index}]: {arg}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Error Message */}
          {transaction.errorMessage && (
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Error Message
              </label>
              <div className="mt-1 flex items-start gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded">
                <AlertCircle className="w-5 h-5 text-red-500 mt-0.5" />
                <p className="text-sm text-red-700 dark:text-red-400 flex-1">
                  {transaction.errorMessage}
                </p>
              </div>
            </div>
          )}

          {/* Endorsing Organizations */}
          {transaction.endorsingOrgs && transaction.endorsingOrgs.length > 0 && (
            <div>
              <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Endorsing Organizations
              </label>
              <div className="mt-1 flex flex-wrap gap-2">
                {transaction.endorsingOrgs.map((org, index) => (
                  <Badge key={index} variant="default">{org}</Badge>
                ))}
              </div>
            </div>
          )}

          {/* Transaction Receipt */}
          {receiptLoading && (
            <div className="p-4 text-center text-gray-600 dark:text-gray-400">
              <Loader className="w-5 h-5 animate-spin mx-auto mb-2" />
              Đang tải receipt...
            </div>
          )}

          {receipt && (
            <div className="border-t border-gray-200 dark:border-gray-700 pt-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
                Transaction Receipt
              </h3>
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Status
                    </label>
                    <div className="mt-1">{getStatusBadge(receipt.status)}</div>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Block Number
                    </label>
                    <div className="mt-1 text-sm text-gray-900 dark:text-white">
                      #{receipt.blockNumber}
                    </div>
                  </div>
                </div>
                {receipt.result !== undefined && receipt.result !== null && (
                  <div>
                    <label className="text-sm font-medium text-gray-500 dark:text-gray-400">
                      Result
                    </label>
                    <pre className="mt-1 text-xs font-mono text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-800 p-2 rounded break-all overflow-auto">
                      {JSON.stringify(receipt.result, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-6 border-t border-gray-200 dark:border-gray-700 flex justify-end">
          <Button onClick={onClose}>Đóng</Button>
        </div>
      </Card>
    </div>
  )
}

