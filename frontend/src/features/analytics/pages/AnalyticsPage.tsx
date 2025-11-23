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
import { Activity, Cpu, Network, BarChart3, Server, Shield, Layers, Clock } from 'lucide-react'
import { Card } from '@shared/components/ui/Card'
import { StatsCard } from '@shared/components/ui/StatsCard'
import { Badge } from '@shared/components/ui/Badge'

const networkMetrics = [
  {
    title: 'Active Channels',
    value: '04',
    subtitle: 'Last 24h',
    icon: <Network className="w-5 h-5" />,
    trend: { value: 8, isPositive: true },
  },
  {
    title: 'Chaincodes',
    value: '12',
    subtitle: 'Running smart contracts',
    icon: <Layers className="w-5 h-5" />,
    trend: { value: 3, isPositive: true },
  },
  {
    title: 'Avg TPS',
    value: '64',
    subtitle: 'Transactions / second',
    icon: <Activity className="w-5 h-5" />,
    trend: { value: 5, isPositive: true },
  },
  {
    title: 'Latency',
    value: '1.2s',
    subtitle: 'Tx confirmation',
    icon: <Clock className="w-5 h-5" />,
    trend: { value: -2, isPositive: true },
  },
]

const chaincodeActivity = [
  { name: 'teaTraceCC', invocations: 482, latency: '920ms', health: 'Healthy' },
  { name: 'authCC', invocations: 207, latency: '1.1s', health: 'Degraded' },
  { name: 'supplyCC', invocations: 156, latency: '1.5s', health: 'Healthy' },
  { name: 'qualityCC', invocations: 88, latency: '1.8s', health: 'Investigate' },
]

export const AnalyticsPage = () => {
  const totalInvocations = useMemo(
    () => chaincodeActivity.reduce((sum, cc) => sum + cc.invocations, 0),
    []
  )

  return (
    <div className="space-y-8 text-white">
      <div className="flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between bg-gradient-to-br from-white/10 via-black/20 to-white/5 border border-white/10 rounded-3xl p-8 shadow-[0_25px_80px_rgba(0,0,0,0.55)]">
        <div className="space-y-3">
          <p className="inline-flex items-center gap-2 text-xs uppercase tracking-[0.3em] text-gray-300">
            <BarChart3 className="w-4 h-4" />
            Analytics
          </p>
          <h1 className="text-4xl font-semibold">Network Intelligence</h1>
          <p className="text-gray-300 max-w-2xl">
            Theo dõi hiệu suất blockchain realtime: throughput, major chaincodes và tình trạng các peers để đảm bảo hệ thống vận hành ổn định.
          </p>
        </div>
        <Card className="p-6 lg:min-w-[280px] text-white">
          <p className="text-sm text-gray-300">Transactions (24h)</p>
          <p className="text-3xl font-semibold mt-2">1,248</p>
          <div className="mt-4 flex items-center justify-between text-sm text-gray-300">
            <span>Success rate</span>
            <Badge variant="success">99.2%</Badge>
          </div>
          <div className="mt-3 h-2 rounded-full bg-white/10 overflow-hidden">
            <div className="h-full bg-gradient-to-r from-emerald-400 to-sky-400 w-[92%]" />
          </div>
        </Card>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
        {networkMetrics.map((metric) => (
          <StatsCard
            key={metric.title}
            title={metric.title}
            value={metric.value}
            subtitle={metric.subtitle}
            icon={metric.icon}
            trend={metric.trend}
          />
        ))}
      </div>

      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        <Card className="p-6 xl:col-span-2 text-white">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold">Chaincode Activity</h2>
              <p className="text-sm text-gray-300">Invocation volume và latency</p>
            </div>
            <Badge variant="primary">{totalInvocations} calls / 24h</Badge>
          </div>
          <div className="space-y-4">
            {chaincodeActivity.map((cc) => (
              <div
                key={cc.name}
                className="flex items-center justify-between rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
              >
                <div>
                  <p className="font-semibold">{cc.name}</p>
                  <p className="text-sm text-gray-300">{cc.invocations} invocations</p>
                </div>
                <div className="flex items-center gap-6 text-sm">
                  <span className="text-gray-300">
                    Latency <strong className="text-white">{cc.latency}</strong>
                  </span>
                  <Badge
                    variant={
                      cc.health === 'Healthy'
                        ? 'success'
                        : cc.health === 'Degraded'
                        ? 'warning'
                        : 'danger'
                    }
                  >
                    {cc.health}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </Card>

        <Card className="p-6 text-white">
          <h2 className="text-xl font-semibold mb-4">Node Performance</h2>
          <div className="space-y-4">
            <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              <Cpu className="w-5 h-5 text-white/80" />
              <div className="flex-1">
                <p className="text-sm text-gray-300">CPU utilization</p>
                <div className="flex items-center justify-between">
                  <span className="text-lg font-semibold">58%</span>
                  <Badge variant="default">Peers</Badge>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              <Server className="w-5 h-5 text-white/80" />
              <div className="flex-1">
                <p className="text-sm text-gray-300">Storage consumption</p>
                <div className="flex items-center justify-between">
                  <span className="text-lg font-semibold">3.4 TB</span>
                  <Badge variant="info">4 nodes</Badge>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-2xl border border-white/10 bg-white/5 px-4 py-3">
              <Shield className="w-5 h-5 text-white/80" />
              <div className="flex-1">
                <p className="text-sm text-gray-300">Security events</p>
                <div className="flex items-center justify-between">
                  <span className="text-lg font-semibold">0 critical</span>
                  <Badge variant="success">Stable</Badge>
                </div>
              </div>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )
}


