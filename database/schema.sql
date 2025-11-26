-- K线数据表（只存储1分钟数据）
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

-- 数据同步状态表
CREATE TABLE IF NOT EXISTS sync_status (
    id INT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL COMMENT '交易对',
    last_sync_time BIGINT NOT NULL DEFAULT 0 COMMENT '最后同步时间（毫秒时间戳）',
    last_kline_time BIGINT NOT NULL DEFAULT 0 COMMENT '最后一条K线时间（毫秒时间戳）',
    sync_count INT NOT NULL DEFAULT 0 COMMENT '同步次数',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_symbol (symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据同步状态表';

