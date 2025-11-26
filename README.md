# Wails + Vue 合约行情分析应用

基于 Wails v2 + Vue 3 + ECharts 的合约行情分析工具，使用 Go+WebView 架构。

## 架构说明

这是 **Go + WebView** 模式，不是纯后台服务：
- **Go 后端**：处理行情数据获取、存储、技术指标计算
- **Vue 前端**：使用 ECharts 渲染 K 线图
- **WebView**：内嵌 Edge/WebKit，显示前端界面
- **通信**：通过 `wails.Invoke()` 调用 Go 函数
- **打包**：输出单个 `.exe` / `.dmg`，包含内嵌 WebView

## 功能特性

- ✅ K 线图展示（使用 ECharts）
- ✅ 实时行情数据更新
- ✅ 技术指标计算（MA5/MA10/MA20、MACD）
- ✅ 多周期切换（1分钟、5分钟、15分钟、1小时）
- ✅ 多交易对支持（BTC/USDT、ETH/USDT）

## 技术栈

- **后端**: Go 1.21+
- **前端**: Vue 3 + ECharts 5
- **框架**: Wails v2
- **构建工具**: Vite

## 开发环境要求

1. **Go** 1.21 或更高版本
2. **Node.js** 16+ 和 npm
3. **Wails CLI** v2

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 快速开始

### 1. 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 2. 开发模式运行

```bash
# 启动开发服务器（前端热重载 + Go 后端）
wails dev
```

### 3. 构建生产版本

```bash
# 构建单个可执行文件
wails build

# Windows 会生成 wails-contract-warn.exe
# macOS 会生成 wails-contract-warn.app
# Linux 会生成 wails-contract-warn
```

## 项目结构

```
wails-contract-warn/
├── main.go              # 应用入口
├── app.go               # 应用逻辑（行情服务、指标计算）
├── go.mod               # Go 模块文件
├── wails.json           # Wails 配置文件
├── frontend/            # 前端代码
│   ├── src/
│   │   ├── main.js      # Vue 入口
│   │   ├── App.vue      # 主组件
│   │   └── components/
│   │       └── KLineChart.vue  # K 线图组件
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
└── README.md
```

## 核心功能说明

### Go 后端

- **MarketService**: 市场数据服务，管理 K 线数据存储和更新
- **CalculateIndicators**: 计算技术指标（移动平均线、MACD）
- **GetMarketData**: 获取 K 线数据（通过 `wails.Invoke()` 调用）
- **GetIndicators**: 获取技术指标数据

### Vue 前端

- **App.vue**: 主界面，包含交易对选择、周期选择、实时数据开关
- **KLineChart.vue**: ECharts K 线图组件，支持：
  - 蜡烛图显示
  - 移动平均线（MA5/MA10/MA20）
  - MACD 指标
  - 成交量柱状图
  - 数据缩放和拖拽

## 前端调用后端方法

在 Vue 组件中，通过 `window.go.main.App` 调用 Go 方法：

```javascript
// 获取市场数据
const data = await window.go.main.App.GetMarketData('BTCUSDT', '1m')

// 获取技术指标
const indicators = await window.go.main.App.GetIndicators('BTCUSDT', '1m')

// 开始实时数据流
await window.go.main.App.StartMarketDataStream('BTCUSDT', '1m')

// 停止实时数据流
await window.go.main.App.StopMarketDataStream('BTCUSDT')
```

## 自定义和扩展

### 添加新的技术指标

在 `app.go` 的 `CalculateIndicators` 函数中添加计算逻辑。

### 连接真实行情 API

修改 `MarketService` 的 `updateData` 方法，替换为真实的 API 调用：

```go
func (m *MarketService) updateData() {
    // 调用真实 API 获取行情数据
    // 例如：币安、OKX、火币等交易所 API
}
```

### 添加更多交易对

在 `App.vue` 的 `selectedSymbol` 选项中添加更多交易对，并在 `MarketService` 中初始化对应数据。

## 注意事项

- 当前使用模拟数据，实际使用时需要连接真实的行情 API
- 数据存储使用内存，应用关闭后数据会丢失
- 可根据需要添加数据库持久化存储

## 许可证

MIT
