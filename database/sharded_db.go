package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// ShardedDB 分表数据库操作
// 当数据量超大时，可以按币种分表存储
type ShardedDB struct {
	db *sql.DB
}

// NewShardedDB 创建分表数据库实例
func NewShardedDB(db *sql.DB) *ShardedDB {
	return &ShardedDB{db: db}
}

// 注意：GetTableName 已在 db.go 中定义，这里不再重复定义

// CreateSymbolTable 为指定币种创建表
func (s *ShardedDB) CreateSymbolTable(symbol string) error {
	tableName := GetTableName(symbol)

	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL COMMENT '交易对',
			open_time BIGINT NOT NULL COMMENT 'K线开盘时间（毫秒时间戳）',
			open DECIMAL(20, 8) NOT NULL COMMENT '开盘价',
			high DECIMAL(20, 8) NOT NULL COMMENT '最高价',
			low DECIMAL(20, 8) NOT NULL COMMENT '最低价',
			close DECIMAL(20, 8) NOT NULL COMMENT '收盘价',
			volume DECIMAL(20, 8) NOT NULL COMMENT '成交量',
			close_time BIGINT NOT NULL COMMENT 'K线收盘时间（毫秒时间戳）',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_open_time (open_time),
			INDEX idx_close_time (close_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='1分钟K线数据表 - %s';
	`, tableName, symbol)

	_, err := s.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("创建表 %s 失败: %w", tableName, err)
	}

	return nil
}

// SaveKLine1mSharded 保存K线数据到分表
func (s *ShardedDB) SaveKLine1mSharded(symbol string, klines []KLine1m) error {
	if len(klines) == 0 {
		return nil
	}

	// 确保表存在
	if err := s.CreateSymbolTable(symbol); err != nil {
		return err
	}

	tableName := GetTableName(symbol)
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf(`
		INSERT IGNORE INTO %s 
		(symbol, open_time, open, high, low, close, volume, close_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName))
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

// GetLatestKLineTimeSharded 获取分表中指定币种的最新K线时间
func (s *ShardedDB) GetLatestKLineTimeSharded(symbol string) (int64, error) {
	tableName := GetTableName(symbol)

	// 检查表是否存在
	var exists bool
	err := s.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) > 0 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		  AND table_name = ?
	`, tableName)).Scan(&exists)

	if err != nil || !exists {
		return 0, nil // 表不存在，返回0
	}

	var lastTime int64
	err = s.db.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(MAX(close_time), 0) 
		FROM %s
	`, tableName)).Scan(&lastTime)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return lastTime, nil
}

// GetKLines1mSharded 从分表获取K线数据
func (s *ShardedDB) GetKLines1mSharded(symbol string, startTime, endTime int64, limit int) ([]KLine1m, error) {
	tableName := GetTableName(symbol)

	query := fmt.Sprintf(`
		SELECT open_time, open, high, low, close, volume, close_time
		FROM %s
		WHERE 1=1
	`, tableName)

	args := []interface{}{}

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

	rows, err := s.db.Query(query, args...)
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

// ListSymbolTables 列出所有币种表
func (s *ShardedDB) ListSymbolTables() ([]string, error) {
	rows, err := s.db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		  AND table_name LIKE 'klines_1m_%'
		ORDER BY table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		// 提取币种名称（去掉前缀 klines_1m_）
		symbol := strings.TrimPrefix(tableName, "klines_1m_")
		tables = append(tables, symbol)
	}

	return tables, rows.Err()
}
