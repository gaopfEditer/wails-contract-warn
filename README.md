# Wails + Vue 数据分析应用

基于 Wails v2 + Vue 3 + ECharts 的数据分析工具，使用 Go+WebView 架构。

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
- ✅ 技术指标计算（MA5/MA10/MA20、MACD、布林带）
- ✅ 多周期切换（1分钟、5分钟、15分钟、1小时）
- ✅ 多交易对支持（BTC/USDT、ETH/USDT）
- ✅ **K线形态信号检测**（十字星、锤子、吞没、吊颈等）
- ✅ **布林带预警系统**（下轨十字星、连续锤子等）
- ✅ **MySQL数据存储**（只存1分钟数据，动态聚合多周期）
- ✅ **增量数据同步**（自动从交易所拉取最新数据）

## 技术栈

- **后端**: Go 1.21+
- **前端**: Vue 3 + ECharts 5
- **框架**: Wails v2
- **构建工具**: Vite
- **数据库**: MySQL 8.0+

## 开发环境要求

1. **Go** 1.21 或更高版本
2. **Node.js** 16+ 和 npm
3. **Wails CLI** v2
4. **MySQL** 8.0+（可选，用于数据持久化）

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

### 2. 数据库设置（可选）

如果使用数据库存储功能：

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE contract_warn CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 执行建表脚本
mysql -u root -p contract_warn < database/schema.sql
```

### 3. 开发模式运行

```bash
# 启动开发服务器（前端热重载 + Go 后端）
wails dev
```

### 4. 构建生产版本

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
├── app.go               # 应用逻辑（行情服务、指标计算、信号检测）
├── go.mod               # Go 模块文件
├── wails.json           # Wails 配置文件
├── database/            # 数据库相关
│   ├── schema.sql       # 数据库表结构
│   └── db.go            # 数据库操作
├── sync/                # 数据同步
│   └── exchange.go      # 交易所API接口
├── utils/               # 工具函数
│   └── aggregate.go     # K线聚合函数
├── service/             # 后台服务
│   └── sync_service.go  # 自动同步服务
├── frontend/            # 前端代码
│   ├── src/
│   │   ├── main.js      # Vue 入口
│   │   ├── App.vue      # 主组件
│   │   ├── components/  # 组件
│   │   │   ├── AppHeader.vue
│   │   │   ├── KLineChart.vue
│   │   │   └── AlertBanner.vue
│   │   └── utils/       # 工具函数
│   │       ├── indicators.js
│   │       └── signalTypes.js
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
└── README.md
```

## 核心功能说明

### Go 后端

- **MarketService**: 市场数据服务，管理 K 线数据存储和更新
- **CalculateIndicators**: 计算技术指标（移动平均线、MACD、布林带）
- **DetectAllSignals**: 检测所有K线形态信号（可扩展）
- **数据库集成**: 支持MySQL存储，只存1分钟数据，动态聚合多周期
- **自动同步**: 后台服务自动从交易所拉取最新数据

### Vue 前端

- **App.vue**: 主界面，包含交易对选择、周期选择、实时数据开关
- **KLineChart.vue**: ECharts K 线图组件，支持：
  - 蜡烛图显示
  - 移动平均线（MA5/MA10/MA20）
  - 布林带（上轨/中轨/下轨）
  - MACD 指标
  - 成交量柱状图
  - 信号标记（不同图标和颜色）
  - 数据缩放和拖拽
- **AlertBanner.vue**: 预警提示组件，显示最新信号

## 数据库功能

### 核心策略

- ✅ **只存储1分钟K线数据**
- ✅ **其他周期通过聚合生成**
- ✅ **增量同步，节省API配额**
- ✅ **支持离线查看历史数据**

详细说明见 [README_DATABASE.md](README_DATABASE.md)

### 使用示例

```javascript
// 1. 初始化数据库
await window.go.main.App.InitDatabase('user:password@tcp(localhost:3306)/contract_warn?charset=utf8mb4&parseTime=True&loc=Local')

// 2. 首次同步历史数据
await window.go.main.App.SyncKlineDataInitial('BTCUSDT', 7)

// 3. 启动自动同步
await window.go.main.App.StartAutoSync('BTCUSDT', 60)

// 4. 获取任意周期数据（自动聚合）
const data = await window.go.main.App.GetMarketData('BTCUSDT', '30m')
```

## 信号系统

### 已实现的信号类型

**看涨信号**:
- 布林带下轨十字星
- 布林带下轨锤子
- 布林带下轨连续锤子
- 布林带下轨看涨吞没

**看跌信号**:
- 布林带上轨吊颈
- 布林带上轨看跌吞没

### 扩展新信号

系统采用可扩展架构，可以轻松添加新的K线形态信号。详细说明见 [SIGNAL_EXTENSION.md](SIGNAL_EXTENSION.md)

## 前端调用后端方法

在 Vue 组件中，通过 `window.go.main.App` 调用 Go 方法：

```javascript
// 获取市场数据
const data = await window.go.main.App.GetMarketData('BTCUSDT', '1m')

// 获取技术指标
const indicators = await window.go.main.App.GetIndicators('BTCUSDT', '1m')

// 获取预警信号
const signals = await window.go.main.App.GetAlertSignals('BTCUSDT', '1m')

// 开始实时数据流
await window.go.main.App.StartMarketDataStream('BTCUSDT', '1m')

// 停止实时数据流
await window.go.main.App.StopMarketDataStream('BTCUSDT')
```

## 自定义和扩展

### 添加新的技术指标

在 `app.go` 的 `CalculateIndicators` 函数中添加计算逻辑。

### 连接真实行情 API

修改 `sync/exchange.go` 中的 `ExchangeAPI`，替换为真实的 API 调用。

### 添加更多交易对

在 `App.vue` 的 `selectedSymbol` 选项中添加更多交易对。

## 注意事项

- 当前使用模拟数据，实际使用时需要连接真实的行情 API
- 数据库功能为可选，不配置数据库时使用内存模式
- 数据存储使用MySQL，应用关闭后数据会持久化保存
- 建议定期备份数据库

## 许可证

MIT
