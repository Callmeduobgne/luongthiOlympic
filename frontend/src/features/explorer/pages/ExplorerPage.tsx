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

import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Search, Hash, Clock, FileText, Blocks, Receipt, ArrowRight } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Badge } from '@shared/components/ui/Badge'
import { Input } from '@shared/components/ui/Input'
import { explorerService } from '../services/explorerService'
import { BlockDetailModal } from '../components/BlockDetailModal'
import { TransactionExplorerPage } from './TransactionExplorerPage'
import type { Block } from '@shared/types/blockchain.types'

const CHANNEL = 'ibnchannel'
const ITEMS_PER_PAGE = 20

type ExplorerTab = 'blocks' | 'transactions'

export const ExplorerPage = () => {
  const [activeTab, setActiveTab] = useState<ExplorerTab>('blocks')
  const [selectedBlock, setSelectedBlock] = useState<Block | null>(null)
  const [page, setPage] = useState(0)
  const [searchQuery, setSearchQuery] = useState('')
  const [blockNumberInput, setBlockNumberInput] = useState('')

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['explorer-blocks', CHANNEL, page],
    queryFn: () => explorerService.listBlocks(CHANNEL, ITEMS_PER_PAGE, page * ITEMS_PER_PAGE),
    refetchInterval: 30000, // Refresh every 30s
  })

  // Query for specific block when searching by number
  const blockNumber = searchQuery && /^\d+$/.test(searchQuery.trim()) ? parseInt(searchQuery.trim(), 10) : null
  const { data: searchedBlock, isLoading: isLoadingBlock } = useQuery({
    queryKey: ['block', CHANNEL, blockNumber],
    queryFn: () => explorerService.getBlock(CHANNEL, blockNumber!),
    enabled: !!blockNumber && blockNumber > 0,
  })

  // Auto-open modal when block is found via search
  useEffect(() => {
    if (searchedBlock && blockNumber) {
      setSelectedBlock(searchedBlock)
    }
  }, [searchedBlock, blockNumber])

  // Filter blocks by search query
  const filteredBlocks = data?.blocks.filter((block) => {
    if (!searchQuery) return true
    const query = searchQuery.toLowerCase()
    // If it's a pure number, don't filter (we'll show the searched block instead)
    if (/^\d+$/.test(searchQuery.trim())) {
      return false
    }
    return (
      block.number.toString().includes(query) ||
      block.hash.toLowerCase().includes(query) ||
      block.transactionCount.toString().includes(query)
    )
  }) || []

  const handleViewBlock = () => {
    const num = parseInt(blockNumberInput.trim(), 10)
    if (!isNaN(num) && num > 0) {
      setSearchQuery(num.toString())
      setBlockNumberInput('')
    }
  }

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
    return `${hash.substring(0, 8)}...${hash.substring(hash.length - 8)}`
  }

  return (
    <div className="space-y-6 text-white">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Blockchain Explorer</h1>
          <p className="text-gray-400 mt-1">
            Khám phá và xem chi tiết các blocks và transactions trên blockchain
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-white/10">
        <nav className="flex space-x-4">
          <button
            onClick={() => setActiveTab('blocks')}
            className={`py-3 px-5 rounded-full text-sm font-semibold transition-all ${
              activeTab === 'blocks'
                ? 'bg-white text-black shadow-lg'
                : 'bg-white/5 text-gray-300 border border-white/10 hover:bg-white/10'
            }`}
          >
            <div className="flex items-center gap-2">
              <Blocks className="w-4 h-4" />
              Blocks
            </div>
          </button>
          <button
            onClick={() => setActiveTab('transactions')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'transactions'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            <div className="flex items-center gap-2">
              <Receipt className="w-4 h-4" />
              Transactions
            </div>
          </button>
        </nav>
      </div>

      {/* Content */}
      {activeTab === 'transactions' ? (
        <TransactionExplorerPage />
      ) : (
        <>

      {/* Search and Filters */}
      <Card className="p-4 text-white">
        <div className="flex items-center gap-4 flex-wrap">
          <div className="flex-1 relative min-w-[300px]">
            <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-white/40 w-5 h-5" />
            <Input
              placeholder="Tìm kiếm theo block number, hash..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-11"
            />
            {isLoadingBlock && blockNumber && (
              <div className="absolute right-4 top-1/2 transform -translate-y-1/2 text-xs text-gray-400">
                Đang tải...
              </div>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              placeholder="Block #"
              value={blockNumberInput}
              onChange={(e) => setBlockNumberInput(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleViewBlock()}
              className="w-24"
              min="0"
            />
            <Button
              variant="secondary"
              size="sm"
              onClick={handleViewBlock}
              disabled={!blockNumberInput.trim() || isNaN(parseInt(blockNumberInput.trim(), 10))}
            >
              <ArrowRight className="w-4 h-4 mr-1" />
              Xem Block
            </Button>
          </div>
          {data && (
            <div className="text-sm text-gray-300 whitespace-nowrap">
              Tổng: <span className="font-semibold">{data.total}</span> blocks
            </div>
          )}
        </div>
        {blockNumber && searchedBlock && (
          <div className="mt-3 p-3 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Hash className="w-4 h-4 text-blue-400" />
                <span className="text-sm text-blue-300">
                  Đã tìm thấy Block #{searchedBlock.number}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setSelectedBlock(searchedBlock)}
              >
                Xem chi tiết
              </Button>
            </div>
          </div>
        )}
        {blockNumber && !searchedBlock && !isLoadingBlock && (
          <div className="mt-3 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
            <span className="text-sm text-yellow-300">
              Không tìm thấy Block #{blockNumber}
            </span>
          </div>
        )}
      </Card>

      {/* Blocks List */}
      <Card className="p-0 text-white">
        {isLoading && (
          <div className="p-8 text-center text-gray-300">
            Đang tải blocks...
          </div>
        )}

        {isError && (
          <div className="p-8 text-center">
            <p className="text-red-400 mb-4">
              Không thể tải blocks: {error instanceof Error ? error.message : 'Unknown error'}
            </p>
            <Button onClick={() => window.location.reload()}>Thử lại</Button>
          </div>
        )}

        {!isLoading && !isError && filteredBlocks.length === 0 && !blockNumber && (
          <div className="p-8 text-center text-gray-300">
            {searchQuery ? 'Không tìm thấy blocks nào phù hợp' : 'Không có blocks nào'}
          </div>
        )}

        {!isLoading && !isError && filteredBlocks.length > 0 && (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-left">
                <thead className="bg-white/5 border-b border-white/10 text-xs uppercase tracking-widest text-white/60">
                  <tr>
                    <th className="px-6 py-3">
                      Block #
                    </th>
                    <th className="px-6 py-3">
                      Hash
                    </th>
                    <th className="px-6 py-3">
                      Transactions
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
                  {filteredBlocks.map((block) => (
                    <tr
                      key={block.number}
                      className="hover:bg-white/5 transition-colors"
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          <Hash className="w-4 h-4 text-white/60" />
                          <span className="text-sm font-medium">
                            {block.number}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <code className="text-xs text-gray-300 font-mono">
                          {formatHash(block.hash)}
                        </code>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <Badge variant="default">{block.transactionCount}</Badge>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                        <div className="flex items-center gap-2">
                          <Clock className="w-4 h-4" />
                          {formatTimestamp(block.timestamp)}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setSelectedBlock(block)}
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

      {/* Block Detail Modal */}
      {selectedBlock && (
        <BlockDetailModal
          block={selectedBlock}
          onClose={() => {
            setSelectedBlock(null)
            // Clear search when closing modal if it was a number search
            if (blockNumber) {
              setSearchQuery('')
            }
          }}
        />
      )}
        </>
      )}
    </div>
  )
}

