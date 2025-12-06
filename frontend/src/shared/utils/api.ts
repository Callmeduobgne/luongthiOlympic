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
import { authService } from '@features/authentication/services/authService'
import { API_CONFIG, API_ENDPOINTS } from '@shared/config/api.config'

const api = axios.create({
  baseURL: API_CONFIG.BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - Add token
api.interceptors.request.use(
  (config) => {
    const token = authService.getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Dev mode: Log all requests
    if (import.meta.env.DEV) {
      console.log('🌐 [DEV] API Request:', {
        method: config.method?.toUpperCase(),
        url: config.url,
        baseURL: config.baseURL,
        headers: { ...config.headers, Authorization: token ? 'Bearer ***' : 'none' },
      })
    }

    return config
  },
  (error) => {
    if (import.meta.env.DEV) {
      console.error('❌ [DEV] Request Error:', error)
    }
    return Promise.reject(error)
  }
)

// Response interceptor - Handle 401 & refresh token
api.interceptors.response.use(
  (response) => {
    // Dev mode: Log successful responses
    if (import.meta.env.DEV) {
      console.log('✅ [DEV] API Response:', {
        status: response.status,
        url: response.config.url,
        data: response.data,
      })
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // Dev mode: Log error responses
    if (import.meta.env.DEV) {
      console.error('❌ [DEV] API Error Response:', {
        status: error.response?.status,
        statusText: error.response?.statusText,
        url: error.config?.url,
        data: error.response?.data,
        message: error.message,
      })
    }

    // If 401 and not already retried and not a refresh token request
    // ALSO: Do not try to refresh if the failed request was a login request (invalid credentials)
    const isRefreshRequest = originalRequest.url?.includes('/auth/refresh')
    const isLoginRequest = originalRequest.url?.includes(API_ENDPOINTS.AUTH.LOGIN)

    if (error.response?.status === 401 && !originalRequest._retry && !isRefreshRequest && !isLoginRequest) {
      originalRequest._retry = true

      if (import.meta.env.DEV) {
        console.log('🔄 [DEV] Attempting token refresh...')
      }

      try {
        const newToken = await authService.refreshToken()
        if (newToken) {
          originalRequest.headers.Authorization = `Bearer ${newToken}`

          if (import.meta.env.DEV) {
            console.log('✅ [DEV] Token refreshed, retrying request')
          }

          return api(originalRequest)
        } else {
          throw new Error('No token received from refresh')
        }
      } catch (refreshError) {
        // Refresh failed, logout user
        if (import.meta.env.DEV) {
          console.error('❌ [DEV] Token refresh failed, logging out')
        }
        authService.logout()
        return Promise.reject(refreshError)
      }
    }

    // If refresh token request fails, logout immediately
    if (error.response?.status === 401 && isRefreshRequest) {
      if (import.meta.env.DEV) {
        console.error('❌ [DEV] Refresh token request failed, logging out')
      }
      authService.logout()
    }

    return Promise.reject(error)
  }
)

export default api

