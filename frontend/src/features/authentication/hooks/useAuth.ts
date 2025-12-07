

import { useState, useEffect } from 'react'
import { authService } from '../services/authService'
import type { User, LoginRequest } from '../types/auth.types'

interface UseAuthReturn {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (credentials: LoginRequest) => Promise<any>
  logout: () => void
}

export const useAuth = (): UseAuthReturn => {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    // Check if user is authenticated on mount
    const checkAuth = async () => {
      const token = authService.getAccessToken()
      if (token) {
        // Start token refresh manager if token exists
        const tokenExpiresAt = localStorage.getItem('tokenExpiresAt')
        if (tokenExpiresAt) {
          const { tokenRefreshManager } = await import('@shared/utils/tokenRefresh')
          tokenRefreshManager.setTokenExpiryTime(tokenExpiresAt)
          tokenRefreshManager.start()
        }

        // TODO: Fetch user profile from API
        // For now, just set authenticated state
        setIsLoading(false)
      } else {
        setIsLoading(false)
      }
    }

    checkAuth()
  }, [])

  const login = async (credentials: LoginRequest) => {
    setIsLoading(true)
    try {
      const response = await authService.login(credentials)
      setUser(response.user)
      setIsLoading(false)
      return response
    } catch (error) {
      setIsLoading(false)
      throw error
    }
  }

  const logout = () => {
    setUser(null)
    authService.logout()
  }

  return {
    user,
    isAuthenticated: !!user || authService.isAuthenticated(),
    isLoading,
    login,
    logout,
  }
}


