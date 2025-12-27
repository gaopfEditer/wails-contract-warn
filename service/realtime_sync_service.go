package service

import (
	"sync"
	"time"

	"wails-contract-warn/config"
	"wails-contract-warn/logger"
	datasync "wails-contract-warn/sync"
)

// RealtimeSyncService 实时数据同步服务
// 每分钟获取一次最新数据
type RealtimeSyncService struct {
	mu           sync.RWMutex
	running      bool
	stopChan     chan struct{}
	syncInterval time.Duration // 同步间隔（默认1分钟）
}

// NewRealtimeSyncService 创建实时数据同步服务
func NewRealtimeSyncService(intervalSeconds int) *RealtimeSyncService {
	if intervalSeconds <= 0 {
		intervalSeconds = 60 // 默认1分钟
	}
	return &RealtimeSyncService{
		stopChan:     make(chan struct{}),
		syncInterval: time.Duration(intervalSeconds) * time.Second,
	}
}

// Start 启动实时数据同步服务
func (s *RealtimeSyncService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Warn("实时数据同步服务已在运行")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Infof("启动实时数据同步服务，执行一次同步后自动停止")

	// 立即执行一次同步
	go s.syncLoop()
}

// Stop 停止实时数据同步服务
func (s *RealtimeSyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		// 注意：syncLoop现在会自动停止，所以这里主要是标记状态
		// 为了安全，尝试关闭channel（如果还没有关闭）
		select {
		case <-s.stopChan:
			// channel已经关闭，不需要再次关闭
		default:
			close(s.stopChan)
		}
		logger.Info("实时数据同步服务已停止")
	} else {
		logger.Debug("实时数据同步服务未运行，无需停止")
	}
}

// IsRunning 检查服务是否运行中
func (s *RealtimeSyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// syncLoop 实时数据同步循环
func (s *RealtimeSyncService) syncLoop() {
	// 立即执行一次同步
	s.syncRealtimeData()

	// 同步完成后自动停止
	s.mu.Lock()
	if s.running {
		s.running = false
		logger.Info("实时数据同步完成，服务已自动停止")
	}
	s.mu.Unlock()
}

// syncRealtimeData 同步实时数据
func (s *RealtimeSyncService) syncRealtimeData() {
	// 获取所有启用的币种（包括热门币种和小币种）
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		logger.Errorf("获取币种配置失败: %v", err)
		return
	}

	if len(allSymbols) == 0 {
		logger.Debug("没有配置币种，跳过实时数据同步")
		return
	}

	logger.Infof("开始实时数据同步: %d 个币种", len(allSymbols))

	// 顺序同步所有币种的实时数据（近期数据模式）
	for _, symbolConfig := range allSymbols {
		logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		logger.Infof("实时同步币种: %s", symbolConfig.Symbol)

		// 实时模式：只同步近期数据（使用时间段状态表，智能跳过已同步的数据）
		if err := datasync.SyncSymbolWithPriority(symbolConfig.Symbol, true); err != nil {
			logger.Errorf("❌ 实时数据同步失败: symbol=%s, error=%v", symbolConfig.Symbol, err)
		} else {
			logger.Infof("✅ 实时数据同步成功: %s", symbolConfig.Symbol)
		}

		logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// 每个币种之间稍作延迟，避免API限流
		time.Sleep(200 * time.Millisecond)
	}

		logger.Debugf("完成实时数据同步: %d 个币种", len(allSymbols))
}
