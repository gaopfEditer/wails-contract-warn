# Vite HMR 测试脚本
Write-Host "=== Vite 文件监听和 HMR 测试 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 检查前端开发服务器是否运行
Write-Host "1. 检查前端开发服务器..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:34115" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    Write-Host "   ✅ 前端开发服务器正在运行" -ForegroundColor Green
} catch {
    Write-Host "   ❌ 前端开发服务器未运行" -ForegroundColor Red
    Write-Host "   请先运行: cd frontend && pnpm run dev" -ForegroundColor Yellow
    exit 1
}

# 2. 测试文件监听
Write-Host ""
Write-Host "2. 测试文件监听..." -ForegroundColor Yellow
Write-Host "   请在浏览器中打开: http://localhost:34115" -ForegroundColor Cyan
Write-Host "   然后修改 frontend/src/App.vue 文件" -ForegroundColor Cyan
Write-Host "   观察浏览器是否自动更新" -ForegroundColor Cyan
Write-Host ""
Write-Host "   按任意键继续测试..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

# 3. 检查 Vite 配置
Write-Host ""
Write-Host "3. 检查 Vite 配置..." -ForegroundColor Yellow
if (Test-Path "frontend/vite.config.js") {
    $config = Get-Content "frontend/vite.config.js" -Raw
    if ($config -match "usePolling.*true") {
        Write-Host "   ✅ 已启用文件轮询 (usePolling: true)" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  未启用文件轮询" -ForegroundColor Yellow
    }
    if ($config -match "interval.*\d+") {
        Write-Host "   ✅ 已配置轮询间隔" -ForegroundColor Green
    }
} else {
    Write-Host "   ❌ vite.config.js 不存在" -ForegroundColor Red
}

# 4. 检查端口
Write-Host ""
Write-Host "4. 检查端口 34115..." -ForegroundColor Yellow
$port = netstat -ano | findstr :34115
if ($port) {
    Write-Host "   ✅ 端口 34115 已被占用" -ForegroundColor Green
} else {
    Write-Host "   ❌ 端口 34115 未被占用" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 测试完成 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "如果浏览器中热更新不工作，请检查：" -ForegroundColor Yellow
Write-Host "1. Vite 终端是否有文件变化日志" -ForegroundColor White
Write-Host "2. 浏览器控制台是否有 HMR 连接信息" -ForegroundColor White
Write-Host "3. 防火墙是否阻止了端口 34115" -ForegroundColor White

