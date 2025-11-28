# WebView 热更新修复指南

## 问题：WebView 使用 build 代码而不是开发服务器

### 根本原因

Wails 在启动时会检查开发服务器是否可用。如果开发服务器在 Wails 启动时不可用，它会回退到使用嵌入的静态文件（`frontend/dist`）。

### 解决方案

**关键：必须在 Wails 启动前确保前端开发服务器已经运行！**

## 正确的启动顺序

### 方法 1: 手动启动（推荐，最可靠）

**步骤 1: 启动前端开发服务器（终端 1）**
```bash
cd frontend
pnpm run dev
```

**等待看到：**
```
VITE v4.x.x  ready in xxx ms
➜  Local:   http://localhost:34115/
```

**步骤 2: 验证开发服务器可访问**
- 在浏览器中打开 `http://localhost:34115`
- 应该能看到应用界面

**步骤 3: 启动 Wails（终端 2）**
```powershell
.\dev.ps1
```

**现在 WebView 应该连接到开发服务器！**

### 方法 2: 使用 Wails 自动启动（可能不稳定）

Wails 可以自动启动前端开发服务器，但有时检测可能失败。

```powershell
.\dev.ps1
```

如果 WebView 仍然使用 build 代码，使用方法 1。

## 验证 WebView 是否连接到开发服务器

### 方法 1: 查看 Wails 启动日志

启动 Wails 时，应该看到类似这样的日志：
```
Using dev server: http://localhost:34115
```

如果看到：
```
Serving assets from disk: D:\...\frontend\dist
```
说明没有使用开发服务器。

### 方法 2: 在 WebView 中检查

1. 在 WebView 中按 `F12` 打开开发者工具
2. 查看 **Network** 标签
3. 刷新页面（F5）
4. 查看第一个请求的 URL：
   - ✅ **正确**：`http://localhost:34115/` 或 `http://localhost:34115/index.html`
   - ❌ **错误**：`wails://wails/` 或文件路径（说明使用嵌入的文件）

### 方法 3: 测试热更新

1. 修改 `frontend/src/App.vue`（比如改个标题）
2. 保存文件
3. **在浏览器中**：应该自动更新 ✅
4. **在 WebView 中**：
   - 如果连接到开发服务器：应该自动更新 ✅
   - 如果使用嵌入文件：不会更新 ❌

## 已做的修改

### 1. 修改了 `main.go`

在开发模式下，将 `AssetServer` 设置为 `nil`，强制 Wails 使用 `devServer`：

```go
if isDev {
    assetServer = nil  // 让 Wails 使用 devServer
} else {
    assetServer = &assetserver.Options{Assets: assets}  // 生产模式使用嵌入文件
}
```

### 2. 更新了启动脚本

- 添加了开发服务器检查
- 设置了 `WAILS_ENV=development` 环境变量

## 完整测试流程

### 1. 完全重启

```powershell
# 停止所有相关进程
Get-Process -Name "wails-contract-warn*","node" -ErrorAction SilentlyContinue | Stop-Process -Force

# 清理 build 目录（可选）
Remove-Item -Recurse -Force build\bin\wails-contract-warn-dev.exe -ErrorAction SilentlyContinue
```

### 2. 启动前端开发服务器（终端 1）

```bash
cd frontend
pnpm run dev
```

**等待看到：**
```
VITE v4.x.x  ready in xxx ms
➜  Local:   http://localhost:34115/
```

### 3. 验证开发服务器（可选但推荐）

在浏览器中打开 `http://localhost:34115`，确认能看到应用。

### 4. 启动 Wails（终端 2）

```powershell
.\dev.ps1
```

### 5. 检查 WebView

1. 查看 Wails 启动日志，确认是否使用开发服务器
2. 在 WebView 中按 F12，检查 Network 标签
3. 修改前端代码，测试热更新

## 如果仍然不工作

### 检查清单

- [ ] 前端开发服务器是否在 Wails 启动前就已经运行？
- [ ] 开发服务器是否在 `http://localhost:34115` 可访问？
- [ ] `wails.json` 中的 `devServer` 配置是否正确？
- [ ] `vite.config.js` 中的端口是否与 `wails.json` 一致？
- [ ] 是否使用了 `wails dev` 而不是 `wails build`？

### 强制使用开发服务器

如果 Wails 仍然不使用开发服务器，可以尝试：

1. **删除 dist 目录**（临时）：
   ```powershell
   Remove-Item -Recurse -Force frontend\dist -ErrorAction SilentlyContinue
   ```
   这样 Wails 无法使用嵌入文件，必须使用开发服务器。

2. **检查 Wails 版本**：
   ```bash
   wails version
   ```
   确保使用 Wails v2.11.0 或更高版本。

3. **查看详细日志**：
   ```powershell
   $env:WAILS_DEBUG="true"
   .\dev.ps1
   ```

## 成功标志

当一切正常时，你应该看到：

1. **Wails 启动日志**：
   ```
   Using dev server: http://localhost:34115
   ```

2. **WebView Network 标签**：
   - 第一个请求是 `http://localhost:34115/`

3. **热更新工作**：
   - 修改前端代码 → WebView 自动更新

4. **浏览器控制台**（F12）：
   - 看到 Vite HMR 连接信息

