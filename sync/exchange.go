package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wails-contract-warn/database"
	"wails-contract-warn/logger"
)

// ExchangeAPI 交易所API接口
type ExchangeAPI struct {
	BaseURL  string
	Exchange string // 交易所名称
}

// NewExchangeAPI 创建交易所API实例（默认使用Gate.io）
func NewExchangeAPI() *ExchangeAPI {
	return &ExchangeAPI{
		BaseURL:  "https://api.gateio.ws/api/v4",
		Exchange: "gateio",
	}
}

// NewGateIOAPI 创建Gate.io API实例
func NewGateIOAPI() *ExchangeAPI {
	return &ExchangeAPI{
		BaseURL:  "https://api.gateio.ws/api/v4",
		Exchange: "gateio",
	}
}

// KlineResponse Binance K线响应
type KlineResponse struct {
	OpenTime  int64  `json:"0"`
	Open      string `json:"1"`
	High      string `json:"2"`
	Low       string `json:"3"`
	Close     string `json:"4"`
	Volume    string `json:"5"`
	CloseTime int64  `json:"6"`
}

// FetchKLines 从交易所获取K线数据
func (api *ExchangeAPI) FetchKLines(symbol, interval string, startTime, endTime int64, limit int) ([]database.KLine1m, error) {
	if api.Exchange == "gateio" {
		return api.fetchGateIOKLines(symbol, interval, startTime, endTime, limit)
	}
	// 默认使用Binance格式（向后兼容）
	return api.fetchBinanceKLines(symbol, interval, startTime, endTime, limit)
}

