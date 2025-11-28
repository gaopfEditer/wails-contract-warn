package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"wails-contract-warn/api"
	"wails-contract-warn/database"
	"wails-contract-warn/indicator"
	"wails-contract-warn/logger"
	"wails-contract-warn/models"
	"wails-contract-warn/service"
	"wails-contract-warn/signal"
	datasync "wails-contract-warn/sync"
	"wails-contract-warn/utils"
)

// App 结构体（控制器层）
type App struct {
	ctx         context.Context
	market      *service.MarketService
	syncService *service.SyncService
	proxyClient *api.ProxyClient
	dbInit      bool
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{
		market:      service.NewMarketService(),
		syncService: service.NewSyncService(60), // 默认60秒同步一次
		proxyClient: api.NewProxyClient(),
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	logger.Info("应用初始化开始")
	a.ctx = ctx

	logger.Debug("启动市场数据服务")
	a.market.Start()
	logger.Info("市场数据服务已启动")

	// 初始化数据库（如果配置了DSN）
	// 注意：这里需要从配置文件读取DSN，暂时注释
	// if dsn := getDBDSN(); dsn != "" {
	// 	if err := database.InitDB(dsn); err != nil {
	// 		logger.Errorf("数据库初始化失败: %v", err)
	// 	} else {
	// 		a.dbInit = true
	// 		logger.Info("数据库初始化成功")
	// 	}
	// }

	logger.Info("应用初始化完成")
}

// domReady DOM 准备就绪时调用
func (a *App) domReady(ctx context.Context) {
	// 可以在这里执行一些初始化操作
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	logger.Info("应用正在关闭...")

	a.market.Stop()
	logger.Debug("市场数据服务已停止")

	if a.syncService != nil {
		a.syncService.Stop()
		logger.Debug("同步服务已停止")
	}

	if a.dbInit {
		database.CloseDB()
		logger.Debug("数据库连接已关闭")
	}

	logger.Info("应用已关闭")
}

// GetMarketData 获取市场数据（从数据库读取并聚合）
func (a *App) GetMarketData(symbol string, period string) (string, error) {
	logger.Debugf("获取市场数据: symbol=%s, period=%s", symbol, period)

	// 如果数据库已初始化，从数据库读取
	if a.dbInit {
		logger.Debug("从数据库读取市场数据")
		return a.getMarketDataFromDB(symbol, period)
	}

	// 否则使用内存数据（兼容模式）
	logger.Debug("从内存读取市场数据")
	data := a.market.GetKLineData(symbol, period)
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("序列化市场数据失败: %v", err)
		return "", err
	}

	logger.Debugf("成功获取市场数据，共 %d 条记录", len(data))
	return string(jsonData), nil
}

// getMarketDataFromDB 从数据库获取市场数据并聚合
func (a *App) getMarketDataFromDB(symbol string, period string) (string, error) {
	logger.Debugf("从数据库获取市场数据: symbol=%s, period=%s", symbol, period)

	// 1. 解析周期
	targetIntervalMin := utils.ParseIntervalToMinutes(period)
	logger.Debugf("目标周期: %d 分钟", targetIntervalMin)

	// 2. 计算需要多少根1分钟K线（假设需要最近1000根目标周期）
	targetCount := 1000
	needed1mCount := utils.CalculateNeeded1mCount(targetCount, targetIntervalMin)
	logger.Debugf("需要 %d 根1分钟K线", needed1mCount)

	// 3. 从数据库获取1分钟K线
	klines1m, err := database.GetKLines1mByCount(symbol, needed1mCount)
	if err != nil {
		logger.Errorf("从数据库获取K线失败: %v", err)
		return "", err
	}
	logger.Debugf("从数据库获取到 %d 根1分钟K线", len(klines1m))

	// 4. 聚合为目标周期
	klines := utils.AggregateKlines(klines1m, targetIntervalMin)
	logger.Debugf("聚合后得到 %d 根K线", len(klines))

	// 5. 转换为前端需要的格式
	result := make([]models.KLineData, len(klines))
	for i, k := range klines {
		result[i] = models.KLineData{
			Time:   k.OpenTime,
			Open:   k.Open,
			High:   k.High,
			Low:    k.Low,
			Close:  k.Close,
			Volume: k.Volume,
		}
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		logger.Errorf("序列化K线数据失败: %v", err)
		return "", err
	}

	logger.Debugf("成功返回 %d 条K线数据", len(result))
	return string(jsonData), nil
}

