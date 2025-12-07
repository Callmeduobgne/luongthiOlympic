import { useState, useEffect, useRef } from 'react'
import { Card } from '@shared/components/ui/Card'
import { Button } from '@shared/components/ui/Button'
import { Input } from '@shared/components/ui/Input'
import { Badge } from '@shared/components/ui/Badge'
import { Select } from '@shared/components/ui/Select'
// ScrollArea not needed, using native div with overflow
import { Download, Pause, Play, Search, X } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'

const normalizeContainerName = (name: string) =>
  name?.toLowerCase().trim().replace(/^\/+/, '') || ''

interface LogEntry {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  container: string
  message: string
  raw: string
}

interface NetworkLogViewerProps {
  selectedPeer?: string
  selectedOrderer?: string
  autoRefresh?: boolean
  refreshInterval?: number
}

export const NetworkLogViewer = ({
  selectedPeer,
  selectedOrderer,
  autoRefresh = true,
  refreshInterval = 2000,
}: NetworkLogViewerProps) => {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [isPaused, setIsPaused] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [serviceFilter, setServiceFilter] = useState<string>('all')
  const [containerFilter, setContainerFilter] = useState<string>('all')
  const [maxLines] = useState(500)
  const scrollRef = useRef<HTMLDivElement>(null)
  const shouldAutoScroll = useRef(true)

  // Map service names to container name patterns
  // Note: Container names in logs might be different from docker container names
  // Loki/Promtail might use different naming conventions
  const serviceContainerMap: Record<string, string[]> = {
    frontend: ['ibn-frontend', 'ibn-frontend-dev', 'frontend', 'frontend-dev'],
    backend: ['ibn-backend', 'backend', 'ibn-backend-1'], // Add variations
    gateway: ['api-gateway-nginx', 'api-gateway-1', 'api-gateway-2', 'api-gateway-3', 'gateway', 'api-gateway'],
    core: ['peer0.org1.ibn.vn', 'peer1.org1.ibn.vn', 'peer2.org1.ibn.vn', 'orderer.ibn.vn', 'orderer1.ibn.vn', 'orderer2.ibn.vn', 'couchdb0', 'couchdb1', 'couchdb2', 'peer', 'orderer', 'couchdb'], // Core Fabric components
  }

  // Query logs from backend (which queries Loki)
  const { data, refetch, isError, error } = useQuery({
    queryKey: ['network-logs', selectedPeer, selectedOrderer, serviceFilter, containerFilter, maxLines],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (selectedPeer) params.append('container', selectedPeer)
      if (selectedOrderer) params.append('container', selectedOrderer)
      // Filter by service (frontend, backend, gateway, core)
      if (serviceFilter !== 'all' && serviceContainerMap[serviceFilter]) {
        // For service filter, we'll filter on client side since Loki query might need multiple container filters
        // We'll pass service name and filter in the response
      }
      if (containerFilter !== 'all') params.append('container', containerFilter)
      params.append('limit', maxLines.toString())
      params.append('since', '1h') // Last 1 hour

      try {
        const response = await api.get(`${API_ENDPOINTS.NETWORK.LOGS}?${params.toString()}`)
        const logs = (response.data?.data || []) as LogEntry[]
        
        // Log for debugging
        if (import.meta.env.DEV && logs.length > 0) {
          console.log(`📋 [Network Logs] Fetched ${logs.length} logs`)
          // Log unique container names for debugging
          const uniqueContainers = Array.from(new Set(logs.map(log => log.container)))
          console.log(`📋 [Network Logs] Unique containers:`, uniqueContainers)
        }
        
        return logs
      } catch (err: any) {
        // Log error for debugging
        if (import.meta.env.DEV) {
          console.error('❌ [Network Logs] Failed to fetch logs:', err)
          console.error('   Endpoint:', `${API_ENDPOINTS.NETWORK.LOGS}`)
          console.error('   Error:', err.response?.data || err.message)
        }
        // Return empty array to avoid breaking UI
        return [] as LogEntry[]
      }
    },
    refetchInterval: autoRefresh && !isPaused ? refreshInterval : false,
    enabled: true,
    retry: 2, // Retry failed requests
    retryDelay: 2000, // Wait 2s between retries
  })

  // Update logs when data changes
  useEffect(() => {
    if (data && !isPaused) {
      // Replace with new data (Loki returns latest logs)
      // Keep only last maxLines
      const newLogs = data.length > maxLines ? data.slice(-maxLines) : data
      setLogs(newLogs)

      // Auto-scroll to bottom
      if (shouldAutoScroll.current && scrollRef.current) {
        setTimeout(() => {
          if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight
          }
        }, 100)
      }
    }
  }, [data, isPaused, maxLines])

  // Filter logs by search query and service filter
  const filteredLogs = logs.filter((log) => {
    // Filter by service (frontend, backend, gateway, core)
    if (serviceFilter !== 'all' && serviceContainerMap[serviceFilter]) {
      const serviceContainers = serviceContainerMap[serviceFilter]
      const logContainerLower = normalizeContainerName(log.container)
      
      // Try multiple matching strategies
      const matchesService = serviceContainers.some(pattern => {
        const patternLower = normalizeContainerName(pattern)
        
        // Strategy 1: Exact match
        if (logContainerLower === patternLower) return true
        
        // Strategy 2: Container name contains pattern
        if (logContainerLower.includes(patternLower)) return true
        
        // Strategy 3: Pattern contains container name (for partial matches)
        if (patternLower.includes(logContainerLower)) return true
        
        // Strategy 4: Check if container name starts with pattern
        if (logContainerLower.startsWith(patternLower)) return true
        
        return false
      })
      
      if (!matchesService) {
        // Debug: Log why filter is excluding this log (only first few to avoid spam)
        if (import.meta.env.DEV && Math.random() < 0.01) { // Log 1% of filtered logs
          console.log(`[NetworkLogViewer] Filtered out log: container="${log.container}", serviceFilter="${serviceFilter}", patterns=`, serviceContainers)
        }
        return false
      }
    }

    // Filter by search query
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      return (
        log.message.toLowerCase().includes(query) ||
        log.container.toLowerCase().includes(query) ||
        log.raw.toLowerCase().includes(query)
      )
    }
    return true
  })

  const handleClear = () => {
    setLogs([])
  }

  const handleExport = () => {
    const content = filteredLogs.map((log) => `${log.timestamp} [${log.level}] ${log.container}: ${log.message}`).join('\n')
    const blob = new Blob([content], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `network-logs-${new Date().toISOString()}.txt`
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleScroll = () => {
    if (scrollRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = scrollRef.current
      // Auto-scroll if user is near bottom (within 100px)
      shouldAutoScroll.current = scrollHeight - scrollTop - clientHeight < 100
    }
  }

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error':
        return 'text-red-400'
      case 'warn':
        return 'text-yellow-400'
      case 'debug':
        return 'text-blue-400'
      default:
        return 'text-gray-300'
    }
  }

  const uniqueContainers = Array.from(new Set(logs.map((log) => log.container)))

  return (
    <Card className="p-6 text-white">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-xl font-semibold">Network Logs</h2>
          <p className="text-sm text-gray-400">Real-time logs từ peers và orderers</p>
        </div>
        <div className="flex items-center gap-2">
          <Badge variant={isPaused ? 'warning' : 'success'}>
            {isPaused ? 'Paused' : 'Live'}
          </Badge>
          <Badge variant="default">{filteredLogs.length} logs</Badge>
        </div>
      </div>

      {/* Controls */}
      <div className="flex flex-wrap gap-2 mb-4">
        <div className="flex-1 min-w-[200px] relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            placeholder="Tìm kiếm logs..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="bg-white/5 border-white/10 pl-10"
          />
        </div>

        <Select
          value={serviceFilter}
          onChange={setServiceFilter}
          options={[
            { value: 'all', label: 'All Services' },
            { value: 'frontend', label: 'Frontend' },
            { value: 'backend', label: 'Backend' },
            { value: 'gateway', label: 'Gateway' },
            { value: 'core', label: 'Core (Peer/Orderer)' },
          ]}
          placeholder="All Services"
        />

        <Select
          value={containerFilter}
          onChange={setContainerFilter}
          options={[
            { value: 'all', label: 'All Containers' },
            ...uniqueContainers.map((container) => ({
              value: container,
              label: container,
            })),
          ]}
          placeholder="All Containers"
        />

        <Button
          variant="secondary"
          size="sm"
          onClick={() => setIsPaused(!isPaused)}
          className="flex items-center gap-2"
        >
          {isPaused ? <Play className="w-4 h-4" /> : <Pause className="w-4 h-4" />}
          {isPaused ? 'Resume' : 'Pause'}
        </Button>

        <Button 
          variant="secondary" 
          size="sm" 
          onClick={handleClear}
          className="flex items-center gap-2"
        >
          <X className="w-4 h-4" />
          Clear
        </Button>

        <Button 
          variant="secondary" 
          size="sm" 
          onClick={handleExport}
          className="flex items-center gap-2"
        >
          <Download className="w-4 h-4" />
          Export
        </Button>
      </div>

      {/* Logs Display */}
      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="h-[500px] overflow-y-auto bg-black/20 rounded-lg p-4 font-mono text-xs border border-white/10"
      >
        {isError ? (
          <div className="text-center text-yellow-400 py-8">
            <p className="font-semibold mb-2">⚠️ Không thể kết nối đến Loki</p>
            <p className="text-sm text-gray-400">
              {error instanceof Error ? error.message : 'Loki service có thể chưa sẵn sàng. Vui lòng thử lại sau.'}
            </p>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => refetch()}
              className="mt-4"
            >
              Thử lại
            </Button>
          </div>
        ) : filteredLogs.length === 0 ? (
          <div className="text-center text-gray-400 py-8">
            {logs.length === 0 ? (
              <div>
                <p className="mb-2">Chưa có logs. Đang chờ dữ liệu...</p>
                <p className="text-xs text-gray-500">
                  Đảm bảo Loki và Promtail đang chạy và collect logs từ Docker containers.
                </p>
              </div>
            ) : (
              'Không tìm thấy logs phù hợp với bộ lọc.'
            )}
          </div>
        ) : (
          <div className="space-y-1">
            {filteredLogs.map((log, index) => (
              <div
                key={`${log.timestamp}-${log.container}-${index}`}
                className="flex gap-3 hover:bg-white/5 px-2 py-1 rounded transition-colors"
              >
                <span className="text-gray-500 flex-shrink-0 w-[180px]">{log.timestamp}</span>
                <span className={`flex-shrink-0 w-[60px] ${getLevelColor(log.level)}`}>
                  [{log.level.toUpperCase()}]
                </span>
                <span className="text-blue-400 flex-shrink-0 w-[150px] truncate">{log.container}</span>
                <span className="text-gray-300 flex-1 break-words">{log.message}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer Info */}
      <div className="mt-4 flex items-center justify-between text-xs text-gray-400">
        <div>
          Hiển thị {filteredLogs.length} / {logs.length} logs
          {searchQuery && ` (filtered by "${searchQuery}")`}
        </div>
        <div>
          Auto-refresh: {autoRefresh && !isPaused ? `Mỗi ${refreshInterval / 1000}s` : 'Tắt'}
        </div>
      </div>
    </Card>
  )
}

