import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 34115,
    strictPort: true,
    host: '0.0.0.0', // 允许外部访问（包括 WebView）
    hmr: {
      port: 34115,
      protocol: 'ws',
      host: 'localhost',
      clientPort: 34115, // 客户端连接端口
    },
    cors: true,
    watch: {
      usePolling: false,
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})

