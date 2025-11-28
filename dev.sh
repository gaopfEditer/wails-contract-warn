#!/bin/bash
# Wails 开发模式启动脚本（Linux/Mac）
# 使用 -reloaddirs 参数启用文件监听和自动重新加载

echo "启动 Wails 开发模式..."
echo "前端开发服务器: http://localhost:34115"
echo "按 Ctrl+C 停止"
echo ""

# 清理可能存在的旧进程
pkill -f wails-contract-warn-dev 2>/dev/null
sleep 1

# 启动 Wails dev，指定需要监听的目录（只监听存在的目录）
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils

