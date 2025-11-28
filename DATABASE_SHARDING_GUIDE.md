# 数据库分表方案 - 按币种后缀分组

## 方案概述

根据配置文件中的币种，**按币种后缀（如 `_USDT`, `_BTC`）自动创建表**，将相同后缀的币种存储在同一个表中。

### 表命名规则

- **表名格式**：`klines_1m_{后缀}`
- **示例**：
  - `BTC_USDT` → `klines_1m_USDT`
  - `ETH_USDT` → `klines_1m_USDT`（与BTC_USDT在同一张表）
  - `BTC_BTC` → `klines_1m_BTC`

### 优势

✅ **数据分离**：不同后缀的币种数据完全独立，互不影响  
✅ **表数量可控**：通常只有几个后缀（USDT, BTC, ETH等），表数量少  
✅ **查询高效**：单表数据量适中，索引效率高  
✅ **自动管理**：根据配置文件自动创建表，无需手动维护

## 实现细节

### 1. 表结构

每个后缀表的结构相同：

```sql
CREATE TABLE klines_1m_USDT (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    open_time BIGINT NOT NULL,
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,
    close_time BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_symbol_time (symbol, open_time),
    INDEX idx_symbol_close_time (symbol, close_time),
    INDEX idx_close_time (close_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2. 自动创建表

系统会在以下时机自动创建表：

1. **应用启动时**：根据 `config/symbols.json` 中的币种，自动创建对应的后缀表
2. **首次保存数据时**：如果表不存在，会自动创建

### 3. 数据存储逻辑

```go
// 示例：保存 BTC_USDT 和 ETH_USDT 的数据
klines := []KLine1m{
    {Symbol: "BTC_USDT", ...},
    {Symbol: "ETH_USDT", ...},
}

// 系统会自动：
// 1. 提取后缀：USDT
// 2. 选择表：klines_1m_USDT
// 3. 批量插入到同一张表
SaveKLine1m(klines)
```

### 4. 查询逻辑

所有查询函数都会自动根据 `symbol` 的后缀选择正确的表：

```go
// 查询 BTC_USDT 的数据
// 自动从 klines_1m_USDT 表中查询
klines, err := GetKLines1m("BTC_USDT", startTime, endTime, 1000)
```

## 配置示例

### `config/symbols.json`

```json
{
  "hot_symbols": [
    {"symbol": "BTC_USDT", "enabled": true},
    {"symbol": "ETH_USDT", "enabled": true},
    {"symbol": "SOL_USDT", "enabled": true}
  ],
  "minor_symbols": [
    {"symbol": "BTC_BTC", "enabled": true},
    {"symbol": "ETH_ETH", "enabled": true}
  ]
}
```

**自动创建的表**：
- `klines_1m_USDT`（存储所有 `*_USDT` 币种）
- `klines_1m_BTC`（存储所有 `*_BTC` 币种）
- `klines_1m_ETH`（存储所有 `*_ETH` 币种）

## 数据量估算

### 单表数据量（以 USDT 为例）

假设配置了 20 个 `*_USDT` 币种：

- **单个币种 1 年数据**：525,600 条
- **20 个币种 1 年数据**：20 × 525,600 = 10,512,000 条
- **20 个币种 5 年数据**：20 × 2,628,000 = 52,560,000 条 ≈ 5GB

**结论**：单表数据量在可接受范围内，索引效率高。

## 性能对比

### 单表方案 vs 分表方案

| 指标 | 单表方案 | 分表方案（按后缀） |
|------|---------|-------------------|
| 表数量 | 1 张 | 3-5 张（通常） |
| 单表数据量 | 2.6亿条（100币种×5年） | 5000万条（20币种×5年） |
| 查询性能 | 较慢（索引大） | 快（索引小） |
| 写入性能 | 有锁竞争 | 无锁竞争（不同后缀） |
| 维护成本 | 低 | 低（自动管理） |

## 使用示例

### 1. 自动初始化

应用启动时会自动：

```go
// 1. 读取配置文件
allSymbols := config.GetAllEnabledSymbols()

// 2. 提取唯一后缀
suffixes := []string{"USDT", "BTC", "ETH"}

// 3. 为每个后缀创建表
for suffix := range suffixes {
    CreateTableForSuffix(suffix)
}
```

### 2. 保存数据

```go
klines := []KLine1m{
    {Symbol: "BTC_USDT", OpenTime: ..., ...},
    {Symbol: "ETH_USDT", OpenTime: ..., ...},
}

// 自动分组并保存到 klines_1m_USDT
err := SaveKLine1m(klines)
```

### 3. 查询数据

```go
// 自动从 klines_1m_USDT 查询
klines, err := GetKLines1m("BTC_USDT", startTime, endTime, 1000)
```

## 迁移说明

### 从单表迁移到分表

如果之前使用的是单表 `klines_1m`，需要迁移数据：

```sql
-- 1. 创建新表
-- （系统会自动创建）

-- 2. 迁移数据（按后缀分组）
INSERT INTO klines_1m_USDT 
SELECT * FROM klines_1m 
WHERE symbol LIKE '%_USDT';

INSERT INTO klines_1m_BTC 
SELECT * FROM klines_1m 
WHERE symbol LIKE '%_BTC';

-- 3. 验证数据
SELECT COUNT(*) FROM klines_1m_USDT;
SELECT COUNT(*) FROM klines_1m_BTC;

-- 4. 删除旧表（谨慎操作！）
-- DROP TABLE klines_1m;
```

## 注意事项

1. **后缀提取规则**：使用最后一个下划线后的部分作为后缀
   - `BTC_USDT` → `USDT` ✅
   - `BTC_USDT_PERP` → `PERP` ✅
   - `BTCUSDT` → 默认 `USDT`（无下划线时）

2. **表名限制**：MySQL 表名不能包含特殊字符，系统会自动处理

3. **自动创建**：表会在首次使用时自动创建，无需手动干预

4. **向后兼容**：所有现有的数据库操作函数都已更新，无需修改业务代码

## 监控建议

### 1. 检查表数量

```sql
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = DATABASE() 
  AND table_name LIKE 'klines_1m_%';
```

### 2. 检查各表数据量

```sql
SELECT 
    'klines_1m_USDT' as table_name,
    COUNT(*) as count
FROM klines_1m_USDT
UNION ALL
SELECT 
    'klines_1m_BTC' as table_name,
    COUNT(*) as count
FROM klines_1m_BTC;
```

### 3. 检查索引使用情况

```sql
SHOW INDEX FROM klines_1m_USDT;
```

## 总结

✅ **推荐使用**：按后缀分表方案  
✅ **自动管理**：根据配置自动创建表  
✅ **性能优秀**：单表数据量适中，查询快速  
✅ **易于维护**：表数量少，管理简单

