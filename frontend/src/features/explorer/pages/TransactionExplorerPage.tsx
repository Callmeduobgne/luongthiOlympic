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

import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, Filter, X, Hash, Clock, FileText, CheckCircle, XCircle, AlertCircle, Loader } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Badge } from '@shared/components/ui/Badge'
import { Input } from '@shared/components/ui/Input'
import { transactionExplorerService, type Transaction, type TransactionStatus } from '../services/transactionExplorerService'
import { TransactionDetailModal } from '../components/TransactionDetailModal'

const ITEMS_PER_PAGE = 20

export const TransactionExplorerPage = () => {
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null)
  const [page, setPage] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')
  const [showFilters, setShowFilters] = useState(false)
  
  // Filters
  const [channelFilter, setChannelFilter] = useState('')
  const [chaincodeFilter, setChaincodeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState<TransactionStatus | ''>('')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  // Build query params
  const queryParams = useMemo(() => {
    const params: Record<string, string> = {
      limit: ITEMS_PER_PAGE.toString(),
      offset: (page * ITEMS_PER_PAGE).toString(),
    }
    
    if (channelFilter) params.channel = channelFilter
    if (chaincodeFilter) params.chaincode = chaincodeFilter
    if (statusFilter) params.status = statusFilter
    if (startDate) {
      const start = new Date(startDate)
      start.setHours(0, 0, 0, 0)
      params.startTime = start.toISOString()
    }
    if (endDate) {
      const end = new Date(endDate)
      end.setHours(23, 59, 59, 999)
      params.endTime = end.toISOString()
    }

    return params
  }, [page, channelFilter, chaincodeFilter, statusFilter, startDate, endDate])

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['transactions', queryParams],
    queryFn: () => transactionExplorerService.listTransactions({
      channel: channelFilter || undefined,
      chaincode: chaincodeFilter || undefined,
      status: statusFilter || undefined,
      limit: ITEMS_PER_PAGE,
      offset: page * ITEMS_PER_PAGE,
      startTime: startDate ? new Date(startDate).toISOString() : undefined,
      endTime: endDate ? new Date(endDate).toISOString() : undefined,
    }),
    refetchInterval: 30000, // Refresh every 30s
  })

  // Filter transactions by search query
  const filteredTransactions = useMemo(() => {
    if (!data?.transactions) return []
    if (!searchQuery) return data.transactions
    
    const query = searchQuery.toLowerCase()
    return data.transactions.filter((tx) => {
      return (
        tx.txId.toLowerCase().includes(query) ||
        tx.id.toLowerCase().includes(query) ||
        tx.functionName.toLowerCase().includes(query) ||
        tx.chaincodeName.toLowerCase().includes(query) ||
        (tx.blockNumber?.toString().includes(query) ?? false)
      )
    })
  }, [data?.transactions, searchQuery])

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

  const formatHash = (hash: string) => {
    if (!hash) return 'N/A'
    return `${hash.substring(0, 12)}...${hash.substring(hash.length - 8)}`
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

  const clearFilters = () => {
    setChannelFilter('')
    setChaincodeFilter('')
    setStatusFilter('')
    setStartDate('')
    setEndDate('')
    setPage(0)
  }

  const hasActiveFilters = channelFilter || chaincodeFilter || statusFilter || startDate || endDate

  return (
    <div className="space-y-6 text-white">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Transaction Explorer</h1>
          <p className="text-gray-400 mt-1">
            Khám phá và xem chi tiết các transactions trên blockchain
          </p>
        </div>
        <Button
          variant="secondary"
          onClick={() => setShowFilters(!showFilters)}
        >
          <Filter className="w-4 h-4 mr-2" />
          {showFilters ? 'Ẩn' : 'Hiện'} Filters
        </Button>
      </div>

      {/* Search and Filters */}
      <Card className="p-4 text-white">
        <div className="space-y-4">
          {/* Search */}
          <div className="flex items-center gap-4">
            <div className="flex-1 relative">
              <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-white/40 w-5 h-5" />
              <Input
                placeholder="Tìm kiếm theo TxID, ID, function, chaincode, block number..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-12"
              />
            </div>
            {data && (
              <div className="text-sm text-gray-300 whitespace-nowrap">
                Tổng: <span className="font-semibold">{data.total}</span> transactions
              </div>
            )}
          </div>

          {/* Advanced Filters */}
          {showFilters && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 pt-4 border-t border-white/10">
              <div>
                <label className="block text-sm font-medium text-gray-200 mb-1">
                  Channel
                </label>
                <input
                  type="text"
                  placeholder="ibnchannel"
                  value={channelFilter}
                  onChange={(e) => {
                    setChannelFilter(e.target.value)
                    setPage(0)
                  }}
                  className="w-full px-3 py-2 rounded-2xl border border-white/15 bg-black/30 text-sm text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-200 mb-1">
                  Chaincode
                </label>
                <input
                  type="text"
                  placeholder="teaTraceCC"
                  value={chaincodeFilter}
                  onChange={(e) => {
                    setChaincodeFilter(e.target.value)
                    setPage(0)
                  }}
                  className="w-full px-3 py-2 rounded-2xl border border-white/15 bg-black/30 text-sm text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-200 mb-1">
                  Status
                </label>
                <select
                  value={statusFilter}
                  onChange={(e) => {
                    setStatusFilter(e.target.value as TransactionStatus | '')
                    setPage(0)
                  }}
                  className="w-full px-3 py-2 rounded-2xl border border-white/15 bg-black/30 text-sm text-white"
                >
                  <option value="">Tất cả</option>
                  <option value="SUBMITTED">Submitted</option>
                  <option value="VALID">Valid</option>
                  <option value="INVALID">Invalid</option>
                  <option value="FAILED">Failed</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-200 mb-1">
                  Từ ngày
                </label>
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => {
                    setStartDate(e.target.value)
                    setPage(0)
                  }}
                  className="w-full px-3 py-2 rounded-2xl border border-white/15 bg-black/30 text-sm text-white"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-200 mb-1">
                  Đến ngày
                </label>
                <input
                  type="date"
                  value={endDate}
                  onChange={(e) => {
                    setEndDate(e.target.value)
                    setPage(0)
                  }}
                  className="w-full px-3 py-2 rounded-2xl border border-white/15 bg-black/30 text-sm text-white"
                />
              </div>
            </div>
          )}

          {/* Clear Filters */}
          {hasActiveFilters && (
            <div className="flex items-center gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
              <Button variant="ghost" size="sm" onClick={clearFilters}>
                <X className="w-4 h-4 mr-1" />
                Xóa filters
              </Button>
              <span className="text-sm text-gray-300">
                {hasActiveFilters && 'Filters đang được áp dụng'}
              </span>
            </div>
          )}
        </div>
      </Card>

      {/* Transactions List */}
      <Card className="p-0 text-white">
        {isLoading && (
          <div className="p-8 text-center text-gray-300">
            <Loader className="w-6 h-6 animate-spin mx-auto mb-2" />
            Đang tải transactions...
          </div>
        )}

        {isError && (
          <div className="p-8 text-center">
            <AlertCircle className="w-12 h-12 text-red-400 mx-auto mb-4" />
            <p className="text-red-400 mb-4">
              Không thể tải transactions: {error instanceof Error ? error.message : 'Unknown error'}
            </p>
            <Button onClick={() => refetch()}>Thử lại</Button>
          </div>
        )}

        {!isLoading && !isError && filteredTransactions.length === 0 && (
          <div className="p-8 text-center text-gray-300">
            Không tìm thấy transactions nào
          </div>
        )}

        {!isLoading && !isError && filteredTransactions.length > 0 && (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="bg-white/5 border-b border-white/10 text-xs uppercase tracking-widest text-white/60">
                  <tr>
                    <th className="px-6 py-3">
                      TxID
                    </th>
                    <th className="px-6 py-3">
                      Chaincode / Function
                    </th>
                    <th className="px-6 py-3">
                      Status
                    </th>
                    <th className="px-6 py-3">
                      Block
                    </th>
                    <th className="px-6 py-3">
                      Timestamp
                    </th>
                    <th className="px-6 py-3">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/10">
                  {filteredTransactions.map((tx) => (
                    <tr
                      key={tx.id}
                      className="hover:bg-white/5 transition-colors"
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          <Hash className="w-4 h-4 text-white/60" />
                          <code className="text-xs text-gray-300 font-mono">
                            {formatHash(tx.txId)}
                          </code>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm">
                          <div className="font-medium text-white">
                            {tx.chaincodeName}
                          </div>
                          <div className="text-gray-300 text-xs">
                            {tx.functionName}
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(tx.status)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {tx.blockNumber !== undefined ? (
                          <span className="text-sm text-gray-900 dark:text-white">
                            #{tx.blockNumber}
                          </span>
                        ) : (
                          <span className="text-sm text-gray-400">-</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                        <div className="flex items-center gap-2">
                          <Clock className="w-4 h-4" />
                          {formatTimestamp(tx.timestamp)}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedTransaction(tx)}
                        >
                          <FileText className="w-4 h-4 mr-2" />
                          Chi tiết
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {data && data.total > ITEMS_PER_PAGE && (
              <div className="px-6 py-4 border-t border-white/10 flex items-center justify-between text-gray-300 text-sm">
                <div>
                  Hiển thị {page * ITEMS_PER_PAGE + 1} - {Math.min((page + 1) * ITEMS_PER_PAGE, data.total)} của {data.total}
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(0, p - 1))}
                    disabled={page === 0}
                  >
                    Trước
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={(page + 1) * ITEMS_PER_PAGE >= data.total}
                  >
                    Sau
                  </Button>
                </div>
              </div>
            )}
          </>
        )}
      </Card>

      {/* Transaction Detail Modal */}
      {selectedTransaction && (
        <TransactionDetailModal
          transaction={selectedTransaction}
          onClose={() => setSelectedTransaction(null)}
        />
      )}
    </div>
  )
}


