import { ref, onMounted, onUnmounted, watch } from 'vue'
import { getMarketData, getIndicators, getAlertSignals, startMarketDataStream, stopMarketDataStream } from '../api/market'
import { getLatestAlert } from '../utils/indicators'

/**
 * 市场数据组合式函数
 */
export function useMarketData(initialSymbol = 'BTCUSDT', initialPeriod = '1m') {
  const selectedSymbol = ref(initialSymbol)
  const selectedPeriod = ref(initialPeriod)
  const klineData = ref([])
  const indicators = ref({})
  const alertSignals = ref([])
  const latestAlert = ref(null)
  const isStreaming = ref(false)
  let updateTimer = null

  // 加载数据
  const loadData = async () => {
    try {
      // 并行获取所有数据
      const [klineDataResult, indicatorsResult, signalsResult] = await Promise.all([
        getMarketData(selectedSymbol.value, selectedPeriod.value),
        getIndicators(selectedSymbol.value, selectedPeriod.value),
        getAlertSignals(selectedSymbol.value, selectedPeriod.value),
      ])

      klineData.value = klineDataResult
      indicators.value = indicatorsResult
      alertSignals.value = signalsResult
      latestAlert.value = getLatestAlert(signalsResult)
    } catch (error) {
      console.error('加载数据失败:', error)
    }
  }

  // 开始/停止数据流
  const toggleStream = async () => {
    if (isStreaming.value) {
      // 停止流
      try {
        await stopMarketDataStream(selectedSymbol.value)
        if (updateTimer) {
          clearInterval(updateTimer)
          updateTimer = null
        }
        isStreaming.value = false
      } catch (error) {
        console.error('停止数据流失败:', error)
      }
    } else {
      // 开始流
      try {
        await startMarketDataStream(selectedSymbol.value, selectedPeriod.value)
        isStreaming.value = true
        // 每秒更新一次数据
        updateTimer = setInterval(() => {
          loadData()
        }, 1000)
      } catch (error) {
        console.error('启动数据流失败:', error)
      }
    }
  }

  // 监听交易对和周期变化
  watch([selectedSymbol, selectedPeriod], () => {
    loadData()
  })

  onMounted(() => {
    loadData()
  })

  onUnmounted(() => {
    if (updateTimer) {
      clearInterval(updateTimer)
    }
    if (isStreaming.value) {
      stopMarketDataStream(selectedSymbol.value).catch(console.error)
    }
  })

  return {
    selectedSymbol,
    selectedPeriod,
    klineData,
    indicators,
    alertSignals,
    latestAlert,
    isStreaming,
    loadData,
    toggleStream,
  }
}

