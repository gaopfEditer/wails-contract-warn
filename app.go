package main

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"
)

// App 结构体
type App struct {
	ctx    context.Context
	mu     sync.RWMutex
	market *MarketService
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{
		market: NewMarketService(),
	}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.market.Start()
}

// domReady DOM 准备就绪时调用
func (a *App) domReady(ctx context.Context) {
	// 可以在这里执行一些初始化操作
}

// shutdown 应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	a.market.Stop()
}

// GetMarketData 获取市场数据
func (a *App) GetMarketData(symbol string, period string) (string, error) {
	data := a.market.GetKLineData(symbol, period)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetIndicators 计算技术指标
func (a *App) GetIndicators(symbol string, period string) (string, error) {
	klineData := a.market.GetKLineData(symbol, period)
	indicators := CalculateIndicators(klineData)
	jsonData, err := json.Marshal(indicators)
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

// KLineData K线数据
type KLineData struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// Indicators 技术指标
type Indicators struct {
	MA5    []float64 `json:"ma5"`
	MA10   []float64 `json:"ma10"`
	MA20   []float64 `json:"ma20"`
	MACD   []float64 `json:"macd"`
	Signal []float64 `json:"signal"`
	Hist   []float64 `json:"hist"`
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
		MA5:    make([]float64, len(data)),
		MA10:   make([]float64, len(data)),
		MA20:   make([]float64, len(data)),
		MACD:   make([]float64, len(data)),
		Signal: make([]float64, len(data)),
		Hist:   make([]float64, len(data)),
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

	return indicators
}
