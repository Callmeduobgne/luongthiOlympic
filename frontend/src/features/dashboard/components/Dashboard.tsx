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

import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Activity,
  Blocks,
  TrendingUp,
  Clock,
  CheckCircle,
  XCircle,
  Wifi,
  WifiOff,
  ShieldCheck,
  Signal,
  Sparkles,
  ArrowUpRight,
} from 'lucide-react'
import { dashboardService, type NetworkInfo } from '../services/dashboardService'
import { useDashboardWebSocket } from '../../../shared/hooks/useDashboardWebSocket'
import toast from 'react-hot-toast'

export const Dashboard = () => {
  const [networkInfo, setNetworkInfo] = useState<NetworkInfo | null>(null)
  const [useWebSocket, setUseWebSocket] = useState(true)

  // WebSocket connection for real-time updates
  const { data: wsData, isConnected: wsConnected, error: wsError } = useDashboardWebSocket('ibnchannel')

  // Determine if we should use polling:
  // - If WebSocket is disabled, always poll
  // - If WebSocket is enabled but not connected or has error, poll as fallback
  // - If WebSocket is connected but no data yet, poll until data arrives
  const shouldUsePolling = !useWebSocket || !wsConnected || wsError || (wsConnected && !wsData?.metrics && !wsData?.blocks)

  // Always enable polling queries (they will be used as fallback or primary source)
  const { data: metricsPolling, isLoading: metricsLoading, isError: metricsError } = useQuery({
    queryKey: ['dashboard-metrics'],
    queryFn: () => dashboardService.getMetricsSummary('ibnchannel'),
    refetchInterval: shouldUsePolling ? 60000 : false, // Poll every 60s if using polling
    staleTime: 30000,
    enabled: true, // Always enabled, will be used as fallback
    retry: 2, // Retry failed requests
  })

  const { data: blocksPolling, isLoading: blocksLoading, isError: blocksError } = useQuery({
    queryKey: ['dashboard-blocks'],
    queryFn: () => dashboardService.getLatestBlocks('ibnchannel', 10),
    refetchInterval: shouldUsePolling ? 20000 : false, // Poll every 20s if using polling
    staleTime: 10000,
    enabled: true, // Always enabled, will be used as fallback
    retry: 2, // Retry failed requests
  })

  // Use WebSocket data if available and connected, otherwise fallback to polling
  const metrics = (useWebSocket && wsConnected && wsData?.metrics) ? wsData.metrics : metricsPolling
  const blocks = (useWebSocket && wsConnected && wsData?.blocks) ? wsData.blocks : blocksPolling
  
  // Loading state: true if WebSocket connected but no data yet, OR if polling is loading
  // Only show loading if we're actually waiting for data (not if we have data from either source)
  const isLoading = 
    (useWebSocket && wsConnected && !wsError && !wsData?.metrics && !wsData?.blocks && !metricsPolling && !blocksPolling) 
    ? true 
    : (metricsLoading || blocksLoading)

  // Handle WebSocket connection status
  useEffect(() => {
    if (wsError && useWebSocket) {
      console.warn('WebSocket connection failed, falling back to polling:', wsError)
      setUseWebSocket(false)
      toast.error('Kết nối WebSocket thất bại, đang dùng polling')
    } else if (wsConnected && !useWebSocket) {
      // Re-enable WebSocket if connection is restored
      setUseWebSocket(true)
      toast.success('Đã kết nối WebSocket, cập nhật real-time')
    }
  }, [wsError, wsConnected, useWebSocket])

  // Handle errors - only show error if polling explicitly failed (isError = true)
  // Don't show error if data is just undefined (WebSocket might still be loading)
  useEffect(() => {
    // Only show error if polling explicitly failed, not just undefined
    if (metricsError && !isLoading) {
      toast.error('Không thể tải metrics')
    }
  }, [metricsError, isLoading])

  useEffect(() => {
    // Only show error if polling explicitly failed, not just undefined
    if (blocksError && !isLoading) {
      toast.error('Không thể tải blocks')
    }
  }, [blocksError, isLoading])

  // Fetch network info (only once, not real-time)
  useEffect(() => {
    dashboardService
      .getNetworkInfo()
      .then((data) => {
        if (data && typeof data === 'object') {
          setNetworkInfo(data)
        }
      })
      .catch((error) => {
        console.error('Failed to fetch network info:', error)
        setNetworkInfo(null)
      })
  }, [])

  // Update network info from WebSocket if available
  useEffect(() => {
    if (wsData?.networkInfo) {
      const wsNetworkInfo = wsData.networkInfo
      if (wsNetworkInfo && typeof wsNetworkInfo === 'object') {
        const channels = Array.isArray(wsNetworkInfo.channels) ? wsNetworkInfo.channels : []
        const channelNames = channels.map((ch: any) => ch.name || ch).filter(Boolean)
        const allChaincodes: string[] = channels
          .flatMap((ch: any) => Array.isArray(ch.chaincodes) ? ch.chaincodes : [])
          .filter((cc: any): cc is string => cc && typeof cc === 'string')
        const uniqueChaincodes: string[] = Array.from(new Set(allChaincodes))
        const peers = Array.isArray(wsNetworkInfo.peers) ? wsNetworkInfo.peers : []
        const orderers = Array.isArray(wsNetworkInfo.orderers) ? wsNetworkInfo.orderers : []

        setNetworkInfo({
          name: wsNetworkInfo.name || 'IBN Network',
          version: wsNetworkInfo.version || '',
          channels: channelNames,
          chaincodes: uniqueChaincodes,
          peers: peers.length,
          orderers: orderers.length,
        })
      }
    }
  }, [wsData?.networkInfo])


  const formatNumber = (num: number) => {
    if (!num) return '0'
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toString()
  }

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleString('vi-VN', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const totalBlocks = metrics?.blocks?.total || 0
  const validTransactions = metrics?.transactions?.valid || 0
  const successRate = metrics?.transactions?.successRate || 0
  const averageBlockTime = metrics?.blocks?.averageBlockTime || 0
  const recentBlocks = Array.isArray(blocks) ? blocks.slice(0, 6) : []
  const connectionState = wsConnected && useWebSocket

  const heroStats = [
    {
      label: 'Kênh hoạt động',
      value: networkInfo?.channels?.length || 0,
      icon: <Signal className="w-5 h-5 text-emerald-400" />, // Xanh lá - Kênh hoạt động
    },
    {
      label: 'Smart contracts',
      value: networkInfo?.chaincodes?.length || 0,
      icon: <ShieldCheck className="w-5 h-5 text-blue-400" />, // Xanh dương - Bảo mật
    },
    {
      label: 'Peers',
      value: networkInfo?.peers || 0,
      icon: <Activity className="w-5 h-5 text-amber-400" />, // Cam - Hoạt động
    },
    {
      label: 'Orderers',
      value: networkInfo?.orderers || 0,
      icon: <Blocks className="w-5 h-5 text-violet-400" />, // Tím - Blocks/Orderers
    },
  ]

  const performanceCards = [
    {
      title: 'Tổng số blocks',
      value: formatNumber(totalBlocks),
      delta: recentBlocks.length ? `+${recentBlocks[0]?.transactionCount || 0} tx` : '+0 tx',
      icon: <Blocks className="w-5 h-5 text-blue-400" />, // Xanh dương - Blocks
      gradient: 'from-blue-500/20 via-blue-500/5 to-transparent',
    },
    {
      title: 'Giao dịch hợp lệ',
      value: formatNumber(validTransactions),
      delta: `${(metrics?.transactions?.last24Hours || 0).toFixed(0)} trong 24h`,
      icon: <CheckCircle className="w-5 h-5 text-emerald-400" />, // Xanh lá - Thành công
      gradient: 'from-emerald-500/20 via-emerald-500/5 to-transparent',
    },
    {
      title: 'Tỉ lệ thành công',
      value: `${successRate.toFixed(1)}%`,
      delta: successRate > 90 ? 'Ổn định' : 'Cần chú ý',
      icon: <TrendingUp className="w-5 h-5 text-amber-400" />, // Cam - Tăng trưởng
      gradient: 'from-amber-500/20 via-amber-500/5 to-transparent',
    },
    {
      title: 'Thời gian block',
      value: `${averageBlockTime.toFixed(1)}s`,
      delta: 'Trung bình',
      icon: <Clock className="w-5 h-5 text-violet-400" />, // Tím - Thời gian
      gradient: 'from-violet-500/20 via-violet-500/5 to-transparent',
    },
  ]

  const transactionInsights = [
    {
      label: 'Thành công',
      value: formatNumber(metrics?.transactions?.valid || 0),
      icon: <CheckCircle className="w-4 h-4 text-emerald-400" />, // Xanh lá - Thành công
      badge: '+100%',
      badgeColor: 'text-emerald-300 bg-emerald-500/20',
    },
    {
      label: 'Thất bại',
      value: formatNumber(metrics?.transactions?.invalid || 0),
      icon: <XCircle className="w-4 h-4 text-red-400" />, // Đỏ - Thất bại
      badge: '+0%',
      badgeColor: 'text-red-300 bg-red-500/20',
    },
    {
      label: '24 giờ qua',
      value: formatNumber(metrics?.transactions?.last24Hours || 0),
      icon: <Activity className="w-4 h-4 text-amber-400" />, // Cam - Hoạt động
      badge: `${formatNumber(metrics?.transactions?.last7Days || 0)} / 7 ngày`,
      badgeColor: 'text-amber-300 bg-amber-500/20',
    },
  ]

  const handleRealtimeToggle = () => {
    if (!wsConnected) {
      toast.error('WebSocket chưa sẵn sàng, vẫn dùng polling')
      return
    }
    setUseWebSocket((prev) => {
      const next = !prev
      if (next) {
        toast.success('Đã bật realtime')
      } else {
        toast('Đang chuyển về chế độ polling')
      }
      return next
    })
  }

  return (
    <div className="min-h-screen text-white">
      <div className="relative isolate overflow-hidden border-b border-white/10">
        {/* Overlay gradient để tạo depth */}
        <div className="absolute inset-0 bg-gradient-to-b from-black/40 via-transparent to-black/60" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 relative z-10">
          <div className="flex flex-col gap-8 lg:flex-row lg:items-center lg:justify-between">
            <div className="space-y-4">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-white/20 bg-white/5 text-xs uppercase tracking-widest text-gray-200">
                <Sparkles className="w-3.5 h-3.5 text-amber-400" /> {/* Cam - Sparkles */}
                Real-time Monitoring
              </div>
              <div>
                <h1 className="text-4xl font-semibold tracking-tight text-white">IBN Dashboard</h1>
                <p className="mt-2 text-gray-300">
                  Theo dõi trạng thái blockchain với dữ liệu realtime, block timeline và thống kê giao dịch chuyên sâu với phong cách
                </p>
              </div>
              <div className="flex flex-wrap items-center gap-3 text-sm text-gray-200">
                <div className="flex items-center gap-2 px-3 py-1 rounded-full border border-white/15 bg-white/5 shadow-[0_5px_25px_rgba(255,255,255,0.05)] backdrop-blur">
                  {connectionState ? (
                    <>
                      <Wifi className="w-4 h-4 text-emerald-400" />
                      <span>Realtime mode</span>
                    </>
                  ) : (
                    <>
                      <WifiOff className="w-4 h-4 text-amber-400" />
                      <span>Polling fallback</span>
                    </>
                  )}
                </div>
                <button
                  onClick={handleRealtimeToggle}
                  className="px-4 py-1.5 text-sm rounded-full border border-white/20 bg-white/10 text-black hover:bg-white/30 transition backdrop-blur"
                  disabled={!wsConnected}
                >
                  {connectionState ? 'Tạm dừng realtime' : 'Kích hoạt realtime'}
                </button>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4 min-w-[260px]">
              {heroStats.map((stat) => (
                <div
                  key={stat.label}
                  className="rounded-2xl border border-white/10 bg-white/10 px-4 py-3 backdrop-blur-2xl shadow-[0_15px_35px_rgba(0,0,0,0.35)] text-black/90"
                >
                  <div className="flex items-center gap-2 text-sm text-gray-700">
                    {stat.icon}
                    {stat.label}
                  </div>
                  <p className="text-2xl font-semibold mt-2 text-black">{stat.value}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 space-y-10">
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          {performanceCards.map((card) => (
            <div
              key={card.title}
              className="group relative overflow-hidden rounded-3xl text-white p-5 transition-all duration-200 ease-out hover:-translate-y-2 hover:scale-[1.02]"
              // ============================================
              // HOVER EFFECTS - ĐIỀU CHỈNH Ở ĐÂY
              // ============================================
              // hover:-translate-y-2: Nâng card lên khi hover
              //   - -translate-y-1: Nâng nhẹ (4px)
              //   - -translate-y-2: Nâng vừa (8px) ✅ hiện tại
              //   - -translate-y-3: Nâng cao (12px)
              //   - -translate-y-4: Nâng rất cao (16px)
              // hover:scale-[1.02]: Phóng to card khi hover
              //   - scale-[1.01]: Phóng nhẹ (1%)
              //   - scale-[1.02]: Phóng vừa (2%) ✅ hiện tại
              //   - scale-[1.03]: Phóng mạnh (3%)
              //   - scale-[1.05]: Phóng rất mạnh (5%)
              // duration-700: Tốc độ animation (500-1000ms)
              // ease-out: Kiểu easing (ease-out, ease-in-out, ease-in)
              style={{
                // ============================================
                // ĐIỀU CHỈNH THÔNG SỐ LIQUID GLASS Ở ĐÂY
                // ============================================
                
                // 1. ĐỘ TRONG SUỐT (Background Opacity)
                //    - 0.01-0.03: Rất trong suốt (khuyến nghị)
                //    - 0.03-0.05: Trong suốt vừa phải
                //    - 0.05-0.08: Ít trong suốt hơn
                //    - 0.08-0.12: Gần đục
                background: 'rgba(255, 255, 255, 0.01)', // Thay đổi số 0.03
                
                // 2. ĐỘ MỜ (Blur)
                //    - 10-15px: Mờ nhẹ
                //    - 15-20px: Mờ vừa (khuyến nghị)
                //    - 20-25px: Mờ mạnh
                //    - 25-30px: Rất mờ
                backdropFilter: 'blur(20px) saturate(180%)', // Thay đổi số 20
                WebkitBackdropFilter: 'blur(20px) saturate(180%)', // Thay đổi số 20
                
                // 3. ĐỘ BÃO HÒA MÀU (Saturate)
                //    - 100%: Không tăng màu
                //    - 150-180%: Tăng màu vừa (khuyến nghị)
                //    - 180-200%: Tăng màu mạnh
                //    - 200%+: Rất bão hòa
                //    (Số 180 trong saturate(180%) - thay đổi ở đây)
                
                // 4. ĐỘ DÀY VIỀN (Border Opacity)
                //    - 0.10-0.15: Viền mỏng, mờ
                //    - 0.15-0.20: Viền vừa (khuyến nghị)
                //    - 0.20-0.30: Viền rõ hơn
                //    - 0.30-0.40: Viền đậm
                border: '1px solid rgba(255, 255, 255, 0.08)', // Thay đổi số 0.18
                
                // 5. BÓNG ĐỔ (Box Shadow)
                //    - Shadow ngoài: 0 8px 32px rgba(0,0,0,0.37)
                //      * 8px: Độ mờ blur (tăng = mờ hơn)
                //      * 32px: Khoảng cách lan tỏa (tăng = rộng hơn)
                //      * 0.37: Độ đậm (0.2-0.4 = nhẹ, 0.4-0.6 = đậm)
                //    - Inset highlight trên: inset 0 1px 0 rgba(255,255,255,0.1)
                //      * 0.1: Độ sáng highlight (0.05-0.15 = nhẹ, 0.15-0.25 = sáng)
                //    - Inset shadow dưới: inset 0 -1px 0 rgba(255,255,255,0.05)
                //      * 0.05: Độ tối shadow (0.02-0.08 = nhẹ, 0.08-0.15 = đậm)
                boxShadow: `
                  0 8px 32px 0 rgba(0, 0, 0, 0.37),
                  inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
                  inset 0 -1px 0 0 rgba(255, 255, 255, 0.05)
                `,
              }}
            >
              {/* ============================================
                  SHIMMER EFFECT - ĐIỀU CHỈNH Ở ĐÂY
                  ============================================ */}
              {/* 
                Shimmer: Hiệu ứng ánh sáng trượt qua khi hover
                - opacity-0 → opacity-100: Độ hiện khi hover
                - duration-700: Tốc độ fade in (500-1000ms)
                - via-white/10: Độ sáng shimmer (white/5 = nhẹ, white/15 = sáng)
              */}
              <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-700 overflow-hidden rounded-3xl">
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/10 to-transparent animate-shimmer" />
                {/* Thay đổi white/10 thành white/5 (nhẹ hơn) hoặc white/15 (sáng hơn) */}
              </div>
              
              {/* ============================================
                  GRADIENT OVERLAY - ĐIỀU CHỈNH Ở ĐÂY
                  ============================================ */}
              {/* 
                Gradient overlay: Màu gradient khi hover
                - opacity-0 → opacity-30: Độ hiện khi hover
                - opacity-30: Có thể thay đổi (20 = nhẹ, 40 = đậm)
                - duration-700: Tốc độ fade (500-1000ms)
              */}
              <div className={`absolute inset-0 bg-gradient-to-br ${card.gradient} opacity-0 group-hover:opacity-30 transition-opacity duration-700`} />
              {/* Thay đổi opacity-30 thành opacity-20 (nhẹ hơn) hoặc opacity-40 (đậm hơn) */}
              
              {/* ============================================
                  GLASS REFLECTION - ĐIỀU CHỈNH Ở ĐÂY
                  ============================================ */}
              {/* 
                Glass reflection: Phản chiếu ánh sáng ở nửa trên
                - h-1/2: Chiều cao (1/2 = nửa trên, 1/3 = 1/3 trên)
                - from-white/5: Độ sáng phản chiếu (white/3 = sáng, white/8 = mờ)
                - opacity-0 → opacity-100: Độ hiện khi hover
                - duration-700: Tốc độ fade (500-1000ms)
              */}
              <div className="absolute top-0 left-0 right-0 h-1/2 bg-gradient-to-b from-white/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700 rounded-t-3xl" />
              {/* Thay đổi from-white/5 thành from-white/3 (sáng hơn) hoặc from-white/8 (mờ hơn) */}
              
              <div className="relative z-10">
                <div className="flex items-center justify-between text-sm text-gray-200/90">
                  <span className="font-medium tracking-wide">{card.title}</span>
                  {/* ============================================
                      ICON CONTAINER - ĐIỀU CHỈNH Ở ĐÂY
                      ============================================ */}
                  {/* 
                    Icon container: Container cho icon
                    - background: rgba(255,255,255,0.08) - Độ trong suốt (0.05-0.12)
                    - border: rgba(255,255,255,0.15) - Độ đậm viền (0.10-0.25)
                    - backdropFilter: blur(10px) - Độ mờ (5-15px)
                    - group-hover:scale-110 - Độ phóng to khi hover (1.05-1.15)
                    - duration-500: Tốc độ animation (300-700ms)
                  */}
                  <span 
                    className="p-2.5 rounded-full transition-all duration-500 group-hover:scale-110"
                    style={{
                      background: 'rgba(255, 255, 255, 0.08)', // Thay đổi số 0.08
                      border: '1px solid rgba(255, 255, 255, 0.15)', // Thay đổi số 0.15
                      backdropFilter: 'blur(10px)', // Thay đổi số 10
                    }}
                  >
                    {card.icon}
                  </span>
                </div>
                <p className="text-3xl font-bold mt-4 text-white drop-shadow-sm">{card.value}</p>
                <div className="mt-3 inline-flex items-center gap-1.5 text-xs text-slate-200/80">
                  <ArrowUpRight className="w-4 h-4 text-emerald-300/90" />
                  <span className="font-medium">{card.delta}</span>
                </div>
              </div>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
          {/* ============================================
              CARD "TÌNH TRẠNG MẠNG" - ĐIỀU CHỈNH Ở ĐÂY
              ============================================ */}
          <div
            className="xl:col-span-1 p-6 space-y-5 text-white rounded-2xl transition-all duration-700 ease-out hover:-translate-y-1 hover:scale-[1.01]"
            style={{
              // Liquid Glass Effect - Cùng thông số như performance cards
              background: 'rgba(255, 255, 255, 0.01)',
              backdropFilter: 'blur(20px) saturate(180%)',
              WebkitBackdropFilter: 'blur(20px) saturate(180%)',
              border: '1px solid rgba(255, 255, 255, 0.08)',
              boxShadow: `
                0 8px 32px 0 rgba(0, 0, 0, 0.37),
                inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
                inset 0 -1px 0 0 rgba(255, 255, 255, 0.05)
              `,
            }}
          >
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-xl font-semibold text-white">Tình trạng mạng</h2>
                <p className="text-sm text-gray-300">Thông tin tổng quan toàn hệ thống</p>
              </div>
              <ShieldCheck className="w-6 h-6 text-blue-400" /> {/* Xanh dương - Bảo mật */}
            </div>
            <div className="space-y-4">
              {heroStats.map((stat) => (
                <div key={stat.label} className="flex items-center justify-between rounded-2xl bg-black/25 border border-white/10 px-4 py-3">
                  <div className="flex items-center gap-3 text-gray-200">
                    {stat.icon}
                    <span>{stat.label}</span>
                  </div>
                  <span className="text-lg font-semibold text-white">{stat.value}</span>
                </div>
              ))}
              <div className="rounded-2xl bg-gradient-to-r from-white/15 to-white/5 border border-white/10 px-4 py-3 shadow-[0_10px_35px_rgba(0,0,0,0.35)]">
                <p className="text-sm text-gray-200">Kết nối</p>
                <p className="text-lg font-semibold mt-1 text-white">
                  {connectionState ? 'Realtime & đồng bộ' : 'Polling dự phòng'}
                </p>
                <p className="text-xs text-gray-300 mt-1">
                  {connectionState ? 'Cập nhật mỗi 2s qua WebSocket' : 'Cập nhật mỗi 20-60s qua REST API'}
                </p>
              </div>
            </div>
          </div>

          <div className="xl:col-span-2 grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* ============================================
                CARD "DÒNG THỜI GIAN BLOCKS" - ĐIỀU CHỈNH Ở ĐÂY
                ============================================ */}
            <div
              className="p-6 flex flex-col text-white rounded-2xl transition-all duration-700 ease-out hover:-translate-y-1 hover:scale-[1.01]"
              style={{
                // Liquid Glass Effect - Cùng thông số như performance cards
                background: 'rgba(255, 255, 255, 0.01)',
                backdropFilter: 'blur(20px) saturate(180%)',
                WebkitBackdropFilter: 'blur(20px) saturate(180%)',
                border: '1px solid rgba(255, 255, 255, 0.08)',
                boxShadow: `
                  0 8px 32px 0 rgba(0, 0, 0, 0.37),
                  inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
                  inset 0 -1px 0 0 rgba(255, 255, 255, 0.05)
                `,
              }}
            >
              <div className="flex items-center justify-between mb-5">
                <div>
                  <h2 className="text-xl font-semibold">Dòng thời gian blocks</h2>
                  <p className="text-sm text-gray-300">Theo dõi block mới nhất</p>
                </div>
                <span className="text-xs uppercase tracking-widest text-gray-400">ibnchannel</span>
              </div>
              {blocksLoading ? (
                <div className="flex-1 flex flex-col items-center justify-center gap-3 text-gray-400">
                  <div className="h-10 w-10 border-b-2 border-white rounded-full animate-spin" />
                  Đang tải dữ liệu blocks...
                </div>
              ) : recentBlocks.length > 0 ? (
                <div className="space-y-4">
                  {recentBlocks.map((block, index) => (
                    <div key={block.number} className="relative pl-6">
                      <span className="absolute left-0 top-4 h-full w-px bg-gradient-to-b from-white/40 to-transparent" />
                      <div className="relative rounded-2xl border border-white/10 bg-black/30 px-4 py-3 hover:border-white/30 transition">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center text-white font-semibold">
                              #{block.number}
                            </div>
                            <div>
                              <p className="text-sm font-semibold">Block #{block.number}</p>
                              <p className="text-xs text-gray-300">{formatTime(block.timestamp)}</p>
                            </div>
                          </div>
                          <div className="text-right">
                            <p className="text-sm font-medium text-white">{block.transactionCount || 0} txs</p>
                            <p className="text-xs text-gray-400 font-mono">{(block.hash || '').substring(0, 10)}...</p>
                          </div>
                        </div>
                        <div className="mt-3 h-1 rounded-full bg-gradient-to-r from-white/80 to-white/30 animate-pulse" style={{ width: `${Math.min((block.transactionCount || 1) * 10, 100)}%` }} />
                      </div>
                      {index === recentBlocks.length - 1 && <span className="absolute left-0 bottom-0 h-4 w-px bg-transparent" />}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="flex-1 flex items-center justify-center text-gray-400 text-sm">
                  Không có block gần đây
                </div>
              )}
            </div>

            {/* ============================================
                CARD "TRANSACTION INSIGHTS" - ĐIỀU CHỈNH Ở ĐÂY
                ============================================ */}
            <div
              className="p-6 flex flex-col text-white rounded-2xl transition-all duration-700 ease-out hover:-translate-y-1 hover:scale-[1.01]"
              style={{
                // Liquid Glass Effect - Cùng thông số như performance cards
                background: 'rgba(255, 255, 255, 0.01)',
                backdropFilter: 'blur(20px) saturate(180%)',
                WebkitBackdropFilter: 'blur(20px) saturate(180%)',
                border: '1px solid rgba(255, 255, 255, 0.08)',
                boxShadow: `
                  0 8px 32px 0 rgba(0, 0, 0, 0.37),
                  inset 0 1px 0 0 rgba(255, 255, 255, 0.1),
                  inset 0 -1px 0 0 rgba(255, 255, 255, 0.05)
                `,
              }}
            >
              <div className="flex items-center justify-between mb-5">
                <div>
                  <h2 className="text-xl font-semibold">Transaction Insights</h2>
                  <p className="text-sm text-gray-300">Chi tiết hiệu suất giao dịch</p>
                </div>
                <TrendingUp className="w-5 h-5 text-amber-400" /> {/* Cam - Tăng trưởng */}
              </div>
              {metricsLoading ? (
                <div className="flex-1 flex flex-col gap-3 justify-center items-center text-gray-400">
                  <div className="h-10 w-10 border-b-2 border-white rounded-full animate-spin" />
                  Đang tải metrics...
                </div>
              ) : metrics?.transactions ? (
                <div className="space-y-4">
                  {transactionInsights.map((item) => (
                    <div key={item.label} className="rounded-2xl border border-white/10 bg-black/30 px-4 py-3 flex items-center justify-between hover:border-white/30 transition">
                      <div className="flex items-center gap-3">
                        {item.icon}
                        <div>
                          <p className="text-sm text-gray-300">{item.label}</p>
                          <p className="text-lg font-semibold">{item.value}</p>
                        </div>
                      </div>
                      <span className={`text-xs px-3 py-1 rounded-full ${item.badgeColor}`}>{item.badge}</span>
                    </div>
                  ))}

                  <div className="mt-4 rounded-2xl border border-white/10 p-4 bg-gradient-to-r from-white/10 to-white/5">
                    <div className="flex items-center justify-between text-sm text-gray-300">
                      <span>Thời gian xử lý trung bình</span>
                      <span className="text-lg font-semibold">
                        {(metrics.transactions.averageDuration || 0).toFixed(2)} ms
                      </span>
                    </div>
                    <div className="mt-3 flex gap-2">
                      {[...Array(10)].map((_, idx) => (
                        <div
                          key={idx}
                          className="flex-1 rounded-full bg-white/5"
                          style={{ height: `${6 + Math.random() * 20}px` }}
                        />
                      ))}
                    </div>
                  </div>
                </div>
              ) : (
                <div className="flex-1 flex items-center justify-center text-gray-400 text-sm">
                  Không có dữ liệu giao dịch
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

