package sync

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"wails-contract-warn/api"
	"wails-contract-warn/database"
	"wails-contract-warn/logger"
)

// syncRecord 同步记录（记录本次运行已拉取的币种和时间范围）
type syncRecord struct {
	symbol    string
	startTime int64
	endTime   int64
	syncTime  time.Time // 同步时间
}

// syncRecorder 同步记录器（全局单例）
type syncRecorder struct {
	mu      sync.RWMutex
	records map[string][]syncRecord // key: symbol, value: 该币种的同步记录列表
}

var globalSyncRecorder = &syncRecorder{
	records: make(map[string][]syncRecord),
}

// hasSynced 检查指定币种在指定时间范围内是否已经同步过
func (sr *syncRecorder) hasSynced(symbol string, startTime, endTime int64) bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	records, exists := sr.records[symbol]
	if !exists {
		return false
	}

	// 检查是否有记录覆盖了当前时间范围
	for _, record := range records {
		// 如果已同步的时间范围覆盖了当前时间范围，则认为已同步
		if record.startTime <= startTime && record.endTime >= endTime {
			return true
		}
		// 如果当前时间范围与已同步的时间范围有重叠，也认为已同步（避免重复拉取）
		if !(record.endTime < startTime || record.startTime > endTime) {
			// 有重叠，检查重叠比例
			overlapStart := max(record.startTime, startTime)
			overlapEnd := min(record.endTime, endTime)
			overlapDuration := overlapEnd - overlapStart
			currentDuration := endTime - startTime
			// 如果重叠超过80%，认为已同步
			if currentDuration > 0 && float64(overlapDuration)/float64(currentDuration) > 0.8 {
				return true
			}
		}
	}

	return false
}

// recordSync 记录同步操作
func (sr *syncRecorder) recordSync(symbol string, startTime, endTime int64) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	records := sr.records[symbol]
	records = append(records, syncRecord{
		symbol:    symbol,
		startTime: startTime,
		endTime:   endTime,
		syncTime:  time.Now(),
	})
	sr.records[symbol] = records

	logger.Debugf("[%s] 记录同步: %s ~ %s", symbol,
		time.Unix(startTime/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(endTime/1000, 0).Format("2006-01-02 15:04:05"))
}

// clearRecords 清空指定币种的记录（可选，用于重置）
func (sr *syncRecorder) clearRecords(symbol string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	delete(sr.records, symbol)
}

// clearAllRecords 清空所有记录
func (sr *syncRecorder) clearAllRecords() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.records = make(map[string][]syncRecord)
}

