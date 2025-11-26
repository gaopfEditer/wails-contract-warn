/**
 * 计算布林带（Bollinger Bands）
 * @param {Array} candles - K线数据数组
 * @param {number} period - 周期，默认20
 * @param {number} multiplier - 标准差倍数，默认2
 * @returns {Array} 布林带数据 [{middle, upper, lower}, ...]
 */
export function calculateBollingerBands(candles, period = 20, multiplier = 2) {
  const bands = []
  
  for (let i = 0; i < candles.length; i++) {
    if (i < period - 1) {
      bands.push(null) // 前 period-1 根无数据
      continue
    }

    // 取最近 period 根的收盘价
    const closes = candles.slice(i - period + 1, i + 1).map(c => c.close)
    const sma = closes.reduce((a, b) => a + b, 0) / period

    // 计算标准差
    const variance = closes.reduce((sum, price) => sum + Math.pow(price - sma, 2), 0) / period
    const stdDev = Math.sqrt(variance)

    bands.push({
      middle: sma,
      upper: sma + multiplier * stdDev,
      lower: sma - multiplier * stdDev,
    })
  }
  
  return bands
}

/**
 * 判断是否为十字星（Doji）
 * @param {Object} candle - K线数据 {open, close, high, low}
 * @param {number} threshold - 实体阈值，默认0.001（0.1%）
 * @returns {boolean}
 */
export function isDoji(candle, threshold = 0.001) {
  if (!candle || candle.high === candle.low) return false
  
  const body = Math.abs(candle.close - candle.open)
  const range = candle.high - candle.low
  
  // 实体很小，且有明显影线
  // 实体/开盘价 < 阈值 且 总波动 > 实体*2
  return range > 0 && body / candle.open < threshold && range > body * 2
}

/**
 * 检测布林带下轨 + 十字星信号
 * @param {Array} candles - K线数据数组
 * @param {Array} bands - 布林带数据数组
 * @param {number} tolerance - 容差，默认0.01（1%）
 * @returns {Array} 预警信号数组 [{index, time, price, lowerBand, type}, ...]
 */
export function detectBollingerDoji(candles, bands, tolerance = 0.01) {
  const alerts = []
  
  for (let i = 0; i < candles.length; i++) {
    const candle = candles[i]
    const band = bands[i]

    if (!band || !isDoji(candle)) continue

    const lower = band.lower
    // 判断：最低价是否在下轨附近（允许容差）
    // 收盘价或最低价 <= 下轨 * (1 + tolerance)
    const isNearLower = candle.low <= lower * (1 + tolerance) || 
                        candle.close <= lower * (1 + tolerance)

    if (isNearLower) {
      alerts.push({
        index: i,
        time: candle.time,
        price: candle.low,
        close: candle.close,
        lowerBand: lower,
        type: 'bollinger_doji_bottom',
      })
    }
  }
  
  return alerts
}

/**
 * 获取最新的预警信号
 * @param {Array} alerts - 预警信号数组
 * @returns {Object|null} 最新的预警信号
 */
export function getLatestAlert(alerts) {
  if (!alerts || alerts.length === 0) return null
  return alerts[alerts.length - 1]
}

