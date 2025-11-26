package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wails-contract-warn/database"
)

// ExchangeAPI 交易所API接口
type ExchangeAPI struct {
	BaseURL string
}

// NewExchangeAPI 创建交易所API实例
func NewExchangeAPI() *ExchangeAPI {
	return &ExchangeAPI{
		BaseURL: "https://api.binance.com/api/v3",
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

	resp, err := http.Get(url)
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

// SyncSymbol 同步指定交易对的K线数据
func SyncSymbol(symbol string) error {
	api := NewExchangeAPI()

	// 1. 获取本地最新K线时间
	lastTime, err := database.GetLatestKLineTime(symbol)
	if err != nil {
		return fmt.Errorf("获取最新K线时间失败: %w", err)
	}

	// 2. 计算需要拉取的时间范围
	now := time.Now().UnixMilli()
	startTime := lastTime + 1 // 从下一条开始

	// 如果本地没有数据，拉取最近1000根（Binance限制）
	if lastTime == 0 {
		startTime = now - 1000*60*1000 // 1000分钟前
	}

	// 3. 从交易所拉取1分钟数据
	klines, err := api.FetchKLines(symbol, "1m", startTime, now, 1000)
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

	// Binance限制每次最多1000根，需要分批拉取
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
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
