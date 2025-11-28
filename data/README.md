# 测试数据说明

## 文件说明

- `test1.json` - 包含 500 条 K 线测试数据
- `generate_test_data.go` - 生成测试数据的脚本

## 数据格式

每条 K 线数据包含以下字段：

```json
{
  "time": 1764269596311,    // 时间戳（毫秒）
  "open": 90000,             // 开盘价
  "high": 90045.19,          // 最高价
  "low": 89845.49,           // 最低价
  "close": 89891.94,         // 收盘价
  "volume": 3328.95          // 成交量
}
```

## 数据特点

1. **价格范围**: 基于 BTC 价格，约 85,000 - 100,000 USDT
2. **时间间隔**: 每条数据间隔 1 分钟
3. **特殊形态**: 
   - 每 50 根 K 线包含一个锤子形态（用于测试预警）
   - 每 30 根 K 线包含价格接近或突破布林带上轨的情况
4. **数据量**: 500 条，足够计算技术指标（需要至少 20 条数据计算 MA20 和布林带）

## 使用方法

### 1. 重新生成测试数据

```bash
go run data/generate_test_data.go
```

### 2. 在 Go 代码中加载测试数据

```go
// 加载测试数据
dataStr, err := app.LoadTestData("test1.json")
if err != nil {
    log.Fatal(err)
}

// 解析为 KLineData 数组
var klines []models.KLineData
json.Unmarshal([]byte(dataStr), &klines)
```

### 3. 测试布林带上轨+锤子形态预警

```go
// 测试预警
resultStr, err := app.TestBollingerHammerAlert("test1.json")
if err != nil {
    log.Fatal(err)
}

// 解析结果
var result map[string]interface{}
json.Unmarshal([]byte(resultStr), &result)
```

### 4. 在前端调用

```javascript
// 加载测试数据
const data = await window.go.main.App.LoadTestData('test1.json');
const klines = JSON.parse(data);

// 测试预警
const result = await window.go.main.App.TestBollingerHammerAlert('test1.json');
const analysis = JSON.parse(result);
console.log('检测到的信号:', analysis.signals);
```

## 数据验证

测试数据已经过验证：
- ✅ 格式正确，符合 KLineData 结构
- ✅ 价格数据合理（High >= max(Open, Close), Low <= min(Open, Close)）
- ✅ 包含锤子形态数据点
- ✅ 包含接近/突破布林带上轨的数据点
- ✅ 时间戳连续（每分钟一条）

## 注意事项

1. 测试数据是随机生成的，每次运行 `generate_test_data.go` 会生成不同的数据
2. 如果需要固定的测试数据，请保存生成的 `test1.json` 文件
3. 数据中的锤子形态和布林带上轨突破是模拟的，实际交易中需要更严格的判断条件

