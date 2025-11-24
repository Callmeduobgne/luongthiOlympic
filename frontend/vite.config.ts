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
        // Trong Docker: proxy tới backend container (ibn-backend:8080)
        // Trong local dev: proxy trực tiếp tới backend (localhost:9090)
        // Sử dụng env variable VITE_API_BASE_URL từ docker-compose hoặc mặc định
        target: process.env.VITE_API_BASE_URL || 'http://localhost:9090',
        changeOrigin: true,
        ws: true, // Enable WebSocket for both HMR and API WebSocket endpoints
        rewrite: (path) => path, // Keep original path
      },
    },
  },
})
