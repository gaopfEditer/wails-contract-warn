<template>
  <div class="app-container">
    <AppHeader
      :symbol="selectedSymbol"
      :period="selectedPeriod"
      :is-streaming="isStreaming"
      @update:symbol="selectedSymbol = $event"
      @update:period="selectedPeriod = $event"
      @toggle-stream="toggleStream"
    />
    <AlertBanner
      v-if="latestAlert"
      :alert="latestAlert"
      @close="latestAlert = null"
    />
    <main class="app-main">
      <KLineChart
        :kline-data="klineData"
        :indicators="indicators"
        :alert-signals="alertSignals"
        :symbol="selectedSymbol"
      />
    </main>
  </div>
</template>

<script>
import AppHeader from './components/AppHeader.vue'
import KLineChart from './components/KLineChart.vue'
import AlertBanner from './components/AlertBanner.vue'
import { useMarketData } from './composables/useMarketData'

export default {
  name: 'App',
  components: {
    AppHeader,
    KLineChart,
    AlertBanner,
  },
  setup() {
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
    } = useMarketData('BTCUSDT', '1m')

    return {
      selectedSymbol,
      selectedPeriod,
      klineData,
      indicators,
      alertSignals,
      latestAlert,
      isStreaming,
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

.app-main {
  flex: 1;
  padding: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
