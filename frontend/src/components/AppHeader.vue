<template>
  <header class="app-header">
    <h1>数据分析225577</h1>
    <div class="controls">
      <select v-model="localSymbol" @change="handleSymbolChange" class="symbol-select">
        <option value="BTCUSDT">BTC/USDT</option>
        <option value="ETHUSDT">ETH/USDT</option>
      </select>
      <select v-model="localPeriod" @change="handlePeriodChange" class="period-select">
        <optgroup label="分钟">
          <option value="1m">1分钟</option>
          <option value="5m">5分钟</option>
          <option value="15m">15分钟</option>
          <option value="30m">30分钟</option>
        </optgroup>
        <optgroup label="小时">
          <option value="1h">1小时</option>
          <option value="2h">2小时</option>
          <option value="3h">3小时</option>
          <option value="4h">4小时</option>
        </optgroup>
        <optgroup label="日/周/月">
          <option value="1d">日线</option>
          <option value="1w">周线</option>
          <option value="1M">月线</option>
        </optgroup>
      </select>
      <button @click="handleToggleStream" class="stream-btn" :class="{ active: isStreaming }">
        {{ isStreaming ? '停止' : '开始' }}实时数据
      </button>
      <button @click="handleTestClick" class="test-btn">
        测试
      </button>
    </div>
  </header>
</template>

<script>
import { ref, watch } from 'vue'

export default {
  name: 'AppHeader',
  props: {
    symbol: {
      type: String,
      default: 'BTCUSDT',
    },
    period: {
      type: String,
      default: '1m',
    },
    isStreaming: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:symbol', 'update:period', 'toggle-stream', 'test-click'],
  setup(props, { emit }) {
    const localSymbol = ref(props.symbol)
    const localPeriod = ref(props.period)

    watch(() => props.symbol, (newVal) => {
      localSymbol.value = newVal
    })

    watch(() => props.period, (newVal) => {
      localPeriod.value = newVal
    })

    const handleSymbolChange = () => {
      emit('update:symbol', localSymbol.value)
    }

    const handlePeriodChange = () => {
      emit('update:period', localPeriod.value)
    }

    const handleToggleStream = () => {
      emit('toggle-stream')
    }

    const handleTestClick = () => {
      emit('test-click')
    }

    return {
      localSymbol,
      localPeriod,
      handleSymbolChange,
      handlePeriodChange,
      handleToggleStream,
      handleTestClick,
    }
  },
}
</script>

<style scoped>
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

.test-btn {
  padding: 8px 16px;
  background: #48bb78;
  border: 1px solid #48bb78;
  border-radius: 4px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.test-btn:hover {
  background: #38a169;
  border-color: #38a169;
}
</style>

