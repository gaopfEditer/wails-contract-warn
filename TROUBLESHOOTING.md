# 开发环境问题排查

## 常见问题及解决方案

### 1. "Serving assets from disk" 而不是使用开发服务器

**问题**：Wails 从 `frontend/dist` 加载资源，而不是使用 Vite 开发服务器。

**原因**：
- 前端开发服务器未启动
- `devServer` 配置未生效
- 端口被占用

**解决方案**：

1. **确保前端开发服务器正在运行**：
   ```bash
   cd frontend
   pnpm run dev
   ```
   应该看到：
   ```
   VITE v4.x.x  ready in xxx ms
   ➜  Local:   http://localhost:34115/
   ```

2. **检查端口是否被占用**：
   ```powershell
   # Windows
   netstat -ano | findstr :34115
   
   # 如果被占用，终止进程或修改端口
   ```

3. **验证开发服务器可访问**：
   - 在浏览器中打开 `http://localhost:34115`
   - 应该能看到前端应用

4. **确保 wails.json 配置正确**：
   ```json
   {
     "frontend": {
       "devServer": "http://localhost:34115",
       "devServerUrl": "ws://localhost:34115"
     }
   }
   ```

### 2. "Unable to create filesystem watcher"

**问题**：文件系统监听器无法创建。

**原因**：
- `-reloaddirs` 中包含了不存在的目录
- 路径格式错误
- 权限问题

**解决方案**：

1. **检查目录是否存在**：
   - `./app` 不存在（应该是 `app.go` 文件）
   - 只监听实际存在的目录

2. **使用正确的 reloaddirs**：
   ```bash
   wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils
   ```

3. **如果仍有问题，尝试只监听根目录**：
   ```bash
   wails dev -reloaddirs=.
   ```

### 3. "Unable to kill process and cleanup binary: Access is denied"

**问题**：无法删除旧的开发二进制文件。

**原因**：
- 旧的进程仍在运行
- 文件被锁定

**解决方案**：

1. **手动终止旧进程**：
   ```powershell
   # Windows PowerShell
   Get-Process -Name "wails-contract-warn-dev" | Stop-Process -Force
   
   # Windows CMD
   taskkill /F /IM wails-contract-warn-dev.exe
   ```

2. **删除 build 目录**：
   ```powershell
   Remove-Item -Recurse -Force build\bin\wails-contract-warn-dev.exe -ErrorAction SilentlyContinue
   ```

3. **使用启动脚本**（已包含自动清理）：
   ```powershell
   .\dev.ps1
   ```

### 4. "CreateFile ... app: The system cannot find the file specified"

**问题**：找不到 `app` 文件/目录。

**原因**：
- `-reloaddirs` 中包含了 `./app`，但实际是 `app.go` 文件，不是目录

**解决方案**：
- 从 `-reloaddirs` 中移除 `./app`
- 使用更新后的启动脚本

### 5. 前端修改后不自动更新

**检查清单**：

1. ✅ 前端开发服务器是否运行（`pnpm run dev`）
2. ✅ 浏览器控制台是否有 WebSocket 连接错误
3. ✅ `vite.config.js` 中是否配置了 `hmr`
4. ✅ 防火墙是否阻止了端口 34115
5. ✅ 是否使用了 `wails dev` 而不是 `wails build`

**验证步骤**：

1. 在浏览器中打开 `http://localhost:34115`
2. 修改 `frontend/src/App.vue`
3. 保存文件
4. 浏览器应该自动刷新

### 6. Go 代码修改后不自动重新编译

**检查清单**：

1. ✅ 是否使用了 `-reloaddirs` 参数
2. ✅ `-reloaddirs` 中是否包含了修改的目录
3. ✅ 文件系统监听器是否成功创建（查看启动日志）

**验证步骤**：

1. 修改 `app.go` 中的日志输出
2. 保存文件
3. 应该看到 Wails 重新编译的日志
4. 应用应该自动重启

## 推荐工作流

### 完整开发流程

1. **启动开发环境**：
   ```powershell
   .\dev.ps1
   ```

2. **验证前端开发服务器**：
   - 查看终端输出，应该看到 Vite 启动信息
   - 在浏览器中访问 `http://localhost:34115`

3. **验证热更新**：
   - 修改前端代码 → 应该自动更新
   - 修改 Go 代码 → 应该自动重新编译

### 如果遇到问题

1. **完全重启**：
   ```powershell
   # 停止所有相关进程
   Get-Process -Name "wails-contract-warn*" | Stop-Process -Force
   
   # 清理 build 目录
   Remove-Item -Recurse -Force build\bin -ErrorAction SilentlyContinue
   
   # 重新启动
   .\dev.ps1
   ```

2. **检查日志**：
   - 查看终端输出的错误信息
   - 检查是否有端口冲突
   - 检查文件路径是否正确

3. **简化配置**：
   - 如果 `-reloaddirs` 有问题，先尝试只监听根目录：
     ```bash
     wails dev -reloaddirs=.
     ```

## 快速修复命令

```powershell
# 1. 停止所有相关进程
Get-Process -Name "wails-contract-warn*","node" -ErrorAction SilentlyContinue | Stop-Process -Force

# 2. 清理 build 目录
Remove-Item -Recurse -Force build\bin -ErrorAction SilentlyContinue

# 3. 确保前端开发服务器运行（新终端）
cd frontend
pnpm run dev

# 4. 启动 Wails（原终端）
cd ..
.\dev.ps1
```