// GetIndicators 计算技术指标
func (a *App) GetIndicators(symbol string, period string) (string, error) {
	var klineData []models.KLineData

	// 如果数据库已初始化，从数据库读取
	if a.dbInit {
		targetIntervalMin := utils.ParseIntervalToMinutes(period)
		targetCount := 1000
		needed1mCount := utils.CalculateNeeded1mCount(targetCount, targetIntervalMin)
		klines1m, err := database.GetKLines1mByCount(symbol, needed1mCount)
		if err != nil {
			return "", err
		}
		klines := utils.AggregateKlines(klines1m, targetIntervalMin)
		klineData = make([]models.KLineData, len(klines))
		for i, k := range klines {
			klineData[i] = models.KLineData{
				Time:   k.OpenTime,
				Open:   k.Open,
				High:   k.High,
				Low:    k.Low,
				Close:  k.Close,
				Volume: k.Volume,
			}
		}
	} else {
		klineData = a.market.GetKLineData(symbol, period)
	}

	indicators := indicator.CalculateIndicators(klineData)
	jsonData, err := json.Marshal(indicators)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetAlertSignals 获取预警信号（根据周期重新计算）
func (a *App) GetAlertSignals(symbol string, period string) (string, error) {
	var klineData []models.KLineData

	// 如果数据库已初始化，从数据库读取
	if a.dbInit {
		targetIntervalMin := utils.ParseIntervalToMinutes(period)
		targetCount := 1000
		needed1mCount := utils.CalculateNeeded1mCount(targetCount, targetIntervalMin)
		klines1m, err := database.GetKLines1mByCount(symbol, needed1mCount)
		if err != nil {
			return "", err
		}
		klines := utils.AggregateKlines(klines1m, targetIntervalMin)
		klineData = make([]models.KLineData, len(klines))
		for i, k := range klines {
			klineData[i] = models.KLineData{
				Time:   k.OpenTime,
				Open:   k.Open,
				High:   k.High,
				Low:    k.Low,
				Close:  k.Close,
				Volume: k.Volume,
			}
		}
	} else {
		klineData = a.market.GetKLineData(symbol, period)
	}

	signals := signal.DetectAllSignals(klineData)
	jsonData, err := json.Marshal(signals)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// StartMarketDataStream 开始市场数据流
func (a *App) StartMarketDataStream(symbol string, period string) error {
	return a.market.Subscribe(symbol, period)
}

// StopMarketDataStream 停止市场数据流
func (a *App) StopMarketDataStream(symbol string) error {
	return a.market.Unsubscribe(symbol)
}

// SyncKlineData 同步K线数据（增量）
func (a *App) SyncKlineData(symbol string) (string, error) {
	logger.Infof("开始同步K线数据: symbol=%s", symbol)

	if !a.dbInit {
		logger.Warn("数据库未初始化，无法同步")
		return "", fmt.Errorf("数据库未初始化")
	}

	err := datasync.SyncSymbol(symbol)
	if err != nil {
		logger.Errorf("同步K线数据失败: symbol=%s, error=%v", symbol, err)
		return "", err
	}

	logger.Infof("K线数据同步成功: symbol=%s", symbol)
	return "同步成功", nil
}

// SyncKlineDataInitial 首次同步K线数据（拉取历史）
func (a *App) SyncKlineDataInitial(symbol string, days int) (string, error) {
	logger.Infof("开始初始同步K线数据: symbol=%s, days=%d", symbol, days)

	if !a.dbInit {
		logger.Warn("数据库未初始化，无法同步")
		return "", fmt.Errorf("数据库未初始化")
	}

	err := datasync.SyncSymbolInitial(symbol, days)
	if err != nil {
		logger.Errorf("初始同步K线数据失败: symbol=%s, days=%d, error=%v", symbol, days, err)
		return "", err
	}

	logger.Infof("初始同步K线数据成功: symbol=%s, days=%d", symbol, days)
	return "初始同步成功", nil
}

// InitDatabase 初始化数据库连接
func (a *App) InitDatabase(dsn string) (string, error) {
	logger.Info("正在初始化数据库连接...")
	logger.Debugf("数据库DSN: %s", dsn)

	err := database.InitDB(dsn)
	if err != nil {
		logger.Errorf("数据库初始化失败: %v", err)
		return "", err
	}

	a.dbInit = true
	logger.Info("数据库初始化成功")
	return "数据库初始化成功", nil
}

// StartAutoSync 启动自动同步服务
func (a *App) StartAutoSync(symbol string, intervalSeconds int) (string, error) {
	if !a.dbInit {
		return "", fmt.Errorf("数据库未初始化")
	}

	if a.syncService == nil {
		a.syncService = service.NewSyncService(intervalSeconds)
	}

	a.syncService.AddSymbol(symbol)

	if !a.syncService.IsRunning() {
		a.syncService.Start()
	}

	return "自动同步已启动", nil
}

// StopAutoSync 停止自动同步
func (a *App) StopAutoSync(symbol string) (string, error) {
	if a.syncService != nil {
		a.syncService.RemoveSymbol(symbol)
		return "已停止同步该交易对", nil
	}
	return "同步服务未运行", nil
}

// ProxyAPI 代理 API 请求（用于绕过浏览器的 CORS 和 DNS 限制）
// url: 要请求的完整 URL
// headers: 可选的请求头（JSON 字符串，格式: {"Header-Name": "value"}）
func (a *App) ProxyAPI(url string, headers string) (string, error) {
	logger.Infof("代理 API 请求: %s", url)

	// 解析请求头
	var headerMap map[string]string
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &headerMap); err != nil {
			logger.Warnf("解析请求头失败，使用默认请求头: %v", err)
			headerMap = make(map[string]string)
		}
	} else {
		headerMap = make(map[string]string)
	}

	// 使用代理客户端获取数据
	data, err := a.proxyClient.FetchAPI(url, headerMap)
	if err != nil {
		logger.Errorf("代理请求失败: %v", err)
		return "", err
	}

	// 转换为 JSON 字符串返回
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("序列化响应失败: %v", err)
		return "", err
	}

	logger.Debugf("代理请求成功，响应大小: %d bytes", len(jsonData))
	return string(jsonData), nil
}

