package service

import (
	"sync"
	"time"

	"wails-contract-warn/config"
	"wails-contract-warn/database"
	"wails-contract-warn/logger"
	datasync "wails-contract-warn/sync"
)

// RealtimePriceService 实时价格服务
// 每分钟获取一次最新价格，并推送到前端
type RealtimePriceService struct {
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
	ctx          interface{} // runtime.Context
	eventEmitter func(event string, data ...interface{}) // EventEmitter函数
}

// NewRealtimePriceService 创建实时价格服务
func NewRealtimePriceService(ctx interface{}, eventEmitter func(event string, data ...interface{})) *RealtimePriceService {
	return &RealtimePriceService{
		stopChan:     make(chan struct{}),
		ctx:          ctx,
		eventEmitter: eventEmitter,
	}
}

// Start 启动实时价格服务
func (s *RealtimePriceService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Warn("实时价格服务已在运行")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Info("启动实时价格服务，每10秒获取一次最新价格")

	// 立即执行一次
	go s.priceLoop()
}

// Stop 停止实时价格服务
func (s *RealtimePriceService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		close(s.stopChan)
		logger.Info("实时价格服务已停止")
	}
}

// IsRunning 检查服务是否运行中
func (s *RealtimePriceService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// priceLoop 实时价格获取循环
func (s *RealtimePriceService) priceLoop() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒更新一次，提高实时性
	defer ticker.Stop()

	// 立即执行一次
	s.fetchAndPushLatestPrice()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.fetchAndPushLatestPrice()
		}
	}
}

// fetchAndPushLatestPrice 获取并推送最新价格
func (s *RealtimePriceService) fetchAndPushLatestPrice() {
	// 获取所有启用的币种（包括热门币种和小币种）
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		logger.Errorf("获取币种配置失败: %v", err)
		return
	}

	if len(allSymbols) == 0 {
		logger.Debug("没有配置币种，跳过实时价格获取")
		return
	}

	logger.Debugf("开始获取实时价格: %d 个币种", len(allSymbols))

	// 获取每个币种的最新价格
	for _, symbolConfig := range allSymbols {
		symbol := symbolConfig.Symbol
		
		// 同步最新数据（只同步最近几分钟的数据）
		if err := datasync.SyncSymbolWithPriority(symbol, true); err != nil {
			logger.Errorf("❌ 实时价格同步失败: symbol=%s, error=%v", symbol, err)
			continue
		}

		// 从数据库获取最新的一条K线数据
		latestKLine, err := database.GetLatestKLine1m(symbol)
		if err != nil {
			logger.Errorf("获取最新K线数据失败: symbol=%s, error=%v", symbol, err)
			continue
		}

		if latestKLine != nil {
			// 构建价格数据
			priceData := map[string]interface{}{
				"symbol":    symbol,
				"time":      latestKLine.CloseTime,
				"open":      latestKLine.Open,
				"high":      latestKLine.High,
				"low":       latestKLine.Low,
				"close":     latestKLine.Close,
				"volume":    latestKLine.Volume,
				"timestamp": time.Now().UnixMilli(),
			}

			// 推送到前端
			if s.eventEmitter != nil {
				s.eventEmitter("realtime-price", priceData)
			} else if s.ctx != nil {
				// 使用runtime.EventsEmit
				if ctx, ok := s.ctx.(interface{ EventsEmit(string, ...interface{}) }); ok {
					ctx.EventsEmit("realtime-price", priceData)
				}
			}

			logger.Debugf("✅ 推送实时价格: symbol=%s, price=%.2f", symbol, latestKLine.Close)
		}

		// 每个币种之间稍作延迟，避免API限流
		time.Sleep(200 * time.Millisecond)
	}
}

// GapFillService 历史空缺补充服务
// 检测当天的空缺并补充
type GapFillService struct {
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
	checkInterval time.Duration // 检查间隔（默认5分钟）
}

// NewGapFillService 创建历史空缺补充服务
func NewGapFillService(intervalMinutes int) *GapFillService {
	if intervalMinutes <= 0 {
		intervalMinutes = 5 // 默认5分钟检查一次
	}
	return &GapFillService{
		stopChan:      make(chan struct{}),
		checkInterval: time.Duration(intervalMinutes) * time.Minute,
	}
}

// Start 启动历史空缺补充服务
func (s *GapFillService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Warn("历史空缺补充服务已在运行")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Infof("启动历史空缺补充服务，检查间隔: %v", s.checkInterval)

	// 立即执行一次
	go s.gapFillLoop()
}

// Stop 停止历史空缺补充服务
func (s *GapFillService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		close(s.stopChan)
		logger.Info("历史空缺补充服务已停止")
	}
}

// IsRunning 检查服务是否运行中
func (s *GapFillService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// gapFillLoop 历史空缺补充循环
func (s *GapFillService) gapFillLoop() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	// 立即执行一次
	s.checkAndFillGaps()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndFillGaps()
		}
	}
}

// checkAndFillGaps 检查并补充空缺
func (s *GapFillService) checkAndFillGaps() {
	// 获取所有启用的币种
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		logger.Errorf("获取币种配置失败: %v", err)
		return
	}

	if len(allSymbols) == 0 {
		logger.Debug("没有配置币种，跳过空缺检查")
		return
	}

	logger.Infof("开始检查历史空缺: %d 个币种", len(allSymbols))

	now := time.Now().UnixMilli()
	// 检查今天的数据（从今天00:00:00到现在）
	today := time.Now().UTC()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
	todayEnd := now

	for _, symbolConfig := range allSymbols {
		symbol := symbolConfig.Symbol

		logger.Debugf("检查币种 %s 的当天空缺", symbol)

		// 查找当天的缺失时间段
		missingRanges, err := database.FindMissingRanges(symbol, todayStart, todayEnd)
		if err != nil {
			logger.Errorf("查找缺失时间段失败: symbol=%s, error=%v", symbol, err)
			continue
		}

		if len(missingRanges) == 0 {
			logger.Debugf("[%s] ✓ 当天数据完整，无空缺", symbol)
			continue
		}

		logger.Infof("[%s] 发现 %d 个当天空缺时间段，开始补充", symbol, len(missingRanges))

		// 补充每个缺失的时间段
		for _, missingRange := range missingRanges {
			rangeStartStr := time.Unix(missingRange.StartTime/1000, 0).Format("2006-01-02 15:04:05")
			rangeEndStr := time.Unix(missingRange.EndTime/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("[%s] 补充空缺时间段: %s ~ %s", symbol, rangeStartStr, rangeEndStr)

			// 使用历史数据同步方法补充
			if err := datasync.SyncSymbolHistoricalBackward(symbol, 2024, 300, 7); err != nil {
				logger.Errorf("❌ 补充空缺失败: symbol=%s, range=%s~%s, error=%v", 
					symbol, rangeStartStr, rangeEndStr, err)
			} else {
				logger.Infof("✅ 补充空缺成功: symbol=%s, range=%s~%s", symbol, rangeStartStr, rangeEndStr)
			}
		}

		// 每个币种之间稍作延迟
		time.Sleep(500 * time.Millisecond)
	}

	logger.Debugf("完成历史空缺检查: %d 个币种", len(allSymbols))
}

