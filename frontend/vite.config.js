import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

import tailwindcss from '@tailwindcss/vite'

import { VitePWA } from 'vite-plugin-pwa'
import pkg from './package.json'

// https://vite.dev/config/
export default defineConfig({
  base: "/",
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version),
  },
  plugins: [
    vue(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.png', 'logo_atiga.png'],
      manifest: {
        name: 'AtigaCBT Evaluation System',
        short_name: 'AtigaCBT',
        description: 'Modern Evaluation Dashboard for AtigaCBT',
        theme_color: '#4f46e5',
        background_color: '#020617',
        display: 'standalone',
        icons: [
          {
            src: 'logo_atiga.png',
            sizes: '192x192',
            type: 'image/png'
          },
          {
            src: 'logo_atiga.png',
            sizes: '512x512',
            type: 'image/png'
          }
        ]
      }
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return

          if (
            id.includes('/vue/') ||
            id.includes('/vue-router/') ||
            id.includes('/pinia/') ||
            id.includes('/@tanstack/')
          ) {
            return 'framework'
          }
          if (id.includes('/@mdi/js/')) return 'icons'
          if (id.includes('/katex/')) return 'katex'
          if (id.includes('/chart.js/')) return 'charts'
          if (
            id.includes('/html5-qrcode/') ||
            id.includes('/qrcode.vue/') ||
            id.includes('/html-to-image/')
          ) {
            return 'media-tools'
          }
        },
      },
    },
  },
})
