/**
 * 市场数据API
 */

/**
 * 等待 Wails 绑定初始化
 * @returns {Promise<void>}
 */
async function waitForWailsBinding() {
  if (window.go && window.go.main && window.go.main.App) {
    return
  }
  
  console.warn('Wails 绑定尚未初始化，等待初始化...')
  return new Promise((resolve, reject) => {
    const checkInterval = setInterval(() => {
      if (window.go && window.go.main && window.go.main.App) {
        clearInterval(checkInterval)
        resolve()
      }
    }, 100)
    
    // 最多等待 5 秒
    setTimeout(() => {
      clearInterval(checkInterval)
      if (!window.go || !window.go.main || !window.go.main.App) {
        reject(new Error('Wails 绑定初始化超时'))
      } else {
        resolve()
      }
    }, 5000)
  })
}

/**
 * 获取K线数据
 * @param {string} symbol - 交易对
 * @param {string} period - 周期
 * @returns {Promise<Array>}
 */
export async function getMarketData(symbol, period) {
  try {
    await waitForWailsBinding()
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
    await waitForWailsBinding()
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
    await waitForWailsBinding()
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
    await waitForWailsBinding()
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
    await waitForWailsBinding()
    await window.go.main.App.StopMarketDataStream(symbol)
  } catch (error) {
    console.error('停止数据流失败:', error)
    throw error
  }
}

