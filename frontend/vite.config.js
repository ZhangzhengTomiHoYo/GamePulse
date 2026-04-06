import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  // --- 新增：跨域代理配置 ---
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // 你的 Go 后端地址
        changeOrigin: true
      }
    }
  }
})