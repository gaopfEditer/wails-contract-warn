# 网络访问问题排查指南

## 问题现象

即使配置了 VPN 并可以访问外网，大部分交易所 API 仍然无法访问，只有部分 API（如 Gate.io）可以正常使用。

## 可能的原因

### 1. **DNS 解析问题**
- VPN 可能只代理了 HTTP/HTTPS 流量，但没有代理 DNS 查询
- 本地 DNS 服务器可能无法解析某些域名
- DNS 缓存问题

### 2. **代理配置问题**
- **浏览器环境**：浏览器可能没有使用系统代理设置
- **Node.js 环境**：Node.js 默认不使用系统代理，需要手动配置环境变量

### 3. **防火墙/安全软件阻止**
- Windows 防火墙可能阻止了某些连接
- 杀毒软件或安全软件可能阻止了网络请求
- 企业网络可能有防火墙规则

### 4. **CORS 限制（浏览器环境）**
- 某些 API 可能不允许浏览器直接访问（CORS 策略）
- 需要通过后端代理访问

### 5. **API 地区限制**
- 某些交易所可能对特定地区有访问限制
- 即使使用 VPN，IP 可能仍被识别为受限地区

### 6. **超时时间过短**
- 网络延迟较高时，5 秒超时可能不够
- 已增加到 15 秒

## 解决方案

### 方案 1: 运行网络诊断（推荐）

首先运行诊断工具，了解具体问题：

```bash
# 在 Node.js 环境中
node frontend/test/api.js diagnose

# 或在浏览器控制台中
# 加载 network-diagnosis.js 后运行
networkDiagnosis.diagnoseAllAPIs()
```

诊断工具会检查：
- DNS 解析是否正常
- 基本连接是否成功
- 环境配置信息

### 方案 2: 配置 Node.js 代理（如果使用 Node.js）

如果是在 Node.js 环境中运行，需要设置代理环境变量：

**Windows PowerShell:**
```powershell
$env:HTTPS_PROXY="http://127.0.0.1:7890"  # 替换为你的代理端口
$env:HTTP_PROXY="http://127.0.0.1:7890"
node frontend/test/api.js
```

**Windows CMD:**
```cmd
set HTTPS_PROXY=http://127.0.0.1:7890
set HTTP_PROXY=http://127.0.0.1:7890
node frontend/test/api.js
```

**Linux/Mac:**
```bash
export HTTPS_PROXY=http://127.0.0.1:7890
export HTTP_PROXY=http://127.0.0.1:7890
node frontend/test/api.js
```

**查找代理端口：**
- 查看 VPN 软件的代理设置
- 常见端口：7890, 1080, 8080, 8888

### 方案 3: 检查 DNS 设置

1. **更换 DNS 服务器**：
   - 使用 Google DNS: `8.8.8.8`, `8.8.4.4`
   - 使用 Cloudflare DNS: `1.1.1.1`, `1.0.0.1`

2. **Windows 修改 DNS**：
   - 控制面板 → 网络和共享中心 → 更改适配器设置
   - 右键网络连接 → 属性 → IPv4 → 使用下面的 DNS 服务器地址

### 方案 4: 检查防火墙设置

1. **临时关闭防火墙测试**：
   - Windows 安全中心 → 防火墙和网络保护 → 关闭防火墙（测试用）

2. **添加防火墙规则**：
   - 允许 Node.js 或浏览器通过防火墙

### 方案 5: 使用后端代理（推荐用于生产环境）

由于浏览器环境的 CORS 限制，建议在后端（Go）中实现 API 代理：

```go
// 在 Go 后端添加代理接口
func (a *App) ProxyAPI(url string) (string, error) {
    // 使用 Go 的 http.Client 访问 API
    // 这样可以绕过浏览器的 CORS 限制
    // 并且可以使用系统代理或配置的代理
}
```

### 方案 6: 增加超时时间和重试

已更新代码：
- 超时时间从 5 秒增加到 15 秒
- 添加了更详细的错误信息
- 改进了错误处理

### 方案 7: 检查 VPN 配置

1. **确保 VPN 使用全局代理模式**：
   - 某些 VPN 可能只代理特定应用
   - 确保使用系统代理模式

2. **检查代理类型**：
   - HTTP/HTTPS 代理
   - SOCKS5 代理（需要额外配置）

3. **测试 VPN 连接**：
   ```bash
   # 测试是否能访问外网
   curl https://www.google.com
   
   # 测试特定 API
   curl https://api.binance.com/api/v3/ping
   ```

## 快速检查清单

- [ ] 运行网络诊断工具
- [ ] 检查 VPN 是否正常工作
- [ ] 检查代理端口和配置
- [ ] 设置 Node.js 代理环境变量（如适用）
- [ ] 检查防火墙设置
- [ ] 尝试更换 DNS 服务器
- [ ] 检查是否有安全软件阻止
- [ ] 增加超时时间（已更新）
- [ ] 查看详细错误信息（已更新）

## 常见错误及解决方法

### "请求超时"
- **原因**：网络延迟高或连接被阻止
- **解决**：增加超时时间、检查网络连接、使用代理

### "fetch failed" 或 "网络连接失败"
- **原因**：DNS 解析失败或无法建立连接
- **解决**：检查 DNS 设置、配置代理、检查防火墙

### "CORS 错误"
- **原因**：浏览器跨域限制
- **解决**：使用后端代理或浏览器扩展（仅开发环境）

### "HTTP 403" 或 "HTTP 429"
- **原因**：API 限制或 IP 被封禁
- **解决**：更换 IP、使用其他 API、添加请求头

## 测试命令

```bash
# 1. 运行网络诊断
node frontend/test/api.js diagnose

# 2. 测试所有 API
node frontend/test/api.js all

# 3. 测试单个 API（自动切换）
node frontend/test/api.js test

# 4. 健康检查
node frontend/test/api.js health
```

## 如果问题仍然存在

1. **查看详细日志**：代码已添加更详细的错误信息
2. **尝试不同的 API**：某些 API 可能更容易访问
3. **使用后端代理**：这是最可靠的解决方案
4. **联系网络管理员**：如果是企业网络，可能需要特殊配置

## 相关文件

- `frontend/test/api.js` - API 测试脚本
- `frontend/test/network-diagnosis.js` - 网络诊断工具
- `sync/exchange.go` - Go 后端 API 访问（可参考实现代理）