// max 返回两个int64中的较大值
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// min 返回两个int64中的较小值
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// SyncSymbolByWeeks 按周分批获取历史数据，按天检查状态
// 从当日向前，先获取一周，再获取一周，按天记录状态
// weeks: 要获取的周数（默认1周）
// 使用 SyncSymbolWithPriority 函数，但按周分批调用，按天检查状态
func SyncSymbolByWeeks(symbol string, weeks int) error {
	if weeks <= 0 {
		weeks = 1 // 默认1周
	}

	now := time.Now().UnixMilli()
	tenMinutesAgo := now - int64(10*60*1000) // 10分钟前

	// 计算目标时间范围（从当日向前weeks周）
	today := time.Now().UTC()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
	targetStart := todayStart - int64(weeks*7*24*60*60*1000) // weeks周前
	targetEnd := tenMinutesAgo

	logger.Infof("开始按周分批同步 %s 的历史数据: 从 %s 到 %s (共 %d 周)", 
		symbol,
		time.Unix(targetStart/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(targetEnd/1000, 0).Format("2006-01-02 15:04:05"),
		weeks)

	// 按周分批处理
	currentWeekEnd := targetEnd
	weekCount := 0

	for currentWeekEnd >= targetStart && weekCount < weeks {
		weekCount++
		// 计算本周的开始时间（7天前）
		weekStart := currentWeekEnd - int64(7*24*60*60*1000)
		if weekStart < targetStart {
			weekStart = targetStart
		}

		logger.Infof("[%s] 开始同步第 %d/%d 周: %s ~ %s", 
			symbol, weekCount, weeks,
			time.Unix(weekStart/1000, 0).Format("2006-01-02 15:04:05"),
			time.Unix(currentWeekEnd/1000, 0).Format("2006-01-02 15:04:05"))

		// 在本周范围内，按天处理
		currentDayEnd := currentWeekEnd
		completedDays := 0

		for currentDayEnd >= weekStart {
			// 计算当天的开始时间（当天00:00:00 UTC）
			currentDay := time.Unix(currentDayEnd/1000, 0).UTC()
			dayKey := currentDay.Format("2006-01-02")
			dayStart := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
			dayEnd := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), 23, 59, 59, 999, time.UTC).UnixMilli()

			// 如果dayEnd超过当前周的范围，调整到currentWeekEnd
			if dayEnd > currentWeekEnd {
				dayEnd = currentWeekEnd
			}

			// 检查这一天是否已经完整同步
			daySynced, err := database.IsDaySynced(symbol, dayStart, dayEnd)
			if err != nil {
				logger.Warnf("[%s] 检查日期 %s 同步状态失败: %v", symbol, dayKey, err)
				// 继续处理，不中断
			}

			if daySynced {
				logger.Debugf("[%s] ✓ 日期 %s 已完整同步，跳过", symbol, dayKey)
				// 跳到前一天
				prevDay := currentDay.AddDate(0, 0, -1)
				currentDayEnd = time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 59, 59, 999, time.UTC).UnixMilli()
				completedDays++
				continue
			}

			// 这一天需要同步，使用 SyncSymbolWithPriority 函数
			// 它会自动检查缺失的时间段并同步
			logger.Infof("[%s] 同步日期 %s 的数据", symbol, dayKey)
			if err := SyncSymbolWithPriority(symbol, true); err != nil {
				logger.Errorf("[%s] 同步日期 %s 的数据失败: %v", symbol, dayKey, err)
				// 继续处理下一天
			} else {
				// 检查是否已同步（通过查询sync_time_ranges表）
				daySyncedAfter, _ := database.IsDaySynced(symbol, dayStart, dayEnd)
				if daySyncedAfter {
					completedDays++
					logger.Debugf("[%s] ✓ 日期 %s 同步完成", symbol, dayKey)
				}
			}

			// 跳到前一天
			prevDay := currentDay.AddDate(0, 0, -1)
			currentDayEnd = time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 59, 59, 999, time.UTC).UnixMilli()
		}

		logger.Infof("[%s] ✓ 第 %d/%d 周同步完成，已同步 %d 天", 
			symbol, weekCount, weeks, completedDays)

		// 跳到前一周
		currentWeekEnd = weekStart - 1
	}

	logger.Infof("[%s] ✓ 按周分批同步完成，处理了 %d 周", 
		symbol, weekCount)

	return nil
}

