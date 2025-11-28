@echo off
REM Wails 开发模式启动脚本（Windows CMD）
REM 使用 -reloaddirs 参数启用文件监听和自动重新加载

echo 启动 Wails 开发模式...
echo 前端开发服务器: http://localhost:34115
echo 按 Ctrl+C 停止
echo.

REM 清理可能存在的旧进程
taskkill /F /IM wails-contract-warn-dev.exe 2>nul
timeout /t 1 /nobreak >nul

REM 检查前端开发服务器是否运行
echo 检查前端开发服务器...
curl -s http://localhost:34115 >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] 前端开发服务器正在运行
) else (
    echo [WARN] 前端开发服务器未运行
    echo 请在另一个终端运行: cd frontend ^&^& pnpm run dev
)

REM 设置开发环境变量
set WAILS_ENV=development
set DEV=true

REM 启动 Wails dev，指定需要监听的目录（只监听存在的目录）
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils

