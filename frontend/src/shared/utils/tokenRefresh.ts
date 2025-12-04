/*
 * Copyright (c) 2025 IBN Network
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 */

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

import axios from 'axios'
import { API_ENDPOINTS } from '../config/api.config'

// API Gateway refresh response format (wrapped)
interface RefreshTokenResponse {
  success: boolean
  data: {
    accessToken: string
    refreshToken: string
    expiresIn: number
  }
}

class TokenRefreshManager {
  private isRefreshing = false
  private refreshPromise: Promise<string> | null = null
  private tokenExpiryTime: string | null = null
  private refreshInterval: number | null = null

  setTokenExpiryTime(expiresAt: string): void {
    this.tokenExpiryTime = expiresAt
  }

  clearTokenExpiryTime(): void {
    this.tokenExpiryTime = null
  }

  start(): void {
    this.stop() // Clear any existing interval
    
    if (!this.tokenExpiryTime) {
      return
    }

    const expiryTime = new Date(this.tokenExpiryTime).getTime()
    const now = Date.now()
    const timeUntilExpiry = expiryTime - now

    // Refresh 5 minutes before expiry
    const refreshTime = Math.max(timeUntilExpiry - 5 * 60 * 1000, 60000) // At least 1 minute

    this.refreshInterval = window.setInterval(() => {
      this.refreshToken().catch((error) => {
        console.error('Auto token refresh failed:', error)
      })
    }, refreshTime)
  }

  stop(): void {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval)
      this.refreshInterval = null
    }
  }

  async refreshToken(): Promise<string> {
    // If already refreshing, return the existing promise
    if (this.isRefreshing && this.refreshPromise) {
      return this.refreshPromise
    }

    this.isRefreshing = true
    this.refreshPromise = this.performRefresh()

    try {
      const newToken = await this.refreshPromise
      return newToken
    } finally {
      this.isRefreshing = false
      this.refreshPromise = null
    }
  }

  private async performRefresh(): Promise<string> {
    // Use unified camelCase key
    const refreshToken = localStorage.getItem('refreshToken')

    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    try {
      // Call API Gateway refresh endpoint
      const response = await axios.post<RefreshTokenResponse>(
        API_ENDPOINTS.AUTH.REFRESH,
        { refreshToken }
      )

      const { accessToken, refreshToken: newRefreshToken } = response.data.data

      // Store new tokens with unified keys
      if (accessToken) {
        localStorage.setItem('accessToken', accessToken)
      }
      if (newRefreshToken) {
        localStorage.setItem('refreshToken', newRefreshToken)
      }

      return accessToken
    } catch (error) {
      // Refresh failed, clear tokens
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
      localStorage.removeItem('tokenExpiresAt')
      throw error
    }
  }
}

export const tokenRefreshManager = new TokenRefreshManager()

