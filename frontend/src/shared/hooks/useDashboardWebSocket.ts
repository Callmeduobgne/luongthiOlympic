/*
 * Copyright (c) 2025 IBN Network
 */

/**
 * Copyright 2024 IBN Network (ICTU Blockchain Network)
 */

import { useEffect, useState } from 'react'
import wsService from '@/services/websocketService'

export interface DashboardWebSocketData {
  type: string
  data?: unknown
  metrics?: unknown
  blocks?: unknown[]
  networkInfo?: unknown
}

export function useDashboardWebSocket(channel: string) {
  const [isConnected, setIsConnected] = useState(false)
  const [data, setData] = useState<DashboardWebSocketData | null>(null)
  const [error, setError] = useState<Error | null>(null)

  useEffect(() => {
    const token = localStorage.getItem('access_token')
    if (!token || !channel) {
      return
    }

    wsService.connect(channel, token)
      .then(() => {
        setIsConnected(true)
        setError(null)
      })
      .catch((err: Error) => {
        setError(err)
        setIsConnected(false)
      })

    const handleUpdate = (update: DashboardWebSocketData) => {
      setData(update)
    }

    wsService.on('dashboard:update', handleUpdate)

    return () => {
      wsService.off('dashboard:update', handleUpdate)
      wsService.disconnect()
      setIsConnected(false)
    }
  }, [channel])

  return {
    isConnected,
    data,
    error,
  }
}

