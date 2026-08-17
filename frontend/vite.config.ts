import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      proxy: {
        '/api': env.VITE_DEV_API_PROXY_TARGET || 'http://localhost:8080',
        '/robots.txt': env.VITE_DEV_API_PROXY_TARGET || 'http://localhost:8080',
        '/sitemap.xml':
          env.VITE_DEV_API_PROXY_TARGET || 'http://localhost:8080',
      },
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/test/setup.ts',
    },
  }
})
