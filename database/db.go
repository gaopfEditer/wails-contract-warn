package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"wails-contract-warn/logger"

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

	// 自动创建表结构（sync_status表）
	if err := InitSchema(); err != nil {
		return fmt.Errorf("创建表结构失败: %w", err)
	}

	// 根据配置文件自动创建K线表（按后缀分组）
	// 注意：这里需要先加载配置，如果配置加载失败，会在首次保存数据时自动创建表
	if err := InitTablesFromConfig(); err != nil {
		// 配置加载失败不影响数据库初始化，表会在首次使用时创建
		// 这里只记录警告，不返回错误
		fmt.Printf("警告: 根据配置创建表失败（将在首次使用时自动创建）: %v\n", err)
	}

	return nil
}

// InitSchema 初始化数据库表结构
func InitSchema() error {
	if DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 创建 sync_status 表（全局状态表，不需要分表）
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

	_, err := DB.Exec(createSyncStatusSQL)
	if err != nil {
		return fmt.Errorf("创建 sync_status 表失败: %w", err)
	}

	// 创建 sync_time_ranges 表（记录每个币种已同步的时间段）
	createTimeRangesSQL := `
		CREATE TABLE IF NOT EXISTS sync_time_ranges (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL COMMENT '交易对',
			start_time BIGINT NOT NULL COMMENT '时间段开始时间（毫秒时间戳）',
			end_time BIGINT NOT NULL COMMENT '时间段结束时间（毫秒时间戳）',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			INDEX idx_symbol_time (symbol, start_time, end_time),
			INDEX idx_symbol_start (symbol, start_time),
			INDEX idx_symbol_end (symbol, end_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据同步时间段记录表';
	`

	_, err = DB.Exec(createTimeRangesSQL)
	if err != nil {
		return fmt.Errorf("创建 sync_time_ranges 表失败: %w", err)
	}

	// 注意：K线表不再在这里创建，而是在需要时按后缀动态创建
	// 这样可以避免创建不必要的表

	return nil
}

// CreateTableForSymbol 为指定币种创建K线表（每个币种一张表）
func CreateTableForSymbol(symbol string) error {
	if DB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	tableName := GetTableName(symbol)

	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL COMMENT '交易对，如 BTC_USDT',
			open_time BIGINT NOT NULL COMMENT 'K线开盘时间（毫秒时间戳）',
			open DECIMAL(20, 8) NOT NULL COMMENT '开盘价',
			high DECIMAL(20, 8) NOT NULL COMMENT '最高价',
			low DECIMAL(20, 8) NOT NULL COMMENT '最低价',
			close DECIMAL(20, 8) NOT NULL COMMENT '收盘价',
			volume DECIMAL(20, 8) NOT NULL COMMENT '成交量',
			close_time BIGINT NOT NULL COMMENT 'K线收盘时间（毫秒时间戳）',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			UNIQUE KEY uk_open_time (open_time) COMMENT '唯一索引：防止重复数据',
			INDEX idx_close_time (close_time) COMMENT '索引：按收盘时间查询'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='1分钟K线数据表 - %s';
	`, tableName, symbol)

	_, err := DB.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("创建表 %s 失败: %w", tableName, err)
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

// GetTableName 根据symbol获取对应的表名（每个币种一张表）
func GetTableName(symbol string) string {
	// 将币种名称转换为表名，替换特殊字符为下划线
	tableName := strings.ReplaceAll(symbol, "-", "_")
	tableName = strings.ReplaceAll(tableName, ".", "_")
	// 表名格式：klines_1m_BTC_USDT, klines_1m_ETH_USDT 等
	return fmt.Sprintf("klines_1m_%s", tableName)
}

// SaveKLine1m 保存1分钟K线数据（批量插入，忽略重复）
// 每个币种存储到独立的表
func SaveKLine1m(klines []KLine1m) error {
	if len(klines) == 0 {
		return nil
	}

	// 按表名（币种）分组K线数据
	klinesByTable := make(map[string][]KLine1m)
	for _, k := range klines {
		tableName := GetTableName(k.Symbol)
		klinesByTable[tableName] = append(klinesByTable[tableName], k)
	}

	// 为每个表批量插入数据
	for tableName, tableKLines := range klinesByTable {
		// 确保表存在（使用第一个K线的币种）
		if len(tableKLines) > 0 {
			symbol := tableKLines[0].Symbol
			if err := CreateTableForSymbol(symbol); err != nil {
				return fmt.Errorf("创建表失败: %w", err)
			}
		}

		tx, err := DB.Begin()
		if err != nil {
			return err
		}

		// 使用 INSERT IGNORE 避免重复数据（基于 UNIQUE KEY uk_open_time）
		stmt, err := tx.Prepare(fmt.Sprintf(`
			INSERT IGNORE INTO %s 
			(symbol, open_time, open, high, low, close, volume, close_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, tableName))
		if err != nil {
			tx.Rollback()
			return err
		}

		insertedCount := 0
		skippedCount := 0
		errorCount := 0
		var firstKline *KLine1m

		for i, k := range tableKLines {
			// 保存第一条数据用于日志输出
			if i == 0 {
				firstKline = &k
			}

			result, err := stmt.Exec(
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
				errorCount++
				logger.Errorf("插入数据失败 [%s]: open_time=%d, error=%v", k.Symbol, k.OpenTime, err)
				// 继续处理下一条，不中断整个批次
				continue
			}

			// 检查是否实际插入了数据
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				insertedCount++
			} else {
				skippedCount++ // 数据已存在，被忽略
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败: %w", err)
		}

		// 打印详细的批次统计信息
		logger.Infof("表 %s 批次统计: 总数=%d, 成功插入=%d, 跳过(已存在)=%d, 失败=%d",
			tableName, len(tableKLines), insertedCount, skippedCount, errorCount)

		// 打印第一条完整数据
		if firstKline != nil {
			openTimeStr := time.Unix(firstKline.OpenTime/1000, 0).Format("2006-01-02 15:04:05")
			closeTimeStr := time.Unix(firstKline.CloseTime/1000, 0).Format("2006-01-02 15:04:05")
			logger.Infof("第一条数据示例 [%s]: open_time=%s, close_time=%s, open=%.8f, high=%.8f, low=%.8f, close=%.8f, volume=%.8f",
				firstKline.Symbol, openTimeStr, closeTimeStr, firstKline.Open, firstKline.High, firstKline.Low, firstKline.Close, firstKline.Volume)
		}
	}

	return nil
}

