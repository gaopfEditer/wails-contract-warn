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

// ==================== 信号检测函数 ====================

// detectBollingerDojiBottom 检测布林带下轨 + 十字星
func detectBollingerDojiBottom(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	tolerance := 0.01
	dojiThreshold := 0.001

	for i := range data {
		if i < 19 || bands[i].lower == 0 {
			continue
		}

		candle := data[i]
		if !IsDoji(candle, dojiThreshold) {
			continue
		}

		lower := bands[i].lower
		isNearLower := candle.Low <= lower*(1+tolerance) ||
			candle.Close <= lower*(1+tolerance)

		if isNearLower {
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
func detectBollingerHammer(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	tolerance := 0.01

	for i := range data {
		if i < 19 || bands[i].lower == 0 {
			continue
		}

		candle := data[i]
		if !IsHammer(candle) {
			continue
		}

		lower := bands[i].lower
		isNearLower := candle.Low <= lower*(1+tolerance) ||
			candle.Close <= lower*(1+tolerance)

		if isNearLower {
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
func detectBollingerConsecutiveHammers(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	tolerance := 0.01
	consecutiveCount := 2

	for i := range data {
		if i < 19 || bands[i].lower == 0 {
			continue
		}

		if !IsConsecutiveHammers(data, i, consecutiveCount) {
			continue
		}

		candle := data[i]
		lower := bands[i].lower
		isNearLower := candle.Low <= lower*(1+tolerance) ||
			candle.Close <= lower*(1+tolerance)

		if isNearLower {
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
func detectBollingerHangingMan(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	tolerance := 0.01

	for i := range data {
		if i < 19 || bands[i].upper == 0 {
			continue
		}

		candle := data[i]
		if !IsHangingMan(candle) {
			continue
		}

		upper := bands[i].upper
		isNearUpper := candle.High >= upper*(1-tolerance) ||
			candle.Close >= upper*(1-tolerance)

		if isNearUpper {
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
func detectBollingerEngulfing(data []models.KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []models.AlertSignal {
	var signals []models.AlertSignal
	tolerance := 0.01

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
		if isBullish && bands[i].lower > 0 {
			lower := bands[i].lower
			isNearLower := curr.Low <= lower*(1+tolerance) ||
				prev.Low <= lower*(1+tolerance)

			if isNearLower {
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
		if !isBullish && bands[i].upper > 0 {
			upper := bands[i].upper
			isNearUpper := curr.High >= upper*(1-tolerance) ||
				prev.High >= upper*(1-tolerance)

			if isNearUpper {
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

