import { ref, onMounted, onUnmounted, watch } from 'vue'
import { getMarketData, getIndicators, getAlertSignals, startMarketDataStream, stopMarketDataStream } from '../api/market'
import { getLatestAlert } from '../utils/indicators'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

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
  let currentStreamPeriod = null // 记录当前数据流的周期
  let isLoading = false // 防止并发加载
  let currentLoadId = 0 // 用于标识加载请求，确保只使用最新的结果

  // 加载数据
  const loadData = async () => {
    // 生成新的加载ID
    const loadId = ++currentLoadId
    
    // 防止并发加载
    if (isLoading) {
      // 如果正在加载，直接返回，让调用者稍后重试
      return
    }
    isLoading = true
    
    try {
      // 记录当前要加载的周期和交易对，用于验证数据是否仍然有效
      const loadPeriod = selectedPeriod.value
      const loadSymbol = selectedSymbol.value
      
      // 并行获取所有数据
      const [klineDataResult, indicatorsResult, signalsResult] = await Promise.all([
        getMarketData(loadSymbol, loadPeriod),
        getIndicators(loadSymbol, loadPeriod),
        getAlertSignals(loadSymbol, loadPeriod),
      ])

      // 检查这个加载请求是否仍然有效（周期或交易对可能已变化）
      if (loadId !== currentLoadId) {
        // 这个请求已经过时，丢弃结果
        console.log(`加载请求已过时，丢弃结果: loadId=${loadId}, currentLoadId=${currentLoadId}`)
        return
      }
      if (loadPeriod !== selectedPeriod.value || loadSymbol !== selectedSymbol.value) {
        // 周期或交易对已变化，丢弃这次加载的结果
        console.log(`周期或交易对已变化，丢弃加载结果: loadPeriod=${loadPeriod}, currentPeriod=${selectedPeriod.value}, loadSymbol=${loadSymbol}, currentSymbol=${selectedSymbol.value}`)
        return
      }

      // 只有在数据有效时才更新
      if (klineDataResult && klineDataResult.length > 0) {
        klineData.value = klineDataResult
        indicators.value = indicatorsResult || {}
        alertSignals.value = signalsResult || []
        latestAlert.value = getLatestAlert(signalsResult || [])
        console.log(`成功加载 ${klineDataResult.length} 条K线数据: symbol=${loadSymbol}, period=${loadPeriod}`)
      } else {
        // 如果没有数据，只更新指标和信号，不清空K线数据
        indicators.value = indicatorsResult || {}
        alertSignals.value = signalsResult || []
        latestAlert.value = getLatestAlert(signalsResult || [])
        console.warn(`没有获取到K线数据: symbol=${loadSymbol}, period=${loadPeriod}，保留现有数据`)
      }
      
    } catch (error) {
      // 检查这个加载请求是否仍然有效
      if (loadId !== currentLoadId) {
        return
      }
      console.error('加载数据失败:', error)
      // 发生错误时，不清空数据，保留现有数据
      // 只有在明确知道是周期切换导致的错误时才清空
      // 这里保留数据，让用户看到之前的数据，而不是空白
    } finally {
      // 只有当前加载完成时才重置isLoading
      if (loadId === currentLoadId) {
        isLoading = false
      }
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
        currentStreamPeriod = null
      } catch (error) {
        console.error('停止数据流失败:', error)
      }
    } else {
      // 开始流
      try {
        await startMarketDataStream(selectedSymbol.value, selectedPeriod.value)
        isStreaming.value = true
        currentStreamPeriod = selectedPeriod.value
        const streamSymbol = selectedSymbol.value
        const streamPeriod = selectedPeriod.value
        
        // 立即触发一次数据同步，确保获取最新数据
        try {
          await window.go.main.App.SyncSymbolData(streamSymbol, 1)
          console.log(`已触发币种 ${streamSymbol} 的实时数据同步`)
        } catch (error) {
          console.error(`触发实时数据同步失败:`, error)
        }
        
        // 立即加载一次数据
        loadData()
        
        // 每2秒更新一次数据，提高实时性（从1秒改为2秒，避免过于频繁）
        updateTimer = setInterval(() => {
          // 确保定时器中使用的是当前周期和交易对
          // 如果周期或交易对已变化，停止定时器
          if (selectedPeriod.value !== streamPeriod || selectedSymbol.value !== streamSymbol) {
            if (updateTimer) {
              clearInterval(updateTimer)
              updateTimer = null
            }
            return
          }
          loadData()
        }, 2000) // 2秒更新一次
      } catch (error) {
        console.error('启动数据流失败:', error)
      }
    }
  }

  // 监听交易对和周期变化
  watch([selectedSymbol, selectedPeriod], async (newValues, oldValues) => {
    const [newSymbol, newPeriod] = newValues
    const [oldSymbol, oldPeriod] = oldValues || [null, null]
    
    // 立即清除旧的定时器，防止使用旧的周期参数
    if (updateTimer) {
      clearInterval(updateTimer)
      updateTimer = null
    }
    
    // 如果周期或交易对变化且实时数据流正在运行，需要重新启动数据流
    if (isStreaming.value && (newPeriod !== currentStreamPeriod || newSymbol !== oldSymbol)) {
      // 先停止旧的数据流（只在交易对变化时停止，周期变化时直接更新）
      if (newSymbol !== oldSymbol) {
        try {
          await stopMarketDataStream(oldSymbol)
        } catch (error) {
          console.error('停止旧数据流失败:', error)
        }
      }
      
      // 重新启动数据流，使用新的交易对和周期
      try {
        await startMarketDataStream(newSymbol, newPeriod)
        currentStreamPeriod = newPeriod
        
        // 立即触发一次数据同步，确保获取最新数据
        try {
          await window.go.main.App.SyncSymbolData(newSymbol, 1)
          console.log(`已触发币种 ${newSymbol} 的实时数据同步`)
        } catch (error) {
          console.error(`触发实时数据同步失败:`, error)
        }
        
        // 重新启动定时器
        const timerSymbol = newSymbol
        const timerPeriod = newPeriod
        updateTimer = setInterval(() => {
          // 确保定时器中使用的是当前周期和交易对
          // 如果周期或交易对已变化，停止定时器
          if (selectedPeriod.value !== timerPeriod || selectedSymbol.value !== timerSymbol) {
            if (updateTimer) {
              clearInterval(updateTimer)
              updateTimer = null
            }
            return
          }
          loadData()
        }, 2000) // 2秒更新一次，提高实时性
        console.log(`交易对或周期已切换，重新启动数据流: symbol=${newSymbol}, period=${newPeriod}`)
      } catch (error) {
        console.error('重新启动数据流失败:', error)
        // 如果重新启动失败，停止流状态
        isStreaming.value = false
        currentStreamPeriod = null
      }
    }
    
    // 立即加载新周期的数据（不等待数据流重新启动）
    loadData()
    
    // 切换币种时，触发数据获取（在后台进行，不阻塞数据加载）
    if (newSymbol && newSymbol !== oldSymbol) {
      // 异步触发数据同步，不阻塞当前数据加载
      window.go.main.App.SyncSymbolData(newSymbol, 1).then(() => {
        console.log(`已触发币种 ${newSymbol} 的数据获取`)
      }).catch((error) => {
        console.error(`触发币种 ${newSymbol} 数据获取失败:`, error)
      })
    }
  })

  // 监听实时价格更新
  let realtimePriceUnsubscribe = null

  onMounted(() => {
    loadData()

    // 监听实时价格更新事件
    realtimePriceUnsubscribe = EventsOn('realtime-price', (priceData) => {
      if (!priceData || !priceData.symbol) return

      // 如果当前选中的交易对匹配，更新数据
      const normalizedSymbol = priceData.symbol.replace('_', '')
      if (normalizedSymbol === selectedSymbol.value || priceData.symbol === selectedSymbol.value) {
        // 严格检查周期：实时价格事件只应该在1m周期时处理
        // 如果当前周期不是1m，或者数据流周期不是1m，则完全忽略
        if (selectedPeriod.value !== '1m') {
          // 非1m周期，定时器会处理数据更新，这里不需要额外处理
          return
        }
        
        // 再次检查数据流周期，确保一致性
        if (currentStreamPeriod !== '1m' && isStreaming.value) {
          // 数据流周期不匹配，忽略
          return
        }
        
        console.log('收到实时价格更新:', priceData)
        
        // 当前周期是1m，直接更新最后一条K线或添加新K线
        if (klineData.value.length > 0) {
          const lastKline = klineData.value[klineData.value.length - 1]
          const priceTime = priceData.time || priceData.timestamp
          
          // 如果是同一分钟的数据，更新最后一条
          if (lastKline && Math.abs(lastKline.time - priceTime) < 60000) {
            lastKline.open = priceData.open
            lastKline.high = Math.max(lastKline.high, priceData.high)
            lastKline.low = Math.min(lastKline.low, priceData.low)
            lastKline.close = priceData.close
            lastKline.volume = priceData.volume
          } else {
            // 添加新的K线
            klineData.value.push({
              time: priceTime,
              open: priceData.open,
              high: priceData.high,
              low: priceData.low,
              close: priceData.close,
              volume: priceData.volume,
            })
            // 保持最多1000条数据
            if (klineData.value.length > 1000) {
              klineData.value = klineData.value.slice(-1000)
            }
          }
          
          // 重新计算指标和信号（只在1m周期时）
          // 但需要再次检查周期，确保没有在更新过程中切换周期
          if (selectedPeriod.value === '1m') {
            loadData()
          }
        }
      }
    })
  })

  onUnmounted(() => {
    if (updateTimer) {
      clearInterval(updateTimer)
    }
    if (isStreaming.value) {
      stopMarketDataStream(selectedSymbol.value).catch(console.error)
    }
    // 取消实时价格监听
    if (realtimePriceUnsubscribe) {
      realtimePriceUnsubscribe()
      realtimePriceUnsubscribe = null
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

