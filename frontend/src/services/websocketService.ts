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

// Native WebSocket implementation (no socket.io-client needed)

export interface DashboardUpdate {
  type: 'initial' | 'metrics:update' | 'blocks:update' | 'network:update'
  metrics?: any
  blocks?: any[]
  networkInfo?: any
}

class WebSocketService {
  private socket: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10 // Increased for production
  private reconnectTimeout: number | null = null
  private isConnecting = false
  private isManualDisconnect = false
  private currentChannel: string | null = null
  private currentToken: string | null = null
  private pingInterval: number | null = null
  private reconnectDelay = 1000 // Initial delay: 1s
  private maxReconnectDelay = 30000 // Max delay: 30s

  connect(channel: string, token: string): Promise<WebSocket> {
    return new Promise((resolve, reject) => {
      // Prevent multiple simultaneous connections
      if (this.isConnecting) {
        console.warn('⚠️ WebSocket connection already in progress, skipping...')
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
          resolve(this.socket)
        } else {
          reject(new Error('Connection already in progress'))
        }
        return
      }

      // If already connected to same channel, return existing connection
      if (this.socket && this.socket.readyState === WebSocket.OPEN && this.currentChannel === channel) {
        console.log('✅ WebSocket already connected to', channel)
        resolve(this.socket)
        return
      }

      // Disconnect existing connection if different channel or not open
      if (this.socket) {
        console.log('🔄 Disconnecting existing WebSocket connection...')
        this.disconnect()
      }

      // Cancel any pending reconnect
      if (this.reconnectTimeout) {
        clearTimeout(this.reconnectTimeout)
        this.reconnectTimeout = null
      }

      this.isConnecting = true
      this.isManualDisconnect = false
      this.currentChannel = channel
      this.currentToken = token

      // Determine base URL for WebSocket connection
      // In dev mode: use window.location.origin (Vite dev server) - Vite proxy will forward WebSocket
      // In prod: use window.location.origin (nginx will proxy WebSocket)
      // Never use localhost:8085 as it won't work in Docker
      let baseURL = ''
      if (typeof window !== 'undefined' && window.location?.origin) {
        // Use same origin - Vite proxy (dev) or nginx (prod) will forward WebSocket
        baseURL = window.location.origin
      } else {
        // Fallback: use relative URL (will use current origin)
        baseURL = ''
      }

      // Convert http/https base URL to ws/wss
      // If baseURL is empty, use relative WebSocket URL (browser will use current origin)
      const wsURL = baseURL 
        ? baseURL.replace(/^http/i, 'ws').replace(/\/$/, '')
        : (typeof window !== 'undefined' && window.location?.protocol === 'https:' ? 'wss:' : 'ws:') + '//' + (typeof window !== 'undefined' ? window.location.host : '')

      // Use native WebSocket - pass token as query param
      // Backend endpoint: /api/v1/dashboard/ws/{channel}
      const ws = new WebSocket(`${wsURL}/api/v1/dashboard/ws/${channel}?token=${encodeURIComponent(token)}`)

      ws.onopen = () => {
        console.log('✅ WebSocket connected to dashboard', channel)
        this.reconnectAttempts = 0
        this.reconnectDelay = 1000 // Reset delay on successful connection
        this.socket = ws
        this.isConnecting = false
        
        // Production best practice: Start heartbeat/ping interval
        this.startHeartbeat()
        
        resolve(ws)
      }

      ws.onerror = (error) => {
        console.error('❌ WebSocket error:', error)
        this.isConnecting = false
        reject(error)
      }

      ws.onclose = (event) => {
        console.log('🔌 WebSocket disconnected', { code: event.code, reason: event.reason })
        this.socket = null
        this.isConnecting = false
        this.stopHeartbeat()

        // Only auto-reconnect if not manually disconnected and connection was established
        if (!this.isManualDisconnect && event.code !== 1000) {
          // Production best practice: Exponential backoff for reconnection
          if (this.reconnectAttempts < this.maxReconnectAttempts && this.currentChannel && this.currentToken) {
            this.reconnectAttempts++
            
            // Exponential backoff: delay = min(initialDelay * 2^(attempts-1), maxDelay)
            const delay = Math.min(
              this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
              this.maxReconnectDelay
            )
            
            console.log(`🔄 Reconnecting in ${delay}ms... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
            this.reconnectTimeout = window.setTimeout(() => {
              if (!this.isManualDisconnect && this.currentChannel && this.currentToken) {
                this.connect(this.currentChannel, this.currentToken).catch((err) => {
                  console.error('Reconnection failed:', err)
                })
              }
            }, delay)
          } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.warn('⚠️ Max reconnection attempts reached')
            // Reset after max attempts to allow manual retry
            this.reconnectAttempts = 0
            this.reconnectDelay = 1000
          }
        } else {
          // Reset reconnect attempts on manual disconnect
          this.reconnectAttempts = 0
          this.reconnectDelay = 1000
        }
      }

      // Handle messages
      ws.onmessage = (event) => {
        try {
          const data: DashboardUpdate = JSON.parse(event.data as string)
          // Trigger custom event
          const customEvent = new CustomEvent('dashboard:update', { detail: data })
          window.dispatchEvent(customEvent)
        } catch (err) {
          console.error('Failed to parse WebSocket message', err)
        }
      }
    })
  }

  disconnect() {
    this.isManualDisconnect = true
    
    // Cancel any pending reconnect
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }

    // Stop heartbeat
    this.stopHeartbeat()

    if (this.socket) {
      // Remove event handlers to prevent reconnect
      this.socket.onclose = null
      this.socket.close()
      this.socket = null
    }

    this.currentChannel = null
    this.currentToken = null
    this.reconnectAttempts = 0
    this.reconnectDelay = 1000
    this.isConnecting = false
    console.log('🔌 WebSocket manually disconnected')
  }

  // Production best practice: Heartbeat/ping to keep connection alive
  private startHeartbeat() {
    this.stopHeartbeat() // Clear any existing interval
    
    // Send ping every 30 seconds (server expects pong)
    this.pingInterval = window.setInterval(() => {
      if (this.socket && this.socket.readyState === WebSocket.OPEN) {
        try {
          // Send ping message (server will respond with pong)
          this.socket.send(JSON.stringify({ type: 'ping' }))
        } catch (err) {
          console.error('Failed to send ping:', err)
        }
      }
    }, 30000) // 30 seconds
  }

  private stopHeartbeat() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  isConnected(): boolean {
    return this.socket?.readyState === WebSocket.OPEN
  }

  on(event: string, callback: (data: any) => void) {
    if (event === 'dashboard:update') {
      window.addEventListener('dashboard:update', ((e: CustomEvent) => {
        callback(e.detail)
      }) as EventListener)
    }
  }

  off(event: string, callback: (data: any) => void) {
    if (event === 'dashboard:update') {
      window.removeEventListener('dashboard:update', callback as EventListener)
    }
  }
}

export default new WebSocketService()

