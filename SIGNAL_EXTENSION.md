# 信号系统扩展指南

## 概述

本系统采用可扩展的架构设计，可以轻松添加新的K线形态信号检测。

## 已实现的信号类型

### 看涨信号
1. **布林带下轨十字星** (`bollinger_doji_bottom`)
2. **布林带下轨锤子** (`bollinger_hammer_bottom`)
3. **布林带下轨连续锤子** (`bollinger_consecutive_hammers`)
4. **布林带下轨看涨吞没** (`bollinger_bullish_engulfing`)

### 看跌信号
1. **布林带上轨吊颈** (`bollinger_hanging_man_top`)
2. **布林带上轨看跌吞没** (`bollinger_bearish_engulfing`)

## 如何添加新的信号类型

### 步骤 1: 在 Go 后端添加形态检测函数

在 `app.go` 中添加新的形态检测函数，例如：

```go
// isNewPattern 判断是否为新的K线形态
func isNewPattern(candle KLineData, prev KLineData) bool {
    // 实现你的形态检测逻辑
    // ...
    return true
}
```

### 步骤 2: 创建信号检测函数

在 `app.go` 中添加新的信号检测函数：

```go
// detectBollingerNewPattern 检测布林带附近的新形态
func detectBollingerNewPattern(data []KLineData, bands []struct {
    upper  float64
    middle float64
    lower  float64
}) []AlertSignal {
    var signals []AlertSignal
    tolerance := 0.01

    for i := range data {
        if i < 19 || bands[i].lower == 0 {
            continue
        }

        candle := data[i]
        if !isNewPattern(candle, data[i-1]) {
            continue
        }

        // 判断是否在布林带附近
        lower := bands[i].lower
        isNearLower := candle.Low <= lower*(1+tolerance)

        if isNearLower {
            signals = append(signals, AlertSignal{
                Index:     i,
                Time:      candle.Time,
                Price:     candle.Low,
                Close:     candle.Close,
                LowerBand: lower,
                Type:      "bollinger_new_pattern", // 新的信号类型
                Strength:  0.8, // 信号强度 0-1
            })
        }
    }

    return signals
}
```

### 步骤 3: 在 DetectAllSignals 中注册新信号

在 `DetectAllSignals` 函数中添加新信号的检测：

```go
func DetectAllSignals(data []KLineData) []AlertSignal {
    // ...
    
    // 添加新信号检测
    allSignals = append(allSignals, detectBollingerNewPattern(data, bands)...)
    
    return allSignals
}
```

### 步骤 4: 在前端注册新信号类型

在 `frontend/src/utils/signalTypes.js` 中添加新信号配置：

```javascript
export const SIGNAL_TYPES = {
  // ... 现有信号
  
  bollinger_new_pattern: {
    name: '布林带下轨新形态',
    icon: '🆕',
    color: '#your-color',
    bgColor: '#your-bg-color',
    borderColor: '#your-border-color',
    description: '新形态的描述',
  },
}
```

## 信号强度说明

- **0.9+**: 非常强的信号（如连续锤子）
- **0.8-0.89**: 强信号（如看涨吞没、锤子）
- **0.7-0.79**: 中等信号（如吊颈、十字星）
- **0.6-0.69**: 弱信号

## 周期切换

系统已自动支持周期切换时重新计算所有信号。当用户切换周期时：

1. `App.vue` 中的 `watch` 监听器会触发
2. 调用 `loadData()` 重新获取数据
3. Go 后端根据新的周期数据重新计算所有指标和信号
4. 前端自动更新显示

## 示例：添加"三只乌鸦"形态

### 1. Go 后端

```go
// isThreeBlackCrows 判断是否为三只乌鸦
func isThreeBlackCrows(data []KLineData, index int) bool {
    if index < 2 {
        return false
    }
    
    // 连续三根阴线，且每根都比前一根低
    for i := index - 2; i <= index; i++ {
        if data[i].Close >= data[i].Open {
            return false
        }
        if i > index-2 && data[i].Close >= data[i-1].Close {
            return false
        }
    }
    
    return true
}

// detectBollingerThreeBlackCrows 检测布林带上轨附近的三只乌鸦
func detectBollingerThreeBlackCrows(data []KLineData, bands []struct {
    upper  float64
    middle float64
    lower  float64
}) []AlertSignal {
    var signals []AlertSignal
    tolerance := 0.01

    for i := range data {
        if i < 19 || bands[i].upper == 0 {
            continue
        }

        if !isThreeBlackCrows(data, i) {
            continue
        }

        upper := bands[i].upper
        isNearUpper := data[i].High >= upper*(1-tolerance)

        if isNearUpper {
            signals = append(signals, AlertSignal{
                Index:     i,
                Time:      data[i].Time,
                Price:     data[i].High,
                Close:     data[i].Close,
                UpperBand: upper,
                Type:      "bollinger_three_black_crows",
                Strength:  0.85,
            })
        }
    }

    return signals
}
```

### 2. 在 DetectAllSignals 中注册

```go
allSignals = append(allSignals, detectBollingerThreeBlackCrows(data, bands)...)
```

### 3. 前端配置

```javascript
bollinger_three_black_crows: {
  name: '布林带上轨三只乌鸦',
  icon: '🐦🐦🐦',
  color: '#ef5350',
  bgColor: '#ffebee',
  borderColor: '#ef5350',
  description: '价格在布林带上轨附近出现三只乌鸦，强烈看跌',
},
```

## 注意事项

1. **性能**: 信号检测函数会在每次数据更新时执行，确保算法高效
2. **准确性**: 形态检测的阈值需要根据实际市场调整
3. **容差**: 布林带附近的容差（tolerance）可根据品种波动性调整
4. **测试**: 添加新信号后，建议用历史数据测试准确性

## 总结

通过以上步骤，你可以轻松添加任何新的K线形态信号。系统设计保证了：
- ✅ 代码结构清晰，易于扩展
- ✅ 前后端分离，职责明确
- ✅ 自动支持周期切换重新计算
- ✅ 统一的信号显示和提示机制

