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
    // Check authentication status immediately
    const checkAuth = () => {
      const authenticated = authService.isAuthenticated()
      setIsAuthenticated(authenticated)
    }

    checkAuth()
    
    // Also check on storage events (in case token is set in another tab)
    const handleStorageChange = () => {
      checkAuth()
    }
    
    window.addEventListener('storage', handleStorageChange)
    
    return () => {
      window.removeEventListener('storage', handleStorageChange)
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



