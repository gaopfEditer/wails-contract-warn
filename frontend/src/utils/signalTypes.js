/**
 * 信号类型配置
 */
export const SIGNAL_TYPES = {
  // 看涨信号
  bollinger_doji_bottom: {
    name: '布林带下轨十字星',
    icon: '⚠️',
    color: '#ff4757',
    bgColor: '#fff5f5',
    borderColor: '#ff4757',
    description: '价格在布林带下轨附近出现十字星，可能反弹',
  },
  bollinger_hammer_bottom: {
    name: '布林带下轨锤子',
    icon: '🔨',
    color: '#ff6b6b',
    bgColor: '#fff5f5',
    borderColor: '#ff6b6b',
    description: '价格在布林带下轨附近出现锤子线，看涨信号',
  },
  bollinger_consecutive_hammers: {
    name: '布林带下轨连续锤子',
    icon: '🔨🔨',
    color: '#ee5a6f',
    bgColor: '#fff5f5',
    borderColor: '#ee5a6f',
    description: '价格在布林带下轨附近连续出现锤子线，强烈看涨',
  },
  bollinger_bullish_engulfing: {
    name: '布林带下轨看涨吞没',
    icon: '📈',
    color: '#26a69a',
    bgColor: '#e8f5e9',
    borderColor: '#26a69a',
    description: '价格在布林带下轨附近出现看涨吞没形态',
  },
  // 看跌信号
  bollinger_hanging_man_top: {
    name: '布林带上轨吊颈',
    icon: '🔻',
    color: '#ff9800',
    bgColor: '#fff3e0',
    borderColor: '#ff9800',
    description: '价格在布林带上轨附近出现吊颈线，看跌信号',
  },
  bollinger_bearish_engulfing: {
    name: '布林带上轨看跌吞没',
    icon: '📉',
    color: '#ef5350',
    bgColor: '#ffebee',
    borderColor: '#ef5350',
    description: '价格在布林带上轨附近出现看跌吞没形态',
  },
  // 组合强信号
  strong_hammer_group: {
    name: '强信号-多锤子组合',
    icon: '🔨🔨',
    color: '#ff1744',
    bgColor: '#ffebee',
    borderColor: '#ff1744',
    description: '在3-5个K线中出现多个锤子线，强烈看涨信号',
  },
  strong_top_pin_group: {
    name: '强信号-多顶部针形',
    icon: '📌📌',
    color: '#ff6f00',
    bgColor: '#fff3e0',
    borderColor: '#ff6f00',
    description: '在3-5个K线中出现多个较长的顶部针形，强烈看跌信号',
  },
  strong_mixed_pattern_group: {
    name: '强信号-混合形态组合',
    icon: '⚡',
    color: '#e91e63',
    bgColor: '#fce4ec',
    borderColor: '#e91e63',
    description: '在3-5个K线中出现多个锤子线或顶部针形的混合组合',
  },
}

/**
 * 获取信号配置
 */
export function getSignalConfig(type) {
  return SIGNAL_TYPES[type] || {
    name: '未知信号',
    icon: '⚠️',
    color: '#999',
    bgColor: '#f5f5f5',
    borderColor: '#999',
    description: '未知信号类型',
  }
}

/**
 * 按强度排序信号
 */
export function sortSignalsByStrength(signals) {
  return [...signals].sort((a, b) => (b.strength || 0) - (a.strength || 0))
}

