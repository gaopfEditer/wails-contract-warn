# Gate.io 数据同步系统使用指南

## 概述

本系统使用 **Gate.io** 交易所的 API 获取实时行情数据，并实现了智能的优先级调度机制：

- ✅ **优先同步热门币种的近期数据**（昨日至今）
- ✅ **空闲时同步历史数据**（从2020年开始）
- ✅ **空闲时同步小币种数据**
- ✅ **配置化的币种列表**

## 配置文件

### `config/symbols.json`

币种配置文件，包含：

1. **hot_symbols**: 热门币种列表（优先同步）
2. **minor_symbols**: 小币种列表（空闲时同步）
3. **sync_config**: 同步配置参数

#### 配置示例

```json
{
  "hot_symbols": [
    {
      "symbol": "BTC_USDT",
      "priority": 1,
      "enabled": true,
      "description": "比特币"
    }
  ],
  "sync_config": {
    "priority_recent_days": 1,
    "historical_start_year": 2020,
    "batch_size": 1000,
    "request_interval_ms": 200,
    "idle_sync_enabled": true,
    "idle_check_interval_seconds": 60
  }
}
```

#### 配置说明

- `symbol`: Gate.io 格式的交易对（如 `BTC_USDT`）
- `priority`: 优先级（数字越小优先级越高）
- `enabled`: 是否启用
- `priority_recent_days`: 优先同步最近N天的数据（默认1天）
- `historical_start_year`: 历史数据起始年份（默认2020）
- `idle_sync_enabled`: 是否启用空闲同步
- `idle_check_interval_seconds`: 空闲同步检查间隔（秒）

## 工作流程

### 1. 优先同步（每60秒）

系统会优先同步热门币种的**近期数据**（昨日至今）：

1. 读取 `hot_symbols` 配置
2. 按优先级排序
3. 并发同步所有热门币种的近期数据
4. 如果币种没有数据，从昨日开始拉取
5. 如果币种已有数据，从最后一条的下一条开始拉取

### 2. 空闲同步（每5分钟）

系统在空闲时会同步：

1. **历史数据**：轮询所有币种，从2020年开始逐步拉取历史数据
2. **小币种数据**：同步小币种的近期数据

### 3. 数据存储

所有数据存储到 MySQL 数据库的 `klines_1m` 表中：

- 使用 `INSERT IGNORE` 避免重复数据
- 自动创建唯一索引 `(symbol, open_time)`
- 支持增量同步，只拉取新数据

## API 方法

### StartPrioritySync()

启动优先级同步服务（从配置文件读取币种）

```javascript
await window.go.main.App.StartPrioritySync()
```

### StartAutoSync(symbol, intervalSeconds)

启动自动同步服务（兼容旧接口）

```javascript
await window.go.main.App.StartAutoSync('BTC_USDT', 60)
```

### StopAutoSync(symbol)

停止自动同步服务

```javascript
await window.go.main.App.StopAutoSync('BTC_USDT')
```

## Gate.io API 说明

### 接口地址

```
GET https://api.gateio.ws/api/v4/spot/candlesticks
```

### 参数

- `currency_pair`: 交易对（如 `BTC_USDT`）
- `interval`: K线周期（`1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`）
- `from`: 起始时间（秒级时间戳）
- `to`: 结束时间（秒级时间戳）
- `limit`: 返回数量（最大1000）

### 返回格式

```json
[
  [timestamp, volume, close, high, low, open, base_volume],
  ...
]
```

## 币种格式说明

**重要**：Gate.io 使用下划线格式，如 `BTC_USDT`，而不是 `BTCUSDT`。

配置文件中的 `symbol` 字段必须使用 Gate.io 格式。

## 自动启动

应用启动时会自动：

1. 连接数据库
2. 创建表结构
3. 启动优先级同步服务
4. 开始同步热门币种的近期数据

## 日志查看

系统会记录详细的同步日志：

```
INFO  优先级同步服务已启动
INFO  开始优先同步 10 个热门币种的近期数据
DEBUG 优先同步币种成功: BTC_USDT
INFO  空闲同步: 同步币种 BTC_USDT 的历史数据（从 2020 年开始）
```

## 性能优化

1. **并发同步**：热门币种并发同步，提高效率
2. **批量拉取**：每次最多拉取1000根K线
3. **请求限流**：每次请求间隔200ms，避免触发API限制
4. **增量同步**：只拉取新数据，避免重复

## 故障处理

### 1. API 请求失败

系统会自动记录错误日志，并在下次同步时重试。

### 2. 数据库连接失败

系统会降级到内存模式，但不会自动同步数据。

### 3. 币种配置错误

检查 `config/symbols.json` 中的 `symbol` 格式是否正确（必须是 Gate.io 格式）。

## 扩展配置

### 添加新币种

在 `config/symbols.json` 中添加：

```json
{
  "symbol": "NEW_USDT",
  "priority": 21,
  "enabled": true,
  "description": "新币种"
}
```

添加到 `hot_symbols` 或 `minor_symbols` 数组中。

### 调整同步间隔

修改 `sync_config` 中的参数：

```json
{
  "sync_config": {
    "idle_check_interval_seconds": 300  // 空闲同步间隔改为5分钟
  }
}
```

## 注意事项

1. **币种格式**：必须使用 Gate.io 格式（`BTC_USDT`），不是 `BTCUSDT`
2. **时间戳**：Gate.io API 使用秒级时间戳，系统会自动转换
3. **API限制**：注意 Gate.io 的 API 调用频率限制
4. **数据量**：历史数据量较大，同步需要时间

## 监控建议

1. 定期检查日志，确认同步正常
2. 监控数据库存储空间
3. 检查同步状态表 `sync_status` 了解同步进度

