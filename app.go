package main

import (
	"context"
	"encoding/json"
	"fmt"

	"wails-contract-warn/database"
	"wails-contract-warn/indicator"
	"wails-contract-warn/models"
	"wails-contract-warn/service"
	datasync "wails-contract-warn/sync"
	"wails-contract-warn/signal"
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
	result := make([]KLineData, len(klines))
	for i, k := range klines {
		result[i] = KLineData{
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

	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// Indicators 技术指标
type Indicators struct {
	MA5      []float64 `json:"ma5"`
	MA10     []float64 `json:"ma10"`
	MA20     []float64 `json:"ma20"`
	MACD     []float64 `json:"macd"`
	Signal   []float64 `json:"signal"`
	Hist     []float64 `json:"hist"`
	BBUpper  []float64 `json:"bbUpper"`
	BBMiddle []float64 `json:"bbMiddle"`
	BBLower  []float64 `json:"bbLower"`
}

// AlertSignal 预警信号
type AlertSignal struct {
	Index     int     `json:"index"`
	Time      int64   `json:"time"`
	Price     float64 `json:"price"`
	Close     float64 `json:"close"`
	LowerBand float64 `json:"lowerBand,omitempty"`
	UpperBand float64 `json:"upperBand,omitempty"`
	Type      string  `json:"type"`               // 信号类型：bollinger_doji_bottom, hanging_man, engulfing, hammer, consecutive_hammers
	Strength  float64 `json:"strength,omitempty"` // 信号强度 0-1
}

// MarketService 市场数据服务
type MarketService struct {
	mu          sync.RWMutex
	data        map[string][]KLineData
	subscribers map[string]bool
	running     bool
	stopChan    chan struct{}
}

// NewMarketService 创建市场服务
func NewMarketService() *MarketService {
	return &MarketService{
		data:        make(map[string][]KLineData),
		subscribers: make(map[string]bool),
		stopChan:    make(chan struct{}),
	}
}

// Start 启动市场服务
func (m *MarketService) Start() {
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	// 初始化一些示例数据
	m.initSampleData()

	// 启动数据更新循环
	go m.updateLoop()
}

// Stop 停止市场服务
func (m *MarketService) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		m.running = false
		close(m.stopChan)
	}
}

// initSampleData 初始化示例数据
func (m *MarketService) initSampleData() {
	// 生成示例 K 线数据
	now := time.Now()
	var data []KLineData
	basePrice := 50000.0

	for i := 0; i < 100; i++ {
		timestamp := now.Add(-time.Duration(100-i)*time.Minute).Unix() * 1000
		change := (math.Sin(float64(i)/10) + math.Cos(float64(i)/7)) * 100
		open := basePrice + change
		close := open + (math.Sin(float64(i)/5) * 50)
		high := math.Max(open, close) + math.Abs(math.Sin(float64(i)/3)*30)
		low := math.Min(open, close) - math.Abs(math.Cos(float64(i)/4)*30)
		volume := 1000 + math.Abs(math.Sin(float64(i)/6)*500)

		data = append(data, KLineData{
			Time:   timestamp,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
		basePrice = close
	}

	m.mu.Lock()
	m.data["BTCUSDT"] = data
	m.mu.Unlock()
}

// updateLoop 数据更新循环
func (m *MarketService) updateLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.mu.RLock()
			running := m.running
			m.mu.RUnlock()

			if !running {
				return
			}

			// 更新数据
			m.updateData()
		}
	}
}

// updateData 更新市场数据
func (m *MarketService) updateData() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 模拟实时数据更新
	for symbol := range m.subscribers {
		if data, ok := m.data[symbol]; ok && len(data) > 0 {
			last := data[len(data)-1]
			now := time.Now().Unix() * 1000

			// 更新最后一根 K 线或创建新的
			if now-last.Time < 60000 { // 1分钟内更新
				change := (math.Sin(float64(time.Now().Unix())/10) * 20)
				last.Close += change
				last.High = math.Max(last.High, last.Close)
				last.Low = math.Min(last.Low, last.Close)
				last.Volume += math.Abs(change) * 10
				data[len(data)-1] = last
			} else {
				// 创建新 K 线
				newKLine := KLineData{
					Time:   now,
					Open:   last.Close,
					High:   last.Close + math.Abs(math.Sin(float64(now)/1000)*30),
					Low:    last.Close - math.Abs(math.Cos(float64(now)/1000)*30),
					Close:  last.Close + (math.Sin(float64(now)/1000) * 20),
					Volume: 1000 + math.Abs(math.Sin(float64(now)/1000)*500),
				}
				data = append(data, newKLine)
				// 保持最多 200 根 K 线
				if len(data) > 200 {
					data = data[len(data)-200:]
				}
			}
			m.data[symbol] = data
		}
	}
}

// GetKLineData 获取 K 线数据
func (m *MarketService) GetKLineData(symbol string, period string) []KLineData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[symbol]
}

// Subscribe 订阅市场数据
func (m *MarketService) Subscribe(symbol string, period string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers[symbol] = true
	return nil
}

