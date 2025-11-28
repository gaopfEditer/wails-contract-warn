package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// SymbolConfig 币种配置
type SymbolConfig struct {
	Symbol      string `json:"symbol"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	PriorityRecentDays       int  `json:"priority_recent_days"`        // 优先同步最近N天的数据
	HistoricalStartYear      int  `json:"historical_start_year"`       // 历史数据起始年份
	BatchSize                int  `json:"batch_size"`                  // 每批拉取数量
	RequestIntervalMs        int  `json:"request_interval_ms"`         // 请求间隔（毫秒）
	IdleSyncEnabled          bool `json:"idle_sync_enabled"`           // 是否启用空闲同步
	IdleCheckIntervalSeconds int  `json:"idle_check_interval_seconds"` // 空闲检查间隔（秒）
}

// SymbolsConfig 币种配置文件结构
type SymbolsConfig struct {
	HotSymbols   []SymbolConfig `json:"hot_symbols"`
	MinorSymbols []SymbolConfig `json:"minor_symbols"`
	SyncConfig   SyncConfig     `json:"sync_config"`
}

var symbolsConfig *SymbolsConfig

// LoadSymbolsConfig 加载币种配置
func LoadSymbolsConfig() (*SymbolsConfig, error) {
	if symbolsConfig != nil {
		return symbolsConfig, nil
	}

	// 默认配置文件路径
	configPath := "config/symbols.json"
	if path := os.Getenv("SYMBOLS_CONFIG_PATH"); path != "" {
		configPath = path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config SymbolsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if config.SyncConfig.BatchSize == 0 {
		config.SyncConfig.BatchSize = 1000
	}
	if config.SyncConfig.RequestIntervalMs == 0 {
		config.SyncConfig.RequestIntervalMs = 200
	}
	if config.SyncConfig.PriorityRecentDays == 0 {
		config.SyncConfig.PriorityRecentDays = 1
	}
	if config.SyncConfig.HistoricalStartYear == 0 {
		config.SyncConfig.HistoricalStartYear = 2020
	}
	if config.SyncConfig.IdleCheckIntervalSeconds == 0 {
		config.SyncConfig.IdleCheckIntervalSeconds = 60
	}

	symbolsConfig = &config
	return symbolsConfig, nil
}

// GetAllEnabledSymbols 获取所有启用的币种（按优先级排序）
func GetAllEnabledSymbols() ([]SymbolConfig, error) {
	config, err := LoadSymbolsConfig()
	if err != nil {
		return nil, err
	}

	var allSymbols []SymbolConfig
	allSymbols = append(allSymbols, config.HotSymbols...)
	allSymbols = append(allSymbols, config.MinorSymbols...)

	// 过滤启用的币种
	var enabled []SymbolConfig
	for _, s := range allSymbols {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}

	// 按优先级排序
	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Priority < enabled[j].Priority
	})

	return enabled, nil
}

// GetHotSymbols 获取热门币种（按优先级排序）
func GetHotSymbols() ([]SymbolConfig, error) {
	config, err := LoadSymbolsConfig()
	if err != nil {
		return nil, err
	}

	var enabled []SymbolConfig
	for _, s := range config.HotSymbols {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}

	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Priority < enabled[j].Priority
	})

	return enabled, nil
}

// GetMinorSymbols 获取小币种（按优先级排序）
func GetMinorSymbols() ([]SymbolConfig, error) {
	config, err := LoadSymbolsConfig()
	if err != nil {
		return nil, err
	}

	var enabled []SymbolConfig
	for _, s := range config.MinorSymbols {
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}

	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Priority < enabled[j].Priority
	})

	return enabled, nil
}

// GetSyncConfig 获取同步配置
func GetSyncConfig() (*SyncConfig, error) {
	config, err := LoadSymbolsConfig()
	if err != nil {
		return nil, err
	}
	return &config.SyncConfig, nil
}
