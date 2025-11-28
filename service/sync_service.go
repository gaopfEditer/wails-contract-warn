package service

import (
	"sync"
	"time"

	"wails-contract-warn/logger"
	datasync "wails-contract-warn/sync"
)

// SyncService 数据同步服务
type SyncService struct {
	mu           sync.RWMutex
	running      bool
	symbols      map[string]bool
	stopChan     chan struct{}
	syncInterval time.Duration
}

// NewSyncService 创建同步服务
func NewSyncService(intervalSeconds int) *SyncService {
	return &SyncService{
		symbols:      make(map[string]bool),
		stopChan:     make(chan struct{}),
		syncInterval: time.Duration(intervalSeconds) * time.Second,
	}
}

// Start 启动同步服务
func (s *SyncService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Warn("同步服务已在运行")
		return
	}
	s.running = true
	s.mu.Unlock()

	logger.Infof("启动同步服务，同步间隔: %v", s.syncInterval)
	go s.syncLoop()
}

// Stop 停止同步服务
func (s *SyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		close(s.stopChan)
		logger.Info("同步服务已停止")
	}
}

// AddSymbol 添加需要同步的交易对
func (s *SyncService) AddSymbol(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols[symbol] = true
	logger.Infof("添加同步交易对: %s", symbol)
}

// RemoveSymbol 移除交易对
func (s *SyncService) RemoveSymbol(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.symbols, symbol)
	logger.Infof("移除同步交易对: %s", symbol)
}

// syncLoop 同步循环
func (s *SyncService) syncLoop() {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	// 立即执行一次
	s.syncAll()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.syncAll()
		}
	}
}

// syncAll 同步所有交易对
func (s *SyncService) syncAll() {
	s.mu.RLock()
	symbols := make([]string, 0, len(s.symbols))
	for symbol := range s.symbols {
		symbols = append(symbols, symbol)
	}
	s.mu.RUnlock()

	if len(symbols) == 0 {
		logger.Debug("没有需要同步的交易对")
		return
	}

	logger.Debugf("开始同步 %d 个交易对: %v", len(symbols), symbols)

	// 并发同步多个交易对
	var wg sync.WaitGroup
	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			if err := datasync.SyncSymbol(sym); err != nil {
				logger.Errorf("同步交易对失败: symbol=%s, error=%v", sym, err)
			} else {
				logger.Debugf("同步交易对成功: %s", sym)
			}
		}(symbol)
	}
	wg.Wait()

	logger.Debugf("完成同步 %d 个交易对", len(symbols))
}

// IsRunning 检查服务是否运行中
func (s *SyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
