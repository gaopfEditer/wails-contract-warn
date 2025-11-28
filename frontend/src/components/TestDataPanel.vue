<template>
  <div v-if="show" class="test-panel-overlay" @click.self="close">
    <div class="test-panel" @click.stop>
      <div class="test-panel-header">
        <h2>测试数据分析</h2>
        <button @click="close" class="close-btn">×</button>
      </div>
      
      <div class="test-panel-content">
        <!-- 测试数据选择 -->
        <div class="test-data-selector">
          <label>选择测试数据：</label>
          <select v-model="selectedTestFile" @change="loadTestData" class="test-select">
            <option value="test1.json">test1.json (混合形态)</option>
            <option value="test2.json">test2.json (上轨+锤子)</option>
            <option value="test3.json">test3.json (下轨+锤子)</option>
          </select>
          <button @click="loadTestData" class="load-btn" :disabled="loading">
            {{ loading ? '加载中...' : '加载数据' }}
          </button>
        </div>

        <!-- 分析结果 -->
        <div v-if="analysisResult" class="analysis-result">
          <div class="result-summary">
            <h3>分析结果汇总</h3>
            <div class="summary-stats">
              <div class="stat-item">
                <span class="stat-label">K线总数：</span>
                <span class="stat-value">{{ analysisResult.totalKlines }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">检测到信号：</span>
                <span class="stat-value">{{ analysisResult.totalSignals }}</span>
              </div>
              <div class="stat-item">
                <span class="stat-label">布林带上轨+锤子：</span>
                <span class="stat-value highlight">{{ analysisResult.bollingerHammerSignals }}</span>
              </div>
            </div>
          </div>

          <!-- 信号统计 -->
          <div v-if="analysisResult.signalStats && Object.keys(analysisResult.signalStats).length > 0" class="signal-stats">
            <h3>信号类型统计</h3>
            <div class="stats-grid">
              <div
                v-for="(count, type) in analysisResult.signalStats"
                :key="type"
                class="stat-card"
              >
                <div class="stat-type">{{ getSignalTypeName(type) }}</div>
                <div class="stat-count">{{ count }}</div>
              </div>
            </div>
          </div>

          <!-- 布林带上轨+锤子形态信号列表 -->
          <div v-if="analysisResult.signals && analysisResult.signals.length > 0" class="signals-list">
            <h3>布林带上轨+锤子形态信号 ({{ analysisResult.signals.length }})</h3>
            <div class="signal-items">
              <div
                v-for="(signal, index) in analysisResult.signals"
                :key="index"
                class="signal-item"
              >
                <div class="signal-header">
                  <span class="signal-type">{{ signal.type }}</span>
                  <span class="signal-index">K线 #{{ signal.index }}</span>
                </div>
                <div class="signal-details">
                  <div class="detail-row">
                    <span>时间：</span>
                    <span>{{ formatTime(signal.time) }}</span>
                  </div>
                  <div class="detail-row">
                    <span>收盘价：</span>
                    <span>${{ signal.close.toFixed(2) }}</span>
                  </div>
                  <div class="detail-row">
                    <span>布林带上轨：</span>
                    <span>${{ signal.upperBand.toFixed(2) }}</span>
                  </div>
                  <div class="detail-row">
                    <span>价格/上轨比率：</span>
                    <span>{{ (signal.analysis.bandRatio * 100).toFixed(2) }}%</span>
                  </div>
                  <div class="detail-row">
                    <span>信号强度：</span>
                    <span>{{ (signal.strength * 100).toFixed(1) }}%</span>
                  </div>
                  <div class="kline-details">
                    <div class="kline-row">
                      <span>开盘：</span><span>${{ signal.kline.open.toFixed(2) }}</span>
                      <span>最高：</span><span>${{ signal.kline.high.toFixed(2) }}</span>
                    </div>
                    <div class="kline-row">
                      <span>最低：</span><span>${{ signal.kline.low.toFixed(2) }}</span>
                      <span>收盘：</span><span>${{ signal.kline.close.toFixed(2) }}</span>
                    </div>
                    <div class="kline-analysis">
                      <span>实体：</span><span>${{ signal.analysis.body.toFixed(2) }}</span>
                      <span>上影线：</span><span>${{ signal.analysis.upperShadow.toFixed(2) }}</span>
                      <span>下影线：</span><span>${{ signal.analysis.lowerShadow.toFixed(2) }}</span>
                      <span class="hammer-indicator" v-if="signal.analysis.isHammer">✓ 锤子形态</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="no-signals">
            未检测到符合条件的信号
          </div>
        </div>

        <!-- 加载状态 -->
        <div v-if="loading" class="loading">
          正在分析数据...
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'

export default {
  name: 'TestDataPanel',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['close', 'load-test-data'],
  setup(props, { emit }) {
    const selectedTestFile = ref('test1.json')
    const loading = ref(false)
    const analysisResult = ref(null)

    const loadTestData = async () => {
      if (loading.value) return

      loading.value = true
      analysisResult.value = null

      try {
        // 调用后端方法分析测试数据（使用完整分析方法）
        const resultStr = await window.go.main.App.AnalyzeTestData(selectedTestFile.value)
        const result = JSON.parse(resultStr)
        
        // 筛选布林带上轨+锤子形态的信号
        const bollingerHammerSignals = []
        if (result.allSignals && result.klines) {
          for (const sig of result.allSignals) {
            // 检查是否是布林带上轨相关的信号
            if (sig.upperBand > 0 && sig.close >= sig.upperBand * 0.98) {
              const idx = sig.index
              if (idx < result.klines.length) {
                const k = result.klines[idx]
                const body = Math.abs(k.close - k.open)
                const upperShadow = k.high - Math.max(k.open, k.close)
                const lowerShadow = Math.min(k.open, k.close) - k.low
                
                // 锤子形态条件
                const isHammer = lowerShadow > body * 2 && upperShadow < body * 0.5
                
                if (isHammer) {
                  bollingerHammerSignals.push({
                    index: sig.index,
                    time: sig.time,
                    price: sig.price,
                    close: sig.close,
                    upperBand: sig.upperBand,
                    type: '布林带上轨+锤子形态',
                    strength: sig.strength || 0.8,
                    kline: {
                      open: k.open,
                      high: k.high,
                      low: k.low,
                      close: k.close,
                    },
                    analysis: {
                      body: body,
                      upperShadow: upperShadow,
                      lowerShadow: lowerShadow,
                      isHammer: isHammer,
                      bandRatio: sig.close / sig.upperBand,
                    },
                  })
                }
              }
            }
          }
        }
        
        analysisResult.value = {
          totalKlines: result.totalKlines || 0,
          totalSignals: result.totalSignals || 0,
          bollingerHammerSignals: bollingerHammerSignals.length,
          signals: bollingerHammerSignals,
          signalStats: result.signalStats || {},
          signalDetails: result.signalDetails || {},
        }
        
        // 通知父组件加载测试数据到图表
        emit('load-test-data', {
          filename: selectedTestFile.value,
          klines: result.klines || [],
          signals: result.allSignals || [],
          indicators: result.indicators || {},
        })
      } catch (error) {
        console.error('加载测试数据失败:', error)
        alert('加载测试数据失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const close = () => {
      emit('close')
    }

    const formatTime = (timestamp) => {
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
      })
    }

    const getSignalTypeName = (type) => {
      const typeMap = {
        'bollinger_doji_bottom': '布林带下轨十字星',
        'bollinger_hammer_bottom': '布林带下轨锤子',
        'bollinger_consecutive_hammers': '布林带下轨连续锤子',
        'bollinger_bullish_engulfing': '布林带下轨看涨吞没',
        'bollinger_hanging_man_top': '布林带上轨吊颈',
        'bollinger_bearish_engulfing': '布林带上轨看跌吞没',
      }
      return typeMap[type] || type
    }

    return {
      selectedTestFile,
      loading,
      analysisResult,
      loadTestData,
      close,
      formatTime,
      getSignalTypeName,
    }
  },
}
</script>

<style scoped>
.test-panel-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.test-panel {
  background: #1b2636;
  border: 1px solid #3e3e42;
  border-radius: 8px;
  width: 90%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
}

.test-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #3e3e42;
}

