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

/**
 * Token Refresh Utility
 * 
 * Manages automatic token refresh before expiry to prevent logout loops.
 * Features:
 * - Monitors token expiry time
 * - Auto-refreshes 5 minutes before expiry
 * - Prevents multiple simultaneous refresh requests
 * - Handles tab visibility and network issues
 */

type TokenRefreshCallback = (newToken: string) => void
type TokenExpiredCallback = () => void

class TokenRefreshManager {
    private refreshTimer: number | null = null
    private isRefreshing = false
    private refreshCallbacks: TokenRefreshCallback[] = []
    private expiredCallbacks: TokenExpiredCallback[] = []

    // Refresh token 5 minutes before expiry
    private readonly REFRESH_BEFORE_EXPIRY_MS = 5 * 60 * 1000

    // Minimum time before refresh (prevent too frequent refreshes)
    private readonly MIN_REFRESH_INTERVAL_MS = 60 * 1000 // 1 minute

    /**
     * Start monitoring token expiry
     */
    start(): void {
        if (import.meta.env.DEV) {
            console.log('🔄 [TokenRefresh] Starting token refresh manager')
        }

        this.scheduleNextRefresh()

        // Re-check when tab becomes visible
        document.addEventListener('visibilitychange', () => {
            if (!document.hidden) {
                this.scheduleNextRefresh()
            }
        })
    }

    /**
     * Stop monitoring
     */
    stop(): void {
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer)
            this.refreshTimer = null
        }

        if (import.meta.env.DEV) {
            console.log('⏹️ [TokenRefresh] Stopped token refresh manager')
        }
    }

    /**
     * Schedule next token refresh
     */
    private scheduleNextRefresh(): void {
        // Clear existing timer
        if (this.refreshTimer) {
            clearTimeout(this.refreshTimer)
            this.refreshTimer = null
        }

        const expiresAt = this.getTokenExpiryTime()
        if (!expiresAt) {
            if (import.meta.env.DEV) {
                console.log('⚠️ [TokenRefresh] No token expiry time found')
            }
            return
        }

        const now = Date.now()
        const expiryTime = new Date(expiresAt).getTime()
        const timeUntilExpiry = expiryTime - now
        const timeUntilRefresh = timeUntilExpiry - this.REFRESH_BEFORE_EXPIRY_MS

        if (import.meta.env.DEV) {
            console.log('⏰ [TokenRefresh] Token expires in:', Math.round(timeUntilExpiry / 1000 / 60), 'minutes')
            console.log('⏰ [TokenRefresh] Will refresh in:', Math.round(timeUntilRefresh / 1000 / 60), 'minutes')
        }

        // If token is already expired or about to expire, refresh immediately
        if (timeUntilRefresh <= 0) {
            if (import.meta.env.DEV) {
                console.log('🚨 [TokenRefresh] Token expired or expiring soon, refreshing immediately')
            }
            this.refreshToken()
            return
        }

        // Schedule refresh
        this.refreshTimer = setTimeout(() => {
            this.refreshToken()
        }, Math.max(timeUntilRefresh, this.MIN_REFRESH_INTERVAL_MS))
    }

    /**
     * Refresh the access token
     */
    private async refreshToken(): Promise<void> {
        // Prevent multiple simultaneous refreshes
        if (this.isRefreshing) {
            if (import.meta.env.DEV) {
                console.log('⏳ [TokenRefresh] Refresh already in progress, skipping')
            }
            return
        }

        this.isRefreshing = true

        try {
            const refreshToken = localStorage.getItem('refreshToken')
            if (!refreshToken) {
                if (import.meta.env.DEV) {
                    console.error('❌ [TokenRefresh] No refresh token found')
                }
                this.notifyExpired()
                return
            }

            if (import.meta.env.DEV) {
                console.log('🔄 [TokenRefresh] Refreshing access token...')
            }

            // Dynamic import to avoid circular dependency
            const { authService } = await import('@features/authentication/services/authService')
            const newAccessToken = await authService.refreshToken()

            if (import.meta.env.DEV) {
                console.log('✅ [TokenRefresh] Token refreshed successfully')
            }

            // Notify callbacks
            this.notifyRefreshed(newAccessToken)

            // Schedule next refresh
            this.scheduleNextRefresh()
        } catch (error) {
            console.error('❌ [TokenRefresh] Failed to refresh token:', error)

            // If refresh fails, notify expired callbacks (will trigger logout)
            this.notifyExpired()
        } finally {
            this.isRefreshing = false
        }
    }

    /**
     * Get token expiry time from localStorage
     */
    private getTokenExpiryTime(): string | null {
        return localStorage.getItem('tokenExpiresAt')
    }

    /**
     * Set token expiry time in localStorage
     */
    setTokenExpiryTime(expiresAt: string | Date): void {
        const expiryTime = typeof expiresAt === 'string' ? expiresAt : expiresAt.toISOString()
        localStorage.setItem('tokenExpiresAt', expiryTime)

        if (import.meta.env.DEV) {
            console.log('💾 [TokenRefresh] Token expiry time set:', expiryTime)
        }

        // Reschedule refresh with new expiry time
        this.scheduleNextRefresh()
    }

    /**
     * Clear token expiry time
     */
    clearTokenExpiryTime(): void {
        localStorage.removeItem('tokenExpiresAt')
        this.stop()
    }

    /**
     * Register callback for when token is refreshed
     */
    onTokenRefreshed(callback: TokenRefreshCallback): void {
        this.refreshCallbacks.push(callback)
    }

    /**
     * Register callback for when token expires
     */
    onTokenExpired(callback: TokenExpiredCallback): void {
        this.expiredCallbacks.push(callback)
    }

    /**
     * Notify all refresh callbacks
     */
    private notifyRefreshed(newToken: string): void {
        this.refreshCallbacks.forEach(callback => {
            try {
                callback(newToken)
            } catch (error) {
                console.error('Error in token refresh callback:', error)
            }
        })
    }

    /**
     * Notify all expired callbacks
     */
    private notifyExpired(): void {
        this.expiredCallbacks.forEach(callback => {
            try {
                callback()
            } catch (error) {
                console.error('Error in token expired callback:', error)
            }
        })
    }

    /**
     * Check if token is expired or about to expire
     */
    isTokenExpiringSoon(): boolean {
        const expiresAt = this.getTokenExpiryTime()
        if (!expiresAt) return true

        const now = Date.now()
        const expiryTime = new Date(expiresAt).getTime()
        const timeUntilExpiry = expiryTime - now

        return timeUntilExpiry <= this.REFRESH_BEFORE_EXPIRY_MS
    }

    /**
     * Get time until token expires (in milliseconds)
     */
    getTimeUntilExpiry(): number | null {
        const expiresAt = this.getTokenExpiryTime()
        if (!expiresAt) return null

        const now = Date.now()
        const expiryTime = new Date(expiresAt).getTime()
        return Math.max(0, expiryTime - now)
    }
}

// Export singleton instance
export const tokenRefreshManager = new TokenRefreshManager()