// Unsubscribe 取消订阅
func (m *MarketService) Unsubscribe(symbol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscribers, symbol)
	return nil
}

// CalculateIndicators 计算技术指标
func CalculateIndicators(data []KLineData) Indicators {
	if len(data) == 0 {
		return Indicators{}
	}

	indicators := Indicators{
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

	// 计算 MACD
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

	// 计算布林带
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

	return indicators
}

// ==================== K线形态检测函数 ====================

// isDoji 判断是否为十字星
func isDoji(candle KLineData, threshold float64) bool {
	if candle.High == candle.Low {
		return false
	}

	body := math.Abs(candle.Close - candle.Open)
	range_ := candle.High - candle.Low

	// 实体很小，且有明显影线
	return range_ > 0 && body/candle.Open < threshold && range_ > body*2
}

// isHammer 判断是否为锤子线（看涨）
func isHammer(candle KLineData) bool {
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

// isHangingMan 判断是否为吊颈线（看跌）
func isHangingMan(candle KLineData) bool {
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

// isEngulfing 判断是否为吞没形态
func isEngulfing(prev, curr KLineData) (bool, bool) {
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

// isConsecutiveHammers 判断是否为连续锤子
func isConsecutiveHammers(data []KLineData, index int, count int) bool {
	if index < count-1 {
		return false
	}

	// 检查最近count根K线是否都是锤子
	for i := index - count + 1; i <= index; i++ {
		if i < 0 || i >= len(data) {
			return false
		}
		if !isHammer(data[i]) {
			return false
		}
	}
	return true
}

// ==================== 信号检测函数 ====================

// DetectAllSignals 检测所有信号（可扩展）
func DetectAllSignals(data []KLineData) []AlertSignal {
	if len(data) == 0 {
		return []AlertSignal{}
	}

	var allSignals []AlertSignal

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
func calculateBollingerBands(data []KLineData, period int, multiplier float64) []struct {
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

// detectBollingerDojiBottom 检测布林带下轨 + 十字星
func detectBollingerDojiBottom(data []KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []AlertSignal {
	var signals []AlertSignal
	tolerance := 0.01
	dojiThreshold := 0.001

	for i := range data {
		if i < 19 || bands[i].lower == 0 {
			continue
		}

		candle := data[i]
		if !isDoji(candle, dojiThreshold) {
			continue
		}

		lower := bands[i].lower
		isNearLower := candle.Low <= lower*(1+tolerance) ||
			candle.Close <= lower*(1+tolerance)

		if isNearLower {
			signals = append(signals, AlertSignal{
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
func detectBollingerHammer(data []KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []AlertSignal {
	var signals []AlertSignal
	tolerance := 0.01

	for i := range data {
		if i < 19 || bands[i].lower == 0 {
			continue
		}

		candle := data[i]
		if !isHammer(candle) {
			continue
		}

		lower := bands[i].lower
		isNearLower := candle.Low <= lower*(1+tolerance) ||
			candle.Close <= lower*(1+tolerance)

		if isNearLower {
			signals = append(signals, AlertSignal{
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
func detectBollingerConsecutiveHammers(data []KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []AlertSignal {
	var signals []AlertSignal
	tolerance := 0.01
	consecutiveCount := 2

	for i := range data {
		if i < 19 || bands[i].lower == 0 {
			continue
		}

		if !isConsecutiveHammers(data, i, consecutiveCount) {
			continue
		}

		candle := data[i]
		lower := bands[i].lower
		isNearLower := candle.Low <= lower*(1+tolerance) ||
			candle.Close <= lower*(1+tolerance)

		if isNearLower {
			signals = append(signals, AlertSignal{
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
func detectBollingerHangingMan(data []KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []AlertSignal {
	var signals []AlertSignal
	tolerance := 0.01

	for i := range data {
		if i < 19 || bands[i].upper == 0 {
			continue
		}

		candle := data[i]
		if !isHangingMan(candle) {
			continue
		}

		upper := bands[i].upper
		isNearUpper := candle.High >= upper*(1-tolerance) ||
			candle.Close >= upper*(1-tolerance)

		if isNearUpper {
			signals = append(signals, AlertSignal{
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
func detectBollingerEngulfing(data []KLineData, bands []struct {
	upper  float64
	middle float64
	lower  float64
}) []AlertSignal {
	var signals []AlertSignal
	tolerance := 0.01

	for i := 1; i < len(data); i++ {
		if i < 19 {
			continue
		}

		prev := data[i-1]
		curr := data[i]

		isEngulfing, isBullish := isEngulfing(prev, curr)
		if !isEngulfing {
			continue
		}

		// 看涨吞没在下轨附近
		if isBullish && bands[i].lower > 0 {
			lower := bands[i].lower
			isNearLower := curr.Low <= lower*(1+tolerance) ||
				prev.Low <= lower*(1+tolerance)

			if isNearLower {
				signals = append(signals, AlertSignal{
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
				signals = append(signals, AlertSignal{
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
