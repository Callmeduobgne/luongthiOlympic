// Local development config - copy to vite.config.ts khi chạy local
// Hoặc sử dụng: cp vite.config.local.ts vite.config.ts

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

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
    host: '0.0.0.0',
    port: 5173,
    strictPort: true,
    hmr: {
      clientPort: 5173,
    },
    watch: {
      usePolling: false, // Không cần polling khi chạy local
    },
    proxy: {
      '/api': {
        // Local: proxy tới localhost
        target: 'http://localhost:8082', // API Gateway port
        changeOrigin: true,
      },
    },
  },
})