// GetLatestKLineTime 获取指定交易对的最新K线时间
func GetLatestKLineTime(symbol string) (int64, error) {
	tableName := GetTableName(symbol)

	// 检查表是否存在
	var exists bool
	err := DB.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) > 0 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		  AND table_name = ?
	`, tableName)).Scan(&exists)

	if err != nil || !exists {
		return 0, nil // 表不存在，返回0
	}

	var lastTime int64
	// 注意：由于每个币种一张表，不需要 WHERE symbol = ? 条件
	err = DB.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(MAX(close_time), 0) 
		FROM %s
	`, tableName)).Scan(&lastTime)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return lastTime, nil
}

// GetKLines1m 获取1分钟K线数据（按时间范围）
func GetKLines1m(symbol string, startTime, endTime int64, limit int) ([]KLine1m, error) {
	tableName := GetTableName(symbol)

	// 注意：由于每个币种一张表，不需要 WHERE symbol = ? 条件
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

	rows, err := DB.Query(query, args...)
	if err != nil {
		// 如果表不存在，返回空数组
		if strings.Contains(err.Error(), "doesn't exist") {
			return []KLine1m{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var klines []KLine1m
	for rows.Next() {
		var k KLine1m
		var symbolFromDB string
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
		// 从表名或查询结果中获取symbol
		k.Symbol = symbol // 使用传入的symbol参数
		if symbolFromDB != "" {
			k.Symbol = symbolFromDB
		}
		klines = append(klines, k)
	}

	return klines, rows.Err()
}

// GetKLines1mByCount 获取最近N根1分钟K线
func GetKLines1mByCount(symbol string, count int) ([]KLine1m, error) {
	tableName := GetTableName(symbol)

	// 注意：由于每个币种一张表，不需要 WHERE symbol = ? 条件
	query := fmt.Sprintf(`
		SELECT open_time, open, high, low, close, volume, close_time
		FROM %s
		ORDER BY open_time DESC
		LIMIT ?
	`, tableName)

	rows, err := DB.Query(query, count)
	if err != nil {
		// 如果表不存在，返回空数组
		if strings.Contains(err.Error(), "doesn't exist") {
			return []KLine1m{}, nil
		}
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

// SyncTimeRange 已同步的时间段
type SyncTimeRange struct {
	StartTime int64
	EndTime   int64
}

// AddSyncTimeRange 添加已同步的时间段
func AddSyncTimeRange(symbol string, startTime, endTime int64) error {
	_, err := DB.Exec(`
		INSERT INTO sync_time_ranges (symbol, start_time, end_time)
		VALUES (?, ?, ?)
	`, symbol, startTime, endTime)
	if err != nil {
		return fmt.Errorf("添加同步时间段失败: %w", err)
	}

	// 尝试合并相邻的时间段（异步优化，不影响主流程）
	go mergeAdjacentRanges(symbol)

	return nil
}

// GetSyncTimeRanges 获取指定币种的所有已同步时间段（按开始时间排序）
func GetSyncTimeRanges(symbol string) ([]SyncTimeRange, error) {
	rows, err := DB.Query(`
		SELECT start_time, end_time
		FROM sync_time_ranges
		WHERE symbol = ?
		ORDER BY start_time ASC
	`, symbol)
	if err != nil {
		return nil, fmt.Errorf("查询同步时间段失败: %w", err)
	}
	defer rows.Close()

	var ranges []SyncTimeRange
	for rows.Next() {
		var r SyncTimeRange
		if err := rows.Scan(&r.StartTime, &r.EndTime); err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}

	return ranges, rows.Err()
}

// FindMissingRanges 找出缺失的时间段
// targetStart: 目标开始时间
// targetEnd: 目标结束时间
// 返回需要同步的时间段列表
func FindMissingRanges(symbol string, targetStart, targetEnd int64) ([]SyncTimeRange, error) {
	// 获取已同步的时间段
	syncedRanges, err := GetSyncTimeRanges(symbol)
	if err != nil {
		return nil, err
	}

	// 如果没有已同步的数据，整个时间段都需要同步
	if len(syncedRanges) == 0 {
		return []SyncTimeRange{{StartTime: targetStart, EndTime: targetEnd}}, nil
	}

	var missingRanges []SyncTimeRange
	currentStart := targetStart

	// 遍历已同步的时间段，找出缺失的部分
	for _, synced := range syncedRanges {
		// 如果当前开始时间在已同步时间段之前，说明有缺失
		if currentStart < synced.StartTime {
			// 缺失的时间段：从 currentStart 到 synced.StartTime - 1
			missingRanges = append(missingRanges, SyncTimeRange{
				StartTime: currentStart,
				EndTime:   synced.StartTime - 1,
			})
		}

		// 更新当前开始时间到已同步时间段的结束时间之后
		if synced.EndTime >= currentStart {
			currentStart = synced.EndTime + 1
		}
	}

	// 检查最后是否还有缺失的时间段
	if currentStart <= targetEnd {
		missingRanges = append(missingRanges, SyncTimeRange{
			StartTime: currentStart,
			EndTime:   targetEnd,
		})
	}

	return missingRanges, nil
}

// mergeAdjacentRanges 合并相邻的时间段（优化存储）
func mergeAdjacentRanges(symbol string) {
	// 获取所有时间段
	ranges, err := GetSyncTimeRanges(symbol)
	if err != nil || len(ranges) <= 1 {
		return
	}

	// 合并相邻的时间段
	var merged []SyncTimeRange
	current := ranges[0]

	for i := 1; i < len(ranges); i++ {
		next := ranges[i]
		// 如果相邻（next的开始时间 <= current的结束时间+1），则合并
		if next.StartTime <= current.EndTime+1 {
			if next.EndTime > current.EndTime {
				current.EndTime = next.EndTime
			}
		} else {
			// 不相邻，保存当前时间段，开始新的时间段
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)

	// 如果合并后数量减少，更新数据库
	if len(merged) < len(ranges) {
		// 删除旧数据
		_, _ = DB.Exec(`DELETE FROM sync_time_ranges WHERE symbol = ?`, symbol)
		// 插入合并后的数据
		for _, r := range merged {
			_, _ = DB.Exec(`
				INSERT INTO sync_time_ranges (symbol, start_time, end_time)
				VALUES (?, ?, ?)
			`, symbol, r.StartTime, r.EndTime)
		}
		logger.Debugf("[%s] 合并时间段: %d -> %d", symbol, len(ranges), len(merged))
	}
}
