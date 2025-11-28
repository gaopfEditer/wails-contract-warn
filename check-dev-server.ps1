# 检查开发服务器状态的脚本

Write-Host "检查前端开发服务器状态..." -ForegroundColor Cyan
Write-Host ""

# 检查端口是否被占用
$port = 34115
$listening = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue

if ($listening) {
    Write-Host "✅ 端口 $port 正在被使用" -ForegroundColor Green
    $process = Get-Process -Id $listening.OwningProcess -ErrorAction SilentlyContinue
    if ($process) {
        Write-Host "   进程: $($process.Name) (PID: $($process.Id))" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ 端口 $port 未被使用" -ForegroundColor Red
    Write-Host "   请先启动前端开发服务器: cd frontend && pnpm run dev" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "测试 HTTP 连接..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:$port" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    Write-Host "✅ HTTP 连接成功 (状态码: $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "❌ HTTP 连接失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   请确保前端开发服务器正在运行" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "检查 WebSocket (HMR)..." -ForegroundColor Cyan
Write-Host "   在浏览器中打开开发者工具 (F12)" -ForegroundColor Yellow
Write-Host "   查看 Console 标签，应该看到 WebSocket 连接信息" -ForegroundColor Yellow
Write-Host "   如果看到 'WebSocket connection failed'，检查防火墙设置" -ForegroundColor Yellow