// SyncSymbolWithPriority 按优先级同步币种数据
// symbol: 交易对符号（如 BTC_USDT）
// priority: true 表示只同步近期数据（最近7天），false 表示同步所有缺失数据
func SyncSymbolWithPriority(symbol string, priority bool) error {
	now := time.Now().UnixMilli()
	// 历史同步时，延迟10分钟（避免获取不完整的数据）
	historyDelay := int64(10 * 60 * 1000)  // 10分钟前（历史数据）

	// 计算目标时间范围
	var targetStart, targetEnd int64
	if priority {
		// 优先模式：只同步最近7天的数据，实时同步到当前时间
		// 实时模式：只延迟10秒，优先保证实时性（避免获取到正在进行的K线）
		realtimeDelay := int64(10 * 1000) // 10秒前，优先保证实时数据
		today := time.Now().UTC()
		todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
		targetStart = todayStart - int64(7*24*60*60*1000) // 7天前
		targetEnd = now - realtimeDelay // 10秒前，优先保证实时性
	} else {
		// 非优先模式：从数据库最新数据开始，向前同步到2020年
		latestTime, err := database.GetLatestKLineTime(symbol)
		if err != nil {
			logger.Warnf("[%s] 获取最新K线时间失败: %v", symbol, err)
			latestTime = 0
		}
		if latestTime > 0 {
			// 从最新数据的时间开始，向前同步
			targetStart = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
			targetEnd = latestTime
		} else {
			// 如果没有数据，从今天开始向前同步7天
			today := time.Now().UTC()
			todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
			targetStart = todayStart - int64(7*24*60*60*1000)
			targetEnd = now - historyDelay // 10分钟前，历史数据使用较长延迟
		}
	}

	// 查找缺失的时间段（不检查同步记录器，直接查找数据库中的缺失时间段）
	missingRanges, err := database.FindMissingRanges(symbol, targetStart, targetEnd)
	if err != nil {
		return fmt.Errorf("查找缺失时间段失败: %w", err)
	}

	if len(missingRanges) == 0 {
		logger.Debugf("[%s] ✓ 无缺失数据，跳过同步", symbol)
		// 即使没有缺失数据，也记录本次同步操作，避免重复检查
		globalSyncRecorder.recordSync(symbol, targetStart, targetEnd)
		return nil
	}

	logger.Infof("[%s] 发现 %d 个缺失时间段，开始同步", symbol, len(missingRanges))

	// 创建 API 客户端
	proxyClient := api.NewProxyClient()

	// 同步每个缺失的时间段
	successCount := 0
	for _, missingRange := range missingRanges {
		if err := syncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime, proxyClient); err != nil {
			logger.Errorf("[%s] 同步时间段失败: %s ~ %s, error=%v",
				symbol,
				time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05"),
				time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05"),
				err)
			// 继续处理下一个时间段，不中断
			continue
		}

		// 记录已同步的时间段
		if err := database.AddSyncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime); err != nil {
			logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
		} else {
			successCount++
		}
	}

	// 记录本次同步操作（即使部分失败，也记录成功的部分）
	if successCount > 0 {
		globalSyncRecorder.recordSync(symbol, targetStart, targetEnd)
	}

	logger.Infof("[%s] ✓ 同步完成: 成功同步 %d/%d 个时间段", symbol, successCount, len(missingRanges))
	return nil
}

