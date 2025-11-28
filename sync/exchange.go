package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"wails-contract-warn/database"
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

	url := fmt.Sprintf("%s/spot/candlesticks", api.BaseURL)

	// 构建查询参数
	params := make(map[string]string)
	params["currency_pair"] = symbol // Gate.io使用 BTC_USDT 格式
	params["interval"] = interval

	if startTime > 0 {
		params["from"] = strconv.FormatInt(startTime/1000, 10) // Gate.io使用秒级时间戳
	}
	if endTime > 0 {
		params["to"] = strconv.FormatInt(endTime/1000, 10)
	}
	if limit > 0 && limit <= 1000 {
		params["limit"] = strconv.Itoa(limit)
	} else if limit > 1000 {
		params["limit"] = "1000" // Gate.io最大限制1000
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

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(body))
	}

	// Gate.io返回格式: [["timestamp", "volume", "close", "high", "low", "open", "base_volume"], ...]
	var klinesRaw [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&klinesRaw); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	var klines []database.KLine1m
	for _, k := range klinesRaw {
		if len(k) < 7 {
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
		default:
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
	}

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

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误: %s, 响应: %s", resp.Status, string(body))
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

	// 1. 获取本地最新K线时间
	lastTime, err := database.GetLatestKLineTime(symbol)
	if err != nil {
		return fmt.Errorf("获取最新K线时间失败: %w", err)
	}

	// 2. 计算需要拉取的时间范围
	now := time.Now().UnixMilli()
	var startTime int64

	if priorityRecent {
		// 优先模式：优先获取昨日至今的数据
		if lastTime == 0 {
			// 如果没有数据，从昨日开始拉取
			yesterday := now - 24*60*60*1000
			startTime = yesterday
		} else {
			// 从最后一条的下一条开始
			startTime = lastTime + 1
		}
	} else {
		// 历史模式：从指定时间开始拉取
		if lastTime == 0 {
			// 如果没有数据，从2020年开始
			start2020 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
			startTime = start2020
		} else {
			// 继续从上次停止的地方开始
			startTime = lastTime + 1
		}
	}

	// 3. 从交易所拉取1分钟数据
	batchSize := 1000
	klines, err := api.FetchKLines(symbol, "1m", startTime, now, batchSize)
	if err != nil {
		return fmt.Errorf("拉取K线数据失败: %w", err)
	}

	if len(klines) == 0 {
		return nil // 没有新数据
	}

	// 4. 保存到数据库
	if err := database.SaveKLine1m(klines); err != nil {
		return fmt.Errorf("保存K线数据失败: %w", err)
	}

	// 5. 更新同步状态
	lastKlineTime := klines[len(klines)-1].CloseTime
	if err := database.UpdateSyncStatus(symbol, now, lastKlineTime); err != nil {
		return fmt.Errorf("更新同步状态失败: %w", err)
	}

	return nil
}

// SyncSymbolInitial 首次同步（拉取历史数据）
func SyncSymbolInitial(symbol string, days int) error {
	api := NewExchangeAPI()

	// 计算时间范围
	now := time.Now().UnixMilli()
	startTime := now - int64(days*24*60*60*1000) // days天前

	// Gate.io限制每次最多1000根，需要分批拉取
	batchSize := 1000
	currentStart := startTime

	for {
		klines, err := api.FetchKLines(symbol, "1m", currentStart, now, batchSize)
		if err != nil {
			return fmt.Errorf("拉取K线数据失败: %w", err)
		}

		if len(klines) == 0 {
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

		// 如果返回的数据少于batchSize，说明已经拉完
		if len(klines) < batchSize {
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

	// 分批拉取
	batchSize := 1000
	currentStart := startTime

	for {
		// 如果已经到达当前时间，停止
		if currentStart >= now {
			break
		}

		klines, err := api.FetchKLines(symbol, "1m", currentStart, now, batchSize)
		if err != nil {
			return fmt.Errorf("拉取K线数据失败: %w", err)
		}

		if len(klines) == 0 {
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

		// 如果返回的数据少于batchSize，说明已经拉完
		if len(klines) < batchSize {
			break
		}

		// 下一批从最后一条的下一条开始
		currentStart = lastKlineTime + 1

		// 避免请求过快，稍作延迟
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}
