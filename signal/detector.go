package signal

import (
	"math"

	"wails-contract-warn/models"
)

// DetectAllSignals 检测所有信号（可扩展）
func DetectAllSignals(data []models.KLineData) []models.AlertSignal {
	if len(data) == 0 {
		return []models.AlertSignal{}
	}

	var allSignals []models.AlertSignal

	// 计算布林带
	bands := calculateBollingerBands(data, 20, 2.0)

	// 1. 布林带下轨 + 十字星
	allSignals = append(allSignals, detectBollingerDojiBottom(data, bands)...)

	// 2. 布林带下轨 + 锤子
	allSignals = append(allSignals, detectBollingerHammer(data, bands)...)

	// 3. 布林带下轨 + 连续锤子
	allSignals = append(allSignals, detectBollingerConsecutiveHammers(data, bands)...)

	// 4. 布林带上轨 + 吊颈
	allSignals = append(allSignals, detectBollingerHangingMan(data, bands)...)

	// 5. 吞没形态（结合布林带）
	allSignals = append(allSignals, detectBollingerEngulfing(data, bands)...)

	// 6. 组合强信号：在3-5个K线中出现多个锤子线或顶部针形
	allSignals = append(allSignals, detectStrongPatternGroup(data, bands)...)

	return allSignals
}

// calculateBollingerBands 计算布林带
func calculateBollingerBands(data []models.KLineData, period int, multiplier float64) []struct {
	upper  float64
	middle float64
	lower  float64
} {
	bands := make([]struct {
		upper  float64
		middle float64
		lower  float64
	}, len(data))

	for i := range data {
		if i < period-1 {
			continue
		}

		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += data[j].Close
		}
		sma := sum / float64(period)

		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			variance += math.Pow(data[j].Close-sma, 2)
		}
		stdDev := math.Sqrt(variance / float64(period))

		bands[i].middle = sma
		bands[i].upper = sma + multiplier*stdDev
		bands[i].lower = sma - multiplier*stdDev
	}

	return bands
}

// ==================== K线形态检测函数 ====================

// IsDoji 判断是否为十字星
func IsDoji(candle models.KLineData, threshold float64) bool {
	if candle.High == candle.Low {
		return false
	}

	body := math.Abs(candle.Close - candle.Open)
	range_ := candle.High - candle.Low

	// 实体很小，且有明显影线
	return range_ > 0 && body/candle.Open < threshold && range_ > body*2
}

// IsHammer 判断是否为锤子线（看涨）
func IsHammer(candle models.KLineData) bool {
	if candle.High == candle.Low {
		return false
	}

	body := math.Abs(candle.Close - candle.Open)
	range_ := candle.High - candle.Low
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)

	// 下影线至少是实体的2倍，上影线很小
	return range_ > 0 && lowerShadow >= body*2 && upperShadow <= body*0.3
}

// IsHangingMan 判断是否为吊颈线（看跌）
func IsHangingMan(candle models.KLineData) bool {
	if candle.High == candle.Low {
		return false
	}

	body := math.Abs(candle.Close - candle.Open)
	range_ := candle.High - candle.Low
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)

	// 下影线至少是实体的2倍，上影线很小（与锤子类似，但位置不同）
	return range_ > 0 && lowerShadow >= body*2 && upperShadow <= body*0.3
}

// IsEngulfing 判断是否为吞没形态
func IsEngulfing(prev, curr models.KLineData) (bool, bool) {
	// 第一个bool表示是否为吞没，第二个bool表示是否为看涨（true）或看跌（false）
	if prev.High == prev.Low || curr.High == curr.Low {
		return false, false
	}

	prevBody := math.Abs(prev.Close - prev.Open)
	currBody := math.Abs(curr.Close - curr.Open)

	// 当前K线实体必须大于前一根
	if currBody <= prevBody {
		return false, false
	}

	// 看涨吞没：前一根是阴线，当前是阳线，且当前实体完全包含前一根
	isBullish := prev.Close < prev.Open && curr.Close > curr.Open &&
		curr.Open < prev.Close && curr.Close > prev.Open

	// 看跌吞没：前一根是阳线，当前是阴线，且当前实体完全包含前一根
	isBearish := prev.Close > prev.Open && curr.Close < curr.Open &&
		curr.Open > prev.Close && curr.Close < prev.Open

	if isBullish {
		return true, true
	}
	if isBearish {
		return true, false
	}
	return false, false
}

