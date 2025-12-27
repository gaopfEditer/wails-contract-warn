# 前端热更新修复指南

## 问题诊断步骤

### 步骤 1: 测试浏览器中的热更新

**这是最重要的测试！** 如果浏览器中热更新不工作，说明是 Vite 配置问题。

1. **启动前端开发服务器**：
   ```powershell
   cd frontend
   pnpm run dev
   ```

2. **在浏览器中打开** `http://localhost:34115`

3. **修改前端文件**（如 `frontend/src/App.vue`）：
   - 添加一行文字或改变颜色
   - 保存文件

4. **观察浏览器**：
   - ✅ **如果浏览器自动更新**：说明 Vite HMR 正常工作，问题在 WebView 连接
   - ❌ **如果浏览器不更新**：说明 Vite 配置有问题，需要修复

### 步骤 2: 检查 WebView 是否连接到开发服务器

1. **启动 Wails**：
   ```powershell
   .\dev.ps1
   ```

2. **在 WebView 中按 `F12` 打开开发者工具**

3. **查看 Network 标签**：
   - 刷新页面（F5）
   - 查看第一个请求的 URL：
     - ✅ **正确**：`http://localhost:34115/` 或 `http://localhost:34115/index.html`
     - ❌ **错误**：`wails://wails/` 或文件路径（说明使用嵌入的文件）

4. **查看 Console 标签**：
   - 应该看到 Vite 的 HMR 连接信息
   - 如果看到 "Serving assets from disk"，说明没有使用开发服务器

### 步骤 3: 检查 Wails 启动日志

启动 Wails 时，查看日志：

- ✅ **正确**：`Using dev server: http://localhost:34115`
- ❌ **错误**：`Serving assets from disk: D:\...\frontend\dist`

## 修复方案

### 方案 1: 确保正确的启动顺序（最重要！）

**必须按以下顺序启动：**

1. **终端 1 - 启动前端开发服务器**：
   ```powershell
   cd frontend
   pnpm run dev
   ```
   
   **等待看到：**
   ```
   VITE v4.x.x  ready in xxx ms
   ➜  Local:   http://localhost:34115/
   ```

2. **验证开发服务器**（在浏览器中打开 `http://localhost:34115`）

3. **终端 2 - 启动 Wails**：
   ```powershell
   .\dev.ps1
   ```

### 方案 2: 如果浏览器中热更新不工作

**检查 Vite 配置**：

1. 确认 `frontend/vite.config.js` 中的配置：
   ```javascript
   server: {
     port: 34115,
     strictPort: true,
     host: '0.0.0.0',
     hmr: {
       port: 34115,
       protocol: 'ws',
       host: 'localhost',
       clientPort: 34115,
     },
     watch: {
       usePolling: true,
       interval: 100,
     },
   }
   ```

2. **重启开发服务器**：
   ```powershell
   # 停止当前服务器（Ctrl+C）
   cd frontend
   pnpm run dev
   ```

### 方案 3: 如果 WebView 没有连接到开发服务器

**检查 main.go**：

确认开发模式下 `assetServer = nil`：
```go
if isDev {
    assetServer = nil  // 强制使用 devServer
}
```

**检查环境变量**：

确认 `WAILS_ENV=development` 已设置：
```powershell
$env:WAILS_ENV = "development"
$env:DEV = "true"
```

### 方案 4: 完全重启

如果以上都不行，完全重启：

```powershell
# 1. 停止所有进程
Get-Process -Name "wails-contract-warn*","node" -ErrorAction SilentlyContinue | Stop-Process -Force

# 2. 清理 build 目录
Remove-Item -Recurse -Force build\bin\wails-contract-warn-dev.exe -ErrorAction SilentlyContinue

# 3. 重新启动（按正确顺序）
# 终端1:
cd frontend
pnpm run dev

# 等待看到 VITE ready 后，终端2:
.\dev.ps1
```

## 验证热更新

### 测试 1: 浏览器测试

1. 在浏览器中打开 `http://localhost:34115`
2. 修改 `frontend/src/App.vue`
3. 保存文件
4. 浏览器应该自动更新 ✅

### 测试 2: WebView 测试

1. 启动 Wails
2. 在 WebView 中按 F12
3. 查看 Network 标签，确认连接到 `http://localhost:34115`
4. 修改前端文件
5. 保存文件
6. WebView 应该自动更新 ✅

## 常见问题

### Q: 浏览器中热更新工作，但 WebView 中不工作

**A**: WebView 没有连接到开发服务器。检查：
- Wails 启动日志是否显示 "Using dev server"
- WebView 的 Network 标签是否显示 `http://localhost:34115`
- 确认 `main.go` 中开发模式下 `assetServer = nil`

### Q: 浏览器和 WebView 中都不工作

**A**: Vite 配置问题。检查：
- `vite.config.js` 中的 HMR 配置
- 开发服务器是否正常运行
- 查看 Vite 终端是否有错误

### Q: 修改文件后没有任何反应

**A**: 文件监听问题。检查：
- `vite.config.js` 中是否启用了 `usePolling: true`
- 文件是否保存在正确的位置
- 查看 Vite 终端是否有文件变化日志

## 调试命令

```powershell
# 检查端口
netstat -ano | findstr :34115

# 测试开发服务器
Invoke-WebRequest -Uri "http://localhost:34115" -UseBasicParsing

# 检查环境变量
echo $env:WAILS_ENV
echo $env:DEV

# 运行诊断脚本
.\check-hmr.ps1
```

