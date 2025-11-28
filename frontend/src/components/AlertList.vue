<template>
  <div class="alert-list-container">
    <div class="alert-header">
      <span class="alert-title">预警信息</span>
      <div class="alert-controls">
        <button @click="clearAlerts" class="btn-clear">清空</button>
        <button @click="toggleAutoRefresh" class="btn-toggle">
          {{ autoRefresh ? '暂停刷新' : '自动刷新' }}
        </button>
      </div>
    </div>
    <div class="alert-content">
      <div v-if="alerts.length === 0" class="empty-alerts">
        暂无预警信息
      </div>
      <div
        v-for="(alert, index) in alerts"
        :key="index"
        :class="['alert-item', `alert-${alert.level}`]"
      >
        <div class="alert-icon">
          <span v-if="alert.level === 'error'">⚠️</span>
          <span v-else-if="alert.level === 'warn'">⚠️</span>
          <span v-else-if="alert.level === 'info'">ℹ️</span>
          <span v-else>✓</span>
        </div>
        <div class="alert-body">
          <div class="alert-header-row">
            <span class="alert-time">{{ formatTime(alert.time) }}</span>
            <span class="alert-level">{{ getLevelText(alert.level) }}</span>
          </div>
          <div class="alert-message">{{ alert.message }}</div>
          <div v-if="alert.symbol" class="alert-meta">
            <span class="alert-symbol">交易对: {{ alert.symbol }}</span>
            <span v-if="alert.period" class="alert-period">周期: {{ alert.period }}</span>
          </div>
        </div>
        <button @click="removeAlert(index)" class="btn-remove">×</button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'

export default {
  name: 'AlertList',
  setup() {
    const alerts = ref([])
    const autoRefresh = ref(true)
    const maxAlerts = 500 // 最多保留500条预警

    // 格式化时间
    const formatTime = (timestamp) => {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
      })
    }

    // 获取级别文本
    const getLevelText = (level) => {
      const levelMap = {
        error: '错误',
        warn: '警告',
        info: '信息',
        success: '成功'
      }
      return levelMap[level] || '未知'
    }

    // 添加预警
    const addAlert = (message, level = 'info', symbol = null, period = null) => {
      const alert = {
        time: Date.now(),
        message,
        level,
        symbol,
        period
      }
      
      alerts.value.unshift(alert) // 新预警添加到顶部
      
      // 限制预警数量
      if (alerts.value.length > maxAlerts) {
        alerts.value.pop()
      }
    }

    // 移除预警
    const removeAlert = (index) => {
      alerts.value.splice(index, 1)
    }

    // 清空预警
    const clearAlerts = () => {
      alerts.value = []
    }

    // 切换自动刷新
    const toggleAutoRefresh = () => {
      autoRefresh.value = !autoRefresh.value
    }

    // 添加测试预警
    const addTestAlerts = () => {
      const testAlerts = [
        { message: '检测到价格突破阻力位', level: 'warn', symbol: 'BTCUSDT', period: '1m' },
        { message: 'RSI 指标超买', level: 'warn', symbol: 'BTCUSDT', period: '5m' },
        { message: 'MACD 金叉信号', level: 'info', symbol: 'ETHUSDT', period: '15m' },
        { message: '连接超时', level: 'error', symbol: null, period: null },
      ]
      
      testAlerts.forEach((alert, index) => {
        setTimeout(() => {
          addAlert(alert.message, alert.level, alert.symbol, alert.period)
        }, index * 200)
      })
    }

    onMounted(() => {
      // 测试：添加一些示例预警
      // addTestAlerts()
      
      // TODO: 从后端获取预警
      // 可以定期轮询或使用 WebSocket
    })

    onUnmounted(() => {
      // 清理定时器
    })

    // 暴露方法供父组件调用
    return {
      alerts,
      autoRefresh,
      formatTime,
      getLevelText,
      addAlert,
      removeAlert,
      clearAlerts,
      toggleAutoRefresh,
    }
  }
}
</script>

<style scoped>
.alert-list-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1b2636;
  color: #d4d4d4;
}

.alert-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #252526;
  border-bottom: 1px solid #3e3e42;
}

.alert-title {
  font-weight: 600;
  color: #cccccc;
  font-size: 14px;
}

.alert-controls {
  display: flex;
  gap: 8px;
}

.btn-clear,
.btn-toggle {
  padding: 4px 12px;
  background: #3e3e42;
  border: 1px solid #3e3e42;
  color: #cccccc;
  border-radius: 3px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
}

.btn-clear:hover,
.btn-toggle:hover {
  background: #4a4a4a;
  border-color: #4a4a4a;
}

.alert-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.alert-content::-webkit-scrollbar {
  width: 10px;
}

.alert-content::-webkit-scrollbar-track {
  background: #1b2636;
}

.alert-content::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 5px;
}

.alert-content::-webkit-scrollbar-thumb:hover {
  background: #4e4e4e;
}

.empty-alerts {
  color: #858585;
  text-align: center;
  padding: 40px;
  font-style: italic;
}

.alert-item {
  display: flex;
  align-items: flex-start;
  padding: 12px;
  margin-bottom: 12px;
  background: #252526;
  border-left: 4px solid;
  border-radius: 4px;
  transition: all 0.2s;
}

.alert-item:hover {
  background: #2d2d30;
}

.alert-error {
  border-left-color: #f48771;
  background: rgba(244, 135, 113, 0.1);
}

.alert-warn {
  border-left-color: #dcdcaa;
  background: rgba(220, 220, 170, 0.1);
}

.alert-info {
  border-left-color: #4ec9b0;
  background: rgba(78, 201, 176, 0.1);
}

.alert-success {
  border-left-color: #6a9955;
  background: rgba(106, 153, 85, 0.1);
}

.alert-icon {
  font-size: 20px;
  margin-right: 12px;
  flex-shrink: 0;
}

.alert-body {
  flex: 1;
  min-width: 0;
}

.alert-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.alert-time {
  color: #858585;
  font-size: 12px;
}

.alert-level {
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 600;
}

.alert-error .alert-level {
  background: rgba(244, 135, 113, 0.2);
  color: #f48771;
}

.alert-warn .alert-level {
  background: rgba(220, 220, 170, 0.2);
  color: #dcdcaa;
}

.alert-info .alert-level {
  background: rgba(78, 201, 176, 0.2);
  color: #4ec9b0;
}

.alert-success .alert-level {
  background: rgba(106, 153, 85, 0.2);
  color: #6a9955;
}

.alert-message {
  color: #d4d4d4;
  font-size: 14px;
  margin-bottom: 6px;
  word-break: break-word;
}

.alert-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #858585;
}

.alert-symbol,
.alert-period {
  padding: 2px 6px;
  background: rgba(62, 62, 66, 0.5);
  border-radius: 3px;
}

.btn-remove {
  background: transparent;
  border: none;
  color: #858585;
  font-size: 20px;
  cursor: pointer;
  padding: 0 8px;
  line-height: 1;
  transition: color 0.2s;
  flex-shrink: 0;
}

.btn-remove:hover {
  color: #f48771;
}
</style>