// syncTimeRange 同步指定时间范围的K线数据（支持分页，确保获取完整数据）
func syncTimeRange(symbol string, startTime, endTime int64, proxyClient *api.ProxyClient) error {
	// Gate.io API 使用秒级时间戳
	from := startTime / 1000
	to := endTime / 1000
	
	// API限制：每次最多返回1000条数据（约16.7小时）
	// 如果时间范围超过1000分钟，需要分页请求
	const maxLimit = 1000
	const maxMinutes = 1000 // 1000分钟 = 约16.7小时
	
	allKlines := make([]database.KLine1m, 0)
	currentFrom := from
	
	// 分页请求，确保获取所有数据
	for currentFrom < to {
		// 计算本次请求的结束时间（不超过1000分钟）
		currentTo := currentFrom + int64(maxMinutes*60)
		if currentTo > to {
			currentTo = to
		}
		
		// 构建 API URL
		url := fmt.Sprintf("https://api.gateio.ws/api/v4/spot/candlesticks?currency_pair=%s&interval=1m&from=%d&to=%d&limit=%d",
			symbol, currentFrom, currentTo, maxLimit)

		logger.Debugf("[%s] 请求K线数据: %s", symbol, url)

		// 调用 API（Gate.io 返回数组格式，使用 FetchAPIRaw 获取原始响应）
		rawBody, err := proxyClient.FetchAPIRaw(url, nil)
		if err != nil {
			return fmt.Errorf("API请求失败: %w", err)
		}

		// 解析响应数据（Gate.io 返回数组格式）
		// Gate.io API v4 实际返回格式: [timestamp, volume, close, high, low, open, base_volume]
		// 索引对应: [0: timestamp, 1: volume, 2: close, 3: high, 4: low, 5: open, 6: base_volume]
		var candlesticks [][]interface{}
		if err := json.Unmarshal(rawBody, &candlesticks); err != nil {
			return fmt.Errorf("解析API响应失败: %w", err)
		}

		if len(candlesticks) == 0 {
			// 如果本次请求没有数据，跳到下一个时间段
			currentFrom = currentTo + 1
			continue
		}

		// 调试：打印第一条数据的原始格式（仅第一次）
		if len(allKlines) == 0 && len(candlesticks) > 0 {
			logger.Debugf("[%s] 第一条K线原始数据: %v", symbol, candlesticks[0])
		}

		// 转换为 KLine1m 格式
		for _, candle := range candlesticks {
			if len(candle) < 7 {
				continue
			}

			// 解析时间戳（秒级）
			var timestamp int64
			switch v := candle[0].(type) {
			case float64:
				timestamp = int64(v)
			case string:
				ts, _ := strconv.ParseInt(v, 10, 64)
				timestamp = ts
			default:
				continue
			}

			// 解析价格和成交量
			// Gate.io API v4 实际返回格式: [timestamp, volume, close, high, low, open, base_volume]
			// 索引: [0: timestamp, 1: volume, 2: close, 3: high, 4: low, 5: open, 6: base_volume]
			open, errOpen := parseFloat(candle[5]) // open (索引5)
			high, errHigh := parseFloat(candle[3]) // high (索引3)
			low, errLow := parseFloat(candle[4])   // low (索引4)
			close, errClose := parseFloat(candle[2]) // close (索引2)
			volume, errVolume := parseFloat(candle[1]) // volume (索引1)
			
			// 验证解析错误
			if errOpen != nil || errHigh != nil || errLow != nil || errClose != nil || errVolume != nil {
				logger.Warnf("[%s] 解析价格数据失败: open=%v, high=%v, low=%v, close=%v, volume=%v", 
					symbol, errOpen, errHigh, errLow, errClose, errVolume)
				continue
			}
			
			// 验证价格合理性（BTC和ETH不应该低于1000）
			// 如果价格异常，记录详细日志以便调试
			if (symbol == "BTC_USDT" || symbol == "ETH_USDT") && (open < 1000 || close < 1000) {
				// 打印原始数据以便调试
				logger.Warnf("[%s] ⚠️ 检测到异常价格: open=%.2f, high=%.2f, low=%.2f, close=%.2f", 
					symbol, open, high, low, close)
				logger.Warnf("[%s] 原始数据: timestamp=%v, volume=%v, [2]=%v, [3]=%v, [4]=%v, [5]=%v", 
					symbol, candle[0], candle[1], candle[2], candle[3], candle[4], candle[5])
			}

			// 转换为毫秒级时间戳
			openTime := timestamp * 1000
			closeTime := openTime + 60000 - 1 // 1分钟K线，close_time = open_time + 60秒 - 1毫秒

			allKlines = append(allKlines, database.KLine1m{
				Symbol:    symbol,
				OpenTime:  openTime,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
				CloseTime: closeTime,
			})
		}

		// 更新下一次请求的起始时间（使用最后一条数据的时间 + 1分钟）
		if len(candlesticks) > 0 {
			var lastTimestamp int64
			switch v := candlesticks[len(candlesticks)-1][0].(type) {
			case float64:
				lastTimestamp = int64(v)
			case string:
				ts, _ := strconv.ParseInt(v, 10, 64)
				lastTimestamp = ts
			}
			// 如果返回的数据少于limit，说明已经获取完这个时间段的数据
			if len(candlesticks) < maxLimit {
				currentFrom = currentTo + 1
			} else {
				// 否则从最后一条数据的时间 + 1分钟开始
				currentFrom = lastTimestamp + 60
			}
		} else {
			currentFrom = currentTo + 1
		}

		// 避免API限流，稍作延迟
		time.Sleep(200 * time.Millisecond)
	}

	if len(allKlines) == 0 {
		logger.Debugf("[%s] 该时间段无数据", symbol)
		return nil
	}

	// 保存到数据库
	result, err := database.SaveKLine1m(allKlines)
	if err != nil {
		return fmt.Errorf("保存K线数据失败: %w", err)
	}

	logger.Infof("[%s] ✓ 成功拉取 %d 条数据 (时间范围: %s ~ %s, 插入=%d, 跳过=%d, 失败=%d)",
		symbol,
		len(allKlines),
		time.Unix(startTime/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(endTime/1000, 0).Format("2006-01-02 15:04:05"),
		result.InsertedCount,
		result.SkippedCount,
		result.ErrorCount)

	return nil
}

// parseFloat 解析浮点数（支持 string 和 float64）
func parseFloat(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("无法转换为浮点数: %v", v)
	}
}

