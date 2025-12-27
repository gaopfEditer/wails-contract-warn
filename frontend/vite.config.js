import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  // 确保根目录正确
  root: resolve(__dirname),
  server: {
    port: 34115,
    strictPort: true,
    host: '0.0.0.0', // 允许外部访问（包括 WebView）
    hmr: {
      port: 34115,
      protocol: 'ws',
      host: 'localhost',
      clientPort: 34115, // 客户端连接端口
      overlay: true, // 显示错误覆盖层
    },
    // 确保开发服务器可以被外部访问
    open: false, // 不自动打开浏览器
    proxy: {}, // 空代理配置
    cors: true,
    watch: {
      // Windows上必须使用轮询才能可靠检测文件变化
      usePolling: true,
      interval: 200, // 轮询间隔（毫秒）- 稍微增加以提高稳定性
      binaryInterval: 300, // 二进制文件轮询间隔
      // 只忽略真正不需要的文件
      ignored: [
        '**/node_modules/**',
        '**/.git/**',
        '**/dist/**',
        '**/build/**',
        '**/.vite/**',
      ],
      // 明确指定要监听的文件
      include: [
        'src/**/*.{js,ts,vue,jsx,tsx,json,css,scss,html}',
        'index.html',
        'vite.config.js',
      ],
    },
    // 确保文件变化时触发重新加载
    fs: {
      strict: false, // 允许访问项目根目录外的文件
      allow: ['..'], // 允许访问父目录（如果需要）
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    // 禁用构建时的文件监听（开发模式不需要）
    watch: null,
  },
  // 确保开发模式下启用热更新
  optimizeDeps: {
    exclude: [],
    include: ['vue', 'echarts'], // 预构建这些依赖
    // 强制重新构建依赖（如果热更新有问题）
    force: false,
  },
  // 确保文件变化时触发重新加载
  clearScreen: false, // 不清屏，保留日志
  // 启用热更新
  appType: 'spa', // 单页应用模式
  // 日志级别
  logLevel: 'info',
})

