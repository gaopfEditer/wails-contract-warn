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
$devServerRunning = $false
try {
    $response = Invoke-WebRequest -Uri "http://localhost:34115" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ 前端开发服务器正在运行" -ForegroundColor Green
    $devServerRunning = $true
} catch {
    Write-Host "⚠️  前端开发服务器未运行" -ForegroundColor Yellow
    Write-Host "   正在启动前端开发服务器..." -ForegroundColor Cyan
    Write-Host ""
    Write-Host "   请在新的终端窗口中运行以下命令：" -ForegroundColor Yellow
    Write-Host "   cd frontend" -ForegroundColor White
    Write-Host "   pnpm run dev" -ForegroundColor White
    Write-Host ""
    Write-Host "   等待看到 'VITE ready' 后，再按任意键继续启动 Wails..." -ForegroundColor Yellow
    Write-Host ""
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}

if (-not $devServerRunning) {
    # 再次检查
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:34115" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
        Write-Host "✅ 前端开发服务器现在正在运行" -ForegroundColor Green
    } catch {
        Write-Host "❌ 前端开发服务器仍未运行，Wails 将使用嵌入的静态文件" -ForegroundColor Red
        Write-Host "   热更新将不可用！" -ForegroundColor Red
    }
}

# 设置开发环境变量
$env:WAILS_ENV = "development"
$env:DEV = "true"

# 等待开发服务器启动（如果还没启动）
Write-Host "等待开发服务器就绪..." -ForegroundColor Cyan
$maxRetries = 15
$retryCount = 0
$serverReady = $false

while ($retryCount -lt $maxRetries -and -not $serverReady) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:34115" -TimeoutSec 1 -UseBasicParsing -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            $serverReady = $true
            Write-Host "✅ 开发服务器已就绪 (状态码: $($response.StatusCode))" -ForegroundColor Green
        }
    } catch {
        $retryCount++
        if ($retryCount -lt $maxRetries) {
            Write-Host "等待开发服务器... ($retryCount/$maxRetries)" -ForegroundColor Yellow
            Start-Sleep -Seconds 1
        }
    }
}

if (-not $serverReady) {
    Write-Host "❌ 开发服务器未启动！" -ForegroundColor Red
    Write-Host "请先运行: cd frontend && pnpm run dev" -ForegroundColor Yellow
    Write-Host "然后等待看到 'VITE ready' 后再运行此脚本" -ForegroundColor Yellow
    exit 1
}

# 启动 Wails dev，指定需要监听的目录（只监听存在的目录）
wails dev -reloaddirs=.,./api,./config,./database,./indicator,./logger,./models,./service,./signal,./sync,./utils