// fetchGateIOKLines 从Gate.io获取K线数据
func (api *ExchangeAPI) fetchGateIOKLines(symbol, interval string, startTime, endTime int64, limit int) ([]database.KLine1m, error) {
	// Gate.io API: GET /api/v4/spot/candlesticks
	// 参数: currency_pair, interval, from, to, limit
	// interval: 1m, 5m, 15m, 30m, 1h, 4h, 1d
	//
	// Gate.io 限制：
	// 1. 最多返回 1000 条数据
	// 2. 最多只能获取最近 10000 个数据点（对于1分钟K线，约6.9天）
	// 3. 如果设置了 from，不能超过 10000 个数据点之前

	url := fmt.Sprintf("%s/spot/candlesticks", api.BaseURL)

	// 构建查询参数
	params := make(map[string]string)
	params["currency_pair"] = symbol // Gate.io使用 BTC_USDT 格式
	params["interval"] = interval

	// Gate.io 限制：最多只能获取最近 10000 个数据点
	// 对于 1 分钟 K 线，10000 点 = 10000 分钟 ≈ 6.9 天
	maxPointsAgo := int64(10000 * 60 * 1000) // 10000分钟的毫秒数
	now := time.Now().UnixMilli()

	// 如果设置了 startTime，检查是否超过限制
	if startTime > 0 {
		minAllowedTime := now - maxPointsAgo
		if startTime < minAllowedTime {
			// 时间太早，不设置 from 参数，让 API 返回最近的数据
			// 只设置 limit，从最新数据开始往前拉取
		} else {
			// 时间在允许范围内，设置 from 参数
			params["from"] = strconv.FormatInt(startTime/1000, 10) // Gate.io使用秒级时间戳
		}
	}

	// 设置 limit（优先使用较小的值，避免一次性拉取太多）
	if limit > 0 && limit <= 1000 {
		params["limit"] = strconv.Itoa(limit)
	} else if limit > 1000 {
		params["limit"] = "1000" // Gate.io最大限制1000
	} else {
		// 默认拉取300条
		params["limit"] = "300"
	}

	// 设置 to 参数（如果提供了 endTime）
	if endTime > 0 {
		params["to"] = strconv.FormatInt(endTime/1000, 10) // Gate.io使用秒级时间戳
	}

	// 构建URL
	url += "?"
	first := true
	for k, v := range params {
		if !first {
			url += "&"
		}
		url += fmt.Sprintf("%s=%s", k, v)
		first = false
	}

	logger.Infof("[Gate.io API] 请求URL: %s", url)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 重试机制：最多重试3次
	maxRetries := 3
	var resp *http.Response
	var bodyBytes []byte
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 重试前等待，使用指数退避：1s, 2s, 4s
			waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
			logger.Warnf("[Gate.io API] 第 %d 次重试，等待 %v 后重试...", attempt, waitTime)
			time.Sleep(waitTime)
		}

		resp, err = client.Get(url)
		if err != nil {
			// 网络错误，可以重试
			logger.Warnf("[Gate.io API] 请求失败 (尝试 %d/%d): %v", attempt+1, maxRetries+1, err)
			if attempt < maxRetries {
				continue // 继续重试
			}
			logger.Errorf("[Gate.io API] 请求失败，已重试 %d 次: %v", maxRetries+1, err)
			return nil, fmt.Errorf("请求失败（已重试%d次）: %w", maxRetries, err)
		}

		// 读取响应体
		bodyBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close() // 立即关闭，避免资源泄漏

		if err != nil {
			logger.Warnf("[Gate.io API] 读取响应失败 (尝试 %d/%d): %v", attempt+1, maxRetries+1, err)
			if attempt < maxRetries {
				continue // 继续重试
			}
			logger.Errorf("[Gate.io API] 读取响应失败，已重试 %d 次: %v", maxRetries+1, err)
			return nil, fmt.Errorf("读取响应失败（已重试%d次）: %w", maxRetries, err)
		}

		logger.Infof("[Gate.io API] 响应状态码: %d, 响应长度: %d 字节", resp.StatusCode, len(bodyBytes))
		if len(bodyBytes) > 0 && len(bodyBytes) < 1000 {
			logger.Infof("[Gate.io API] 响应内容: %s", string(bodyBytes))
		}

		// 检查HTTP状态码
		if resp.StatusCode == http.StatusOK {
			// 成功，跳出重试循环
			if attempt > 0 {
				logger.Infof("[Gate.io API] ✓ 重试成功（第 %d 次尝试）", attempt+1)
			}
			break
		}

		// HTTP错误状态码处理
		if resp.StatusCode >= 500 {
			// 5xx 服务器错误，可以重试
			logger.Warnf("[Gate.io API] 服务器错误 %d (尝试 %d/%d): %s", resp.StatusCode, attempt+1, maxRetries+1, string(bodyBytes))
			if attempt < maxRetries {
				continue // 继续重试
			}
			logger.Errorf("[Gate.io API] 服务器错误，已重试 %d 次: %s", maxRetries+1, string(bodyBytes))
			return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(bodyBytes))
		} else if resp.StatusCode == 429 {
			// 429 Too Many Requests，可以重试
			logger.Warnf("[Gate.io API] 请求频率限制 (尝试 %d/%d): %s", attempt+1, maxRetries+1, string(bodyBytes))
			if attempt < maxRetries {
				// 对于429错误，等待更长时间
				waitTime := time.Duration(2<<uint(attempt)) * time.Second
				logger.Infof("[Gate.io API] 频率限制，等待 %v 后重试...", waitTime)
				time.Sleep(waitTime)
				continue // 继续重试
			}
			logger.Errorf("[Gate.io API] 请求频率限制，已重试 %d 次: %s", maxRetries+1, string(bodyBytes))
			return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(bodyBytes))
		} else {
			// 4xx 客户端错误（除了429），通常不需要重试
			logger.Errorf("[Gate.io API] 客户端错误 %d: %s", resp.StatusCode, string(bodyBytes))
			return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(bodyBytes))
		}
	}

	// Gate.io返回格式: [["timestamp", "volume", "close", "high", "low", "open", "base_volume"], ...]
	var klinesRaw [][]interface{}
	if err := json.Unmarshal(bodyBytes, &klinesRaw); err != nil {
		logger.Errorf("[Gate.io API] JSON解析失败: %v, 原始响应: %s", err, string(bodyBytes))
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	logger.Infof("[Gate.io API] 解析成功，原始数据条数: %d", len(klinesRaw))
	if len(klinesRaw) > 0 {
		logger.Infof("[Gate.io API] 第一条原始数据: %+v", klinesRaw[0])
	}

	var klines []database.KLine1m
	parseErrors := 0
	for i, k := range klinesRaw {
		if len(k) < 7 {
			logger.Warnf("[Gate.io API] 数据项 #%d 长度不足: %d (需要7)", i, len(k))
			parseErrors++
			continue
		}

		// Gate.io格式: [timestamp, volume, close, high, low, open, base_volume]
		// 注意：Gate.io返回的时间戳可能是秒或毫秒，需要判断
		var timestamp int64
		switch v := k[0].(type) {
		case float64:
			timestamp = int64(v)
			// 如果时间戳小于1e12，认为是秒级，需要转换为毫秒
			if timestamp < 1e12 {
				timestamp = timestamp * 1000
			}
		case int64:
			timestamp = v
			if timestamp < 1e12 {
				timestamp = timestamp * 1000
			}
		case string:
			// 尝试解析字符串时间戳
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				timestamp = ts
				if timestamp < 1e12 {
					timestamp = timestamp * 1000
				}
			} else {
				logger.Warnf("[Gate.io API] 无法解析时间戳: %v (类型: %T)", k[0], k[0])
				parseErrors++
				continue
			}
		default:
			logger.Warnf("[Gate.io API] 未知的时间戳类型: %v (类型: %T)", k[0], k[0])
			parseErrors++
			continue
		}

		volume := parseFloat(fmt.Sprintf("%v", k[1]))
		close := parseFloat(fmt.Sprintf("%v", k[2]))
		high := parseFloat(fmt.Sprintf("%v", k[3]))
		low := parseFloat(fmt.Sprintf("%v", k[4]))
		open := parseFloat(fmt.Sprintf("%v", k[5]))

		// 计算close_time（假设1分钟K线）
		closeTime := timestamp + 60*1000 - 1

		klines = append(klines, database.KLine1m{
			Symbol:    symbol,
			OpenTime:  timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: closeTime,
		})

		// 只打印第一条解析后的数据
		if i == 0 {
			timestampStr := time.Unix(timestamp/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[Gate.io API] 第一条解析后的数据: timestamp=%s (%d), open=%.8f, high=%.8f, low=%.8f, close=%.8f, volume=%.8f",
				timestampStr, timestamp, open, high, low, close, volume)
		}
	}

	if parseErrors > 0 {
		logger.Warnf("[Gate.io API] 解析过程中有 %d 条数据解析失败", parseErrors)
	}

	logger.Infof("[Gate.io API] 最终解析成功 %d 条K线数据", len(klines))
	return klines, nil
}

