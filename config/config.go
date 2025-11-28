package config

import (
	"fmt"
	"os"
)

// GetDBDSN 从环境变量或配置文件获取数据库DSN
func GetDBDSN() string {
	// 优先从环境变量读取
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}

	// 默认配置（远程MySQL数据库）
	// 格式：user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	host := "8.155.10.218"
	port := "3306"
	user := "root"
	password := "123456"
	dbname := "wails-contract-warn"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	return dsn
}

// GetSyncInterval 获取同步间隔（秒）
func GetSyncInterval() int {
	// 默认60秒同步一次
	return 60
}
