# 前端热更新调试指南

## 问题：修改前端代码后 WebView 不自动更新

### 诊断步骤

#### 1. 确认前端开发服务器正在运行

**检查方法：**
```powershell
# 运行检查脚本
.\check-dev-server.ps1
```

**或手动检查：**
```powershell
# 检查端口
netstat -ano | findstr :34115

# 测试 HTTP 连接
Invoke-WebRequest -Uri "http://localhost:34115" -UseBasicParsing
```

**如果未运行，启动它：**
```bash
cd frontend
pnpm run dev
```

应该看到：
```
VITE v4.x.x  ready in xxx ms
➜  Local:   http://localhost:34115/
```

#### 2. 确认 WebView 连接到开发服务器

**检查方法：**
1. 在 WebView 中按 `F12` 打开开发者工具
2. 查看 **Console** 标签
3. 应该看到 Vite 的 HMR 连接信息

**如果看到 "Serving assets from disk"：**
- WebView 没有连接到开发服务器
- 检查 `wails.json` 中的 `devServer` 配置
- 确保使用 `wails dev` 而不是 `wails build`

#### 3. 检查 WebSocket 连接（HMR）

**在浏览器开发者工具中：**
1. 打开 **Network** 标签
2. 筛选 **WS** (WebSocket)
3. 应该看到连接到 `ws://localhost:34115`

**如果连接失败：**
- 检查防火墙设置
- 检查 `vite.config.js` 中的 `hmr` 配置
- 尝试修改 `host` 为 `0.0.0.0`

#### 4. 验证 Vite HMR 配置

**检查 `frontend/vite.config.js`：**
```javascript
server: {
  port: 34115,
  strictPort: true,
  host: '0.0.0.0', // 允许外部访问
  hmr: {
    port: 34115,
    protocol: 'ws',
    host: 'localhost',
    clientPort: 34115,
  },
}
```

#### 5. 测试热更新

1. **修改前端文件**（如 `frontend/src/App.vue`）
2. **保存文件**
3. **查看终端**：应该看到 Vite 的 HMR 日志
4. **查看浏览器控制台**：应该看到 HMR 更新信息

### 常见问题及解决方案

#### 问题 1: WebView 显示 "Serving assets from disk"

**原因**：WebView 没有使用开发服务器，而是使用嵌入的静态文件。

**解决方案**：
1. 确保前端开发服务器正在运行
2. 确保使用 `wails dev` 启动
3. 检查 `wails.json` 中的 `devServer` 配置
4. 重启 Wails 应用

#### 问题 2: WebSocket 连接失败

**原因**：防火墙或网络配置阻止了 WebSocket 连接。

**解决方案**：
1. 检查 Windows 防火墙设置
2. 尝试修改 `vite.config.js` 中的 `host` 为 `0.0.0.0`
3. 检查是否有代理软件干扰

#### 问题 3: 修改后需要手动刷新

**原因**：HMR 没有正常工作。

**解决方案**：
1. 检查浏览器控制台是否有错误
2. 检查 WebSocket 连接是否建立
3. 尝试在浏览器中直接访问 `http://localhost:34115` 测试 HMR

#### 问题 4: 只有部分文件更新

**原因**：Vue 组件的 HMR 可能有问题。

**解决方案**：
1. 检查 `vite.config.js` 中是否正确配置了 Vue 插件
2. 尝试完全刷新页面（Ctrl+R）
3. 检查是否有语法错误阻止了 HMR

### 完整调试流程

1. **启动前端开发服务器**（独立终端）：
   ```bash
   cd frontend
   pnpm run dev
   ```

2. **验证开发服务器**：
   - 在浏览器中打开 `http://localhost:34115`
   - 应该能看到应用界面

3. **启动 Wails**（另一个终端）：
   ```powershell
   .\dev.ps1
   ```

4. **检查 WebView 连接**：
   - 在 WebView 中按 F12
   - 查看 Console，应该看到 Vite 的 HMR 信息
   - 查看 Network → WS，应该看到 WebSocket 连接

5. **测试热更新**：
   - 修改 `frontend/src/App.vue`
   - 保存文件
   - WebView 应该自动更新

### 如果仍然不工作

1. **完全重启**：
   ```powershell
   # 停止所有进程
   Get-Process -Name "wails-contract-warn*","node" -ErrorAction SilentlyContinue | Stop-Process -Force
   
   # 重新启动
   cd frontend
   pnpm run dev
   # 新终端
   cd ..
   .\dev.ps1
   ```

2. **检查日志**：
   - 查看 Vite 终端输出
   - 查看 Wails 终端输出
   - 查看浏览器控制台

3. **尝试简化配置**：
   - 临时移除 `hmr` 配置，只使用基本的文件监听
   - 或者使用 `watch` 模式而不是 HMR

