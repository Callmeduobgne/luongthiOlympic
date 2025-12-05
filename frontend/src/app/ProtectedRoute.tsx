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

import { Navigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { authService } from '@features/authentication/services/authService'
import { Layout } from '@shared/components/layout/Layout'
import { LoadingState } from '@shared/components/common/LoadingState'

interface ProtectedRouteProps {
  children: React.ReactNode
}

export const ProtectedRoute = ({ children }: ProtectedRouteProps) => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null)

  useEffect(() => {
    // Declare pollInterval at useEffect scope so it's accessible in cleanup
    let pollInterval: NodeJS.Timeout | null = null
    let initialCheckTimeout: NodeJS.Timeout | null = null

    // Check authentication status
    const checkAuth = () => {
      const token = localStorage.getItem('accessToken')
      const authenticated = authService.isAuthenticated()

      // Always log for debugging (not just DEV mode)
      console.log('🔒 [ProtectedRoute] Auth check:', {
        authenticated,
        hasToken: !!token,
        tokenLength: token?.length || 0,
        tokenPreview: token ? `${token.substring(0, 20)}...` : 'null',
        timestamp: new Date().toISOString(),
      })

      setIsAuthenticated(authenticated)
      return authenticated
    }

    // Check on storage events (for cross-tab changes)
    const handleStorageChange = (e: StorageEvent) => {
      // Only react to accessToken changes
      if (e.key === 'accessToken' || e.key === null) {
        const auth = checkAuth()
        // If authenticated, stop polling
        if (auth && pollInterval) {
          clearInterval(pollInterval)
          pollInterval = null
        }
      }
    }

    // Custom event for same-tab localStorage changes
    const handleCustomStorageChange = () => {
      const auth = checkAuth()
      // If authenticated, stop polling
      if (auth && pollInterval) {
        clearInterval(pollInterval)
        pollInterval = null
      }
    }

    // Add event listeners
    window.addEventListener('storage', handleStorageChange)
    window.addEventListener('localStorageChange', handleCustomStorageChange)

    // Add a small delay before initial check to avoid race condition
    // This gives time for token to be set after login
    initialCheckTimeout = setTimeout(() => {
      const initialAuth = checkAuth()

      // Only poll if not authenticated initially (fallback for edge cases)
      // Poll less frequently (2 seconds) and stop once authenticated
      if (!initialAuth) {
        pollInterval = setInterval(() => {
          const auth = checkAuth()
          // Stop polling once authenticated
          if (auth && pollInterval) {
            clearInterval(pollInterval)
            pollInterval = null
          }
        }, 2000)
      }
    }, 200) // 200ms delay to ensure token is set

    return () => {
      if (initialCheckTimeout) {
        clearTimeout(initialCheckTimeout)
      }
      window.removeEventListener('storage', handleStorageChange)
      window.removeEventListener('localStorageChange', handleCustomStorageChange)
      if (pollInterval) {
        clearInterval(pollInterval)
      }
    }
  }, [])

  // Show loading while checking
  if (isAuthenticated === null) {
    return <LoadingState text="Đang kiểm tra..." />
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  return <Layout>{children}</Layout>
}