// IsConsecutiveHammers 判断是否为连续锤子
func IsConsecutiveHammers(data []models.KLineData, index int, count int) bool {
	if index < count-1 {
		return false
	}

	// 检查最近count根K线是否都是锤子
	for i := index - count + 1; i <= index; i++ {
		if i < 0 || i >= len(data) {
			return false
		}
		if !IsHammer(data[i]) {
			return false
		}
	}
	return true
}

// IsTopPin 判断是否为顶部针形（上影线很长，下影线很短）
func IsTopPin(candle models.KLineData) bool {
	if candle.High == candle.Low {
		return false
	}

	body := math.Abs(candle.Close - candle.Open)
	range_ := candle.High - candle.Low
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)

	// 上影线至少是实体的2倍，下影线很小
	// 且上影线要足够长（至少是总波动的50%）
	return range_ > 0 && upperShadow >= body*2 && lowerShadow <= body*0.3 && upperShadow >= range_*0.5
}

// IsLongTopPin 判断是否为较长的顶部针形（上影线更长）
func IsLongTopPin(candle models.KLineData) bool {
	if candle.High == candle.Low {
		return false
	}

	body := math.Abs(candle.Close - candle.Open)
	range_ := candle.High - candle.Low
	lowerShadow := math.Min(candle.Open, candle.Close) - candle.Low
	upperShadow := candle.High - math.Max(candle.Open, candle.Close)

	// 上影线至少是实体的3倍，下影线很小
	// 且上影线要足够长（至少是总波动的60%）
	return range_ > 0 && upperShadow >= body*3 && lowerShadow <= body*0.2 && upperShadow >= range_*0.6
}

// ==================== 信号检测函数 ====================

// detectBollingerDojiBottom 检测布林带下轨 + 十字星
// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
func detectBollingerDojiBottom(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	dojiThreshold := 0.001
	bandToleranceRatio := 0.1 // 上下轨高度的10%

	for i := range data {
		if i < 19 || bands[i].lower == 0 || bands[i].upper == 0 {
			continue
		}

		candle := data[i]
		if !IsDoji(candle, dojiThreshold) {
			continue
		}

		lower := bands[i].lower
		upper := bands[i].upper
		bandHeight := upper - lower

		// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
		// 即：candle.Low - lower <= bandHeight * 0.1
		priceDiff := candle.Low - lower
		isAtLowerBand := priceDiff >= 0 && priceDiff <= bandHeight*bandToleranceRatio

		if isAtLowerBand {
			signals = append(signals, models.AlertSignal{
				Index:     i,
				Time:      candle.Time,
				Price:     candle.Low,
				Close:     candle.Close,
				LowerBand: lower,
				Type:      "bollinger_doji_bottom",
				Strength:  0.8,
			})
		}
	}

	return signals
}

// detectBollingerHammer 检测布林带下轨 + 锤子
// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
func detectBollingerHammer(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	bandToleranceRatio := 0.1 // 上下轨高度的10%

	for i := range data {
		if i < 19 || bands[i].lower == 0 || bands[i].upper == 0 {
			continue
		}

		candle := data[i]
		if !IsHammer(candle) {
			continue
		}

		lower := bands[i].lower
		upper := bands[i].upper
		bandHeight := upper - lower

		// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
		// 即：candle.Low - lower <= bandHeight * 0.1
		priceDiff := candle.Low - lower
		isAtLowerBand := priceDiff >= 0 && priceDiff <= bandHeight*bandToleranceRatio

		if isAtLowerBand {
			signals = append(signals, models.AlertSignal{
				Index:     i,
				Time:      candle.Time,
				Price:     candle.Low,
				Close:     candle.Close,
				LowerBand: lower,
				Type:      "bollinger_hammer_bottom",
				Strength:  0.85,
			})
		}
	}

	return signals
}

