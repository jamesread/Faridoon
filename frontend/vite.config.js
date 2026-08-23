import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'
import { supplementalThemesPlugin } from './vite.supplementalThemes.js'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  resolve: {
    alias: {
      '../../gen/faridoon/v1/faridoon_pb.js': resolve(__dirname, 'gen/faridoon/v1/faridoon_pb.ts'),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
  },
  plugins: [
    vue(),
    supplementalThemesPlugin(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['faridoon.png', 'icon.svg'],
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}', 'supplemental-themes/**/*.css'],
        navigateFallbackDenylist: [/^\/faridoon\.v1\./],
      },
      manifest: {
        name: 'Faridoon',
        short_name: 'Faridoon',
        description: 'Easily save and publish your favourite chat quotes for others to see.',
        theme_color: '#dee3e7',
        background_color: '#ffffff',
        display: 'standalone',
        orientation: 'portrait',
        start_url: '/',
        scope: '/',
        icons: [
          {
            src: 'pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
          },
          {
            src: 'pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
          },
          {
            src: 'icon.svg',
            sizes: 'any',
            type: 'image/svg+xml',
          },
        ],
      },
    }),
  ],
  server: {
    port: 3000,
    proxy: {
      '/faridoon.v1.FaridoonService': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
