# 快速开始 - Mac 环境

## 是的，需要先安装 Wails！

这个项目使用 **Wails v2** 框架，必须安装 Wails CLI 才能运行。

## 完整运行步骤

### 第一步：安装 Go（如果还没有）

```bash
# 检查是否已安装
go version

# 如果未安装，使用 Homebrew 安装
brew install go
```

### 第二步：配置 Go 代理（解决网络问题）

```bash
# 设置国内代理镜像（避免下载超时）
go env -w GOPROXY=https://goproxy.cn,direct

# 永久配置（可选）
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.zshrc
source ~/.zshrc
```

### 第三步：安装 Wails CLI

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 配置 PATH（如果提示找不到命令）
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# 验证安装
wails version
```

### 第四步：安装项目依赖

```bash
# 进入项目目录
cd /Users/mac/frontend/code/go/wails-contract-warn

# 安装 Go 依赖
go mod download

# 安装前端依赖
cd frontend
pnpm install  # 如果没有 pnpm，先运行: npm install -g pnpm
cd ..
```

### 第五步：运行项目

```bash
# 方法 1: 使用开发脚本（推荐）
chmod +x dev.sh
./dev.sh

# 方法 2: 直接使用 wails 命令
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils
```

## 一键安装脚本

如果你想一次性完成所有安装，可以运行：

```bash
#!/bin/bash
# 一键安装脚本

echo "🔍 检查 Go..."
if ! command -v go &> /dev/null; then
    echo "📦 安装 Go..."
    brew install go
else
    echo "✅ Go 已安装: $(go version)"
fi

echo "🔧 配置 Go 代理..."
go env -w GOPROXY=https://goproxy.cn,direct
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.zshrc

echo "📦 安装 Wails CLI..."
go install github.com/wailsapp/wails/v2/cmd/wails@latest

echo "🔧 配置 PATH..."
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

echo "✅ 安装完成！"
echo "📝 运行项目: ./dev.sh"
```

## 常见问题

### 1. `wails: command not found`

**解决**：
```bash
# 确保 GOPATH/bin 在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# 重新验证
wails version
```

### 2. `pnpm: command not found`

**解决**：
```bash
npm install -g pnpm
```

### 3. `go mod download` 超时

**解决**：
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

### 4. 端口 34115 被占用

**解决**：
```bash
# 查找并关闭占用端口的进程
lsof -i :34115
kill -9 <PID>
```

## 运行后的效果

运行 `./dev.sh` 后：
- ✅ 会自动启动前端开发服务器（Vite）
- ✅ 会自动启动 Go 后端
- ✅ 会打开一个桌面应用窗口（WebView）
- ✅ 支持前端热更新（修改前端代码自动刷新）
- ✅ 支持后端自动重新编译（修改 Go 代码自动重启）

## 下一步

- 查看 [README.md](README.md) 了解项目功能
- 查看 [MAC_SETUP_GUIDE.md](MAC_SETUP_GUIDE.md) 了解详细配置
- 查看 [DEVELOPMENT.md](DEVELOPMENT.md) 了解开发指南