// detectBollingerConsecutiveHammers 检测布林带下轨 + 连续锤子
// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
func detectBollingerConsecutiveHammers(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	bandToleranceRatio := 0.1 // 上下轨高度的10%
	consecutiveCount := 2

	for i := range data {
		if i < 19 || bands[i].lower == 0 || bands[i].upper == 0 {
			continue
		}

		if !IsConsecutiveHammers(data, i, consecutiveCount) {
			continue
		}

		candle := data[i]
		lower := bands[i].lower
		upper := bands[i].upper
		bandHeight := upper - lower

		// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
		// 即：candle.Low - lower <= bandHeight * 0.1
		priceDiff := candle.Low - lower
		isAtLowerBand := priceDiff >= 0 && priceDiff <= bandHeight*bandToleranceRatio

		if isAtLowerBand {
			signals = append(signals, models.AlertSignal{
				Index:     i,
				Time:      candle.Time,
				Price:     candle.Low,
				Close:     candle.Close,
				LowerBand: lower,
				Type:      "bollinger_consecutive_hammers",
				Strength:  0.9,
			})
		}
	}

	return signals
}

// detectBollingerHangingMan 检测布林带上轨 + 吊颈
// 上轨信号：K线最高价与上轨价差 < 上下轨高度的10%
func detectBollingerHangingMan(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	bandToleranceRatio := 0.1 // 上下轨高度的10%

	for i := range data {
		if i < 19 || bands[i].upper == 0 || bands[i].lower == 0 {
			continue
		}

		candle := data[i]
		if !IsHangingMan(candle) {
			continue
		}

		upper := bands[i].upper
		lower := bands[i].lower
		bandHeight := upper - lower

		// 上轨信号：K线最高价与上轨价差 < 上下轨高度的10%
		// 即：upper - candle.High <= bandHeight * 0.1
		priceDiff := upper - candle.High
		isAtUpperBand := priceDiff >= 0 && priceDiff <= bandHeight*bandToleranceRatio

		if isAtUpperBand {
			signals = append(signals, models.AlertSignal{
				Index:     i,
				Time:      candle.Time,
				Price:     candle.High,
				Close:     candle.Close,
				UpperBand: upper,
				Type:      "bollinger_hanging_man_top",
				Strength:  0.75,
			})
		}
	}

	return signals
}

// detectBollingerEngulfing 检测布林带附近的吞没形态
// 下轨信号：K线最低价与下轨价差 < 上下轨高度的10%
// 上轨信号：K线最高价与上轨价差 < 上下轨高度的10%
func detectBollingerEngulfing(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	bandToleranceRatio := 0.1 // 上下轨高度的10%

	for i := 1; i < len(data); i++ {
		if i < 19 {
			continue
		}

		prev := data[i-1]
		curr := data[i]

		isEngulfing, isBullish := IsEngulfing(prev, curr)
		if !isEngulfing {
			continue
		}

		// 看涨吞没在下轨附近
		if isBullish && bands[i].lower > 0 && bands[i].upper > 0 {
			lower := bands[i].lower
			upper := bands[i].upper
			bandHeight := upper - lower

			// 检查当前K线或前一根K线是否在下轨附近
			currPriceDiff := curr.Low - lower
			prevPriceDiff := prev.Low - lower
			isAtLowerBand := (currPriceDiff >= 0 && currPriceDiff <= bandHeight*bandToleranceRatio) ||
				(prevPriceDiff >= 0 && prevPriceDiff <= bandHeight*bandToleranceRatio)

			if isAtLowerBand {
				signals = append(signals, models.AlertSignal{
					Index:     i,
					Time:      curr.Time,
					Price:     curr.Low,
					Close:     curr.Close,
					LowerBand: lower,
					Type:      "bollinger_bullish_engulfing",
					Strength:  0.88,
				})
			}
		}

		// 看跌吞没在上轨附近
		if !isBullish && bands[i].upper > 0 && bands[i].lower > 0 {
			upper := bands[i].upper
			lower := bands[i].lower
			bandHeight := upper - lower

			// 检查当前K线或前一根K线是否在上轨附近
			currPriceDiff := upper - curr.High
			prevPriceDiff := upper - prev.High
			isAtUpperBand := (currPriceDiff >= 0 && currPriceDiff <= bandHeight*bandToleranceRatio) ||
				(prevPriceDiff >= 0 && prevPriceDiff <= bandHeight*bandToleranceRatio)

			if isAtUpperBand {
				signals = append(signals, models.AlertSignal{
					Index:     i,
					Time:      curr.Time,
					Price:     curr.High,
					Close:     curr.Close,
					UpperBand: upper,
					Type:      "bollinger_bearish_engulfing",
					Strength:  0.88,
				})
			}
		}
	}

	return signals
}

