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

import { useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card } from '@shared/components/ui/Card'
import { Badge } from '@shared/components/ui/Badge'
import { Button } from '@shared/components/ui/Button'
import { LoadingState } from '@shared/components/common/LoadingState'
import { EmptyState } from '@shared/components/common/EmptyState'
import { networkService } from '../services/networkService'
import { NetworkLogViewer } from '../components/NetworkLogViewer'
import { Network, ShieldCheck, Server, Wifi, Workflow, AlertTriangle } from 'lucide-react'

const getPeerStatusVariant = (status?: string) => {
  const normalized = status?.toLowerCase()
  if (normalized === 'connected' || normalized === 'healthy') return 'success'
  if (normalized === 'unknown') return 'warning'
  return 'danger'
}

const getOrdererStatusVariant = (status?: string) => {
  const normalized = status?.toLowerCase()
  if (normalized === 'healthy') return 'success'
  if (normalized === 'unknown') return 'warning'
  return 'danger'
}

export const NetworkPage = () => {
  const {
    data: overview,
    isLoading: overviewLoading,
    isError: overviewError,
    error: overviewErrorObj,
    refetch: refetchOverview,
  } = useQuery({
    queryKey: ['network-overview'],
    queryFn: networkService.getOverview,
    refetchInterval: 60000,
  })

  const {
    data: peersData,
    isLoading: peersLoading,
    isError: peersError,
    error: peersErrorObj,
    refetch: refetchPeers,
  } = useQuery({
    queryKey: ['network-peers'],
    queryFn: networkService.listPeers,
    refetchInterval: 45000,
  })

  const {
    data: orderersData,
    isLoading: orderersLoading,
    isError: orderersError,
    error: orderersErrorObj,
    refetch: refetchOrderers,
  } = useQuery({
    queryKey: ['network-orderers'],
    queryFn: networkService.listOrderers,
    refetchInterval: 60000,
  })

  const peers = peersData ?? overview?.peers ?? []
  const orderers = orderersData ?? overview?.orderers ?? []
  const channels = overview?.channels ?? []

  const chaincodeCount = useMemo(() => {
    const set = new Set<string>()
    channels.forEach((channel) => {
      channel.chaincodes?.forEach((cc) => set.add(cc))
    })
    return set.size
  }, [channels])

  const onlinePeers = peers.filter((p) => (p.status || '').toLowerCase() === 'connected').length
  const degradedPeers = peers.filter(
    (p) => ['disconnected', 'unknown', 'unhealthy'].includes((p.status || '').toLowerCase())
  ).length

  const securityScore = peers.length ? ((onlinePeers / peers.length) * 100).toFixed(1) : '100'

  const isLoading = overviewLoading || peersLoading || orderersLoading
  const hasError = overviewError || peersError || orderersError

  const handleRetry = () => {
    refetchOverview()
    refetchPeers()
    refetchOrderers()
  }

  if (isLoading) {
    return <LoadingState text="Đang tải thông tin mạng..." fullScreen />
  }

  if (hasError) {
    const message =
      overviewErrorObj?.message || peersErrorObj?.message || orderersErrorObj?.message || 'Không thể tải dữ liệu mạng.'
    return (
      <EmptyState
        title="Không thể tải thông tin mạng"
        description={message}
        action={{
          label: 'Thử lại',
          onClick: handleRetry,
        }}
      />
    )
  }

  return (
    <div className="space-y-8 text-white">
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <Card className="xl:col-span-2 p-8 bg-gradient-to-br from-white/10 via-black/20 to-white/5 border border-white/10 shadow-[0_25px_80px_rgba(0,0,0,0.55)] text-white">
          <div className="flex flex-col gap-6">
            <div className="flex items-center gap-3 text-sm uppercase tracking-[0.4em] text-gray-300">
              <Network className="w-4 h-4" />
              Network Topology
            </div>
            <div>
              <h1 className="text-4xl font-semibold">IBN Fabric Network</h1>
              <p className="text-gray-200 mt-2 max-w-3xl">
                Theo dõi trạng thái realtime của toàn bộ peers, orderers và chaincodes. Hệ thống sẽ cảnh báo khi có node offline hoặc hiệu năng giảm.
              </p>
            </div>
            <div className="flex flex-wrap gap-4">
              <Badge variant="primary">{channels.length} Channels</Badge>
              <Badge variant="default">{chaincodeCount} Chaincodes</Badge>
              <Badge variant="success">{onlinePeers} Peers online</Badge>
              {degradedPeers > 0 && <Badge variant="warning">{degradedPeers} Peers cần kiểm tra</Badge>}
            </div>
          </div>
        </Card>

        <Card className="p-6 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-300">Security posture</p>
              <p className="text-3xl font-semibold mt-2">{securityScore}%</p>
            </div>
            <ShieldCheck className="w-10 h-10 text-emerald-300" />
          </div>
          <p className="text-sm text-gray-400">
            {peers.length
              ? `Đang giám sát ${peers.length} nodes, ${onlinePeers} node hoạt động bình thường.`
              : 'Chưa có dữ liệu nodes để đánh giá.'}
          </p>
          <Button variant="secondary" className="w-full" onClick={handleRetry}>
            View Alerts
          </Button>
        </Card>
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <Card className="p-6 text-white">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold">Peers</h2>
              <p className="text-sm text-gray-400">Chi tiết từng node</p>
            </div>
            <Badge variant="default">{peers.length} nodes</Badge>
          </div>
          <div className="space-y-4">
            {peers.map((peer) => (
              <div key={peer.name} className="rounded-2xl border border-white/10 bg-white/5 p-4 flex flex-col gap-3">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-semibold">{peer.name}</p>
                    <p className="text-sm text-gray-300">{peer.mspId || 'Unknown MSP'}</p>
                  </div>
                  <Badge variant={getPeerStatusVariant(peer.status)}>{peer.status || 'Unknown'}</Badge>
                </div>
                <div className="grid grid-cols-2 gap-4 text-sm text-gray-300">
                  <div className="flex items-center gap-2">
                    <Wifi className="w-4 h-4" />
                    Endpoint <span className="font-semibold text-white break-all">{peer.address || '-'}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Server className="w-4 h-4" />
                    Block height <span className="font-semibold text-white">{peer.blockHeight ?? '-'}</span>
                  </div>
                  <div className="flex items-center gap-2 col-span-2">
                    <Workflow className="w-4 h-4" />
                    Channels{' '}
                    <span className="font-semibold text-white">
                      {peer.channels?.length ? peer.channels.join(', ') : '-'}
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </Card>

        <Card className="p-6 text-white space-y-6">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-semibold">Orderers</h2>
              <p className="text-sm text-gray-400">Raft cluster status</p>
            </div>
            <Badge variant="info">Raft</Badge>
          </div>
          <div className="space-y-4">
            {orderers.map((orderer) => (
              <div
                key={orderer.name}
                className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 flex items-center justify-between"
              >
                <div>
                  <p className="font-semibold">{orderer.name}</p>
                  <p className="text-sm text-gray-300">
                    {orderer.isLeader ? 'Leader' : 'Follower'} • {orderer.mspId || 'OrdererMSP'}
                  </p>
                </div>
                <div className="text-right text-sm text-gray-300">
                  <Badge variant={getOrdererStatusVariant(orderer.status)}>{orderer.status || 'Unknown'}</Badge>
                  <p className="mt-2 break-all">{orderer.address}</p>
                </div>
              </div>
            ))}
          </div>
          <div className="rounded-2xl border border-white/10 bg-white/5 p-4 flex items-center gap-3">
            <AlertTriangle className="w-5 h-5 text-amber-300" />
            <div>
              <p className="text-sm font-semibold">Failover readiness</p>
              <p className="text-xs text-gray-300">
                Theo dõi trạng thái các follower và cảnh báo khi leader không phản hồi.
              </p>
            </div>
          </div>
        </Card>
      </div>

      {/* Network Logs Section */}
      <div className="mt-8">
        <NetworkLogViewer
          autoRefresh={true}
          refreshInterval={3000}
        />
      </div>
    </div>
  )
}


