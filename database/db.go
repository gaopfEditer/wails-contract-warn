package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB 初始化数据库连接并创建表结构
func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动创建表结构
	if err := InitSchema(); err != nil {
		return fmt.Errorf("创建表结构失败: %w", err)
	}

	return nil
}

// InitSchema 初始化数据库表结构
func InitSchema() error {
	if DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 创建 klines_1m 表
	createKlines1mSQL := `
		CREATE TABLE IF NOT EXISTS klines_1m (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL COMMENT '交易对，如 BTCUSDT',
			open_time BIGINT NOT NULL COMMENT 'K线开盘时间（毫秒时间戳）',
			open DECIMAL(20, 8) NOT NULL COMMENT '开盘价',
			high DECIMAL(20, 8) NOT NULL COMMENT '最高价',
			low DECIMAL(20, 8) NOT NULL COMMENT '最低价',
			close DECIMAL(20, 8) NOT NULL COMMENT '收盘价',
			volume DECIMAL(20, 8) NOT NULL COMMENT '成交量',
			close_time BIGINT NOT NULL COMMENT 'K线收盘时间（毫秒时间戳）',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_symbol_time (symbol, open_time),
			INDEX idx_symbol_close_time (symbol, close_time),
			INDEX idx_close_time (close_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='1分钟K线数据表（原始数据）';
	`

	_, err := DB.Exec(createKlines1mSQL)
	if err != nil {
		return fmt.Errorf("创建 klines_1m 表失败: %w", err)
	}

	// 创建 sync_status 表
	createSyncStatusSQL := `
		CREATE TABLE IF NOT EXISTS sync_status (
			id INT AUTO_INCREMENT PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL COMMENT '交易对',
			last_sync_time BIGINT NOT NULL DEFAULT 0 COMMENT '最后同步时间（毫秒时间戳）',
			last_kline_time BIGINT NOT NULL DEFAULT 0 COMMENT '最后一条K线时间（毫秒时间戳）',
			sync_count INT NOT NULL DEFAULT 0 COMMENT '同步次数',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			UNIQUE KEY uk_symbol (symbol)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据同步状态表';
	`

	_, err = DB.Exec(createSyncStatusSQL)
	if err != nil {
		return fmt.Errorf("创建 sync_status 表失败: %w", err)
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// KLine1m 1分钟K线数据
type KLine1m struct {
	ID        int64
	Symbol    string
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

// SaveKLine1m 保存1分钟K线数据（批量插入，忽略重复）
func SaveKLine1m(klines []KLine1m) error {
	if len(klines) == 0 {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT IGNORE INTO klines_1m 
		(symbol, open_time, open, high, low, close, volume, close_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, k := range klines {
		_, err := stmt.Exec(
			k.Symbol,
			k.OpenTime,
			k.Open,
			k.High,
			k.Low,
			k.Close,
			k.Volume,
			k.CloseTime,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLatestKLineTime 获取指定交易对的最新K线时间
func GetLatestKLineTime(symbol string) (int64, error) {
	var lastTime int64
	err := DB.QueryRow(`
		SELECT COALESCE(MAX(close_time), 0) 
		FROM klines_1m 
		WHERE symbol = ?
	`, symbol).Scan(&lastTime)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return lastTime, nil
}

// GetKLines1m 获取1分钟K线数据（按时间范围）
func GetKLines1m(symbol string, startTime, endTime int64, limit int) ([]KLine1m, error) {
	query := `
		SELECT open_time, open, high, low, close, volume, close_time
		FROM klines_1m
		WHERE symbol = ?
	`
	args := []interface{}{symbol}

	if startTime > 0 {
		query += " AND open_time >= ?"
		args = append(args, startTime)
	}

	if endTime > 0 {
		query += " AND open_time <= ?"
		args = append(args, endTime)
	}

	query += " ORDER BY open_time ASC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klines []KLine1m
	for rows.Next() {
		var k KLine1m
		k.Symbol = symbol
		err := rows.Scan(
			&k.OpenTime,
			&k.Open,
			&k.High,
			&k.Low,
			&k.Close,
			&k.Volume,
			&k.CloseTime,
		)
		if err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}

	return klines, rows.Err()
}

// GetKLines1mByCount 获取最近N根1分钟K线
func GetKLines1mByCount(symbol string, count int) ([]KLine1m, error) {
	query := `
		SELECT open_time, open, high, low, close, volume, close_time
		FROM klines_1m
		WHERE symbol = ?
		ORDER BY open_time DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, symbol, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var klines []KLine1m
	for rows.Next() {
		var k KLine1m
		k.Symbol = symbol
		err := rows.Scan(
			&k.OpenTime,
			&k.Open,
			&k.High,
			&k.Low,
			&k.Close,
			&k.Volume,
			&k.CloseTime,
		)
		if err != nil {
			return nil, err
		}
		klines = append(klines, k)
	}

	// 反转顺序（从旧到新）
	for i, j := 0, len(klines)-1; i < j; i, j = i+1, j-1 {
		klines[i], klines[j] = klines[j], klines[i]
	}

	return klines, rows.Err()
}

// UpdateSyncStatus 更新同步状态
func UpdateSyncStatus(symbol string, lastSyncTime, lastKlineTime int64) error {
	_, err := DB.Exec(`
		INSERT INTO sync_status (symbol, last_sync_time, last_kline_time, sync_count)
		VALUES (?, ?, ?, 1)
		ON DUPLICATE KEY UPDATE
			last_sync_time = ?,
			last_kline_time = ?,
			sync_count = sync_count + 1
	`, symbol, lastSyncTime, lastKlineTime, lastSyncTime, lastKlineTime)
	return err
}

// GetSyncStatus 获取同步状态
func GetSyncStatus(symbol string) (lastSyncTime, lastKlineTime int64, err error) {
	err = DB.QueryRow(`
		SELECT last_sync_time, last_kline_time
		FROM sync_status
		WHERE symbol = ?
	`, symbol).Scan(&lastSyncTime, &lastKlineTime)

	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	return
}
