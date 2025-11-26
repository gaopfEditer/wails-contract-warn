# 模块化架构说明

## 后端模块结构

### 📁 models/ - 数据模型层
- `kline.go` - K线数据、技术指标、预警信号的数据结构定义

### 📁 service/ - 业务服务层
- `market_service.go` - 市场数据服务（内存数据管理）
- `sync_service.go` - 数据同步服务（自动同步）

### 📁 indicator/ - 技术指标计算层
- `calculator.go` - 技术指标计算（MA、MACD、布林带）

### 📁 signal/ - 信号检测层
- `detector.go` - K线形态检测和信号识别（可扩展）

### 📁 database/ - 数据访问层
- `db.go` - 数据库操作（MySQL）
- `schema.sql` - 数据库表结构

### 📁 sync/ - 数据同步层
- `exchange.go` - 交易所API接口（数据拉取）

### 📁 utils/ - 工具层
- `aggregate.go` - K线聚合工具（多周期转换）

### 📁 config/ - 配置层
- `config.go` - 配置管理

### 📁 main/ - 控制器层
- `app.go` - Wails应用控制器（暴露给前端的方法）
- `main.go` - 应用入口

## 前端模块结构

### 📁 api/ - API调用层
- `market.js` - 市场数据API封装
- `database.js` - 数据库API封装

### 📁 composables/ - 组合式函数层
- `useMarketData.js` - 市场数据管理组合式函数

### 📁 components/ - 组件层
- `AppHeader.vue` - 应用头部组件
- `KLineChart.vue` - K线图组件
- `AlertBanner.vue` - 预警横幅组件

### 📁 utils/ - 工具层
- `indicators.js` - 指标计算工具（前端）
- `signalTypes.js` - 信号类型配置

### 📁 App.vue - 主组件
- 使用组合式函数管理状态
- 组合各个子组件

## 模块依赖关系

```
前端:
App.vue
  ├── composables/useMarketData.js
  │     ├── api/market.js
  │     └── utils/indicators.js
  ├── components/AppHeader.vue
  ├── components/KLineChart.vue
  │     └── utils/signalTypes.js
  └── components/AlertBanner.vue
        └── utils/signalTypes.js

后端:
main.go
  └── app.go (控制器)
        ├── service/market_service.go
        ├── service/sync_service.go
        ├── indicator/calculator.go
        ├── signal/detector.go
        ├── database/db.go
        ├── sync/exchange.go
        └── utils/aggregate.go
```

## 模块职责

### 后端

| 模块 | 职责 |
|------|------|
| `models/` | 定义数据结构，不包含业务逻辑 |
| `service/` | 业务逻辑服务，管理数据流和状态 |
| `indicator/` | 技术指标计算，纯函数 |
| `signal/` | 信号检测，可扩展的检测器 |
| `database/` | 数据持久化，数据库操作 |
| `sync/` | 外部数据同步，API调用 |
| `utils/` | 通用工具函数 |
| `app.go` | 控制器，连接前端和业务层 |

### 前端

| 模块 | 职责 |
|------|------|
| `api/` | API调用封装，统一错误处理 |
| `composables/` | 可复用的组合式函数，状态管理 |
| `components/` | UI组件，展示和交互 |
| `utils/` | 前端工具函数 |

## 扩展指南

### 添加新的技术指标

1. 在 `indicator/calculator.go` 中添加计算函数
2. 在 `models/kline.go` 的 `Indicators` 结构体中添加字段
3. 在 `CalculateIndicators` 中调用新函数

### 添加新的信号类型

1. 在 `signal/detector.go` 中添加形态检测函数
2. 在 `DetectAllSignals` 中注册新信号
3. 在前端 `utils/signalTypes.js` 中添加配置

### 添加新的API方法

1. 在 `app.go` 中添加方法
2. 在前端 `api/` 中添加对应的封装函数
3. 在 `composables/` 中使用（如需要）

## 优势

✅ **职责清晰**: 每个模块只负责一个功能领域  
✅ **易于测试**: 模块独立，便于单元测试  
✅ **易于扩展**: 新功能只需在对应模块添加  
✅ **易于维护**: 代码组织清晰，便于定位问题  
✅ **代码复用**: 工具函数和组合式函数可复用  

