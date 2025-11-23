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
import { useQuery, useQueryClient, type UseQueryOptions, type UseQueryResult } from '@tanstack/react-query'
import { Package, CheckCircle, Upload, TrendingUp, Clock, RotateCcw, Play, Search } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Badge } from '@shared/components/ui/Badge'
import { chaincodeService } from '../services/chaincodeService'
import { rollbackService } from '../services/rollbackService'
import { LoadingState } from '@shared/components/common/LoadingState'
import { EmptyState } from '@shared/components/common/EmptyState'
import { DeployChaincodeModal } from '../components/DeployChaincodeModal'
import { InvokeQueryChaincodeModal } from '../components/InvokeQueryChaincodeModal'
import { ErrorBoundary } from '@shared/components/common/ErrorBoundary'

const CHANNEL = 'ibnchannel'

// Safe wrapper for useQuery to prevent crashes
type SafeQueryOptions<TData, TError = unknown> = UseQueryOptions<TData, TError> & {
  queryFn: () => Promise<TData>
}

const useSafeQuery = <TData, TError = unknown>(
  options: SafeQueryOptions<TData, TError>
): UseQueryResult<TData, TError> => {
  const { queryFn, initialData, ...rest } = options

  return useQuery({
    ...rest,
    queryFn: async () => {
      try {
        return await queryFn()
      } catch (error) {
        if (import.meta.env.DEV) {
          console.error('[SafeQuery] Error in query:', error)
          // Log detailed error for debugging
          if (error instanceof Error) {
            console.error('[SafeQuery] Error message:', error.message)
            console.error('[SafeQuery] Error stack:', error.stack)
          }
        }
        // Don't swallow errors - let React Query handle them
        // This allows UI to show error state
        throw error
      }
    },
    initialData,
    retry: 1,
    retryDelay: 1000,
    staleTime: 30000,
    gcTime: 5 * 60 * 1000,
  })
}

