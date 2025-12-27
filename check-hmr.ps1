# 热更新诊断脚本
Write-Host "=== 热更新诊断 ===" -ForegroundColor Cyan
Write-Host ""

# 1. 检查前端开发服务器
Write-Host "1. 检查前端开发服务器..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:34115" -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
    Write-Host "   ✅ 前端开发服务器正在运行" -ForegroundColor Green
    Write-Host "   状态码: $($response.StatusCode)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ 前端开发服务器未运行" -ForegroundColor Red
    Write-Host "   请运行: cd frontend && pnpm run dev" -ForegroundColor Yellow
}

# 2. 检查端口占用
Write-Host ""
Write-Host "2. 检查端口 34115..." -ForegroundColor Yellow
$port = netstat -ano | findstr :34115
if ($port) {
    Write-Host "   ✅ 端口 34115 已被占用" -ForegroundColor Green
    Write-Host "   $port" -ForegroundColor Gray
} else {
    Write-Host "   ❌ 端口 34115 未被占用" -ForegroundColor Red
}

# 3. 检查环境变量
Write-Host ""
Write-Host "3. 检查环境变量..." -ForegroundColor Yellow
$wailsEnv = $env:WAILS_ENV
$dev = $env:DEV
if ($wailsEnv -eq "development" -or $dev -eq "true") {
    Write-Host "   ✅ 开发模式已启用" -ForegroundColor Green
    Write-Host "   WAILS_ENV: $wailsEnv" -ForegroundColor Gray
    Write-Host "   DEV: $dev" -ForegroundColor Gray
} else {
    Write-Host "   ⚠️  开发模式未启用" -ForegroundColor Yellow
    Write-Host "   请设置: `$env:WAILS_ENV = 'development'" -ForegroundColor Yellow
}

# 4. 检查配置文件
Write-Host ""
Write-Host "4. 检查配置文件..." -ForegroundColor Yellow
if (Test-Path "wails.json") {
    $wailsConfig = Get-Content "wails.json" | ConvertFrom-Json
    $devServer = $wailsConfig.frontend.devServer
    Write-Host "   ✅ wails.json 存在" -ForegroundColor Green
    Write-Host "   devServer: $devServer" -ForegroundColor Gray
} else {
    Write-Host "   ❌ wails.json 不存在" -ForegroundColor Red
}

if (Test-Path "frontend/vite.config.js") {
    Write-Host "   ✅ vite.config.js 存在" -ForegroundColor Green
} else {
    Write-Host "   ❌ vite.config.js 不存在" -ForegroundColor Red
}

# 5. 检查main.go配置
Write-Host ""
Write-Host "5. 检查 main.go 配置..." -ForegroundColor Yellow
if (Test-Path "main.go") {
    $mainGo = Get-Content "main.go" -Raw
    if ($mainGo -match "assetServer = nil") {
        Write-Host "   ✅ main.go 已配置为开发模式（assetServer = nil）" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  main.go 可能未正确配置开发模式" -ForegroundColor Yellow
    }
} else {
    Write-Host "   ❌ main.go 不存在" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== 诊断完成 ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "建议的启动步骤：" -ForegroundColor Yellow
Write-Host "1. 终端1: cd frontend && pnpm run dev" -ForegroundColor White
Write-Host "2. 等待看到 'VITE ready in xxx ms'" -ForegroundColor White
Write-Host "3. 终端2: .\dev.ps1" -ForegroundColor White
Write-Host ""
Write-Host "验证热更新：" -ForegroundColor Yellow
Write-Host "1. 修改 frontend/src/App.vue" -ForegroundColor White
Write-Host "2. 保存文件" -ForegroundColor White
Write-Host "3. 在浏览器中打开 http://localhost:34115 查看是否自动更新" -ForegroundColor White
Write-Host "4. 在 WebView 中按 F12，查看 Network 标签，确认连接到 http://localhost:34115" -ForegroundColor White

