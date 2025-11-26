/**
 * 市场数据API
 */

/**
 * 获取K线数据
 * @param {string} symbol - 交易对
 * @param {string} period - 周期
 * @returns {Promise<Array>}
 */
export async function getMarketData(symbol, period) {
  try {
    const dataStr = await window.go.main.App.GetMarketData(symbol, period)
    return JSON.parse(dataStr)
  } catch (error) {
    console.error('获取市场数据失败:', error)
    throw error
  }
}

/**
 * 获取技术指标
 * @param {string} symbol - 交易对
 * @param {string} period - 周期
 * @returns {Promise<Object>}
 */
export async function getIndicators(symbol, period) {
  try {
    const indicatorsStr = await window.go.main.App.GetIndicators(symbol, period)
    return JSON.parse(indicatorsStr)
  } catch (error) {
    console.error('获取技术指标失败:', error)
    throw error
  }
}

/**
 * 获取预警信号
 * @param {string} symbol - 交易对
 * @param {string} period - 周期
 * @returns {Promise<Array>}
 */
export async function getAlertSignals(symbol, period) {
  try {
    const signalsStr = await window.go.main.App.GetAlertSignals(symbol, period)
    return JSON.parse(signalsStr)
  } catch (error) {
    console.error('获取预警信号失败:', error)
    throw error
  }
}

/**
 * 开始实时数据流
 * @param {string} symbol - 交易对
 * @param {string} period - 周期
 * @returns {Promise<void>}
 */
export async function startMarketDataStream(symbol, period) {
  try {
    await window.go.main.App.StartMarketDataStream(symbol, period)
  } catch (error) {
    console.error('启动数据流失败:', error)
    throw error
  }
}

/**
 * 停止实时数据流
 * @param {string} symbol - 交易对
 * @returns {Promise<void>}
 */
export async function stopMarketDataStream(symbol) {
  try {
    await window.go.main.App.StopMarketDataStream(symbol)
  } catch (error) {
    console.error('停止数据流失败:', error)
    throw error
  }
}

