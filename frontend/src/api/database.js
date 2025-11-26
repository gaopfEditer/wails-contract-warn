/**
 * 数据库API
 */

/**
 * 初始化数据库
 * @param {string} dsn - 数据库连接字符串
 * @returns {Promise<string>}
 */
export async function initDatabase(dsn) {
  try {
    return await window.go.main.App.InitDatabase(dsn)
  } catch (error) {
    console.error('数据库初始化失败:', error)
    throw error
  }
}

/**
 * 同步K线数据（增量）
 * @param {string} symbol - 交易对
 * @returns {Promise<string>}
 */
export async function syncKlineData(symbol) {
  try {
    return await window.go.main.App.SyncKlineData(symbol)
  } catch (error) {
    console.error('同步数据失败:', error)
    throw error
  }
}

/**
 * 首次同步K线数据（拉取历史）
 * @param {string} symbol - 交易对
 * @param {number} days - 天数
 * @returns {Promise<string>}
 */
export async function syncKlineDataInitial(symbol, days) {
  try {
    return await window.go.main.App.SyncKlineDataInitial(symbol, days)
  } catch (error) {
    console.error('初始同步失败:', error)
    throw error
  }
}

/**
 * 启动自动同步
 * @param {string} symbol - 交易对
 * @param {number} intervalSeconds - 同步间隔（秒）
 * @returns {Promise<string>}
 */
export async function startAutoSync(symbol, intervalSeconds) {
  try {
    return await window.go.main.App.StartAutoSync(symbol, intervalSeconds)
  } catch (error) {
    console.error('启动自动同步失败:', error)
    throw error
  }
}

/**
 * 停止自动同步
 * @param {string} symbol - 交易对
 * @returns {Promise<string>}
 */
export async function stopAutoSync(symbol) {
  try {
    return await window.go.main.App.StopAutoSync(symbol)
  } catch (error) {
    console.error('停止自动同步失败:', error)
    throw error
  }
}

