import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', '')
  const apiProxyTarget =
    env.VITE_DEV_API_PROXY_TARGET || 'http://localhost:8080'
  const proxy = {
    '/api': apiProxyTarget,
    '/robots.txt': apiProxyTarget,
    '/sitemap.xml': apiProxyTarget,
  }

  return {
    plugins: [react(), tailwindcss()],
    build: {
      sourcemap: false,
      manifest: true,
      rollupOptions: {
        output: {
          // Keep framework code in a stable cacheable file instead of making
          // every application edit invalidate one oversized entry bundle.
          // The initial-transfer budget still counts this chunk, so this is a
          // delivery improvement rather than a hidden budget exception.
          manualChunks(id) {
            if (id.includes('/node_modules/')) {
              return 'vendor'
            }
          },
        },
      },
    },
    server: {
      proxy,
    },
    preview: { proxy },
    test: {
      environment: 'jsdom',
      exclude: ['e2e/**', 'node_modules/**', 'dist/**'],
      setupFiles: './src/test/setup.ts',
    },
  }
})
