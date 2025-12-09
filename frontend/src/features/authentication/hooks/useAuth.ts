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
 * Copyright 2025 IBN Network (ICTU Blockchain Network)
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


