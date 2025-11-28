# 开发环境配置指南

## 前端热更新配置

Wails 支持在开发模式下使用前端开发服务器，实现热更新（HMR）。

### 配置说明

1. **wails.json** 配置：
   - `devServer`: 前端开发服务器地址（HTTP）
   - `devServerUrl`: WebSocket 地址（用于 HMR）

2. **vite.config.js** 配置：
   - `server.port`: 必须与 `wails.json` 中的端口一致（34115）
   - `server.hmr`: 配置 HMR 的 WebSocket 连接

### 使用方法

#### 方法 1: 使用 `wails dev`（推荐）

**Windows PowerShell:**
```powershell
.\dev.ps1
```

**Windows CMD:**
```cmd
dev.bat
```

**Linux/Mac:**
```bash
chmod +x dev.sh
./dev.sh
```

**或者直接使用命令（需要指定 reloaddirs）:**
```bash
wails dev -reloaddirs=.,./api,./app,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils
```

这会自动：
1. 启动前端开发服务器（Vite）
2. 启动 Go 后端
3. 在 WebView 中加载开发服务器地址
4. 启用热更新（HMR）
5. 监听 Go 文件变化并自动重新编译

**优点**：
- ✅ 自动管理前后端
- ✅ 支持前端热更新（Vite HMR）
- ✅ 支持后端自动重新编译（Go 文件变化）
- ✅ 无需手动启动多个服务

#### 方法 2: 手动启动（调试用）

如果需要分别控制前后端：

**终端 1 - 启动前端开发服务器：**
```bash
cd frontend
pnpm run dev
```

**终端 2 - 启动 Wails 应用：**
```bash
wails dev
```

### 验证热更新

1. 启动 `wails dev`
2. 修改前端代码（如 `frontend/src/App.vue`）
3. 保存文件
4. 在 WebView 中应该看到自动刷新

### 常见问题

#### 1. 热更新不工作

**检查项**：
- ✅ 确保 `wails.json` 中的 `devServer` 端口与 `vite.config.js` 中的 `server.port` 一致
- ✅ 确保 `vite.config.js` 中配置了 `hmr`
- ✅ 检查防火墙是否阻止了端口 34115
- ✅ 查看终端是否有错误信息

**解决方案**：
```bash
# 检查端口是否被占用
netstat -ano | findstr :34115

# 如果被占用，修改 vite.config.js 和 wails.json 中的端口
```

#### 2. WebView 显示空白

**可能原因**：
- 前端开发服务器未启动
- 端口配置错误
- CORS 问题

**解决方案**：
1. 检查前端开发服务器是否运行
2. 在浏览器中访问 `http://localhost:34115` 验证前端是否正常
3. 检查 `vite.config.js` 中的 `server.cors` 配置

#### 3. 修改后需要手动刷新

**检查项**：
- ✅ 确保 `vite.config.js` 中配置了 `hmr`
- ✅ 检查浏览器控制台是否有 WebSocket 连接错误
- ✅ 确保使用的是 `wails dev` 而不是 `wails build`

### 开发 vs 生产模式

#### 开发模式（`wails dev`）
- 使用前端开发服务器（Vite）
- 支持热更新（HMR）
- 代码未压缩，便于调试
- 自动重新编译

#### 生产模式（`wails build`）
- 使用嵌入的静态文件（`frontend/dist`）
- 需要先运行 `pnpm run build` 构建前端
- 代码已压缩优化
- 无热更新

### 推荐工作流

1. **开发时**：
   ```bash
   # Windows
   .\dev.ps1
   # 或
   dev.bat
   
   # Linux/Mac
   ./dev.sh
   ```
   - 修改前端代码 → Vite HMR 自动热更新（无需刷新）
   - 修改后端 Go 代码 → Wails 自动重新编译并重启应用

2. **构建生产版本**：
   ```bash
   # 构建前端
   cd frontend
   pnpm run build
   
   # 构建应用
   cd ..
   wails build
   ```

### 调试技巧

1. **查看前端日志**：
   - 打开浏览器开发者工具（在 WebView 中按 F12）
   - 查看 Console 和 Network 标签

2. **查看后端日志**：
   - 查看运行 `wails dev` 的终端输出
   - 使用 `logger` 包输出的日志

3. **检查网络连接**：
   - 在浏览器中直接访问 `http://localhost:34115`
   - 验证前端开发服务器是否正常

### 配置文件位置

- `wails.json` - Wails 配置文件
- `frontend/vite.config.js` - Vite 配置文件
- `frontend/package.json` - 前端依赖配置

