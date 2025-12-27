<template>
  <div class="tab-view-container">
    <div class="tab-header">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="['tab-button', { active: activeTab === tab.id }]"
        @click="activeTab = tab.id"
      >
        <span class="tab-icon">{{ tab.icon }}</span>
        <span class="tab-label">{{ tab.label }}</span>
        <span v-if="tab.badge" class="tab-badge">{{ tab.badge }}</span>
      </button>
    </div>
    <div class="tab-content">
      <Terminal v-if="activeTab === 'terminal'" />
      <KLineChart
        v-if="activeTab === 'chart'"
        :kline-data="klineData"
        :indicators="indicators"
        :alert-signals="alertSignals"
        :symbol="symbol"
        :period="period"
      />
      <AlertList v-if="activeTab === 'alert'" />
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import Terminal from './Terminal.vue'
import KLineChart from './KLineChart.vue'
import AlertList from './AlertList.vue'

export default {
  name: 'TabView',
  components: {
    Terminal,
    KLineChart,
    AlertList,
  },
  props: {
    klineData: {
      type: Array,
      default: () => [],
    },
    indicators: {
      type: Object,
      default: () => ({}),
    },
    alertSignals: {
      type: Array,
      default: () => [],
    },
    symbol: {
      type: String,
      default: 'BTCUSDT',
    },
    period: {
      type: String,
      default: '1m',
    },
  },
  setup() {
    const activeTab = ref('terminal')

    const tabs = [
      { id: 'terminal', label: '终端', icon: '💻' },
      { id: 'chart', label: '图表', icon: '📊' },
      { id: 'alert', label: '预警', icon: '⚠️' },
    ]

    return {
      activeTab,
      tabs,
    }
  }
}
</script>

<style scoped>
.tab-view-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1b2636;
}

.tab-header {
  display: flex;
  background: #252526;
  border-bottom: 1px solid #3e3e42;
  padding: 0 8px;
}

.tab-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: #858585;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
  position: relative;
}

.tab-button:hover {
  color: #cccccc;
  background: rgba(255, 255, 255, 0.05);
}

.tab-button.active {
  color: #4ec9b0;
  border-bottom-color: #4ec9b0;
  background: rgba(78, 201, 176, 0.1);
}

.tab-icon {
  font-size: 16px;
}

.tab-label {
  font-weight: 500;
}

.tab-badge {
  background: #f48771;
  color: white;
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 10px;
  min-width: 18px;
  text-align: center;
}

.tab-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}
</style>

