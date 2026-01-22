#!/bin/bash
# Wails 开发模式启动脚本（Linux/Mac）
# 使用 -reloaddirs 参数启用文件监听和自动重新加载

echo "启动 Wails 开发模式..."
echo "前端开发服务器: http://localhost:34115"
echo "按 Ctrl+C 停止"
echo ""

# 清理可能存在的旧进程
pkill -f wails-contract-warn-dev 2>/dev/null
pkill -f "vite" 2>/dev/null
sleep 1

# 检查前端开发服务器是否在运行
check_dev_server() {
  if curl -s http://localhost:34115 > /dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

# 如果前端服务器未运行，提示用户手动启动
if ! check_dev_server; then
  echo "⚠️  前端开发服务器未运行"
  echo ""
  echo "Vite 开发服务器需要在单独的终端运行，请："
  echo "  1. 打开新的终端窗口"
  echo "  2. 运行: cd frontend && pnpm run dev"
  echo "  3. 等待看到 'VITE ready' 后，回到这里按回车继续"
  echo ""
  read -p "按回车继续启动 Wails..." -r
  echo ""
  
  # 再次检查
  if ! check_dev_server; then
    echo "❌ 前端开发服务器仍未运行！"
    echo "请确保在另一个终端运行: cd frontend && pnpm run dev"
    exit 1
  fi
  echo "✅ 前端开发服务器已就绪"
else
  echo "✅ 前端开发服务器已在运行"
fi

# 设置开发环境变量
export WAILS_ENV=development
export DEV=true

# 启动 Wails dev，指定需要监听的目录（只监听存在的目录）
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils

# 注意：Vite 进程在另一个终端运行，需要手动停止
