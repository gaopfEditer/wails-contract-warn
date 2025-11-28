package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

// KLineData K线数据结构
type KLineData struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

func main() {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成 500 条 K 线数据
	klines := generateKLineData(500)

	// 转换为 JSON
	jsonData, err := json.MarshalIndent(klines, "", "  ")
	if err != nil {
		fmt.Printf("序列化 JSON 失败: %v\n", err)
		os.Exit(1)
	}

	// 确保 data 目录存在
	if err := os.MkdirAll("data", 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		os.Exit(1)
	}

	// 写入文件
	filename := "data/test1.json"
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功生成测试数据: %s (共 %d 条)\n", filename, len(klines))
}

// generateKLineData 生成 K 线数据
func generateKLineData(count int) []KLineData {
	klines := make([]KLineData, count)

	// 起始价格（BTC 价格范围）
	basePrice := 90000.0
	currentPrice := basePrice

	// 起始时间（从现在往前推）
	now := time.Now()
	startTime := now.Add(-time.Duration(count) * time.Minute)

	for i := 0; i < count; i++ {
		// 计算时间戳（毫秒）
		timestamp := startTime.Add(time.Duration(i) * time.Minute).UnixMilli()

		// 生成价格波动（随机游走 + 趋势）
		trend := math.Sin(float64(i)/50.0) * 2000  // 长期趋势
		volatility := (rand.Float64() - 0.5) * 500 // 随机波动
		priceChange := trend + volatility

		// 开盘价
		open := currentPrice

		// 收盘价（基于趋势和随机波动）
		close := open + priceChange

		// 最高价和最低价（确保 High >= max(Open, Close), Low <= min(Open, Close)）
		bodyRange := math.Abs(close - open)
		upperShadow := rand.Float64() * bodyRange * 0.5 // 上影线
		lowerShadow := rand.Float64() * bodyRange * 0.5 // 下影线

		high := math.Max(open, close) + upperShadow
		low := math.Min(open, close) - lowerShadow

		// 确保价格合理（不能为负）
		if low < basePrice*0.8 {
			low = basePrice*0.8 + rand.Float64()*100
		}
		if high > basePrice*1.2 {
			high = basePrice*1.2 - rand.Float64()*100
		}

		// 成交量（随机，但有一定相关性）
		volume := 1000 + rand.Float64()*2000 + math.Abs(priceChange)*10

		// 特殊处理：在某些位置生成锤子形态（用于测试预警）
		// 锤子形态特征：下影线很长，实体很小，上影线很短或没有
		if i%50 == 0 && i > 20 { // 每50根K线生成一个锤子形态
			// 确保有足够的数据计算布林带
			hammerClose := open + (rand.Float64()-0.3)*100                // 实体较小
			hammerHigh := math.Max(open, hammerClose) + rand.Float64()*50 // 很小的上影线
			hammerLow := math.Min(open, hammerClose) - rand.Float64()*500 // 很长的下影线（锤子特征）

			// 确保价格在合理范围
			if hammerLow < basePrice*0.85 {
				hammerLow = basePrice * 0.85
			}

			close = hammerClose
			high = hammerHigh
			low = hammerLow
		}

		// 特殊处理：在某些位置让价格接近或突破布林带上轨
		// 布林带上轨通常是 MA20 + 2*标准差
		// 我们让某些K线的收盘价接近或超过上轨
		if i%30 == 0 && i > 20 {
			// 模拟价格突破上轨的情况
			upperBandEstimate := basePrice * 1.05 // 估算的上轨位置
			if rand.Float64() > 0.5 {
				close = upperBandEstimate + rand.Float64()*500 // 突破上轨
				high = close + rand.Float64()*200
				low = close - rand.Float64()*300
			}
		}

		// 创建 K 线数据
		klines[i] = KLineData{
			Time:   timestamp,
			Open:   roundTo(open, 2),
			High:   roundTo(high, 2),
			Low:    roundTo(low, 2),
			Close:  roundTo(close, 2),
			Volume: roundTo(volume, 2),
		}

		// 更新当前价格
		currentPrice = close
	}

	return klines
}

// roundTo 四舍五入到指定小数位
func roundTo(val float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(val*multiplier) / multiplier
}
