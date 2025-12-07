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

