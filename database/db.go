package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB 初始化数据库连接
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
