package service

import (
	"math"
	"sync"
	"time"

	"wails-contract-warn/logger"
	"wails-contract-warn/models"
)

// MarketService 市场数据服务
type MarketService struct {
	mu          sync.RWMutex
	data        map[string][]models.KLineData
	subscribers map[string]bool
	running     bool
	stopChan    chan struct{}
}

// NewMarketService 创建市场服务
func NewMarketService() *MarketService {
	return &MarketService{
		data:        make(map[string][]models.KLineData),
		subscribers: make(map[string]bool),
		stopChan:    make(chan struct{}),
	}
}

// Start 启动市场服务
func (m *MarketService) Start() {
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	logger.Debug("初始化市场数据服务示例数据")
	// 初始化一些示例数据
	m.initSampleData()

	// 启动数据更新循环
	logger.Debug("启动市场数据更新循环")
	go m.updateLoop()
}

// Stop 停止市场服务
func (m *MarketService) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		m.running = false
		close(m.stopChan)
		logger.Debug("市场数据服务已停止")
	}
}

// initSampleData 初始化示例数据
func (m *MarketService) initSampleData() {
	// 生成示例 K 线数据
	now := time.Now()
	var data []models.KLineData
	basePrice := 50000.0

	for i := 0; i < 100; i++ {
		timestamp := now.Add(-time.Duration(100-i)*time.Minute).Unix() * 1000
		change := (math.Sin(float64(i)/10) + math.Cos(float64(i)/7)) * 100
		open := basePrice + change
		close := open + (math.Sin(float64(i)/5) * 50)
		high := math.Max(open, close) + math.Abs(math.Sin(float64(i)/3)*30)
		low := math.Min(open, close) - math.Abs(math.Cos(float64(i)/4)*30)
		volume := 1000 + math.Abs(math.Sin(float64(i)/6)*500)

		data = append(data, models.KLineData{
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
				newKLine := models.KLineData{
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
func (m *MarketService) GetKLineData(symbol string, period string) []models.KLineData {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[symbol]
}

// Subscribe 订阅市场数据
func (m *MarketService) Subscribe(symbol string, period string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers[symbol] = true
	logger.Infof("订阅市场数据: symbol=%s, period=%s", symbol, period)
	return nil
}

// Unsubscribe 取消订阅
func (m *MarketService) Unsubscribe(symbol string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscribers, symbol)
	logger.Infof("取消订阅市场数据: symbol=%s", symbol)
	return nil
}
