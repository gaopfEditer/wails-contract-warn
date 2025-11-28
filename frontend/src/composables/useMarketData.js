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

  // 加载测试数据到图表
  const loadTestData = async (testKlines, testSignals, testIndicators) => {
    try {
      klineData.value = testKlines || []
      alertSignals.value = testSignals || []
      
      // 使用提供的指标或重新计算
      if (testIndicators) {
        indicators.value = testIndicators
      } else if (testKlines && testKlines.length > 0) {
        // 如果没有提供指标，尝试从后端计算
        // 注意：这里需要将测试数据发送到后端计算指标
        // 或者在前端计算（如果前端有指标计算逻辑）
        console.warn('测试数据未包含技术指标，可能需要重新计算')
      }
      
      latestAlert.value = getLatestAlert(testSignals || [])
      
      console.log('测试数据已加载:', {
        klines: testKlines?.length || 0,
        signals: testSignals?.length || 0,
      })
    } catch (error) {
      console.error('加载测试数据失败:', error)
    }
  }

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
    loadTestData,
  }
}

