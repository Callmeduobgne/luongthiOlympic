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

import { useEffect, useState, useRef } from 'react'
import websocketService from '../../services/websocketService'

export interface DashboardData {
  metrics: any | null
  blocks: any[] | null
  networkInfo: any | null
}

export const useDashboardWebSocket = (channel: string = 'ibnchannel') => {
  const [data, setData] = useState<DashboardData>({
    metrics: null,
    blocks: null,
    networkInfo: null,
  })
  const [isConnected, setIsConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)
  const isMountedRef = useRef(true)
  const connectionCheckIntervalRef = useRef<number | null>(null)

  useEffect(() => {
    isMountedRef.current = true
    const token = localStorage.getItem('accessToken')
    if (!token) {
      setError('No authentication token')
      return
    }

    let mounted = true

    const connect = async () => {
      // Don't connect if already connected to same channel
      if (websocketService.isConnected() && mounted) {
        setIsConnected(true)
        setError(null)
        return
      }

      try {
        await websocketService.connect(channel, token)
        if (mounted) {
          setIsConnected(true)
          setError(null)
        }
      } catch (err) {
        if (mounted) {
          setIsConnected(false)
          setError(err instanceof Error ? err.message : 'Failed to connect')
          // Note: Don't retry here - let WebSocketService handle reconnection
          // This prevents duplicate reconnection logic
        }
      }
    }

    // Check connection status periodically (WebSocketService handles reconnection)
    const checkConnection = () => {
      if (!mounted) return
      const connected = websocketService.isConnected()
      if (connected !== isConnected) {
        setIsConnected(connected)
        if (connected) {
          setError(null)
        }
      }
    }

    // Handle updates
    const handleUpdate = (update: any) => {
      if (!mounted) return

      if (update.type === 'initial') {
        setData({
          metrics: update.metrics || null,
          blocks: update.blocks || null,
          networkInfo: update.networkInfo || null,
        })
      } else if (update.type === 'metrics:update') {
        setData((prev) => ({
          ...prev,
          metrics: update.metrics || prev.metrics,
        }))
      } else if (update.type === 'blocks:update') {
        setData((prev) => ({
          ...prev,
          blocks: update.blocks || prev.blocks,
        }))
      } else if (update.type === 'network:update') {
        setData((prev) => ({
          ...prev,
          networkInfo: update.networkInfo || prev.networkInfo,
        }))
      }
    }

    websocketService.on('dashboard:update', handleUpdate)

    // Connect
    connect()

    // Check connection status every 2 seconds
    connectionCheckIntervalRef.current = window.setInterval(checkConnection, 2000)

    // Cleanup
    return () => {
      mounted = false
      isMountedRef.current = false
      
      // Clear reconnect timeout
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }

      // Clear connection check interval
      if (connectionCheckIntervalRef.current) {
        clearInterval(connectionCheckIntervalRef.current)
        connectionCheckIntervalRef.current = null
      }

      // Remove event listener
      websocketService.off('dashboard:update', handleUpdate)
      
      // Only disconnect if this is the last component using WebSocket
      // In a real app, you might want to track usage count
      // For now, we'll disconnect to prevent leaks
      websocketService.disconnect()
      setIsConnected(false)
    }
  }, [channel])

  return { data, isConnected, error }
}