.test-panel-header h2 {
  margin: 0;
  color: #ffffff;
  font-size: 18px;
}

.close-btn {
  background: transparent;
  border: none;
  color: #858585;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
  line-height: 1;
  transition: color 0.2s;
}

.close-btn:hover {
  color: #ffffff;
}

.test-panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.test-data-selector {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
  padding: 16px;
  background: #252526;
  border-radius: 6px;
}

.test-data-selector label {
  color: #cccccc;
  font-size: 14px;
}

.test-select {
  flex: 1;
  padding: 8px 12px;
  background: #2d3748;
  border: 1px solid #4a5568;
  border-radius: 4px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
}

.load-btn {
  padding: 8px 16px;
  background: #48bb78;
  border: 1px solid #48bb78;
  border-radius: 4px;
  color: #ffffff;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.load-btn:hover:not(:disabled) {
  background: #38a169;
}

.load-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.analysis-result {
  margin-top: 20px;
}

.result-summary {
  background: #252526;
  padding: 16px;
  border-radius: 6px;
  margin-bottom: 20px;
}

.result-summary h3 {
  margin: 0 0 12px 0;
  color: #ffffff;
  font-size: 16px;
}

.summary-stats {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  color: #858585;
  font-size: 12px;
}

.stat-value {
  color: #4ec9b0;
  font-size: 18px;
  font-weight: 600;
}