// detectStrongPatternGroup 检测组合强信号
// 在3-5个K线中检测出现多个锤子线或较长的顶部针形，记为一组强信号
func detectStrongPatternGroup(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	minWindowSize := 3  // 最小窗口：3个K线
	maxWindowSize := 5  // 最大窗口：5个K线
	minPatternCount := 2 // 最少需要2个特定形态

	for i := maxWindowSize - 1; i < len(data); i++ {
		if i < 19 || bands[i].lower == 0 || bands[i].upper == 0 {
			continue
		}

		// 在3-5个K线窗口中检测
		for windowSize := minWindowSize; windowSize <= maxWindowSize; windowSize++ {
			if i < windowSize-1 {
				continue
			}

			// 统计窗口内的形态
			hammerCount := 0
			longTopPinCount := 0
			patternIndices := []int{} // 记录形态出现的位置

			// 检查窗口内的每个K线
			for j := i - windowSize + 1; j <= i; j++ {
				if j < 0 || j >= len(data) {
					continue
				}

				candle := data[j]
				if IsHammer(candle) {
					hammerCount++
					patternIndices = append(patternIndices, j)
				} else if IsLongTopPin(candle) {
					longTopPinCount++
					patternIndices = append(patternIndices, j)
				}
			}

			// 计算总形态数
			totalPatternCount := hammerCount + longTopPinCount

			// 如果满足条件：至少有minPatternCount个形态
			if totalPatternCount >= minPatternCount {
				// 确定信号类型和强度
				signalType := ""
				strength := 0.0
				price := 0.0
				lower := bands[i].lower
				upper := bands[i].upper

				// 根据形态组合确定信号类型
				if hammerCount >= 2 {
					// 多个锤子线：看涨信号
					signalType = "strong_hammer_group"
					strength = 0.92 + float64(hammerCount-2)*0.02 // 2个锤子0.92，3个0.94，4个0.96，5个0.98
					if strength > 0.98 {
						strength = 0.98
					}
					// 使用最后一个锤子线的价格
					lastCandle := data[i]
					price = lastCandle.Low
				} else if longTopPinCount >= 2 {
					// 多个顶部针形：看跌信号
					signalType = "strong_top_pin_group"
					strength = 0.90 + float64(longTopPinCount-2)*0.02 // 2个针形0.90，3个0.92，4个0.94，5个0.96
					if strength > 0.96 {
						strength = 0.96
					}
					// 使用最后一个针形的价格
					lastCandle := data[i]
					price = lastCandle.High
				} else if totalPatternCount >= 2 {
					// 混合形态：根据数量确定强度
					signalType = "strong_mixed_pattern_group"
					strength = 0.88 + float64(totalPatternCount-2)*0.02
					if strength > 0.94 {
						strength = 0.94
					}
					// 使用最后一个形态的价格
					lastCandle := data[i]
					if hammerCount > longTopPinCount {
						price = lastCandle.Low
					} else {
						price = lastCandle.High
					}
				}

				// 如果确定了信号类型，创建信号
				if signalType != "" {
					lastCandle := data[i]
					signals = append(signals, models.AlertSignal{
						Index:     i,
						Time:      lastCandle.Time,
						Price:     price,
						Close:     lastCandle.Close,
						LowerBand: lower,
						UpperBand: upper,
						Type:      signalType,
						Strength:  strength,
					})

					// 找到一个窗口后，不再检查更大的窗口（避免重复）
					break
				}
			}
		}
	}

	return signals
}

