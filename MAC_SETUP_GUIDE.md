# Mac 环境运行指南

## 是的，需要安装 Go！

这个项目是基于 **Wails v2** 的 Go 桌面应用，必须安装 Go 才能运行。

## 环境要求

1. **Go 1.22.0 或更高版本**（项目要求 Go 1.22.0）
2. **Node.js 16+** 和 **pnpm**（前端依赖管理）
3. **Wails CLI v2**（Go 桌面应用框架）
4. **MySQL 8.0+**（可选，用于数据持久化）

## 安装步骤

### 1. 安装 Go

#### 方法 1: 使用 Homebrew（推荐）

```bash
# 安装 Homebrew（如果还没有）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 Go
brew install go

# 验证安装·
go version
# 应该显示 go version go1.22.x 或更高版本
```

#### 方法 2: 从官网下载

1. 访问 [Go 官网](https://go.dev/dl/)
2. 下载 macOS 安装包（.pkg 文件）
3. 双击安装
4. 验证安装：`go version`

### 2. 配置 Go 环境变量（如果需要）

通常 Homebrew 安装的 Go 会自动配置，但如果没有，需要设置：

```bash
# 编辑 ~/.zshrc（Mac 默认使用 zsh）
nano ~/.zshrc

# 添加以下内容（如果不存在）
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# 保存后重新加载
source ~/.zshrc
```

### 2.5. 配置 Go 代理（重要！解决网络超时问题）

**如果遇到 `dial tcp ... i/o timeout` 错误，说明无法访问 Go 官方代理服务器。**

在中国大陆，建议使用国内 Go 代理镜像：

```bash
# 设置 Go 代理（使用 goproxy.cn，国内镜像）
go env -w GOPROXY=https://goproxy.cn,direct

# 验证配置
go env GOPROXY
# 应该显示：https://goproxy.cn,direct

# 如果需要永久生效，添加到 ~/.zshrc
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.zshrc
source ~/.zshrc
```

**其他可用的 Go 代理镜像：**
- `https://goproxy.cn` - 七牛云（推荐，速度快）
- `https://goproxy.io` - GoProxy.io
- `https://mirrors.aliyun.com/goproxy/` - 阿里云镜像

**配置完成后，重新尝试下载依赖：**
```bash
go mod download
```

### 3. 安装 Node.js 和 pnpm

```bash
# 安装 Node.js（如果还没有）
brew install node

# 验证安装
node --version
npm --version

# 安装 pnpm（项目使用 pnpm 作为包管理器）
npm install -g pnpm

# 验证安装
pnpm --version
```

### 4. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 验证安装
wails version
```

**注意**：如果提示 `wails: command not found`，需要确保 `$GOPATH/bin` 在 PATH 中：

```bash
# 检查 GOPATH
go env GOPATH

# 将 GOPATH/bin 添加到 PATH（如果还没有）
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# 再次验证
wails version
```

### 5. 安装 MySQL（可选）

如果不需要数据库功能，可以跳过这一步。

```bash
# 安装 MySQL
brew install mysql

# 启动 MySQL 服务
brew services start mysql

# 设置 root 密码（首次安装会提示）
mysql_secure_installation
```

## 运行项目

### 1. 克隆/进入项目目录

```bash
cd /Users/mac/frontend/code/go/wails-contract-warn
```

### 2. 安装 Go 依赖

```bash
go mod download
```

### 3. 安装前端依赖

```bash
cd frontend
pnpm install
cd ..
```

### 4. 配置数据库（可选）

如果使用数据库功能：

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE contract_warn CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行建表脚本
mysql -u root -p contract_warn < database/schema.sql
```

### 5. 启动开发模式

#### 方法 1: 使用开发脚本（推荐）

```bash
# 给脚本添加执行权限
chmod +x dev.sh

# 运行开发脚本
./dev.sh
```

#### 方法 2: 直接使用 wails 命令

```bash
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils
```

### 6. 构建生产版本（可选）

```bash
# 构建前端
cd frontend
pnpm run build
cd ..

# 构建应用（会生成 .app 文件）
wails build
```

构建完成后，可以在 `build/bin` 目录找到 `wails-contract-warn.app`。

## 常见问题

### 1. `go: command not found`

**原因**：Go 未安装或 PATH 未配置

**解决**：
- 检查是否安装：`which go`
- 如果已安装但找不到，检查 PATH：`echo $PATH`
- 确保 `/usr/local/go/bin` 或 `$GOPATH/bin` 在 PATH 中

### 2. `wails: command not found`

**原因**：Wails CLI 未安装或 `$GOPATH/bin` 不在 PATH 中

**解决**：
```bash
# 重新安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 确保 GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```

### 3. `pnpm: command not found`

**原因**：pnpm 未安装

**解决**：
```bash
npm install -g pnpm
```

### 4. 端口 34115 被占用

**原因**：前端开发服务器端口被占用

**解决**：
```bash
# 查找占用端口的进程
lsof -i :34115

# 杀死进程（替换 PID 为实际进程 ID）
kill -9 <PID>

# 或者修改端口（需要同时修改 wails.json 和 vite.config.js）
```

### 5. `go mod download` 超时错误

**错误信息**：
```
go: Get "https://proxy.golang.org/...": dial tcp ...:443: i/o timeout
```

**原因**：无法访问 Go 官方代理服务器（在中国大陆很常见）

**解决**：
```bash
# 配置使用国内 Go 代理镜像
go env -w GOPROXY=https://goproxy.cn,direct

# 验证配置
go env GOPROXY

# 重新下载依赖
go mod download
```

**永久配置**（推荐）：
```bash
# 添加到 ~/.zshrc
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.zshrc
source ~/.zshrc
```

### 6. MySQL 连接失败

**原因**：MySQL 未启动或配置错误

**解决**：
```bash
# 启动 MySQL
brew services start mysql

# 检查 MySQL 状态
brew services list

# 测试连接
mysql -u root -p
```

## 快速检查清单

运行项目前，确保以下命令都能正常执行：

```bash
# ✅ Go 版本检查
go version
# 应该显示 go1.22.x 或更高

# ✅ Node.js 版本检查
node --version
# 应该显示 v16.x 或更高

# ✅ pnpm 版本检查
pnpm --version

# ✅ Wails 版本检查
wails version

# ✅ Go 依赖检查
go mod download

# ✅ 前端依赖检查
cd frontend && pnpm install && cd ..
```

## 下一步

安装完成后，可以：

1. 运行 `./dev.sh` 启动开发模式
2. 查看 [README.md](README.md) 了解项目功能
3. 查看 [DEVELOPMENT.md](DEVELOPMENT.md) 了解开发指南

## 需要帮助？

如果遇到问题，可以：

1. 查看项目文档（README.md、DEVELOPMENT.md 等）
2. 检查终端错误信息
3. 查看 Wails 官方文档：https://wails.io/docs/
