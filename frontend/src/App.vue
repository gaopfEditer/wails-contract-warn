<template>
  <div class="app-container">
    <AppHeader
      :symbol="selectedSymbol"
      :period="selectedPeriod"
      :is-streaming="isStreaming"
      @update:symbol="selectedSymbol = $event"
      @update:period="selectedPeriod = $event"
      @toggle-stream="toggleStream"
      @test-click="showTestPanel = true"
    />
    <AlertBanner
      v-if="latestAlert"
      :alert="latestAlert"
      @close="latestAlert = null"
    />
    <main class="app-main">
      <TabView
        :kline-data="klineData"
        :indicators="indicators"
        :alert-signals="alertSignals"
        :symbol="selectedSymbol"
      />
    </main>
    <TestDataPanel
      :show="showTestPanel"
      @close="showTestPanel = false"
      @load-test-data="handleLoadTestData"
    />
  </div>
</template>

<script>
import { ref } from 'vue'
import AppHeader from './components/AppHeader.vue'
import TabView from './components/TabView.vue'
import AlertBanner from './components/AlertBanner.vue'
import TestDataPanel from './components/TestDataPanel.vue'
import { useMarketData } from './composables/useMarketData'

export default {
  name: 'App',
  components: {
    AppHeader,
    TabView,
    AlertBanner,
    TestDataPanel,
  },
  setup() {
    const showTestPanel = ref(false)
    
    // 使用组合式函数
    const {
      selectedSymbol,
      selectedPeriod,
      klineData,
      indicators,
      alertSignals,
      latestAlert,
      isStreaming,
      toggleStream,
      loadTestData: loadTestDataToChart,
    } = useMarketData('BTCUSDT', '1m')

    const handleLoadTestData = async (data) => {
      // 将测试数据加载到图表
      if (loadTestDataToChart) {
        await loadTestDataToChart(data.klines, data.signals, data.indicators)
      }
      // 切换到图表标签页查看结果
      // 可以通过 TabView 组件的方法切换，这里先保持面板打开
    }

    return {
      selectedSymbol,
      selectedPeriod,
      klineData,
      indicators,
      alertSignals,
      latestAlert,
      isStreaming,
      toggleStream,
      showTestPanel,
      handleLoadTestData,
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

.app-main {
  flex: 1;
  padding: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
