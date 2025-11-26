package main

import (
	"context"
	"encoding/json"
	"fmt"

	"wails-contract-warn/database"
	"wails-contract-warn/indicator"
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
	dbInit      bool
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{
		market:      service.NewMarketService(),
		syncService: service.NewSyncService(60), // 默认60秒同步一次
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.market.Start()

	// 初始化数据库（如果配置了DSN）
	// 注意：这里需要从配置文件读取DSN，暂时注释
	// if dsn := getDBDSN(); dsn != "" {
	// 	if err := database.InitDB(dsn); err != nil {
	// 		fmt.Printf("数据库初始化失败: %v\n", err)
	// 	} else {
	// 		a.dbInit = true
	// 	}
	// }
}

// domReady DOM 准备就绪时调用
func (a *App) domReady(ctx context.Context) {
	// 可以在这里执行一些初始化操作
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	a.market.Stop()
	if a.syncService != nil {
		a.syncService.Stop()
	}
	if a.dbInit {
		database.CloseDB()
	}
}

// GetMarketData 获取市场数据（从数据库读取并聚合）
func (a *App) GetMarketData(symbol string, period string) (string, error) {
	// 如果数据库已初始化，从数据库读取
	if a.dbInit {
		return a.getMarketDataFromDB(symbol, period)
	}

	// 否则使用内存数据（兼容模式）
	data := a.market.GetKLineData(symbol, period)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// getMarketDataFromDB 从数据库获取市场数据并聚合
func (a *App) getMarketDataFromDB(symbol string, period string) (string, error) {
	// 1. 解析周期
	targetIntervalMin := utils.ParseIntervalToMinutes(period)

	// 2. 计算需要多少根1分钟K线（假设需要最近1000根目标周期）
	targetCount := 1000
	needed1mCount := utils.CalculateNeeded1mCount(targetCount, targetIntervalMin)

	// 3. 从数据库获取1分钟K线
	klines1m, err := database.GetKLines1mByCount(symbol, needed1mCount)
	if err != nil {
		return "", err
	}

	// 4. 聚合为目标周期
	klines := utils.AggregateKlines(klines1m, targetIntervalMin)

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
		return "", err
	}
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
	if !a.dbInit {
		return "", fmt.Errorf("数据库未初始化")
	}

	err := datasync.SyncSymbol(symbol)
	if err != nil {
		return "", err
	}

	return "同步成功", nil
}

// SyncKlineDataInitial 首次同步K线数据（拉取历史）
func (a *App) SyncKlineDataInitial(symbol string, days int) (string, error) {
	if !a.dbInit {
		return "", fmt.Errorf("数据库未初始化")
	}

	err := datasync.SyncSymbolInitial(symbol, days)
	if err != nil {
		return "", err
	}

	return "初始同步成功", nil
}

// InitDatabase 初始化数据库连接
func (a *App) InitDatabase(dsn string) (string, error) {
	err := database.InitDB(dsn)
	if err != nil {
		return "", err
	}
	a.dbInit = true
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
