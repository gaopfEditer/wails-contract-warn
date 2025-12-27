package utils

import (
	"wails-contract-warn/database"
)

// KLine K线数据结构（用于聚合）
type KLine struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

// ParseIntervalToMinutes 将周期字符串转换为分钟数
func ParseIntervalToMinutes(interval string) int {
	switch interval {
	case "1m":
		return 1
	case "5m":
		return 5
	case "15m":
		return 15
	case "30m":
		return 30
	case "1h":
		return 60
	case "2h":
		return 120
	case "3h":
		return 180
	case "4h":
		return 240
	case "1d":
		return 1440 // 日线：24小时 = 1440分钟
	case "1w":
		return 10080 // 周线：7天 = 10080分钟
	case "1M":
		return 43200 // 月线：30天 = 43200分钟（简化处理，实际月份天数不同）
	default:
		// 尝试解析数字+m/h/d格式
		// 这里简化处理，实际可以更复杂
		return 1
	}
}

// AggregateKlines 将1分钟K线聚合为指定周期
func AggregateKlines(klines1m []database.KLine1m, targetIntervalMin int) []KLine {
	if len(klines1m) == 0 {
		return []KLine{}
	}

	if targetIntervalMin == 1 {
		// 直接转换，无需聚合
		result := make([]KLine, len(klines1m))
		for i, k := range klines1m {
			result[i] = KLine{
				OpenTime:  k.OpenTime,
				Open:      k.Open,
				High:      k.High,
				Low:       k.Low,
				Close:     k.Close,
				Volume:    k.Volume,
				CloseTime: k.CloseTime,
			}
		}
		return result
	}

	intervalMs := int64(targetIntervalMin * 60 * 1000)
	var result []KLine
	var group []database.KLine1m

	for i, k := range klines1m {
		// 计算当前K线所属的周期起始时间
		periodStart := (k.OpenTime / intervalMs) * intervalMs

		// 判断是否开始新的周期
		if i == 0 {
			group = []database.KLine1m{k}
		} else {
			prevPeriodStart := (klines1m[i-1].OpenTime / intervalMs) * intervalMs
			if periodStart != prevPeriodStart {
				// 新周期开始，处理上一组
				if len(group) > 0 {
					result = append(result, mergeKlines(group))
					group = nil
				}
			}
			group = append(group, k)
		}
	}

	// 处理最后一组
	if len(group) > 0 {
		result = append(result, mergeKlines(group))
	}

	return result
}

// mergeKlines 合并一组K线（如5根1m → 1根5m）
func mergeKlines(group []database.KLine1m) KLine {
	if len(group) == 0 {
		return KLine{}
	}

	first := group[0]
	last := group[len(group)-1]

	high := first.High
	low := first.Low
	volume := 0.0

	for _, k := range group {
		if k.High > high {
			high = k.High
		}
		if k.Low < low {
			low = k.Low
		}
		volume += k.Volume
	}

	// 计算周期结束时间（下一个周期的开始时间 - 1ms）
	intervalMs := int64((last.CloseTime - first.OpenTime + 1) / int64(len(group)) * int64(len(group)))
	closeTime := first.OpenTime + intervalMs - 1

	return KLine{
		OpenTime:  first.OpenTime, // 保留第一个的开盘时间
		Open:      first.Open,
		High:      high,
		Low:       low,
		Close:     last.Close, // 保留最后一个的收盘价
		Volume:    volume,
		CloseTime: closeTime,
	}
}

// CalculateNeeded1mCount 计算需要多少根1分钟K线才能生成指定数量的目标周期K线
func CalculateNeeded1mCount(targetCount int, targetIntervalMin int) int {
	return targetCount * targetIntervalMin
}

// GetKLineTimeRange 根据目标周期和数量，计算需要的1分钟K线时间范围
// 注意：需要导入time包才能使用
// func GetKLineTimeRange(targetIntervalMin int, targetCount int) (startTime, endTime int64) {
// 	now := time.Now()
// 	intervalMs := int64(targetIntervalMin * 60 * 1000)
//
// 	// 计算需要的时间跨度
// 	totalMs := int64(targetCount) * intervalMs
//
// 	endTime = now.UnixMilli()
// 	startTime = endTime - totalMs
//
// 	return startTime, endTime
// }
