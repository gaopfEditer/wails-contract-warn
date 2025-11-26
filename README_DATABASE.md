# 数据库集成说明

## 概述

本系统采用 **"只存1分钟数据 + 动态聚合"** 的策略，支持多周期K线数据查询。

## 数据库设计

### 表结构

1. **klines_1m** - 存储1分钟K线原始数据
2. **sync_status** - 存储数据同步状态

详细SQL见 `database/schema.sql`

## 快速开始

### 1. 创建数据库

```sql
CREATE DATABASE IF NOT EXISTS contract_warn CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE contract_warn;
```

### 2. 执行建表脚本

```bash
mysql -u your_user -p contract_warn < database/schema.sql
```

### 3. 配置数据库连接

在应用启动时，通过前端调用初始化数据库：

```javascript
// 在 Vue 中
await window.go.main.App.InitDatabase('user:password@tcp(localhost:3306)/contract_warn?charset=utf8mb4&parseTime=True&loc=Local')
```

或者设置环境变量：

```bash
export DB_DSN="user:password@tcp(localhost:3306)/contract_warn?charset=utf8mb4&parseTime=True&loc=Local"
```

### 4. 首次同步数据

```javascript
// 同步最近7天的历史数据
await window.go.main.App.SyncKlineDataInitial('BTCUSDT', 7)
```

### 5. 定期增量同步

```javascript
// 增量同步（只拉取新数据）
await window.go.main.App.SyncKlineData('BTCUSDT')
```

建议每分钟执行一次增量同步。

## 工作原理

### 数据存储

- ✅ **只存储1分钟K线数据**到MySQL
- ✅ 其他周期（5m, 30m, 1h, 1d）通过聚合生成
- ✅ 存储空间：5年BTC数据约40MB

### 数据聚合

```go
// 前端请求30分钟K线
GetMarketData("BTCUSDT", "30m")

// 后端流程：
// 1. 从数据库读取1分钟数据
// 2. 聚合为30分钟K线
// 3. 返回给前端
```

### 增量同步

- 每次只拉取本地最新K线之后的数据
- 避免重复拉取，节省API配额
- 支持离线查看历史数据

## API 方法

### InitDatabase(dsn string)
初始化数据库连接

### SyncKlineDataInitial(symbol string, days int)
首次同步，拉取指定天数的历史数据

### SyncKlineData(symbol string)
增量同步，只拉取新数据

### GetMarketData(symbol string, period string)
获取市场数据（自动从数据库读取并聚合）

## 性能优化

### 1. 按需加载

系统会根据目标周期和数量，只加载必要的1分钟数据：

```go
// 请求100根1小时K线
// 只需加载 100 * 60 = 6000 根1分钟K线
```

### 2. 缓存策略（可选）

可以添加内存缓存，缓存常用周期的聚合结果：

```go
// 缓存5分钟、15分钟、1小时的聚合结果
// 新数据到来时增量更新缓存
```

## 注意事项

1. **API限制**: Binance等交易所对K线接口有频率限制，避免过于频繁请求
2. **数据完整性**: 首次同步建议拉取足够的历史数据（至少7天）
3. **定期同步**: 建议每分钟执行一次增量同步
4. **错误处理**: 网络错误时，系统会使用本地缓存数据

## 存储估算

| 交易对 | 1分钟数据量/年 | 5年数据量 |
|--------|---------------|----------|
| BTC/USDT | ~52万根 | ~260万根 ≈ 400MB |
| ETH/USDT | ~52万根 | ~260万根 ≈ 400MB |

**结论**: 存储成本极低，完全可接受。

## 故障恢复

如果数据库连接失败，系统会自动降级到内存模式（使用模拟数据），确保应用可用。

