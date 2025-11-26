package service

import (
	"sync"
	"time"

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
		return
	}
	s.running = true
	s.mu.Unlock()

	go s.syncLoop()
}

// Stop 停止同步服务
func (s *SyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		s.running = false
		close(s.stopChan)
	}
}

// AddSymbol 添加需要同步的交易对
func (s *SyncService) AddSymbol(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.symbols[symbol] = true
}

// RemoveSymbol 移除交易对
func (s *SyncService) RemoveSymbol(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.symbols, symbol)
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

	// 并发同步多个交易对
	var wg sync.WaitGroup
	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			if err := datasync.SyncSymbol(sym); err != nil {
				// 记录错误但不中断其他同步
				// 可以在这里添加日志
			}
		}(symbol)
	}
	wg.Wait()
}

// IsRunning 检查服务是否运行中
func (s *SyncService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
