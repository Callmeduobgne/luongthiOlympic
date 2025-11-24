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

import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import type { LoginRequest, AuthResponse } from '../types/auth.types'

type WrappedAuthResponse = {
  success: boolean
  data: AuthResponse
}

const isWrappedResponse = (payload: unknown): payload is WrappedAuthResponse => {
  return (
    typeof payload === 'object' &&
    payload !== null &&
    'success' in payload &&
    (payload as WrappedAuthResponse).success !== undefined &&
    'data' in payload
  )
}

export const authService = {
  async login(credentials: LoginRequest): Promise<AuthResponse> {
    // Backend expects 'email' field (not username)
    const requestBody = {
      email: credentials.email,
      password: credentials.password,
    }
    
    // Dev mode: Log request details
    if (import.meta.env.DEV) {
      console.log('📤 [DEV] API Request:', {
        url: API_ENDPOINTS.AUTH.LOGIN,
        method: 'POST',
        body: { ...requestBody, password: '***' },
        baseURL: import.meta.env.VITE_API_BASE_URL || 'relative',
      })
    }
    
    try {
      const response = await api.post<AuthResponse | WrappedAuthResponse>(
        API_ENDPOINTS.AUTH.LOGIN,
        requestBody
      )
      const payload = response.data
      const wrappedPayload = isWrappedResponse(payload) ? payload : null
      const directPayload = wrappedPayload ? null : (payload as AuthResponse)
      
      // Dev mode: Log response
      if (import.meta.env.DEV) {
        console.log('📥 [DEV] API Response:', {
          status: response.status,
          success: wrappedPayload ? wrappedPayload.success : undefined,
          hasData: wrappedPayload ? !!wrappedPayload.data : !!directPayload?.user,
          hasToken: wrappedPayload ? !!wrappedPayload.data?.accessToken : !!directPayload?.accessToken,
          payloadKeys: Object.keys(payload || {}),
          payload: payload, // Log full payload for debugging
        })
      }
      
      // API Gateway wraps response in { success: true, data: {...} }
      // Handle both wrapped and unwrapped responses
      // Backend returns snake_case (access_token) but frontend expects camelCase (accessToken)
      let authData: AuthResponse
      if (wrappedPayload) {
        authData = wrappedPayload.data
      } else if (directPayload) {
        // Handle both camelCase and snake_case formats
        const payload = directPayload as any
        authData = {
          accessToken: payload.accessToken || payload.access_token,
          refreshToken: payload.refreshToken || payload.refresh_token,
          expiresAt: payload.expiresAt || payload.expires_at,
          user: payload.user,
        }
      } else {
        // Debug: Log the actual payload structure
        if (import.meta.env.DEV) {
          console.error('❌ [DEV] Invalid response format. Payload:', payload)
          console.error('❌ [DEV] Payload type:', typeof payload)
          console.error('❌ [DEV] Payload keys:', payload ? Object.keys(payload) : 'null')
        }
        throw new Error('Invalid response format from server')
      }
      
      // Validate that we have required fields
      if (!authData.accessToken) {
        if (import.meta.env.DEV) {
          console.error('❌ [DEV] No accessToken found in response:', authData)
        }
        throw new Error('Invalid response format: missing accessToken')
      }
      
      // Store tokens
      if (authData.accessToken) {
        localStorage.setItem('accessToken', authData.accessToken)
        if (import.meta.env.DEV) {
          console.log('✅ [DEV] Token stored in localStorage')
        }
      } else {
        if (import.meta.env.DEV) {
          console.error('❌ [DEV] No accessToken in response:', authData)
        }
      }
      
      if (authData.refreshToken) {
        localStorage.setItem('refreshToken', authData.refreshToken)
      }

      return authData
    } catch (error) {
      const axiosError = error as {
        message?: string
        response?: { data?: { message?: string }; status?: number }
        config?: { url?: string; method?: string; data?: unknown }
      }
      // Dev mode: Log error details
      if (import.meta.env.DEV) {
        console.error('❌ [DEV] API Error:', {
          message: axiosError.message,
          response: axiosError.response?.data,
          status: axiosError.response?.status,
          request: {
            url: axiosError.config?.url,
            method: axiosError.config?.method,
            data: axiosError.config?.data,
          },
        })
      }
      throw axiosError
    }
  },

  async refreshToken(): Promise<string> {
    const refreshToken = localStorage.getItem('refreshToken')
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    const response = await api.post<{ success: boolean; data: { accessToken: string } } | { accessToken: string }>(
      API_ENDPOINTS.AUTH.REFRESH,
      { refreshToken }
    )

    const accessToken = isWrappedResponse(response.data)
      ? response.data.data?.accessToken
      : (response.data as { accessToken?: string }).accessToken

    if (!accessToken) {
      throw new Error('No access token returned from refresh endpoint')
    }

    if (accessToken) {
      localStorage.setItem('accessToken', accessToken)
    }
    return accessToken
  },

  logout() {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    window.location.href = '/login'
  },

  getAccessToken(): string | null {
    return localStorage.getItem('accessToken')
  },

  getRefreshToken(): string | null {
    return localStorage.getItem('refreshToken')
  },

  isAuthenticated(): boolean {
    return !!this.getAccessToken()
  },
}