// fetchBinanceKLines 从Binance获取K线数据（向后兼容）
func (api *ExchangeAPI) fetchBinanceKLines(symbol, interval string, startTime, endTime int64, limit int) ([]database.KLine1m, error) {
	url := fmt.Sprintf("%s/klines?symbol=%s&interval=%s", api.BaseURL, symbol, interval)

	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 重试机制：最多重试3次
	maxRetries := 3
	var resp *http.Response
	var body []byte
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// 重试前等待，使用指数退避：1s, 2s, 4s
			waitTime := time.Duration(1<<uint(attempt-1)) * time.Second
			logger.Warnf("[Binance API] 第 %d 次重试，等待 %v 后重试...", attempt, waitTime)
			time.Sleep(waitTime)
		}

		resp, err = client.Get(url)
		if err != nil {
			// 网络错误，可以重试
			logger.Warnf("[Binance API] 请求失败 (尝试 %d/%d): %v", attempt+1, maxRetries+1, err)
			if attempt < maxRetries {
				continue // 继续重试
			}
			logger.Errorf("[Binance API] 请求失败，已重试 %d 次: %v", maxRetries+1, err)
			return nil, fmt.Errorf("请求失败（已重试%d次）: %w", maxRetries, err)
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close() // 立即关闭，避免资源泄漏

		if err != nil {
			logger.Warnf("[Binance API] 读取响应失败 (尝试 %d/%d): %v", attempt+1, maxRetries+1, err)
			if attempt < maxRetries {
				continue // 继续重试
			}
			logger.Errorf("[Binance API] 读取响应失败，已重试 %d 次: %v", maxRetries+1, err)
			return nil, fmt.Errorf("读取响应失败（已重试%d次）: %w", maxRetries, err)
		}

		if resp.StatusCode == http.StatusOK {
			// 成功，跳出重试循环
			if attempt > 0 {
				logger.Infof("[Binance API] ✓ 重试成功（第 %d 次尝试）", attempt+1)
			}
			break
		}

		// HTTP错误状态码处理
		if resp.StatusCode >= 500 {
			// 5xx 服务器错误，可以重试
			logger.Warnf("[Binance API] 服务器错误 %d (尝试 %d/%d): %s", resp.StatusCode, attempt+1, maxRetries+1, string(body))
			if attempt < maxRetries {
				continue // 继续重试
			}
			logger.Errorf("[Binance API] 服务器错误，已重试 %d 次: %s", maxRetries+1, string(body))
			return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(body))
		} else if resp.StatusCode == 429 {
			// 429 Too Many Requests，可以重试
			logger.Warnf("[Binance API] 请求频率限制 (尝试 %d/%d): %s", attempt+1, maxRetries+1, string(body))
			if attempt < maxRetries {
				// 对于429错误，等待更长时间
				waitTime := time.Duration(2<<uint(attempt)) * time.Second
				logger.Infof("[Binance API] 频率限制，等待 %v 后重试...", waitTime)
				time.Sleep(waitTime)
				continue // 继续重试
			}
			logger.Errorf("[Binance API] 请求频率限制，已重试 %d 次: %s", maxRetries+1, string(body))
			return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(body))
		} else {
			// 4xx 客户端错误（除了429），通常不需要重试
			logger.Errorf("[Binance API] 客户端错误 %d: %s", resp.StatusCode, string(body))
			return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(body))
		}
	}

	var klinesRaw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&klinesRaw); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var klines []database.KLine1m
	for _, k := range klinesRaw {
		if len(k) < 7 {
			continue
		}

		openTime := int64(k[0].(float64))
		open := parseFloat(k[1].(string))
		high := parseFloat(k[2].(string))
		low := parseFloat(k[3].(string))
		close := parseFloat(k[4].(string))
		volume := parseFloat(k[5].(string))
		closeTime := int64(k[6].(float64))

		klines = append(klines, database.KLine1m{
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

	return klines, nil
}

// parseFloat 解析字符串为float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// SyncSymbol 同步指定交易对的K线数据（优先同步近期数据）
func SyncSymbol(symbol string) error {
	return SyncSymbolWithPriority(symbol, true)
}

// SyncSymbolWithPriority 同步指定交易对的K线数据
// priorityRecent: true表示优先同步近期数据（昨日至今），false表示同步历史数据
func SyncSymbolWithPriority(symbol string, priorityRecent bool) error {
	api := NewExchangeAPI()

	// 1. 计算目标时间范围
	now := time.Now().UnixMilli()
	var targetStart, targetEnd int64

	if priorityRecent {
		// 实时模式：同步最近2天的数据（用于实时数据同步）
		targetStart = now - 2*24*60*60*1000 // 2天前
		targetEnd = now
	} else {
		// 历史模式：从2020年开始到当前（用于历史数据同步）
		targetStart = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
		targetEnd = now
	}

	// 2. 查找缺失的时间段
	missingRanges, err := database.FindMissingRanges(symbol, targetStart, targetEnd)
	if err != nil {
		return fmt.Errorf("查找缺失时间段失败: %w", err)
	}

	// 显示同步开始信息
	var mode string
	if priorityRecent {
		mode = "近期数据"
	} else {
		mode = "历史数据"
	}
	targetStartStr := time.Unix(targetStart/1000, 0).Format("2006-01-02 15:04:05")
	targetEndStr := time.Unix(targetEnd/1000, 0).Format("2006-01-02 15:04:05")
	logger.Infof("开始同步 %s [%s] 目标范围: %s ~ %s", symbol, mode, targetStartStr, targetEndStr)

	// 如果没有缺失的时间段，说明已经全部同步完成
	if len(missingRanges) == 0 {
		logger.Infof("[%s] ✓ 所有时间段已同步，无需同步", symbol)
		return nil
	}

	logger.Infof("[%s] 发现 %d 个缺失的时间段需要同步", symbol, len(missingRanges))
	for i, r := range missingRanges {
		rangeStartStr := time.Unix(r.StartTime/1000, 0).Format("2006-01-02 15:04:05")
		rangeEndStr := time.Unix(r.EndTime/1000, 0).Format("2006-01-02 15:04:05")
		logger.Infof("[%s] 缺失时间段 #%d: %s ~ %s", symbol, i+1, rangeStartStr, rangeEndStr)
	}

	// 3. 对每个缺失的时间段进行同步
	totalFetched := 0
	batchSize := 300 // 每次只拉取300条

	for rangeIdx, missingRange := range missingRanges {
		logger.Infof("[%s] 开始同步缺失时间段 #%d/%d: %s ~ %s", symbol, rangeIdx+1, len(missingRanges),
			time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05"),
			time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05"))

		// 在当前缺失时间段内分批拉取
		currentStart := missingRange.StartTime
		batchCount := 0
		rangeFetched := 0

		for currentStart <= missingRange.EndTime {
			// 检查是否超过 Gate.io 的限制（最多10000个数据点之前）
			maxPointsAgo := int64(10000 * 60 * 1000) // 10000分钟的毫秒数
			minAllowedTime := now - maxPointsAgo

			// 如果起始时间太早，调整到允许的最早时间
			if currentStart < minAllowedTime {
				currentStart = minAllowedTime
			}

			// 如果已经超过限制，停止同步历史数据
			if currentStart >= now || currentStart > missingRange.EndTime {
				break
			}

			// 显示当前批次信息
			batchCount++
			currentStartStr := time.Unix(currentStart/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[%s] 批次 #%d: 从 %s 开始拉取 (最多 %d 条)", symbol, batchCount, currentStartStr, batchSize)

			// 从交易所拉取数据（每次300条，不设置endTime，让API返回到当前时间）
			logger.Infof("[%s] 正在请求API: startTime=%d (from=%s), limit=%d", symbol, currentStart, currentStartStr, batchSize)
			klines, err := api.FetchKLines(symbol, "1m", currentStart, 0, batchSize)
			if err != nil {
				logger.Errorf("[%s] ❌ API请求失败: %v", symbol, err)
				// 如果是时间太早的错误，说明已经无法获取更早的数据
				if strings.Contains(err.Error(), "too long ago") || strings.Contains(err.Error(), "10000 points") {
					logger.Warnf("[%s] 无法获取更早的数据（超过10000个数据点限制），停止同步", symbol)
					// 只能获取最近的数据，停止同步更早的历史数据
					break
				}
				// 如果是时间范围错误，尝试不设置startTime，只拉取最近的数据
				if strings.Contains(err.Error(), "range too broad") || strings.Contains(err.Error(), "INVALID_PARAM_VALUE") {
					logger.Infof("[%s] 时间范围错误，尝试拉取最近的数据（不设置startTime）", symbol)
					// 不设置startTime，只拉取最近的数据
					klines, err = api.FetchKLines(symbol, "1m", 0, 0, batchSize)
					if err != nil {
						logger.Errorf("[%s] ❌ 重试拉取最近数据也失败: %v", symbol, err)
						return fmt.Errorf("拉取K线数据失败: %w", err)
					}
					logger.Infof("[%s] ✓ 重试成功，拉取到最近的数据", symbol)
				} else {
					logger.Errorf("[%s] ❌ 拉取K线数据失败（未知错误）: %v", symbol, err)
					return fmt.Errorf("拉取K线数据失败: %w", err)
				}
			}

			if len(klines) == 0 {
				// 没有新数据，可能已经同步到最新
				logger.Infof("[%s] ⚠️ API返回空数据，可能已经同步到最新或没有数据", symbol)
				break
			}

			// 显示拉取成功信息
			firstTime := time.Unix(klines[0].OpenTime/1000, 0).Format("2006-01-02 15:04:05")
			lastTimeStr := time.Unix(klines[len(klines)-1].OpenTime/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[%s] ✓ 成功拉取 %d 条数据 (时间范围: %s ~ %s)", symbol, len(klines), firstTime, lastTimeStr)

			// 打印第一条完整数据（拉取阶段）
			if len(klines) > 0 {
				firstKline := klines[0]
				firstOpenTimeStr := time.Unix(firstKline.OpenTime/1000, 0).Format("2006-01-02 15:04:05")
				firstCloseTimeStr := time.Unix(firstKline.CloseTime/1000, 0).Format("2006-01-02 15:04:05")
				logger.Infof("[%s] 第一条数据详情: open_time=%s, close_time=%s, open=%.8f, high=%.8f, low=%.8f, close=%.8f, volume=%.8f",
					symbol, firstOpenTimeStr, firstCloseTimeStr, firstKline.Open, firstKline.High, firstKline.Low, firstKline.Close, firstKline.Volume)
			}

			// 保存到数据库
			logger.Infof("[%s] 开始保存 %d 条数据到数据库...", symbol, len(klines))
			if err := database.SaveKLine1m(klines); err != nil {
				logger.Errorf("[%s] ❌ 保存K线数据失败: %v", symbol, err)
				return fmt.Errorf("保存K线数据失败: %w", err)
			}
			logger.Infof("[%s] ✓ 数据保存完成", symbol)

			// 更新统计
			totalFetched += len(klines)
			rangeFetched += len(klines)

			// 更新同步状态
			lastKlineTime := klines[len(klines)-1].CloseTime
			if err := database.UpdateSyncStatus(symbol, time.Now().UnixMilli(), lastKlineTime); err != nil {
				return fmt.Errorf("更新同步状态失败: %w", err)
			}

			// 如果返回的数据少于batchSize，说明已经拉完这个时间段
			if len(klines) < batchSize {
				break
			}

			// 下一批从最后一条的下一条开始
			currentStart = lastKlineTime + 1

			// 避免请求过快，稍作延迟
			time.Sleep(200 * time.Millisecond)
		}

		// 记录已同步的时间段
		if rangeFetched > 0 {
			// 计算实际同步的时间范围（从第一批的第一条到最后一批的最后一条）
			// 这里简化处理，使用缺失时间段的边界
			actualEnd := currentStart - 1
			if actualEnd < missingRange.StartTime {
				actualEnd = missingRange.EndTime
			}
			if err := database.AddSyncTimeRange(symbol, missingRange.StartTime, actualEnd); err != nil {
				logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
			} else {
				logger.Infof("[%s] ✓ 时间段 #%d 同步完成，获取 %d 条数据，已记录时间段", symbol, rangeIdx+1, rangeFetched)
			}
		}
	}

	// 显示同步完成信息
	if totalFetched > 0 {
		logger.Infof("[%s] ✓ 同步完成，共获取 %d 条数据", symbol, totalFetched)
	} else {
		logger.Debugf("[%s] 同步完成，无新数据", symbol)
	}

	return nil
}

// SyncSymbolHistoricalBackward 历史数据同步（从近到远倒推，按天为单位，直到指定年份）
// 从1小时前开始，以天为单位逐步向前（倒推）获取数据，直到startYear
func SyncSymbolHistoricalBackward(symbol string, startYear int, batchSize int) error {
	api := NewExchangeAPI()

	// 计算目标时间范围
	// 从10分钟前开始，避免每次都获取到最新的一条数据而无法继续倒推
	// 跳过最近10分钟，防止最近时间增加导致的死循环
	now := time.Now().UnixMilli()
	tenMinutesAgo := now - int64(10*60*1000) // 10分钟前
	targetStart := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	targetEnd := tenMinutesAgo // 从10分钟前开始，跳过最近10分钟

	// 查找缺失的时间段
	missingRanges, err := database.FindMissingRanges(symbol, targetStart, targetEnd)
	if err != nil {
		return fmt.Errorf("查找缺失时间段失败: %w", err)
	}

	// 如果没有缺失的时间段，说明已经全部同步完成
	if len(missingRanges) == 0 {
		logger.Infof("[%s] ✓ 所有历史数据已同步，无需同步", symbol)
		return nil
	}

	// 对缺失时间段按时间倒序排序（从近到远）
	// 这样优先同步最新的缺失数据
	for i, j := 0, len(missingRanges)-1; i < j; i, j = i+1, j-1 {
		missingRanges[i], missingRanges[j] = missingRanges[j], missingRanges[i]
	}

	logger.Infof("开始同步 %s [历史数据-按天倒推] 目标范围: %s ~ %s (起始年份: %d, 跳过最近10分钟)",
		symbol,
		time.Unix(targetStart/1000, 0).Format("2006-01-02 15:04:05"),
		time.Unix(targetEnd/1000, 0).Format("2006-01-02 15:04:05"),
		startYear)
	logger.Infof("[%s] 发现 %d 个缺失的时间段需要同步（从近到远，按天倒推）", symbol, len(missingRanges))

	// 从近到远同步缺失的时间段
	totalFetched := 0

	for rangeIdx, missingRange := range missingRanges {
		rangeStartStr := time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05")
		rangeEndStr := time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05")
		logger.Infof("[%s] 开始同步缺失时间段 #%d/%d: %s ~ %s (从近到远)", symbol, rangeIdx+1, len(missingRanges), rangeStartStr, rangeEndStr)

		// 在当前缺失时间段内，按天为单位倒推拉取
		// 策略：从 missingRange.EndTime 开始，每天向前倒推一天
		currentDayEnd := missingRange.EndTime
		dayCount := 0
		rangeFetched := 0

		for currentDayEnd >= missingRange.StartTime {
			// 检查是否超过 Gate.io 的限制（最多10000个数据点之前）
			// 注意：这里使用 tenMinutesAgo 而不是 now，因为我们从10分钟前开始同步
			maxPointsAgo := int64(10000 * 60 * 1000) // 10000分钟的毫秒数
			minAllowedTime := tenMinutesAgo - maxPointsAgo

			// 如果结束时间太早，调整到允许的最早时间
			if currentDayEnd < minAllowedTime {
				currentDayEnd = minAllowedTime
			}

			// 如果已经超过限制或到达起始时间，停止
			if currentDayEnd < missingRange.StartTime {
				break
			}

			// 计算当天的开始时间（当天00:00:00 UTC）
			currentDay := time.Unix(currentDayEnd/1000, 0).UTC()
			dayStart := time.Date(currentDay.Year(), currentDay.Month(), currentDay.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()

			// 如果 dayStart >= currentDayEnd，说明已经到达当天的开始时间之前
			// 应该直接跳到前一天，而不是拉取无效的时间范围
			if dayStart >= currentDayEnd {
				// 跳到前一天：前一天的 23:59:59.999 UTC
				prevDay := currentDay.AddDate(0, 0, -1)
				prevDayEnd := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 59, 59, 999, time.UTC).UnixMilli()

				// 检查是否已经到达或超过起始时间
				if prevDayEnd < missingRange.StartTime {
					break
				}

				// 更新 currentDayEnd 为前一天，并添加调试日志
				oldDayEnd := currentDayEnd
				oldDayEndStr := time.Unix(oldDayEnd/1000, 0).UTC().Format("2006-01-02 15:04:05")
				currentDayEnd = prevDayEnd
				newDayEndStr := time.Unix(currentDayEnd/1000, 0).UTC().Format("2006-01-02 15:04:05")

				// 防止死循环：如果跳转后的日期和跳转前相同，说明有问题，直接break
				if prevDayEnd >= oldDayEnd {
					logger.Errorf("[%s] ❌ 跳转逻辑错误：跳转后的时间 %s 大于等于跳转前的时间 %s，停止同步", symbol, newDayEndStr, oldDayEndStr)
					break
				}

				logger.Infof("[%s] 当天数据已获取完毕，从 %s 跳到前一天: %s", symbol, oldDayEndStr, newDayEndStr)
				continue
			}

			// 确保不超过缺失时间段的起始时间
			if dayStart < missingRange.StartTime {
				dayStart = missingRange.StartTime
			}

			// 显示当前批次信息（按天，从00:00:00开始）
			dayCount++
			dayStartStr := time.Unix(dayStart/1000, 0).Format("2006-01-02 15:04:05")
			dayEndStr := time.Unix(currentDayEnd/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[%s] 第 %d 天: 倒推拉取 %s (00:00:00) ~ %s (最多 %d 条)", symbol, dayCount, time.Unix(dayStart/1000, 0).Format("2006-01-02"), dayEndStr, batchSize)

			// 从交易所拉取数据（从 dayStart 开始，到 currentDayEnd 结束）
			logger.Infof("[%s] 正在请求API: startTime=%d (from=%s), endTime=%d (to=%s), limit=%d",
				symbol, dayStart, dayStartStr, currentDayEnd, dayEndStr, batchSize)
			klines, err := api.FetchKLines(symbol, "1m", dayStart, currentDayEnd, batchSize)
			if err != nil {
				logger.Errorf("[%s] ❌ API请求失败: %v", symbol, err)
				// 如果是时间太早的错误，说明已经无法获取更早的数据
				if strings.Contains(err.Error(), "too long ago") || strings.Contains(err.Error(), "10000 points") {
					logger.Warnf("[%s] 无法获取更早的数据（超过10000个数据点限制），停止倒推", symbol)
					break
				}
				// 如果是时间范围错误，尝试缩小范围
				if strings.Contains(err.Error(), "range too broad") || strings.Contains(err.Error(), "INVALID_PARAM_VALUE") {
					logger.Infof("[%s] 时间范围错误，缩小范围继续倒推", symbol)
					// 缩小时间范围，只拉取最近的数据（缩小到半天）
					smallerEnd := dayStart + int64(12*60*60*1000) // 12小时
					if smallerEnd > currentDayEnd {
						smallerEnd = currentDayEnd
					}
					klines, err = api.FetchKLines(symbol, "1m", dayStart, smallerEnd, batchSize)
					if err != nil {
						logger.Errorf("[%s] ❌ 重试拉取也失败: %v", symbol, err)
						// 继续下一个时间段
						break
					}
					logger.Infof("[%s] ✓ 重试成功，拉取到数据", symbol)
				} else {
					logger.Errorf("[%s] ❌ 拉取K线数据失败（未知错误）: %v", symbol, err)
					// 继续下一个时间段
					break
				}
			}

			if len(klines) == 0 {
				// 没有新数据，当天的数据已获取完毕，跳到前一天
				logger.Infof("[%s] ⚠️ API返回空数据，当天数据已获取完毕，跳到前一天", symbol)
				// 使用 currentDay 而不是 dayStart，因为 dayStart 可能已经被调整过
				currentDay := time.Unix(currentDayEnd/1000, 0)
				prevDay := currentDay.AddDate(0, 0, -1)
				prevDayEnd := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 59, 59, 999, time.UTC).UnixMilli()
				if prevDayEnd < missingRange.StartTime {
					break
				}
				currentDayEnd = prevDayEnd
				logger.Infof("[%s] 跳到前一天: %s", symbol, time.Unix(currentDayEnd/1000, 0).Format("2006-01-02 15:04:05"))
				continue
			}

			// 显示拉取成功信息
			firstTime := time.Unix(klines[0].OpenTime/1000, 0).Format("2006-01-02 15:04:05")
			lastTimeStr := time.Unix(klines[len(klines)-1].OpenTime/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[%s] ✓ 成功拉取 %d 条数据 (时间范围: %s ~ %s)", symbol, len(klines), firstTime, lastTimeStr)

			// 打印第一条完整数据（拉取阶段）
			if len(klines) > 0 {
				firstKline := klines[0]
				firstOpenTimeStr := time.Unix(firstKline.OpenTime/1000, 0).Format("2006-01-02 15:04:05")
				firstCloseTimeStr := time.Unix(firstKline.CloseTime/1000, 0).Format("2006-01-02 15:04:05")
				logger.Infof("[%s] 第一条数据详情: open_time=%s, close_time=%s, open=%.8f, high=%.8f, low=%.8f, close=%.8f, volume=%.8f",
					symbol, firstOpenTimeStr, firstCloseTimeStr, firstKline.Open, firstKline.High, firstKline.Low, firstKline.Close, firstKline.Volume)
			}

			// 保存到数据库
			logger.Infof("[%s] 开始保存 %d 条数据到数据库...", symbol, len(klines))
			if err := database.SaveKLine1m(klines); err != nil {
				logger.Errorf("[%s] ❌ 保存K线数据失败: %v", symbol, err)
				return fmt.Errorf("保存K线数据失败: %w", err)
			}
			logger.Infof("[%s] ✓ 数据保存完成", symbol)

			// 更新统计
			totalFetched += len(klines)
			rangeFetched += len(klines)

			// 更新同步状态
			lastKlineTime := klines[len(klines)-1].CloseTime
			if err := database.UpdateSyncStatus(symbol, time.Now().UnixMilli(), lastKlineTime); err != nil {
				return fmt.Errorf("更新同步状态失败: %w", err)
			}

			// 检查是否已经获取完当天的所有数据
			firstKlineTime := klines[0].OpenTime

			// 重新计算当天的开始时间（用于判断是否已获取到00:00:00）
			currentDayForCheck := time.Unix(firstKlineTime/1000, 0)
			dayStartForCheck := time.Date(currentDayForCheck.Year(), currentDayForCheck.Month(), currentDayForCheck.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()

			// 如果第一条数据的时间戳 <= 当天的开始时间，说明已经获取到当天的开始（00:00:00），跳到前一天
			// 或者如果返回的数据少于batchSize，说明已经拉完当天的数据，跳到前一天
			if firstKlineTime <= dayStartForCheck || len(klines) < batchSize {
				// 当天的数据已经获取完毕，跳到前一天
				prevDay := currentDayForCheck.AddDate(0, 0, -1)
				prevDayEnd := time.Date(prevDay.Year(), prevDay.Month(), prevDay.Day(), 23, 59, 59, 999, time.UTC).UnixMilli()

				// 如果已经到达或超过起始时间，停止这个时间段
				if prevDayEnd < missingRange.StartTime {
					break
				}

				// 更新 currentDayEnd 为前一天
				currentDayEnd = prevDayEnd
				logger.Infof("[%s] 当天数据获取完毕（已获取到 %s），跳到前一天: %s",
					symbol,
					time.Unix(firstKlineTime/1000, 0).Format("2006-01-02 15:04:05"),
					time.Unix(currentDayEnd/1000, 0).Format("2006-01-02 15:04:05"))
			} else {
				// 当天的数据还没获取完，继续从第一条数据的前一分钟开始
				currentDayEnd = firstKlineTime - 1

				// 如果已经到达或超过起始时间，停止这个时间段
				if currentDayEnd < missingRange.StartTime {
					break
				}
			}

			// 避免请求过快，稍作延迟
			time.Sleep(200 * time.Millisecond)
		}

		// 记录已同步的时间段
		if rangeFetched > 0 {
			// 计算实际同步的时间范围
			actualStart := missingRange.StartTime
			actualEnd := missingRange.EndTime
			if currentDayEnd >= missingRange.StartTime {
				actualStart = currentDayEnd + 1
			}
			if err := database.AddSyncTimeRange(symbol, actualStart, actualEnd); err != nil {
				logger.Warnf("[%s] 记录同步时间段失败: %v", symbol, err)
			} else {
				logger.Infof("[%s] ✓ 时间段 #%d 同步完成，获取 %d 条数据，已记录时间段", symbol, rangeIdx+1, rangeFetched)
			}
		}
	}

	// 显示同步完成信息
	if totalFetched > 0 {
		logger.Infof("[%s] ✓ 历史数据倒推同步完成，共获取 %d 条数据", symbol, totalFetched)
	} else {
		logger.Debugf("[%s] 历史数据倒推同步完成，无新数据", symbol)
	}

	return nil
}

// SyncSymbolInitial 首次同步（拉取历史数据）
func SyncSymbolInitial(symbol string, days int) error {
	api := NewExchangeAPI()

	// 计算时间范围
	now := time.Now().UnixMilli()
	startTime := now - int64(days*24*60*60*1000) // days天前

	// 分批拉取（每次只拉取少量数据，逐步推进）
	// Gate.io限制：最多只能获取最近 10000 个数据点（约6.9天）
	// 策略：每次只拉取 300 条，逐步推进
	batchSize := 300 // 每次只拉取300条
	currentStart := startTime

	for {
		// 如果已经到达当前时间，停止
		if currentStart >= now {
			break
		}

		// 检查是否超过 Gate.io 的限制（最多10000个数据点之前）
		maxPointsAgo := int64(10000 * 60 * 1000) // 10000分钟的毫秒数
		minAllowedTime := now - maxPointsAgo

		// 如果起始时间太早，调整到允许的最早时间
		if currentStart < minAllowedTime {
			currentStart = minAllowedTime
		}

		// 如果已经超过限制，停止同步历史数据
		if currentStart >= now {
			break
		}

		// 从交易所拉取数据（每次300条，不设置endTime，让API返回到当前时间）
		klines, err := api.FetchKLines(symbol, "1m", currentStart, 0, batchSize)
		if err != nil {
			// 如果是时间太早的错误，说明已经无法获取更早的数据
			if strings.Contains(err.Error(), "too long ago") || strings.Contains(err.Error(), "10000 points") {
				// 只能获取最近的数据，停止同步更早的历史数据
				break
			}
			// 如果是时间范围错误，尝试不设置startTime，只拉取最近的数据
			if strings.Contains(err.Error(), "range too broad") || strings.Contains(err.Error(), "INVALID_PARAM_VALUE") {
				// 不设置startTime，只拉取最近的数据
				klines, err = api.FetchKLines(symbol, "1m", 0, 0, batchSize)
				if err != nil {
					return fmt.Errorf("拉取K线数据失败: %w", err)
				}
			} else {
				return fmt.Errorf("拉取K线数据失败: %w", err)
			}
		}

		if len(klines) == 0 {
			// 没有新数据，可能已经同步到最新
			break
		}

		// 保存到数据库
		if err := database.SaveKLine1m(klines); err != nil {
			return fmt.Errorf("保存K线数据失败: %w", err)
		}

		// 更新进度
		lastKlineTime := klines[len(klines)-1].CloseTime
		if err := database.UpdateSyncStatus(symbol, time.Now().UnixMilli(), lastKlineTime); err != nil {
			return fmt.Errorf("更新同步状态失败: %w", err)
		}

		// 如果返回的数据少于batchSize，说明已经拉完这个时间段
		if len(klines) < batchSize {
			// 检查是否还有更多数据需要拉取
			if lastKlineTime < now {
				currentStart = lastKlineTime + 1
				continue
			}
			break
		}

		// 下一批从最后一条的下一条开始
		currentStart = lastKlineTime + 1

		// 避免请求过快，稍作延迟
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

// SyncSymbolHistorical 同步历史数据（从指定年份开始）
func SyncSymbolHistorical(symbol string, startYear int) error {
	api := NewExchangeAPI()

	// 计算起始时间
	startTime := time.Date(startYear, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	now := time.Now().UnixMilli()

	// 获取本地最新K线时间
	lastTime, err := database.GetLatestKLineTime(symbol)
	if err != nil {
		return fmt.Errorf("获取最新K线时间失败: %w", err)
	}

	// 如果已有数据，从最后一条的下一条开始
	if lastTime > 0 && lastTime > startTime {
		startTime = lastTime + 1
	}

	// 如果已经同步到最新，直接返回
	if startTime >= now {
		return nil
	}

	// 显示同步开始信息
	startTimeStr := time.Unix(startTime/1000, 0).Format("2006-01-02 15:04:05")
	logger.Infof("开始同步 %s [历史数据] 从 %s 开始 (年份: %d)", symbol, startTimeStr, startYear)

	// 分批拉取（每次只拉取少量数据，逐步推进）
	// Gate.io限制：最多只能获取最近 10000 个数据点（约6.9天）
	// 策略：每次只拉取 300 条，逐步推进
	batchSize := 300 // 每次只拉取300条
	currentStart := startTime
	totalFetched := 0
	batchCount := 0

	for {
		// 如果已经到达当前时间，停止
		if currentStart >= now {
			break
		}

		// 检查是否超过 Gate.io 的限制（最多10000个数据点之前）
		maxPointsAgo := int64(10000 * 60 * 1000) // 10000分钟的毫秒数
		minAllowedTime := now - maxPointsAgo

		// 如果起始时间太早，调整到允许的最早时间
		if currentStart < minAllowedTime {
			currentStart = minAllowedTime
		}

		// 如果已经超过限制，停止同步历史数据
		if currentStart >= now {
			break
		}

		// 显示当前批次信息
		batchCount++
		currentStartStr := time.Unix(currentStart/1000, 0).Format("2006-01-02 15:04:05")
		logger.Infof("[%s] 批次 #%d: 从 %s 开始拉取 (最多 %d 条)", symbol, batchCount, currentStartStr, batchSize)

		// 从交易所拉取数据（每次300条，不设置endTime，让API返回到当前时间）
		logger.Infof("[%s] 正在请求API: startTime=%d (from=%s), limit=%d", symbol, currentStart, currentStartStr, batchSize)
		klines, err := api.FetchKLines(symbol, "1m", currentStart, 0, batchSize)
		if err != nil {
			logger.Errorf("[%s] ❌ API请求失败: %v", symbol, err)
			// 如果是时间太早的错误，说明已经无法获取更早的数据
			if strings.Contains(err.Error(), "too long ago") || strings.Contains(err.Error(), "10000 points") {
				logger.Warnf("[%s] 无法获取更早的数据（超过10000个数据点限制），停止同步", symbol)
				// 只能获取最近的数据，停止同步更早的历史数据
				break
			}
			// 如果是时间范围错误，尝试不设置startTime，只拉取最近的数据
			if strings.Contains(err.Error(), "range too broad") || strings.Contains(err.Error(), "INVALID_PARAM_VALUE") {
				logger.Infof("[%s] 时间范围错误，尝试拉取最近的数据（不设置startTime）", symbol)
				// 不设置startTime，只拉取最近的数据
				klines, err = api.FetchKLines(symbol, "1m", 0, 0, batchSize)
				if err != nil {
					logger.Errorf("[%s] ❌ 重试拉取最近数据也失败: %v", symbol, err)
					return fmt.Errorf("拉取K线数据失败: %w", err)
				}
				logger.Infof("[%s] ✓ 重试成功，拉取到最近的数据", symbol)
			} else {
				logger.Errorf("[%s] ❌ 拉取K线数据失败（未知错误）: %v", symbol, err)
				return fmt.Errorf("拉取K线数据失败: %w", err)
			}
		}

		if len(klines) == 0 {
			// 没有新数据，可能已经同步到最新
			logger.Infof("[%s] ⚠️ API返回空数据，可能已经同步到最新或没有数据", symbol)
			break
		}

		// 显示拉取成功信息
		firstTime := time.Unix(klines[0].OpenTime/1000, 0).Format("2006-01-02 15:04:05")
		lastTimeStr := time.Unix(klines[len(klines)-1].OpenTime/1000, 0).Format("2006-01-02 15:04:05")
		logger.Infof("[%s] ✓ 成功拉取 %d 条数据 (时间范围: %s ~ %s)", symbol, len(klines), firstTime, lastTimeStr)

		// 打印第一条完整数据（拉取阶段）
		if len(klines) > 0 {
			firstKline := klines[0]
			firstOpenTimeStr := time.Unix(firstKline.OpenTime/1000, 0).Format("2006-01-02 15:04:05")
			firstCloseTimeStr := time.Unix(firstKline.CloseTime/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[%s] 第一条数据详情: open_time=%s, close_time=%s, open=%.8f, high=%.8f, low=%.8f, close=%.8f, volume=%.8f",
				symbol, firstOpenTimeStr, firstCloseTimeStr, firstKline.Open, firstKline.High, firstKline.Low, firstKline.Close, firstKline.Volume)
		}

		// 保存到数据库
		logger.Infof("[%s] 开始保存 %d 条数据到数据库...", symbol, len(klines))
		if err := database.SaveKLine1m(klines); err != nil {
			logger.Errorf("[%s] ❌ 保存K线数据失败: %v", symbol, err)
			return fmt.Errorf("保存K线数据失败: %w", err)
		}
		logger.Infof("[%s] ✓ 数据保存完成", symbol)

		// 更新统计
		totalFetched += len(klines)

		// 更新进度
		lastKlineTime := klines[len(klines)-1].CloseTime
		if err := database.UpdateSyncStatus(symbol, time.Now().UnixMilli(), lastKlineTime); err != nil {
			return fmt.Errorf("更新同步状态失败: %w", err)
		}

		// 如果返回的数据少于batchSize，说明已经拉完这个时间段
		if len(klines) < batchSize {
			// 检查是否还有更多数据需要拉取
			if lastKlineTime < now {
				currentStart = lastKlineTime + 1
				continue
			}
			break
		}

		// 下一批从最后一条的下一条开始
		currentStart = lastKlineTime + 1

		// 避免请求过快，稍作延迟
		time.Sleep(200 * time.Millisecond)
	}

	// 显示同步完成信息
	if totalFetched > 0 {
		logger.Infof("[%s] ✓ 历史数据同步完成，共获取 %d 条数据", symbol, totalFetched)
	} else {
		logger.Debugf("[%s] 历史数据同步完成，无新数据", symbol)
	}

	return nil
}