// SyncSymbol 同步指定币种的数据（同步近期数据）
func SyncSymbol(symbol string) error {
	return SyncSymbolWithPriority(symbol, true)
}

// SyncSymbolHistorical 从指定年份开始同步历史数据
func SyncSymbolHistorical(symbol string, startYear int) error {
	// 计算目标时间范围：从指定年份的1月1日到现在
	startTime := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	now := time.Now().UnixMilli()
	tenMinutesAgo := now - int64(10*60*1000) // 10分钟前

	// 查找缺失的时间段
	missingRanges, err := database.FindMissingRanges(symbol, startTime, tenMinutesAgo)
	if err != nil {
		return fmt.Errorf("查找缺失时间段失败: %w", err)
	}

	if len(missingRanges) == 0 {
		logger.Debugf("[%s] ✓ 无缺失数据，跳过同步", symbol)
		return nil
	}

	logger.Infof("[%s] 发现 %d 个缺失时间段，开始同步历史数据", symbol, len(missingRanges))

	// 创建 API 客户端
	proxyClient := api.NewProxyClient()

	// 同步每个缺失的时间段
	for _, missingRange := range missingRanges {
		if err := syncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime, proxyClient); err != nil {
			logger.Errorf("[%s] 同步时间段失败: %s ~ %s, error=%v",
				symbol,
				time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05"),
				time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05"),
				err)
			// 继续处理下一个时间段，不中断
			continue
		}

		// 记录已同步的时间段
		if err := database.AddSyncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime); err != nil {
			logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
		}
	}

	logger.Infof("[%s] ✓ 历史数据同步完成", symbol)
	return nil
}

