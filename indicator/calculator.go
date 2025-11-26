package indicator

import (
	"math"

	"wails-contract-warn/models"
)

// CalculateIndicators 计算技术指标
func CalculateIndicators(data []models.KLineData) models.Indicators {
	if len(data) == 0 {
		return models.Indicators{}
	}

	indicators := models.Indicators{
		MA5:      make([]float64, len(data)),
		MA10:     make([]float64, len(data)),
		MA20:     make([]float64, len(data)),
		MACD:     make([]float64, len(data)),
		Signal:   make([]float64, len(data)),
		Hist:     make([]float64, len(data)),
		BBUpper:  make([]float64, len(data)),
		BBMiddle: make([]float64, len(data)),
		BBLower:  make([]float64, len(data)),
	}

	// 计算移动平均线
	calculateMA(data, indicators)

	// 计算 MACD
	calculateMACD(data, indicators)

	// 计算布林带
	calculateBollingerBands(data, indicators)

	return indicators
}

// calculateMA 计算移动平均线
func calculateMA(data []models.KLineData, indicators models.Indicators) {
	for i := range data {
		if i >= 4 {
			sum := 0.0
			for j := i - 4; j <= i; j++ {
				sum += data[j].Close
			}
			indicators.MA5[i] = sum / 5
		}

		if i >= 9 {
			sum := 0.0
			for j := i - 9; j <= i; j++ {
				sum += data[j].Close
			}
			indicators.MA10[i] = sum / 10
		}

		if i >= 19 {
			sum := 0.0
			for j := i - 19; j <= i; j++ {
				sum += data[j].Close
			}
			indicators.MA20[i] = sum / 20
		}
	}
}

// calculateMACD 计算MACD
func calculateMACD(data []models.KLineData, indicators models.Indicators) {
	ema12 := make([]float64, len(data))
	ema26 := make([]float64, len(data))

	for i := range data {
		if i == 0 {
			ema12[i] = data[i].Close
			ema26[i] = data[i].Close
		} else {
			ema12[i] = ema12[i-1]*11/13 + data[i].Close*2/13
			ema26[i] = ema26[i-1]*25/27 + data[i].Close*2/27
		}

		if i >= 25 {
			indicators.MACD[i] = ema12[i] - ema26[i]
		}
	}

	// 计算信号线（MACD 的 9 日 EMA）
	for i := range indicators.MACD {
		if i == 26 {
			indicators.Signal[i] = indicators.MACD[i]
		} else if i > 26 {
			indicators.Signal[i] = indicators.Signal[i-1]*8/10 + indicators.MACD[i]*2/10
			indicators.Hist[i] = indicators.MACD[i] - indicators.Signal[i]
		}
	}
}

// calculateBollingerBands 计算布林带
func calculateBollingerBands(data []models.KLineData, indicators models.Indicators) {
	bbPeriod := 20
	bbMultiplier := 2.0

	for i := range data {
		if i < bbPeriod-1 {
			continue
		}

		// 计算 SMA
		sum := 0.0
		for j := i - bbPeriod + 1; j <= i; j++ {
			sum += data[j].Close
		}
		sma := sum / float64(bbPeriod)

		// 计算标准差
		variance := 0.0
		for j := i - bbPeriod + 1; j <= i; j++ {
			variance += math.Pow(data[j].Close-sma, 2)
		}
		stdDev := math.Sqrt(variance / float64(bbPeriod))

		indicators.BBMiddle[i] = sma
		indicators.BBUpper[i] = sma + bbMultiplier*stdDev
		indicators.BBLower[i] = sma - bbMultiplier*stdDev
	}
}

