package database

import (
	"fmt"

	"wails-contract-warn/config"
)

// InitTablesFromConfig 根据配置文件中的币种，自动创建对应的表（每个币种一张表）
func InitTablesFromConfig() error {
	if DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 获取所有启用的币种
	allSymbols, err := config.GetAllEnabledSymbols()
	if err != nil {
		return fmt.Errorf("获取币种配置失败: %w", err)
	}

	// 为每个币种创建独立的表
	for _, symbolConfig := range allSymbols {
		if err := CreateTableForSymbol(symbolConfig.Symbol); err != nil {
			return fmt.Errorf("创建表失败 (币种: %s): %w", symbolConfig.Symbol, err)
		}
	}

	return nil
}