const DeployChaincodePageContent = () => {
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'installed' | 'committed' | 'rollback'>('installed')
  const [showDeployModal, setShowDeployModal] = useState(false)
  const [showInvokeModal, setShowInvokeModal] = useState(false)
  const [showQueryModal, setShowQueryModal] = useState(false)
  const [selectedChaincodeForDeploy, setSelectedChaincodeForDeploy] = useState<{
    packageId: string
    label: string
    name: string
    version: string
  } | null>(null)

  // Use safe queries with comprehensive error handling
  const { data: installedChaincodes, isLoading: isLoadingInstalled, error: installedError } = useSafeQuery({
    queryKey: ['chaincodes-installed'],
    queryFn: async () => {
      try {
        const result = await chaincodeService.listInstalled()
        if (import.meta.env.DEV) {
          console.log('[DeployChaincodePage] Received chaincodes:', result)
          console.log('[DeployChaincodePage] Is array?', Array.isArray(result))
          console.log('[DeployChaincodePage] Count:', Array.isArray(result) ? result.length : 0)
        }
        return Array.isArray(result) ? result : []
      } catch (error) {
        if (import.meta.env.DEV) {
          console.error('[DeployChaincodePage] Error in queryFn:', error)
        }
        throw error
      }
    },
    refetchInterval: 30000,
    initialData: [],
  })

  const { data: committedChaincodes, isLoading: isLoadingCommitted, error: committedError } = useSafeQuery({
    queryKey: ['chaincodes-committed', CHANNEL],
    queryFn: async () => {
      const result = await chaincodeService.listCommitted(CHANNEL)
      return Array.isArray(result) ? result : []
    },
    refetchInterval: 30000,
    initialData: [],
  })

  // Phase 4: Rollback operations - with error handling
  const { data: rollbackOperations, isLoading: isLoadingRollbacks, error: rollbackError } = useSafeQuery({
    queryKey: ['rollback-operations', CHANNEL],
    queryFn: async () => {
      try {
        const result = await rollbackService.listRollbacks({ channel_name: CHANNEL })
        return Array.isArray(result) ? result : []
      } catch (error) {
        if (import.meta.env.DEV) {
          console.warn('[DeployChaincodePage] Rollback service not available:', error)
        }
        return []
      }
    },
    refetchInterval: 30000,
    initialData: [],
    enabled: true,
  })

  // Statistics - with comprehensive safety checks using useMemo
  const installedChaincodesSafe = useMemo(() => {
    try {
      if (import.meta.env.DEV) {
        console.log('[DeployChaincodePage] Filtering chaincodes, input:', installedChaincodes)
      }
      if (!Array.isArray(installedChaincodes)) {
        if (import.meta.env.DEV) {
          console.warn('[DeployChaincodePage] installedChaincodes is not an array:', typeof installedChaincodes)
        }
        return []
      }
      const filtered = installedChaincodes.filter(cc => {
        try {
          const isValid = cc && 
                 typeof cc === 'object' && 
                 cc.packageId && 
                 cc.chaincode && 
                 typeof cc.chaincode === 'object'
          if (!isValid && import.meta.env.DEV) {
            console.warn('[DeployChaincodePage] Filtered out chaincode:', cc)
          }
          return isValid
        } catch (err) {
          if (import.meta.env.DEV) {
            console.warn('[DeployChaincodePage] Error checking chaincode:', err, cc)
          }
          return false
        }
      })
      if (import.meta.env.DEV) {
        console.log('[DeployChaincodePage] Filtered chaincodes count:', filtered.length)
      }
      return filtered
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DeployChaincodePage] Error filtering installed chaincodes:', error)
      }
      return []
    }
  }, [installedChaincodes])

  const committedChaincodesSafe = useMemo(() => {
    try {
      if (!Array.isArray(committedChaincodes)) return []
      return committedChaincodes.filter(cc => {
        try {
          return cc && 
                 typeof cc === 'object' && 
                 cc.name && 
                 cc.version
        } catch {
          return false
        }
      })
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DeployChaincodePage] Error filtering committed chaincodes:', error)
      }
      return []
    }
  }, [committedChaincodes])

  const rollbackOperationsSafe = useMemo(() => {
    try {
      if (!Array.isArray(rollbackOperations)) return []
      return rollbackOperations.filter(op => {
        try {
          return op && typeof op === 'object' && op.id
        } catch {
          return false
        }
      })
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[DeployChaincodePage] Error filtering rollback operations:', error)
      }
      return []
    }
  }, [rollbackOperations])
  
  const installedCount = installedChaincodesSafe.length
  const committedCount = committedChaincodesSafe.length
  const pendingCount = Math.max(0, installedCount - committedCount)

  const statsCards = [
    {
      label: 'Installed',
      value: installedCount,
      icon: Package,
      color: 'text-blue-400',
      gradient: 'from-blue-500/20 via-blue-500/5 to-transparent',
      description: 'Chaincode đã được install',
    },
    {
      label: 'Committed',
      value: committedCount,
      icon: CheckCircle,
      color: 'text-emerald-400',
      gradient: 'from-emerald-500/20 via-emerald-500/5 to-transparent',
      description: 'Chaincode đã được commit',
    },
    {
      label: 'Pending',
      value: pendingCount,
      icon: Clock,
      color: 'text-amber-400',
      gradient: 'from-amber-500/20 via-amber-500/5 to-transparent',
      description: 'Chờ approve/commit',
    },
    {
      label: 'Success Rate',
      value: installedCount > 0 ? `${Math.round((committedCount / installedCount) * 100)}%` : '0%',
      icon: TrendingUp,
      color: 'text-violet-400',
      gradient: 'from-violet-500/20 via-violet-500/5 to-transparent',
      description: 'Tỉ lệ deploy thành công',
    },
  ]

  return (
    <div className="space-y-6 text-white">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Deploy Chaincode</h1>
          <p className="mt-1 text-sm text-gray-400">
            Quản lý và triển khai chaincode trên blockchain network
          </p>
        </div>
        <div className="flex gap-3">
          <Button variant="primary" onClick={() => setShowDeployModal(true)}>
            <Upload className="h-4 w-4 mr-2" />
            Deploy Chaincode
          </Button>
          <Button variant="secondary" onClick={() => setShowInvokeModal(true)}>
            <Play className="h-4 w-4 mr-2" />
            Invoke Chaincode
          </Button>
          <Button variant="secondary" onClick={() => setShowQueryModal(true)}>
            <Search className="h-4 w-4 mr-2" />
            Query Chaincode
          </Button>
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        {statsCards.map((stat) => {
          const Icon = stat.icon
          return (
            <div
              key={stat.label}
              className="group relative overflow-hidden rounded-3xl text-white p-5 transition-all duration-700 ease-out hover:-translate-y-2 hover:scale-[1.02]"
              style={{
                background: 'rgba(255, 255, 255, 0.01)',
                backdropFilter: 'blur(20px) saturate(180%)',
                WebkitBackdropFilter: 'blur(20px) saturate(180%)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                boxShadow: '0 8px 32px 0 rgba(0, 0, 0, 0.37)',
              }}
            >
              <div className={`absolute inset-0 bg-gradient-to-br ${stat.gradient} opacity-0 group-hover:opacity-100 transition-opacity duration-700`} />
              <div className="relative z-10">
                <div className="flex items-center justify-between mb-3">
                  <div className={`p-2.5 rounded-xl bg-white/5 border border-white/10 ${stat.color} bg-opacity-10`}>
                    <Icon className="w-5 h-5" />
                  </div>
                </div>
                <div className="space-y-1">
                  <p className="text-2xl font-bold">{stat.value}</p>
                  <p className="text-sm font-medium text-gray-300">{stat.label}</p>
                  <p className="text-xs text-gray-400">{stat.description}</p>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Tabs */}
      <div className="border-b border-white/10">
        <nav className="flex space-x-4">
          <button
            onClick={() => setActiveTab('installed')}
            className={`py-3 px-5 rounded-full text-sm font-semibold transition-all ${
              activeTab === 'installed'
                ? 'bg-white text-black shadow-lg'
                : 'bg-white/5 text-gray-300 border border-white/10 hover:bg-white/10'
            }`}
          >
            <div className="flex items-center gap-2">
              <Package className="w-4 h-4" />
              Installed
            </div>
          </button>
          <button
            onClick={() => setActiveTab('committed')}
            className={`py-3 px-5 rounded-full text-sm font-semibold transition-all ${
              activeTab === 'committed'
                ? 'bg-white text-black shadow-lg'
                : 'bg-white/5 text-gray-300 border border-white/10 hover:bg-white/10'
            }`}
          >
            <div className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4" />
              Committed
            </div>
          </button>
          <button
            onClick={() => setActiveTab('rollback')}
            className={`py-3 px-5 rounded-full text-sm font-semibold transition-all ${
              activeTab === 'rollback'
                ? 'bg-white text-black shadow-lg'
                : 'bg-white/5 text-gray-300 border border-white/10 hover:bg-white/10'
            }`}
          >
            <div className="flex items-center gap-2">
              <RotateCcw className="w-4 h-4" />
              Rollback
            </div>
          </button>
        </nav>
      </div>

      {/* Installed Chaincodes */}
      {activeTab === 'installed' && (
        <Card className="p-0">
          <div className="p-6 border-b border-white/10">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-white">Installed Chaincodes</h2>
                <p className="text-sm text-gray-400 mt-1">
                  Danh sách chaincode đã được install trên peer
                </p>
              </div>
              {installedChaincodesSafe.length > 0 && (
                <Badge variant="default" className="text-sm">
                  {installedChaincodesSafe.length} chaincode{installedChaincodesSafe.length !== 1 ? 's' : ''}
                </Badge>
              )}
            </div>
          </div>

          <div className="p-6">
            {isLoadingInstalled && <LoadingState text="Đang tải installed chaincodes..." />}

            {installedError ? (
              <div className="text-center py-8">
                <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-6 max-w-2xl mx-auto">
                  <p className="text-red-400 mb-2 font-semibold">⚠️ Lỗi khi tải installed chaincodes</p>
                  <p className="text-sm text-gray-300 mb-4">
                    {installedError instanceof Error ? installedError.message : 'Unknown error'}
                  </p>
                  <div className="text-xs text-gray-400 space-y-1 text-left bg-black/20 p-3 rounded">
                    <p><strong>Nguyên nhân có thể:</strong></p>
                    <ul className="list-disc list-inside space-y-1 ml-2">
                      {installedError instanceof Error && installedError.message.includes('401') ? (
                        <>
                          <li>Token đã hết hạn hoặc không hợp lệ</li>
                          <li>Vui lòng đăng nhập lại</li>
                        </>
                      ) : (
                        <>
                          <li>Peer CLI chưa được cài đặt trong admin-service container</li>
                          <li>Admin Service không thể kết nối đến peer</li>
                          <li>Lỗi khi query chaincode từ peer</li>
                        </>
                      )}
                    </ul>
                    <p className="mt-2"><strong>Giải pháp:</strong> {
                      installedError instanceof Error && installedError.message.includes('401') 
                        ? 'Đăng nhập lại để lấy token mới'
                        : 'Kiểm tra logs của admin-service container'
                    }</p>
                  </div>
                  <div className="flex gap-2 justify-center mt-4">
                    {installedError instanceof Error && installedError.message.includes('401') ? (
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => {
                          // Redirect to login
                          window.location.href = '/login'
                        }}
                      >
                        Đăng nhập lại
                      </Button>
                    ) : (
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => window.location.reload()}
                      >
                        Thử lại
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            ) : null}

            {!isLoadingInstalled && !installedError && installedChaincodesSafe.length === 0 && (
              <EmptyState
                icon="package"
                title="Chưa có chaincode nào được install"
                description="Deploy chaincode đầu tiên để bắt đầu"
                action={{
                  label: 'Deploy Chaincode',
                  onClick: () => setShowDeployModal(true),
                }}
              />
            )}

            {!isLoadingInstalled && !installedError && installedChaincodesSafe.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead className="bg-white/5 border-b border-white/10 text-xs uppercase tracking-widest text-white/60">
                    <tr>
                      <th className="px-6 py-3">Package ID</th>
                      <th className="px-6 py-3">Label</th>
                      <th className="px-6 py-3">Name</th>
                      <th className="px-6 py-3">Version</th>
                      <th className="px-6 py-3">Path</th>
                      <th className="px-6 py-3 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/10">
                    {installedChaincodesSafe.map((cc) => {
                      if (!cc || !cc.packageId || !cc.chaincode) return null
                      return (
                        <tr key={cc.packageId} className="hover:bg-white/5 transition-colors">
                          <td className="px-6 py-4">
                            <code className="text-xs font-mono text-gray-300">
                              {cc.packageId?.substring?.(0, 20) || 'N/A'}...
                            </code>
                          </td>
                          <td className="px-6 py-4">
                            <Badge variant="default">{cc.label || 'N/A'}</Badge>
                          </td>
                          <td className="px-6 py-4 font-medium">{cc.chaincode?.name || 'N/A'}</td>
                          <td className="px-6 py-4">
                            <Badge variant="default">{cc.chaincode?.version || 'N/A'}</Badge>
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-400">{cc.chaincode?.path || 'N/A'}</td>
                          <td className="px-6 py-4 text-right">
                            <Button
                              variant="primary"
                              size="sm"
                              onClick={() => {
                                // Parse label to extract name and version
                                // Label format: <name>_<version> (e.g., "teaTraceCC_1.0")
                                const label = cc.label || ''
                                const lastUnderscoreIndex = label.lastIndexOf('_')
                                const parsedName = lastUnderscoreIndex > 0 
                                  ? label.substring(0, lastUnderscoreIndex) 
                                  : (cc.chaincode?.name || label)
                                const parsedVersion = lastUnderscoreIndex > 0 
                                  ? label.substring(lastUnderscoreIndex + 1) 
                                  : (cc.chaincode?.version || '1.0')
                                
                                // Open deploy modal with pre-filled data from installed chaincode
                                setSelectedChaincodeForDeploy({
                                  packageId: cc.packageId,
                                  label: label,
                                  name: parsedName,
                                  version: parsedVersion,
                                })
                                setShowDeployModal(true)
                              }}
                            >
                              <CheckCircle className="w-4 h-4 mr-1" />
                              Deploy
                            </Button>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Committed Chaincodes */}
      {activeTab === 'committed' && (
        <Card className="p-0">
          <div className="p-6 border-b border-white/10">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-white">Committed Chaincodes</h2>
                <p className="text-sm text-gray-400 mt-1">
                  Danh sách chaincode đã được commit trên channel {CHANNEL}
                </p>
              </div>
              {committedChaincodesSafe.length > 0 && (
                <Badge variant="default" className="text-sm">
                  {committedChaincodesSafe.length} chaincode{committedChaincodesSafe.length !== 1 ? 's' : ''}
                </Badge>
              )}
            </div>
          </div>

          <div className="p-6">
            {isLoadingCommitted && <LoadingState text="Đang tải committed chaincodes..." />}

            {committedError ? (
              <div className="text-center py-8">
                <p className="text-red-400 mb-2">Lỗi khi tải committed chaincodes</p>
                <p className="text-sm text-gray-400">{committedError instanceof Error ? committedError.message : 'Unknown error'}</p>
              </div>
            ) : null}

            {!isLoadingCommitted && !committedError && committedChaincodesSafe.length === 0 && (
              <EmptyState
                icon="check-circle"
                title="Chưa có chaincode nào được commit"
                description="Approve và commit chaincode để sử dụng trên channel"
                action={{
                  label: 'Deploy Chaincode',
                  onClick: () => setShowDeployModal(true),
                }}
              />
            )}

            {!isLoadingCommitted && !committedError && committedChaincodesSafe.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead className="bg-white/5 border-b border-white/10 text-xs uppercase tracking-widest text-white/60">
                    <tr>
                      <th className="px-6 py-3">Name</th>
                      <th className="px-6 py-3">Version</th>
                      <th className="px-6 py-3">Sequence</th>
                      <th className="px-6 py-3">Endorsement Plugin</th>
                      <th className="px-6 py-3">Validation Plugin</th>
                      <th className="px-6 py-3">Init Required</th>
                      <th className="px-6 py-3">Approved Orgs</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/10">
                    {committedChaincodesSafe.map((cc) => {
                      if (!cc || !cc.name || !cc.version) return null
                      return (
                        <tr key={`${cc.name}-${cc.version}`} className="hover:bg-white/5 transition-colors">
                          <td className="px-6 py-4 font-medium">{cc.name || 'N/A'}</td>
                          <td className="px-6 py-4">
                            <Badge variant="default">{cc.version || 'N/A'}</Badge>
                          </td>
                          <td className="px-6 py-4">{cc.sequence ?? 'N/A'}</td>
                          <td className="px-6 py-4 text-sm text-gray-400">{cc.endorsementPlugin || 'N/A'}</td>
                          <td className="px-6 py-4 text-sm text-gray-400">{cc.validationPlugin || 'N/A'}</td>
                          <td className="px-6 py-4">
                            {cc.initRequired ? (
                              <Badge variant="warning">Yes</Badge>
                            ) : (
                              <Badge variant="default">No</Badge>
                            )}
                          </td>
                          <td className="px-6 py-4">
                            {cc.approvedOrganizations && Array.isArray(cc.approvedOrganizations) && cc.approvedOrganizations.length > 0 ? (
                              <div className="flex flex-wrap gap-1">
                                {cc.approvedOrganizations.map((org, idx) => (
                                  <Badge key={org || idx} variant="default" className="text-xs">
                                    {org || 'N/A'}
                                  </Badge>
                                ))}
                              </div>
                            ) : (
                              <span className="text-sm text-gray-500">Chưa có phê duyệt</span>
                            )}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Phase 4: Rollback Tab */}
      {activeTab === 'rollback' && (
        <Card className="p-0">
          <div className="p-6 border-b border-white/10">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-white">Rollback Operations</h2>
                <p className="text-sm text-gray-400 mt-1">
                  Quản lý rollback operations cho chaincode trên channel {CHANNEL}
                </p>
              </div>
              {rollbackOperationsSafe.length > 0 && (
                <Badge variant="default" className="text-sm">
                  {rollbackOperationsSafe.length} operation{rollbackOperationsSafe.length !== 1 ? 's' : ''}
                </Badge>
              )}
            </div>
          </div>

          <div className="p-6">
            {isLoadingRollbacks && <LoadingState text="Đang tải rollback operations..." />}

            {rollbackError ? (
              <div className="text-center py-8">
                <p className="text-red-400 mb-2">Lỗi khi tải rollback operations</p>
                <p className="text-sm text-gray-400">{rollbackError instanceof Error ? rollbackError.message : 'Unknown error'}</p>
              </div>
            ) : null}

            {!isLoadingRollbacks && !rollbackError && rollbackOperationsSafe.length === 0 && (
              <EmptyState
                icon="rotate-ccw"
                title="Chưa có rollback operation nào"
                description="Rollback operations sẽ được hiển thị ở đây"
              />
            )}

            {!isLoadingRollbacks && !rollbackError && rollbackOperationsSafe.length > 0 && (
              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead className="bg-white/5 border-b border-white/10 text-xs uppercase tracking-widest text-white/60">
                    <tr>
                      <th className="px-6 py-3">Chaincode</th>
                      <th className="px-6 py-3">From Version</th>
                      <th className="px-6 py-3">To Version</th>
                      <th className="px-6 py-3">Status</th>
                      <th className="px-6 py-3">Type</th>
                      <th className="px-6 py-3">Created At</th>
                      <th className="px-6 py-3">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/10">
                    {rollbackOperationsSafe.map((op) => {
                      if (!op || !op.id) return null
                      return (
                        <tr key={op.id} className="hover:bg-white/5 transition-colors">
                          <td className="px-6 py-4 font-medium">{op.chaincode_name || 'N/A'}</td>
                          <td className="px-6 py-4">
                            <Badge variant="default">{op.from_version || 'N/A'}</Badge>
                          </td>
                          <td className="px-6 py-4">
                            <Badge variant="primary">{op.to_version || 'N/A'}</Badge>
                          </td>
                          <td className="px-6 py-4">
                            <Badge
                              variant={
                                op.status === 'completed'
                                  ? 'success'
                                  : op.status === 'failed'
                                  ? 'danger'
                                  : op.status === 'in_progress'
                                  ? 'warning'
                                  : 'default'
                              }
                            >
                              {op.status || 'unknown'}
                            </Badge>
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-400">{op.rollback_type || 'N/A'}</td>
                          <td className="px-6 py-4 text-sm text-gray-400">
                            {op.created_at ? new Date(op.created_at).toLocaleString() : 'N/A'}
                          </td>
                          <td className="px-6 py-4">
                            {op.status === 'pending' && (
                              <Button
                                variant="secondary"
                                size="sm"
                                onClick={async () => {
                                  try {
                                    if (op.id) {
                                      await rollbackService.executeRollback(op.id)
                                      // Refetch rollback operations
                                      queryClient.invalidateQueries({ queryKey: ['rollback-operations'] })
                                    }
                                  } catch (error) {
                                    if (import.meta.env.DEV) {
                                      console.error('Failed to execute rollback:', error)
                                    }
                                  }
                                }}
                              >
                                Execute
                              </Button>
                            )}
                            {op.status === 'pending' && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={async () => {
                                  try {
                                    if (op.id) {
                                      await rollbackService.cancelRollback(op.id)
                                      queryClient.invalidateQueries({ queryKey: ['rollback-operations'] })
                                    }
                                  } catch (error) {
                                    if (import.meta.env.DEV) {
                                      console.error('Failed to cancel rollback:', error)
                                    }
                                  }
                                }}
                                className="ml-2"
                              >
                                Cancel
                              </Button>
                            )}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Deploy Modal */}
      {showDeployModal && (
        <ErrorBoundary
          fallback={
            <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-red-400">Không thể mở modal deploy. Vui lòng thử lại.</p>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setShowDeployModal(false)}
                className="mt-2"
              >
                Đóng
              </Button>
            </div>
          }
        >
          <DeployChaincodeModal 
            onClose={() => {
              setShowDeployModal(false)
              setSelectedChaincodeForDeploy(null)
            }}
            initialData={selectedChaincodeForDeploy || undefined}
          />
        </ErrorBoundary>
      )}

      {/* Invoke Modal */}
      {showInvokeModal && (
        <ErrorBoundary
          fallback={
            <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-red-400">Không thể mở modal invoke. Vui lòng thử lại.</p>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setShowInvokeModal(false)}
                className="mt-2"
              >
                Đóng
              </Button>
            </div>
          }
        >
          <InvokeQueryChaincodeModal
            mode="invoke"
            onClose={() => setShowInvokeModal(false)}
          />
        </ErrorBoundary>
      )}

      {/* Query Modal */}
      {showQueryModal && (
        <ErrorBoundary
          fallback={
            <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
              <p className="text-red-400">Không thể mở modal query. Vui lòng thử lại.</p>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setShowQueryModal(false)}
                className="mt-2"
              >
                Đóng
              </Button>
            </div>
          }
        >
          <InvokeQueryChaincodeModal
            mode="query"
            onClose={() => setShowQueryModal(false)}
          />
        </ErrorBoundary>
      )}
    </div>
  )
}

// Export with Error Boundary wrapper for production safety
export const DeployChaincodePage = () => {
  return (
    <ErrorBoundary
      onError={(error, errorInfo) => {
        // Log to error tracking service in production
        if (import.meta.env.PROD) {
          // Example: Sentry.captureException(error, { extra: errorInfo })
          console.error('[Production] DeployChaincodePage error:', error, errorInfo)
        }
      }}
    >
      <DeployChaincodePageContent />
    </ErrorBoundary>
  )
}

