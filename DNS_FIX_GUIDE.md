# DNS 解析问题修复指南

## 问题诊断

从诊断结果看，只有 `api.gateio.ws` 可以解析，其他所有域名都 DNS 解析失败。这是典型的 **DNS 污染/限制** 问题。

## 解决方案（按推荐顺序）

### ✅ 方案 1: 使用 Go 后端代理（最推荐）

**这是最可靠的解决方案！** Go 后端的网络库可以更好地处理 DNS 和代理。

我已经为你实现了后端代理功能，现在你可以：

1. **在 Go 后端调用**（已实现）：
   ```go
   // 已经在 app.go 中实现了以下方法：
   // - ProxyAPI(url, headers) - 通用代理
   // - GetMarketPrice(exchange, symbol) - 获取市场价格
   ```

2. **在前端调用**：
   ```javascript
   // 使用 Wails 调用后端代理
   import { GetMarketPrice } from '../wailsjs/go/main/App';
   
   // 获取 Binance 的 BTC 价格
   const result = await GetMarketPrice('binance', 'BTCUSDT');
   const data = JSON.parse(result);
   console.log('价格:', data.price);
   
   // 或者使用通用代理
   import { ProxyAPI } from '../wailsjs/go/main/App';
   const response = await ProxyAPI('https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT', '{}');
   ```

3. **优势**：
   - ✅ 绕过浏览器的 DNS 限制
   - ✅ 绕过 CORS 限制
   - ✅ 可以使用系统代理或环境变量代理
   - ✅ Go 的网络库更稳定

### 方案 2: 更换 DNS 服务器

#### Windows 修改 DNS

1. **打开网络设置**：
   - 右键点击任务栏的网络图标
   - 选择"网络和 Internet 设置"
   - 点击"更改适配器选项"

2. **修改 DNS**：
   - 右键点击你的网络连接（Wi-Fi 或以太网）
   - 选择"属性"
   - 双击"Internet 协议版本 4 (TCP/IPv4)"
   - 选择"使用下面的 DNS 服务器地址"
   - 输入：
     ```
     首选 DNS: 8.8.8.8
     备用 DNS: 8.8.4.4
     ```
     或
     ```
     首选 DNS: 1.1.1.1
     备用 DNS: 1.0.0.1
     ```

3. **清除 DNS 缓存**：
   ```powershell
   # 以管理员身份运行 PowerShell
   ipconfig /flushdns
   ```

#### 使用命令行快速修改（需要管理员权限）

```powershell
# 查看当前网络适配器
Get-NetAdapter

# 设置 DNS（替换 "以太网" 为你的适配器名称）
Set-DnsClientServerAddress -InterfaceAlias "以太网" -ServerAddresses "8.8.8.8","8.8.4.4"

# 或使用 Cloudflare DNS
Set-DnsClientServerAddress -InterfaceAlias "以太网" -ServerAddresses "1.1.1.1","1.0.0.1"
```

### 方案 3: 配置 VPN 的 DNS

1. **检查 VPN 设置**：
   - 打开你的 VPN 软件
   - 找到 DNS 设置
   - 确保使用 VPN 的 DNS 服务器，而不是本地 DNS

2. **某些 VPN 软件**：
   - 可能需要启用"使用 VPN DNS"选项
   - 或手动设置 VPN 提供的 DNS 服务器

### 方案 4: 使用 hosts 文件（临时方案）

**注意：** 这是临时方案，因为 IP 地址可能会变化。

1. **编辑 hosts 文件**（需要管理员权限）：
   - 路径：`C:\Windows\System32\drivers\etc\hosts`
   - 使用记事本以管理员身份打开

2. **添加域名映射**（需要先查询当前 IP）：
   ```
   # 查询当前 IP（在能访问的机器上）
   nslookup api.binance.com 8.8.8.8
   nslookup api.coingecko.com 8.8.8.8
   
   # 然后添加到 hosts 文件
   52.84.49.xxx api.binance.com
   104.248.xxx.xxx api.coingecko.com
   ```

**不推荐**：IP 地址会变化，维护麻烦。

### 方案 5: 配置系统代理环境变量

如果 VPN 提供了 HTTP/HTTPS 代理：

```powershell
# 设置代理（替换为你的实际代理地址和端口）
$env:HTTP_PROXY="http://127.0.0.1:7890"
$env:HTTPS_PROXY="http://127.0.0.1:7890"
$env:NO_PROXY="localhost,127.0.0.1"

# 然后运行你的应用
wails dev
```

## 推荐操作步骤

### 立即操作（5分钟）

1. **使用后端代理**（最简单）：
   - 前端代码已经可以通过 Wails 调用后端
   - 后端会自动使用系统代理和 DNS
   - 无需任何配置

2. **更换 DNS**（如果需要）：
   - 按照上面的步骤更换为 8.8.8.8
   - 清除 DNS 缓存

### 验证修复

运行诊断工具：
```bash
node frontend/test/api.js diagnose
```

或测试后端代理：
```javascript
// 在前端代码中
const result = await GetMarketPrice('binance', 'BTCUSDT');
console.log(result);
```

## 为什么 Go 后端可以工作？

1. **系统 DNS**：Go 使用系统的 DNS 解析，可能比浏览器的 DNS 更可靠
2. **代理支持**：Go 的 HTTP 客户端会自动使用系统代理环境变量
3. **无 CORS 限制**：服务器端没有跨域限制
4. **更好的错误处理**：可以获取更详细的网络错误信息

## 常见问题

### Q: 为什么只有 Gate.io 能解析？
A: Gate.io 可能在中国有 CDN 节点，或者它的域名没有被 DNS 污染。

### Q: 更换 DNS 后仍然不行？
A: 
1. 确保清除了 DNS 缓存
2. 确保 VPN 已连接
3. 尝试使用后端代理（最可靠）

### Q: 后端代理也失败？
A: 
1. 检查系统代理设置
2. 设置环境变量 `HTTPS_PROXY`
3. 检查防火墙是否阻止了 Go 程序的网络访问

## 总结

**最佳方案**：使用 Go 后端代理（已实现）
- ✅ 无需配置
- ✅ 最可靠
- ✅ 绕过所有浏览器限制

**备选方案**：更换 DNS 服务器
- 使用 8.8.8.8 或 1.1.1.1
- 清除 DNS 缓存

