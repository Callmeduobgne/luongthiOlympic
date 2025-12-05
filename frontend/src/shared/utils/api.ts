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

import axios, { type AxiosInstance, type AxiosResponse } from 'axios'
import { tokenRefreshManager } from './tokenRefresh'

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - Add token to requests
api.interceptors.request.use(
  (config) => {
    // Use unified camelCase key for access token
    const token = localStorage.getItem('accessToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
      // Debug log
      if (import.meta.env.DEV) {
        console.log('🔐 [API Interceptor] Token attached to request:', {
          url: config.url,
          hasToken: !!token,
          tokenPreview: token ? `${token.substring(0, 20)}...` : 'null',
        })
      }
    } else {
      // Debug log when no token
      if (import.meta.env.DEV) {
        console.warn('⚠️ [API Interceptor] No token found for request:', config.url)
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor - Handle token refresh on 401
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // If error is 401 and we haven't retried yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      // Check if we have a refresh token before attempting refresh
      const refreshToken = localStorage.getItem('refreshToken')
      const accessToken = localStorage.getItem('accessToken')
      const tokenExpiresAt = localStorage.getItem('tokenExpiresAt')

      if (import.meta.env.DEV) {
        console.log('🔐 [API] 401 error detected:', {
          url: originalRequest.url,
          hasRefreshToken: !!refreshToken,
          hasAccessToken: !!accessToken,
          willAttemptRefresh: !!refreshToken,
        })
      }

      // Check if token was just created (within last 5 seconds)
      // This prevents refresh attempts immediately after login
      if (tokenExpiresAt && accessToken) {
        const expiryTime = new Date(tokenExpiresAt).getTime()
        const now = Date.now()
        const tokenAge = expiryTime - now
        const maxAge = 24 * 60 * 60 * 1000 // 24 hours (typical token lifetime)
        const tokenJustCreated = tokenAge > (maxAge - 5000) // Created within last 5 seconds

        if (tokenJustCreated) {
          if (import.meta.env.DEV) {
            console.warn('⚠️ [API] Token was just created, skipping refresh. This is likely a race condition.')
          }
          // Don't delete tokens, just reject the error
          // The next request will succeed once the delay passes
          return Promise.reject(error)
        }
      }

      // Only try to refresh if we have a refresh token
      if (refreshToken) {
        try {
          // Try to refresh token
          await tokenRefreshManager.refreshToken()

          // Retry original request with new token
          const newToken = localStorage.getItem('accessToken')
          if (newToken) {
            originalRequest.headers.Authorization = `Bearer ${newToken}`
          }

          if (import.meta.env.DEV) {
            console.log('✅ [API] Token refreshed, retrying request')
          }

          return api(originalRequest)
        } catch (refreshError) {
          // Refresh failed, only clear tokens if refresh actually failed
          if (import.meta.env.DEV) {
            console.error('❌ [API] Token refresh failed:', refreshError)
          }
          localStorage.removeItem('accessToken')
          localStorage.removeItem('refreshToken')
          localStorage.removeItem('tokenExpiresAt')
          window.location.href = '/login'
          return Promise.reject(refreshError)
        }
      } else {
        // No refresh token, just reject (don't clear tokens - might be a temporary network issue)
        if (import.meta.env.DEV) {
          console.warn('⚠️ [API] 401 error but no refresh token available')
        }
        return Promise.reject(error)
      }
    }

    return Promise.reject(error)
  }
)

export default api