// GetMarketPrice 通过后端代理获取市场价格（支持多个交易所）
// exchange: 交易所名称 (coingecko, okx, kraken, gateio, mexc, bitget, binance, bybit)
// symbol: 交易对符号 (如 bitcoin, BTC, BTCUSDT)
func (a *App) GetMarketPrice(exchange string, symbol string) (string, error) {
	logger.Infof("获取市场价格: exchange=%s, symbol=%s", exchange, symbol)

	// 构建 API URL
	var url string
	switch exchange {
	case "coingecko":
		coinId := symbol
		if symbol == "BTC" || symbol == "BTCUSDT" {
			coinId = "bitcoin"
		} else if symbol == "ETH" || symbol == "ETHUSDT" {
			coinId = "ethereum"
		}
		url = fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coinId)

	case "okx":
		pair := symbol
		if pair == "BTC" || pair == "bitcoin" {
			pair = "BTC-USDT"
		} else if !contains(pair, "-") {
			pair = pair + "-USDT"
		}
		url = fmt.Sprintf("https://www.okx.com/api/v5/market/ticker?instId=%s", pair)

	case "kraken":
		pair := symbol
		if pair == "BTC" || pair == "bitcoin" || pair == "BTCUSDT" {
			pair = "XBTUSDT"
		} else if pair == "ETH" || pair == "ethereum" || pair == "ETHUSDT" {
			pair = "ETHUSDT"
		}
		url = fmt.Sprintf("https://api.kraken.com/0/public/Ticker?pair=%s", pair)

	case "gateio":
		pair := symbol
		if !contains(pair, "_") {
			pair = pair + "_USDT"
		}
		url = fmt.Sprintf("https://api.gateio.ws/api/v4/spot/tickers?currency_pair=%s", pair)

	case "mexc":
		pair := symbol
		if pair == "BTC" || pair == "bitcoin" {
			pair = "BTCUSDT"
		}
		url = fmt.Sprintf("https://api.mexc.com/api/v3/ticker/price?symbol=%s", pair)

	case "bitget":
		pair := symbol
		if pair == "BTC" || pair == "bitcoin" {
			pair = "BTCUSDT"
		}
		url = fmt.Sprintf("https://api.bitget.com/api/spot/v1/market/ticker?symbol=%s", pair)

	case "binance":
		pair := symbol
		if pair == "BTC" || pair == "bitcoin" {
			pair = "BTCUSDT"
		}
		url = fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%s", pair)

	case "bybit":
		pair := symbol
		if pair == "BTC" || pair == "bitcoin" {
			pair = "BTCUSDT"
		}
		url = fmt.Sprintf("https://api.bybit.com/v5/market/tickers?category=spot&symbol=%s", pair)

	default:
		return "", fmt.Errorf("不支持的交易所: %s", exchange)
	}

	// 使用代理获取数据
	data, err := a.proxyClient.FetchAPI(url, nil)
	if err != nil {
		logger.Errorf("获取市场价格失败: %v", err)
		return "", err
	}

	// 解析并标准化响应
	result := a.parsePriceResponse(exchange, data)

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// parsePriceResponse 解析不同交易所的价格响应
func (a *App) parsePriceResponse(exchange string, data map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{
		"exchange": exchange,
		"success":  false,
	}

	switch exchange {
	case "coingecko":
		// CoinGecko 返回格式: {"bitcoin": {"usd": 90000}}
		for coinId, priceData := range data {
			if priceMap, ok := priceData.(map[string]interface{}); ok {
				if usd, ok := priceMap["usd"].(float64); ok {
					result["price"] = usd
					result["success"] = true
					result["coinId"] = coinId
					break
				}
			}
		}

	case "okx":
		// OKX 返回格式: {"code":"0","data":[{"last":"90000",...}]}
		if code, ok := data["code"].(string); ok && code == "0" {
			if dataArr, ok := data["data"].([]interface{}); ok && len(dataArr) > 0 {
				if ticker, ok := dataArr[0].(map[string]interface{}); ok {
					if last, ok := ticker["last"].(string); ok {
						var price float64
						fmt.Sscanf(last, "%f", &price)
						result["price"] = price
						result["success"] = true
					}
				}
			}
		}

	case "kraken":
		// Kraken 返回格式: {"result": {"XBTUSDT": {"c": ["90000",...]}}}
		if resultMap, ok := data["result"].(map[string]interface{}); ok {
			for pair, tickerData := range resultMap {
				if ticker, ok := tickerData.(map[string]interface{}); ok {
					if c, ok := ticker["c"].([]interface{}); ok && len(c) > 0 {
						if priceStr, ok := c[0].(string); ok {
							var price float64
							fmt.Sscanf(priceStr, "%f", &price)
							result["price"] = price
							result["success"] = true
							result["pair"] = pair
							break
						}
					}
				}
			}
		}

	case "gateio":
		// Gate.io 返回格式: [{"last":"90000",...}]
		if dataArr, ok := data["raw"].(string); ok {
			// 如果是原始字符串，尝试解析
			var arr []interface{}
			if err := json.Unmarshal([]byte(dataArr), &arr); err == nil && len(arr) > 0 {
				if ticker, ok := arr[0].(map[string]interface{}); ok {
					if last, ok := ticker["last"].(string); ok {
						var price float64
						fmt.Sscanf(last, "%f", &price)
						result["price"] = price
						result["success"] = true
					}
				}
			}
		} else if dataArr, ok := data["data"].([]interface{}); ok && len(dataArr) > 0 {
			if ticker, ok := dataArr[0].(map[string]interface{}); ok {
				if last, ok := ticker["last"].(string); ok {
					var price float64
					fmt.Sscanf(last, "%f", &price)
					result["price"] = price
					result["success"] = true
				}
			}
		}

	case "mexc", "binance":
		// MEXC/Binance 返回格式: {"price": "90000"}
		if priceStr, ok := data["price"].(string); ok {
			var price float64
			fmt.Sscanf(priceStr, "%f", &price)
			result["price"] = price
			result["success"] = true
		}

	case "bitget":
		// Bitget 返回格式: {"code":"00000","data":{"close":"90000",...}}
		if code, ok := data["code"].(string); ok && code == "00000" {
			if dataMap, ok := data["data"].(map[string]interface{}); ok {
				if close, ok := dataMap["close"].(string); ok {
					var price float64
					fmt.Sscanf(close, "%f", &price)
					result["price"] = price
					result["success"] = true
				}
			}
		}

	case "bybit":
		// Bybit 返回格式: {"retCode":0,"result":{"list":[{"lastPrice":"90000",...}]}}
		if retCode, ok := data["retCode"].(float64); ok && retCode == 0 {
			if resultMap, ok := data["result"].(map[string]interface{}); ok {
				if list, ok := resultMap["list"].([]interface{}); ok && len(list) > 0 {
					if ticker, ok := list[0].(map[string]interface{}); ok {
						if lastPrice, ok := ticker["lastPrice"].(string); ok {
							var price float64
							fmt.Sscanf(lastPrice, "%f", &price)
							result["price"] = price
							result["success"] = true
						}
					}
				}
			}
		}
	}

	return result
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetNetworkLogs 获取网络连接日志（用于终端显示）
// 这里返回模拟数据，实际应该从日志系统或文件读取
func (a *App) GetNetworkLogs(limit int) (string, error) {
	logger.Debugf("获取网络日志，限制: %d 条", limit)

	// TODO: 实际实现中应该从日志文件或日志系统读取
	// 这里返回示例数据
	logs := []map[string]interface{}{
		{
			"time":    time.Now().UnixMilli() - 10000,
			"content": "2025/11/28 09:48:57.802245 from tcp:127.0.0.1:65295 accepted tcp:104.128.62.173:443 [socks >> proxy]",
			"type":    "info",
		},
		{
			"time":    time.Now().UnixMilli() - 5000,
			"content": "+0800 2025-11-28 09:49:54 ERROR [94688961 4m8s] connection: connection download closed",
			"type":    "error",
		},
	}

	jsonData, err := json.Marshal(logs)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetAlerts 获取预警信息列表
func (a *App) GetAlerts(limit int) (string, error) {
	logger.Debugf("获取预警信息，限制: %d 条", limit)

	// TODO: 实际实现中应该从数据库或内存中读取预警
	// 这里返回示例数据
	alerts := []map[string]interface{}{
		{
			"time":    time.Now().UnixMilli() - 30000,
			"message": "检测到价格突破阻力位",
			"level":   "warn",
			"symbol":  "BTCUSDT",
			"period":  "1m",
		},
		{
			"time":    time.Now().UnixMilli() - 20000,
			"message": "RSI 指标超买",
			"level":   "warn",
			"symbol":  "BTCUSDT",
			"period":  "5m",
		},
		{
			"time":    time.Now().UnixMilli() - 10000,
			"message": "MACD 金叉信号",
			"level":   "info",
			"symbol":  "ETHUSDT",
			"period":  "15m",
		},
	}

	jsonData, err := json.Marshal(alerts)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// LoadTestData 从测试文件加载 K 线数据
// filename: 测试数据文件名（如 "test1.json"）
func (a *App) LoadTestData(filename string) (string, error) {
	logger.Infof("加载测试数据: %s", filename)

	// 构建文件路径
	filePath := fmt.Sprintf("data/%s", filename)

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		logger.Errorf("读取测试数据文件失败: %v", err)
		return "", fmt.Errorf("读取测试数据文件失败: %w", err)
	}

	// 解析 JSON
	var klines []models.KLineData
	if err := json.Unmarshal(data, &klines); err != nil {
		logger.Errorf("解析测试数据 JSON 失败: %v", err)
		return "", fmt.Errorf("解析测试数据 JSON 失败: %w", err)
	}

	logger.Infof("成功加载测试数据: %d 条 K 线", len(klines))

	// 返回 JSON 字符串
	jsonData, err := json.Marshal(klines)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// TestBollingerHammerAlert 使用测试数据测试布林带上轨+锤子形态预警
// filename: 测试数据文件名
func (a *App) TestBollingerHammerAlert(filename string) (string, error) {
	logger.Infof("测试布林带上轨+锤子形态预警: %s", filename)

	// 1. 加载测试数据
	dataStr, err := a.LoadTestData(filename)
	if err != nil {
		return "", err
	}

	var klines []models.KLineData
	if err := json.Unmarshal([]byte(dataStr), &klines); err != nil {
		return "", fmt.Errorf("解析 K 线数据失败: %w", err)
	}

	if len(klines) == 0 {
		return "", fmt.Errorf("测试数据为空")
	}

	logger.Debugf("加载了 %d 条 K 线数据", len(klines))

	// 2. 计算技术指标
	indicators := indicator.CalculateIndicators(klines)

	// 3. 检测预警信号
	signals := signal.DetectAllSignals(klines)

	// 4. 筛选布林带上轨+锤子形态的信号
	bollingerHammerSignals := []map[string]interface{}{}

	for _, sig := range signals {
		// 检查是否是布林带上轨相关的信号
		if sig.UpperBand > 0 && sig.Close >= sig.UpperBand*0.98 { // 接近或超过上轨（98%以上）
			// 检查是否是锤子形态
			idx := sig.Index
			if idx < len(klines) {
				k := klines[idx]
				// 锤子形态判断：下影线长度 > 实体长度 * 2，上影线很短
				body := math.Abs(k.Close - k.Open)
				upperShadow := k.High - math.Max(k.Open, k.Close)
				lowerShadow := math.Min(k.Open, k.Close) - k.Low

				// 锤子形态条件
				isHammer := lowerShadow > body*2 && upperShadow < body*0.5

				if isHammer {
					bollingerHammerSignals = append(bollingerHammerSignals, map[string]interface{}{
						"index":     sig.Index,
						"time":      sig.Time,
						"price":     sig.Price,
						"close":     sig.Close,
						"upperBand": sig.UpperBand,
						"type":      "布林带上轨+锤子形态",
						"strength":  sig.Strength,
						"kline": map[string]interface{}{
							"open":  k.Open,
							"high":  k.High,
							"low":   k.Low,
							"close": k.Close,
						},
						"analysis": map[string]interface{}{
							"body":        body,
							"upperShadow": upperShadow,
							"lowerShadow": lowerShadow,
							"isHammer":    isHammer,
							"bandRatio":   sig.Close / sig.UpperBand, // 价格与上轨的比率
						},
					})

					logger.Infof("检测到布林带上轨+锤子形态信号: index=%d, time=%d, close=%.2f, upperBand=%.2f",
						sig.Index, sig.Time, sig.Close, sig.UpperBand)
				}
			}
		}
	}

	// 5. 构建结果
	result := map[string]interface{}{
		"totalKlines":            len(klines),
		"totalSignals":           len(signals),
		"bollingerHammerSignals": len(bollingerHammerSignals),
		"signals":                bollingerHammerSignals,
		"indicators": map[string]interface{}{
			"hasBBUpper":   len(indicators.BBUpper) > 0,
			"bbUpperCount": len(indicators.BBUpper),
		},
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	logger.Infof("测试完成: 共检测到 %d 个布林带上轨+锤子形态信号", len(bollingerHammerSignals))
	return string(jsonData), nil
}

// AnalyzeTestData 分析测试数据并返回完整结果（包括所有信号类型）
// filename: 测试数据文件名
func (a *App) AnalyzeTestData(filename string) (string, error) {
	logger.Infof("分析测试数据: %s", filename)

	// 1. 加载测试数据
	dataStr, err := a.LoadTestData(filename)
	if err != nil {
		return "", err
	}

	var klines []models.KLineData
	if err := json.Unmarshal([]byte(dataStr), &klines); err != nil {
		return "", fmt.Errorf("解析 K 线数据失败: %w", err)
	}

	if len(klines) == 0 {
		return "", fmt.Errorf("测试数据为空")
	}

	logger.Debugf("加载了 %d 条 K 线数据", len(klines))

	// 2. 计算技术指标
	indicators := indicator.CalculateIndicators(klines)

	// 3. 检测所有预警信号
	allSignals := signal.DetectAllSignals(klines)

	// 4. 按信号类型分类统计
	signalStats := make(map[string]int)
	signalDetails := make(map[string][]map[string]interface{})

	for _, sig := range allSignals {
		signalType := sig.Type
		signalStats[signalType]++

		// 获取K线详情
		var klineInfo map[string]interface{}
		if sig.Index < len(klines) {
			k := klines[sig.Index]
			body := math.Abs(k.Close - k.Open)
			upperShadow := k.High - math.Max(k.Open, k.Close)
			lowerShadow := math.Min(k.Open, k.Close) - k.Low

			klineInfo = map[string]interface{}{
				"open":        k.Open,
				"high":        k.High,
				"low":         k.Low,
				"close":       k.Close,
				"volume":      k.Volume,
				"body":        body,
				"upperShadow": upperShadow,
				"lowerShadow": lowerShadow,
			}
		}

		signalDetail := map[string]interface{}{
			"index":     sig.Index,
			"time":      sig.Time,
			"price":     sig.Price,
			"close":     sig.Close,
			"type":      sig.Type,
			"strength":  sig.Strength,
			"upperBand": sig.UpperBand,
			"lowerBand": sig.LowerBand,
			"kline":     klineInfo,
		}

		if signalDetails[signalType] == nil {
			signalDetails[signalType] = []map[string]interface{}{}
		}
		signalDetails[signalType] = append(signalDetails[signalType], signalDetail)
	}

	// 5. 构建完整结果
	result := map[string]interface{}{
		"totalKlines":   len(klines),
		"totalSignals":  len(allSignals),
		"signalStats":   signalStats,
		"signalDetails": signalDetails,
		"klines":        klines,
		"allSignals":    allSignals,
		"indicators":    indicators,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	logger.Infof("分析完成: 共检测到 %d 个信号，分布在 %d 种类型", len(allSignals), len(signalStats))
	return string(jsonData), nil
}
