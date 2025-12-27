package service

import (
	"sync"
	"time"

	"wails-contract-warn/config"
	"wails-contract-warn/logger"
	datasync "wails-contract-warn/sync"
)

// HistoricalSyncService 历史数据同步服务
// 从2020年开始倒推获取历史数据，批量获取，每次300条
type HistoricalSyncService struct {
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
	syncInterval time.Duration // 同步间隔（每个币种之间的间隔）
	batchSize    int           // 每批获取的数据量
	startYear    int           // 起始年份
}

// NewHistoricalSyncService 创建历史数据同步服务
func NewHistoricalSyncService(intervalSeconds int, batchSize int, startYear int) *HistoricalSyncService {
	return &HistoricalSyncService{
		stopChan:     make(chan struct{}),
		syncInterval: time.Duration(intervalSeconds) * time.Second,
		batchSize:    batchSize,
		startYear:    startYear,
	}
}

// Start 启动历史数据同步服务
func (s *HistoricalSyncService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Warn("历史数据同步服务已在运行")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Infof("启动历史数据同步服务，同步间隔: %v, 批次大小: %d, 起始年份: %d",
		s.syncInterval, s.batchSize, s.startYear)

	// 立即执行一次同步
	go s.syncLoop()
}

// Stop 停止历史数据同步服务
func (s *HistoricalSyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		close(s.stopChan)
		logger.Info("历史数据同步服务已停止")
	}
}

// IsRunning 检查服务是否运行中
func (s *HistoricalSyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// syncLoop 历史数据同步循环
func (s *HistoricalSyncService) syncLoop() {
	// 获取所有启用的币种
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		logger.Errorf("获取币种配置失败: %v", err)
		return
	}

	if len(allSymbols) == 0 {
		logger.Warn("没有配置币种，历史数据同步服务退出")
		return
	}

	logger.Infof("历史数据同步服务: 开始同步 %d 个币种的历史数据", len(allSymbols))

	// 轮询同步：每次同步一个币种的历史数据
	currentIdx := 0

	for {
		select {
		case <-s.stopChan:
			return
		default:
			// 选择当前币种
			symbolConfig := allSymbols[currentIdx]
			currentIdx = (currentIdx + 1) % len(allSymbols)

			logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			logger.Infof("历史数据同步: 开始同步币种 %s (从近到远按天倒推，直到 %d 年)", symbolConfig.Symbol, s.startYear)

			// 同步历史数据（按天倒推，使用时间段状态表，智能跳过已同步的数据）
			// 限制只拉取最近7天的数据
			if err := datasync.SyncSymbolHistoricalBackward(symbolConfig.Symbol, s.startYear, s.batchSize, 7); err != nil {
				logger.Errorf("❌ 历史数据同步失败: symbol=%s, error=%v", symbolConfig.Symbol, err)
			} else {
				logger.Infof("✅ 历史数据同步成功: %s", symbolConfig.Symbol)
			}

			logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

			// 等待指定间隔后继续下一个币种
			select {
			case <-s.stopChan:
				return
			case <-time.After(s.syncInterval):
				// 继续下一个币种
			}
		}
	}
}
