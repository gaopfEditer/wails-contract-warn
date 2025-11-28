# Wails 开发模式启动脚本（Windows PowerShell）
# 使用 -reloaddirs 参数启用文件监听和自动重新加载

Write-Host "启动 Wails 开发模式..." -ForegroundColor Green
Write-Host "前端开发服务器: http://localhost:34115" -ForegroundColor Cyan
Write-Host "按 Ctrl+C 停止" -ForegroundColor Yellow
Write-Host ""

# 清理可能存在的旧进程
$processes = Get-Process -Name "wails-contract-warn-dev" -ErrorAction SilentlyContinue
if ($processes) {
    Write-Host "正在停止旧进程..." -ForegroundColor Yellow
    $processes | Stop-Process -Force
    Start-Sleep -Seconds 1
}

# 检查前端开发服务器是否运行
Write-Host "检查前端开发服务器..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:34115" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ 前端开发服务器正在运行" -ForegroundColor Green
} catch {
    Write-Host "⚠️  前端开发服务器未运行，正在启动..." -ForegroundColor Yellow
    Write-Host "   请在另一个终端运行: cd frontend && pnpm run dev" -ForegroundColor Yellow
    Write-Host "   或者等待 Wails 自动启动..." -ForegroundColor Yellow
}

# 设置开发环境变量
$env:WAILS_ENV = "development"
$env:DEV = "true"

# 启动 Wails dev，指定需要监听的目录（只监听存在的目录）
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils

