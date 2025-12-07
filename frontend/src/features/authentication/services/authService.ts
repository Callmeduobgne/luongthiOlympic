import api from '@shared/utils/api'
import { API_ENDPOINTS } from '@shared/config/api.config'
import type { LoginRequest, SignupRequest, AuthResponse } from '../types/auth.types'

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
  async signup(data: SignupRequest): Promise<void> {
    const payload = {
      email: data.email,
      password: data.password,
      full_name: data.name,
      role: 'user', // Default role for public signup
    }
    await api.post(API_ENDPOINTS.AUTH.REGISTER, payload)
  },

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

      // Store token expiry time and start auto-refresh
      if (authData.expiresAt) {
        localStorage.setItem('tokenExpiresAt', authData.expiresAt)

        // Start token refresh manager
        const { tokenRefreshManager } = await import('@shared/utils/tokenRefresh')
        tokenRefreshManager.setTokenExpiryTime(authData.expiresAt)
        tokenRefreshManager.start()

        if (import.meta.env.DEV) {
          console.log('✅ [DEV] Token expiry time stored and auto-refresh started')
        }
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

    // Backend expects snake_case: refresh_token
    const response = await api.post<{ success: boolean; data: { accessToken: string; expiresAt?: string } } | { accessToken: string; access_token?: string; expiresAt?: string; expires_at?: string }>(
      API_ENDPOINTS.AUTH.REFRESH,
      { refresh_token: refreshToken }
    )

    // Backend returns snake_case (access_token, expires_at)
    const rawData = isWrappedResponse(response.data)
      ? response.data.data
      : (response.data as { accessToken?: string; access_token?: string; expiresAt?: string; expires_at?: string })

    // Handle both camelCase and snake_case formats
    const accessToken = rawData?.accessToken || (rawData as any)?.access_token
    const expiresAt = rawData?.expiresAt || (rawData as any)?.expires_at

    if (!accessToken) {
      throw new Error('No access token returned from refresh endpoint')
    }

    localStorage.setItem('accessToken', accessToken)

    // Store new expiry time if provided
    if (expiresAt) {
      localStorage.setItem('tokenExpiresAt', expiresAt)

      // Update token refresh manager
      import('@shared/utils/tokenRefresh').then(({ tokenRefreshManager }) => {
        tokenRefreshManager.setTokenExpiryTime(expiresAt)
      })
    }

    return accessToken
  },

  logout() {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
    localStorage.removeItem('tokenExpiresAt')

    // Stop token refresh manager
    import('@shared/utils/tokenRefresh').then(({ tokenRefreshManager }) => {
      tokenRefreshManager.stop()
      tokenRefreshManager.clearTokenExpiryTime()
    })

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


