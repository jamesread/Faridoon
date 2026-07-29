import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

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
  plugins: [vue()],
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