.stat-value.highlight {
  color: #ff6b6b;
}

.signals-list h3 {
  color: #ffffff;
  font-size: 16px;
  margin-bottom: 12px;
}

.signal-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.signal-item {
  background: #252526;
  border: 1px solid #3e3e42;
  border-radius: 6px;
  padding: 16px;
  transition: all 0.2s;
}

.signal-item:hover {
  border-color: #4ec9b0;
  background: #2d2d30;
}

.signal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #3e3e42;
}

.signal-type {
  color: #4ec9b0;
  font-weight: 600;
  font-size: 14px;
}

.signal-index {
  color: #858585;
  font-size: 12px;
}

.signal-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
}

.detail-row span:first-child {
  color: #858585;
}

.detail-row span:last-child {
  color: #cccccc;
}

.kline-details {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #3e3e42;
}

.kline-row {
  display: flex;
  gap: 16px;
  font-size: 12px;
  margin-bottom: 6px;
}

.kline-row span {
  color: #858585;
}

.kline-row span:nth-child(even) {
  color: #cccccc;
  margin-left: 4px;
}

.kline-analysis {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  font-size: 12px;
  margin-top: 8px;
}

.kline-analysis span {
  color: #858585;
}

.kline-analysis span:nth-child(even) {
  color: #cccccc;
  margin-left: 4px;
}

.hammer-indicator {
  color: #48bb78 !important;
  font-weight: 600;
}

.no-signals {
  text-align: center;
  padding: 40px;
  color: #858585;
  font-style: italic;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #4ec9b0;
}

.signal-stats {
  margin-bottom: 24px;
}

.signal-stats h3 {
  color: #ffffff;
  font-size: 16px;
  margin-bottom: 12px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.stat-card {
  background: #252526;
  border: 1px solid #3e3e42;
  border-radius: 6px;
  padding: 12px;
  text-align: center;
}

.stat-type {
  color: #858585;
  font-size: 12px;
  margin-bottom: 6px;
}

.stat-count {
  color: #4ec9b0;
  font-size: 20px;
  font-weight: 600;
}
</style>