// SyncSymbolHistoricalBackward 从指定年份开始向后同步历史数据
// startYear: 起始年份
// batchSize: 每批获取的数据量
// maxDays: 最多拉取的天数（限制数据量）
func SyncSymbolHistoricalBackward(symbol string, startYear int, batchSize int, maxDays int) error {
	if batchSize <= 0 {
		batchSize = 300 // 默认300条
	}
	if maxDays <= 0 {
		maxDays = 7 // 默认7天
	}

	now := time.Now().UnixMilli()
	tenMinutesAgo := now - int64(10*60*1000) // 10分钟前

	// 计算目标时间范围：从指定年份的1月1日到现在，但限制最多 maxDays 天
	today := time.Now().UTC()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
	targetStart := todayStart - int64(maxDays*24*60*60*1000) // maxDays 天前
	startYearTime := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	
	// 如果 maxDays 天前的时间早于起始年份，则从起始年份开始
	if targetStart < startYearTime {
		targetStart = startYearTime
	}
	
	targetEnd := tenMinutesAgo

	logger.Infof("[%s] 开始向后同步历史数据: 从 %s 到 %s (最多 %d 天)",
		symbol,
		time.Unix(targetStart/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(targetEnd/1000, 0).Format("2006-01-02 15:04:05"),
		maxDays)

	// 查找缺失的时间段
	missingRanges, err := database.FindMissingRanges(symbol, targetStart, targetEnd)
	if err != nil {
		return fmt.Errorf("查找缺失时间段失败: %w", err)
	}

	if len(missingRanges) == 0 {
		logger.Debugf("[%s] ✓ 无缺失数据，跳过同步", symbol)
		return nil
	}

	logger.Infof("[%s] 发现 %d 个缺失时间段，开始同步", symbol, len(missingRanges))

	// 创建 API 客户端
	proxyClient := api.NewProxyClient()

	// 同步每个缺失的时间段（从新到旧）
	for i := len(missingRanges) - 1; i >= 0; i-- {
		missingRange := missingRanges[i]
		
		// 限制每个时间段的长度，分批处理
		rangeDuration := missingRange.EndTime - missingRange.StartTime
		maxRangeDuration := int64(batchSize * 60 * 1000) // batchSize 分钟的数据

		if rangeDuration > maxRangeDuration {
			// 如果时间段太长，分批处理
			currentStart := missingRange.StartTime
			for currentStart < missingRange.EndTime {
				currentEnd := currentStart + maxRangeDuration
				if currentEnd > missingRange.EndTime {
					currentEnd = missingRange.EndTime
				}

				if err := syncTimeRange(symbol, currentStart, currentEnd, proxyClient); err != nil {
					logger.Errorf("[%s] 同步时间段失败: %s ~ %s, error=%v",
						symbol,
						time.Unix(currentStart/1000, 0).Format("2006-01-02 15:04:05"),
						time.Unix(currentEnd/1000, 0).Format("2006-01-02 15:04:05"),
						err)
					// 继续处理下一个时间段
					currentStart = currentEnd + 1
					continue
				}

				// 记录已同步的时间段
				if err := database.AddSyncTimeRange(symbol, currentStart, currentEnd); err != nil {
					logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
				}

				currentStart = currentEnd + 1
				
				// 避免API限流，稍作延迟
				time.Sleep(200 * time.Millisecond)
			}
		} else {
			// 时间段不长，直接同步
			if err := syncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime, proxyClient); err != nil {
				logger.Errorf("[%s] 同步时间段失败: %s ~ %s, error=%v",
					symbol,
					time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05"),
					time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05"),
					err)
				// 继续处理下一个时间段，不中断
				continue
			}

			// 记录已同步的时间段
			if err := database.AddSyncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime); err != nil {
				logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
			}
		}

		// 避免API限流，稍作延迟
		time.Sleep(200 * time.Millisecond)
	}

	logger.Infof("[%s] ✓ 向后同步历史数据完成", symbol)
	return nil
}

// SyncSymbolInitial 初始同步指定币种的数据（同步最近N天的数据）
func SyncSymbolInitial(symbol string, days int) error {
	if days <= 0 {
		days = 7 // 默认7天
	}

	now := time.Now().UnixMilli()
	tenMinutesAgo := now - int64(10*60*1000) // 10分钟前

	// 计算目标时间范围：从 days 天前到现在
	today := time.Now().UTC()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
	targetStart := todayStart - int64(days*24*60*60*1000) // days 天前
	targetEnd := tenMinutesAgo

	logger.Infof("[%s] 开始初始同步: 从 %s 到 %s (共 %d 天)",
		symbol,
		time.Unix(targetStart/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(targetEnd/1000, 0).Format("2006-01-02 15:04:05"),
		days)

	// 查找缺失的时间段
	missingRanges, err := database.FindMissingRanges(symbol, targetStart, targetEnd)
	if err != nil {
		return fmt.Errorf("查找缺失时间段失败: %w", err)
	}

	if len(missingRanges) == 0 {
		logger.Debugf("[%s] ✓ 无缺失数据，跳过同步", symbol)
		return nil
	}

	logger.Infof("[%s] 发现 %d 个缺失时间段，开始同步", symbol, len(missingRanges))

	// 创建 API 客户端
	proxyClient := api.NewProxyClient()

	// 同步每个缺失的时间段
	for _, missingRange := range missingRanges {
		if err := syncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime, proxyClient); err != nil {
			logger.Errorf("[%s] 同步时间段失败: %s ~ %s, error=%v",
				symbol,
				time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05"),
				time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05"),
				err)
			// 继续处理下一个时间段，不中断
			continue
		}

		// 记录已同步的时间段
		if err := database.AddSyncTimeRange(symbol, missingRange.StartTime, missingRange.EndTime); err != nil {
			logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
		}

		// 避免API限流，稍作延迟
		time.Sleep(200 * time.Millisecond)
	}

	logger.Infof("[%s] ✓ 初始同步完成", symbol)
	return nil
}

