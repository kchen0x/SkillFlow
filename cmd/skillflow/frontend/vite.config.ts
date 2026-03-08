import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return
          }
          if (id.includes('react-router') || id.includes('\\react\\') || id.includes('/react/') || id.includes('react-dom')) {
            return 'react-vendor'
          }
          if (id.includes('framer-motion') || id.includes('motion-dom') || id.includes('motion-utils')) {
            return 'motion-vendor'
          }
          if (id.includes('@radix-ui') || id.includes('react-remove-scroll') || id.includes('aria-hidden')) {
            return 'radix-vendor'
          }
          if (id.includes('lucide-react')) {
            return 'icons-vendor'
          }
        },
      },
      onwarn(warning, warn) {
        if (warning.code === 'MODULE_LEVEL_DIRECTIVE' && warning.message.includes('use client')) {
          return
        }
        warn(warning)
      },
    },
  },
})
