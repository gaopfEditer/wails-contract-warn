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

	// 生成 test1.json（已有，跳过）
	// generateTestData("test1.json", 500, 90000.0, 0)

	// 生成 test2.json - 包含更多布林带上轨+锤子形态
	generateTestData("test2.json", 500, 95000.0, 1)

	// 生成 test3.json - 包含更多布林带下轨+锤子形态
	generateTestData("test3.json", 500, 85000.0, 2)

	fmt.Println("所有测试数据生成完成！")
}

// generateTestData 生成测试数据
// filename: 文件名
// count: K线数量
// basePrice: 基础价格
// patternType: 0=混合, 1=上轨+锤子, 2=下轨+锤子
func generateTestData(filename string, count int, basePrice float64, patternType int) {
	klines := generateKLineData(count, basePrice, patternType)

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
	filePath := fmt.Sprintf("data/%s", filename)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("成功生成测试数据: %s (共 %d 条)\n", filePath, len(klines))
}

// generateKLineData 生成 K 线数据
func generateKLineData(count int, basePrice float64, patternType int) []KLineData {
	klines := make([]KLineData, count)
	currentPrice := basePrice

	// 起始时间（从现在往前推）
	now := time.Now()
	startTime := now.Add(-time.Duration(count) * time.Minute)

	for i := 0; i < count; i++ {
		// 计算时间戳（毫秒）
		timestamp := startTime.Add(time.Duration(i) * time.Minute).UnixMilli()

		// 生成价格波动
		trend := math.Sin(float64(i)/50.0) * 2000
		volatility := (rand.Float64() - 0.5) * 500
		priceChange := trend + volatility

		open := currentPrice
		close := open + priceChange

		bodyRange := math.Abs(close - open)
		upperShadow := rand.Float64() * bodyRange * 0.5
		lowerShadow := rand.Float64() * bodyRange * 0.5

		high := math.Max(open, close) + upperShadow
		low := math.Min(open, close) - lowerShadow

		// 确保价格合理
		if low < basePrice*0.8 {
			low = basePrice*0.8 + rand.Float64()*100
		}
		if high > basePrice*1.2 {
			high = basePrice*1.2 - rand.Float64()*100
		}

		volume := 1000 + rand.Float64()*2000 + math.Abs(priceChange)*10

		// 根据 patternType 生成不同的形态
		switch patternType {
		case 1: // test2: 布林带上轨+锤子形态
			if i%40 == 0 && i > 20 {
				// 生成上轨附近的锤子形态
				upperBandEstimate := basePrice * 1.08
				hammerClose := upperBandEstimate + (rand.Float64()-0.3)*200
				hammerHigh := math.Max(open, hammerClose) + rand.Float64()*50
				hammerLow := math.Min(open, hammerClose) - rand.Float64()*400 // 长下影线

				if hammerLow < basePrice*0.9 {
					hammerLow = basePrice * 0.9
				}

				close = hammerClose
				high = hammerHigh
				low = hammerLow
			}
		case 2: // test3: 布林带下轨+锤子形态
			if i%40 == 0 && i > 20 {
				// 生成下轨附近的锤子形态
				lowerBandEstimate := basePrice * 0.92
				hammerClose := lowerBandEstimate + (rand.Float64()-0.2)*150
				hammerHigh := math.Max(open, hammerClose) + rand.Float64()*50
				hammerLow := math.Min(open, hammerClose) - rand.Float64()*400 // 长下影线

				if hammerLow < basePrice*0.85 {
					hammerLow = basePrice * 0.85
				}

				close = hammerClose
				high = hammerHigh
				low = hammerLow
			}
		default: // test1: 混合形态
			if i%50 == 0 && i > 20 {
				hammerClose := open + (rand.Float64()-0.3)*100
				hammerHigh := math.Max(open, hammerClose) + rand.Float64()*50
				hammerLow := math.Min(open, hammerClose) - rand.Float64()*500

				if hammerLow < basePrice*0.85 {
					hammerLow = basePrice * 0.85
				}

				close = hammerClose
				high = hammerHigh
				low = hammerLow
			}
		}

		klines[i] = KLineData{
			Time:   timestamp,
			Open:   roundTo(open, 2),
			High:   roundTo(high, 2),
			Low:    roundTo(low, 2),
			Close:  roundTo(close, 2),
			Volume: roundTo(volume, 2),
		}

		currentPrice = close
	}

	return klines
}

// roundTo 四舍五入到指定小数位
func roundTo(val float64, precision int) float64 {
	multiplier := math.Pow(10, float64(precision))
	return math.Round(val*multiplier) / multiplier
}
