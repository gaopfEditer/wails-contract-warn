package service

import (
	"strings"
	"sync"
	"time"

	"wails-contract-warn/config"
	"wails-contract-warn/logger"
	datasync "wails-contract-warn/sync"
)

// PrioritySyncService 优先级同步服务
// 优先同步热门币种的近期数据，空闲时同步历史数据和小币种
type PrioritySyncService struct {
	mu                   sync.RWMutex
	running              bool
	stopChan             chan struct{}
	prioritySyncInterval time.Duration // 优先同步间隔（近期数据）
	idleSyncInterval     time.Duration // 空闲同步间隔（历史数据）
	lastPrioritySyncTime time.Time
	lastIdleSyncTime     time.Time
	currentIdleSymbolIdx int // 当前空闲同步的币种索引
}

// NewPrioritySyncService 创建优先级同步服务
func NewPrioritySyncService(priorityIntervalSeconds, idleIntervalSeconds int) *PrioritySyncService {
	return &PrioritySyncService{
		stopChan:             make(chan struct{}),
		prioritySyncInterval: time.Duration(priorityIntervalSeconds) * time.Second,
		idleSyncInterval:     time.Duration(idleIntervalSeconds) * time.Second,
	}
}

// Start 启动优先级同步服务
func (s *PrioritySyncService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Warn("优先级同步服务已在运行")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Infof("启动优先级同步服务，优先同步间隔: %v, 空闲同步间隔: %v",
		s.prioritySyncInterval, s.idleSyncInterval)

	// 立即执行一次优先同步
	go s.prioritySyncLoop()
	go s.idleSyncLoop()
}

// Stop 停止优先级同步服务
func (s *PrioritySyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		close(s.stopChan)
		logger.Info("优先级同步服务已停止")
	}
}

// IsRunning 检查服务是否运行中
func (s *PrioritySyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// prioritySyncLoop 优先同步循环（热门币种的近期数据）
func (s *PrioritySyncService) prioritySyncLoop() {
	ticker := time.NewTicker(s.prioritySyncInterval)
	defer ticker.Stop()

	// 立即执行一次
	s.syncPrioritySymbols()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.syncPrioritySymbols()
		}
	}
}

// idleSyncLoop 空闲同步循环（历史数据和小币种）
func (s *PrioritySyncService) idleSyncLoop() {
	ticker := time.NewTicker(s.idleSyncInterval)
	defer ticker.Stop()

	// 延迟启动，让优先同步先执行
	time.Sleep(10 * time.Second)

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.syncIdleSymbols()
		}
	}
}

// syncPrioritySymbols 同步所有币种的近期数据
func (s *PrioritySyncService) syncPrioritySymbols() {
	s.mu.Lock()
	s.lastPrioritySyncTime = time.Now()
	s.mu.Unlock()

	// 获取所有启用的币种（包括热门币种和小币种）
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		logger.Errorf("获取币种配置失败: %v", err)
		return
	}

	if len(allSymbols) == 0 {
		logger.Debug("没有配置币种")
		return
	}

	logger.Infof("开始优先同步 %d 个币种的近期数据", len(allSymbols))

	// 顺序同步所有币种（避免并发请求过多导致API限流）
	// 优先模式：只同步近期数据
	for _, symbolConfig := range allSymbols {
		logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		logger.Infof("开始同步币种: %s", symbolConfig.Symbol)
		if err := datasync.SyncSymbolWithPriority(symbolConfig.Symbol, true); err != nil {
			// 忽略小币种的错误（用户要求）
			if strings.Contains(err.Error(), "INVALID_CURRENCY_PAIR") {
				logger.Warnf("币种可能不存在，跳过: symbol=%s", symbolConfig.Symbol)
				continue
			}
			logger.Errorf("❌ 优先同步币种失败: symbol=%s, error=%v", symbolConfig.Symbol, err)
		} else {
			logger.Infof("✅ 优先同步币种成功: %s", symbolConfig.Symbol)
		}
		logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		// 每个币种之间稍作延迟，避免API限流
		time.Sleep(300 * time.Millisecond)
	}

	logger.Debugf("完成优先同步 %d 个币种", len(allSymbols))
}

// syncIdleSymbols 空闲时同步历史数据和小币种
func (s *PrioritySyncService) syncIdleSymbols() {
	s.mu.Lock()
	s.lastIdleSyncTime = time.Now()
	s.mu.Unlock()

	syncConfig, err := config.GetSyncConfig()
	if err != nil {
		logger.Errorf("获取同步配置失败: %v", err)
		return
	}

	if !syncConfig.IdleSyncEnabled {
		logger.Debug("空闲同步已禁用")
		return
	}

	// 获取所有启用的币种
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		logger.Errorf("获取币种配置失败: %v", err)
		return
	}

	if len(allSymbols) == 0 {
		logger.Debug("没有配置币种")
		return
	}

	// 轮询同步：每次同步一个币种的历史数据
	s.mu.Lock()
	currentIdx := s.currentIdleSymbolIdx
	s.currentIdleSymbolIdx = (s.currentIdleSymbolIdx + 1) % len(allSymbols)
	s.mu.Unlock()

	if currentIdx >= len(allSymbols) {
		currentIdx = 0
	}

	symbolConfig := allSymbols[currentIdx]
	logger.Infof("空闲同步: 同步币种 %s 的历史数据（从 %d 年开始）",
		symbolConfig.Symbol, syncConfig.HistoricalStartYear)

	// 同步历史数据
	if err := datasync.SyncSymbolHistorical(symbolConfig.Symbol, syncConfig.HistoricalStartYear); err != nil {
		logger.Errorf("空闲同步历史数据失败: symbol=%s, error=%v", symbolConfig.Symbol, err)
	} else {
		logger.Debugf("空闲同步历史数据成功: %s", symbolConfig.Symbol)
	}

	// 如果还有小币种，也同步它们的近期数据
	minorSymbols, err := config.GetMinorSymbols()
	if err == nil && len(minorSymbols) > 0 {
		// 每次空闲同步时，同步一个小币种的近期数据
		minorIdx := currentIdx % len(minorSymbols)
		if minorIdx < len(minorSymbols) {
			minorSymbol := minorSymbols[minorIdx]
			logger.Debugf("空闲同步: 同步小币种 %s 的近期数据", minorSymbol.Symbol)
			if err := datasync.SyncSymbolWithPriority(minorSymbol.Symbol, true); err != nil {
				logger.Errorf("空闲同步小币种失败: symbol=%s, error=%v", minorSymbol.Symbol, err)
			}
		}
	}
}
