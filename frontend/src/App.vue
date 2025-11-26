<template>
  <div class="app-container">
    <header class="app-header">
      <h1>合约行情分析</h1>
      <div class="controls">
        <select v-model="selectedSymbol" @change="loadData" class="symbol-select">
          <option value="BTCUSDT">BTC/USDT</option>
          <option value="ETHUSDT">ETH/USDT</option>
        </select>
        <select v-model="selectedPeriod" @change="loadData" class="period-select">
          <option value="1m">1分钟</option>
          <option value="5m">5分钟</option>
          <option value="15m">15分钟</option>
          <option value="1h">1小时</option>
        </select>
        <button @click="toggleStream" class="stream-btn" :class="{ active: isStreaming }">
          {{ isStreaming ? '停止' : '开始' }}实时数据
        </button>
      </div>
    </header>
    <main class="app-main">
      <KLineChart 
        :kline-data="klineData" 
        :indicators="indicators"
        :symbol="selectedSymbol"
      />
    </main>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import KLineChart from './components/KLineChart.vue'

export default {
  name: 'App',
  components: {
    KLineChart,
  },
  setup() {
    const selectedSymbol = ref('BTCUSDT')
    const selectedPeriod = ref('1m')
    const klineData = ref([])
    const indicators = ref({})
    const isStreaming = ref(false)
    let updateTimer = null

    // 加载数据
    const loadData = async () => {
      try {
        // 调用 Go 后端获取 K 线数据
        const dataStr = await window.go.main.App.GetMarketData(
          selectedSymbol.value,
          selectedPeriod.value
        )
        klineData.value = JSON.parse(dataStr)

        // 获取技术指标
        const indicatorsStr = await window.go.main.App.GetIndicators(
          selectedSymbol.value,
          selectedPeriod.value
        )
        indicators.value = JSON.parse(indicatorsStr)
      } catch (error) {
        console.error('加载数据失败:', error)
      }
    }

    // 开始/停止数据流
    const toggleStream = async () => {
      if (isStreaming.value) {
        // 停止流
        try {
          await window.go.main.App.StopMarketDataStream(selectedSymbol.value)
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
          await window.go.main.App.StartMarketDataStream(
            selectedSymbol.value,
            selectedPeriod.value
          )
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

    onMounted(() => {
      loadData()
    })

    onUnmounted(() => {
      if (updateTimer) {
        clearInterval(updateTimer)
      }
      if (isStreaming.value) {
        window.go.main.App.StopMarketDataStream(selectedSymbol.value).catch(console.error)
      }
    })

    return {
      selectedSymbol,
      selectedPeriod,
      klineData,
      indicators,
      isStreaming,
      loadData,
      toggleStream,
    }
  },
}
</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1b2636;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: #252d3a;
  border-bottom: 1px solid #2d3748;
}

.app-header h1 {
  font-size: 20px;
  font-weight: 600;
  color: #ffffff;
}

.controls {
  display: flex;
  gap: 12px;
  align-items: center;
}

.symbol-select,
.period-select {
  padding: 8px 12px;
  background: #2d3748;
  border: 1px solid #4a5568;
  border-radius: 4px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  outline: none;
}

.symbol-select:hover,
.period-select:hover {
  border-color: #63b3ed;
}

.stream-btn {
  padding: 8px 16px;
  background: #2d3748;
  border: 1px solid #4a5568;
  border-radius: 4px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.stream-btn:hover {
  background: #374151;
  border-color: #63b3ed;
}

.stream-btn.active {
  background: #3182ce;
  border-color: #3182ce;
}

.app-main {
  flex: 1;
  padding: 16px;
  overflow: hidden;
}
</style>

