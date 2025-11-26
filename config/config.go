package config

import (
	"os"
)

// GetDBDSN 从环境变量或配置文件获取数据库DSN
func GetDBDSN() string {
	// 优先从环境变量读取
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}

	// 默认配置（格式：user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local）
	return ""
}

// GetSyncInterval 获取同步间隔（秒）
func GetSyncInterval() int {
	// 默认60秒同步一次
	return 60
}
