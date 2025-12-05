import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@features': path.resolve(__dirname, './src/features'),
      '@shared': path.resolve(__dirname, './src/shared'),
    },
  },
  server: {
    host: '0.0.0.0', // Allow external connections (for Docker)
    port: 5173,
    strictPort: false, // Allow port fallback in Docker
    hmr: {
      // Let Vite auto-detect host from browser connection
      // This works for both localhost and remote access
      port: 5173,
      protocol: 'ws',
    },
    watch: {
      usePolling: true, // Enable for Docker volume mounts
      interval: 1000, // Polling interval
    },
    proxy: {
      '/api': {
        // Frontend dev và prod đều dùng relative URL trong code
        // Dev: Vite proxy forward /api/* tới backend (localhost:9900)
        // Prod: Nginx proxy forward /api/* tới backend (ibn-backend:8080)
        // Cả 2 đều hoạt động giống nhau, chỉ khác proxy layer
        target: process.env.VITE_API_BASE_URL || 'http://localhost:9900',
        changeOrigin: true,
        ws: true, // Enable WebSocket for both HMR and API WebSocket endpoints
        rewrite: (path) => path, // Keep original path (including query parameters)
        configure: (proxy, _options) => {
          // Production best practice: Configure WebSocket upgrade properly
          proxy.on('error', (err, _req, _res) => {
            console.log('Proxy error:', err)
          })
          proxy.on('proxyReqWs', (proxyReq, req) => {
            // CRITICAL: Ensure query parameters are preserved in WebSocket URL
            // Vite proxy should preserve query params automatically, but we ensure it here

            // Ensure WebSocket upgrade headers are set correctly
            proxyReq.setHeader('Upgrade', req.headers.upgrade || 'websocket')
            proxyReq.setHeader('Connection', req.headers.connection || 'Upgrade')

            // Forward Authorization header if present (fallback for token)
            if (req.headers.authorization) {
              proxyReq.setHeader('Authorization', req.headers.authorization)
            }
          })
        },
      },
    },
  },
})
