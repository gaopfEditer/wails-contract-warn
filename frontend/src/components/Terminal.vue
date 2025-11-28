<template>
  <div class="terminal-container">
    <div class="terminal-header">
      <span class="terminal-title">网络连接日志</span>
      <div class="terminal-controls">
        <button @click="clearLogs" class="btn-clear">清空</button>
        <button @click="toggleAutoScroll" class="btn-toggle">
          {{ autoScroll ? '暂停滚动' : '自动滚动' }}
        </button>
      </div>
    </div>
    <div ref="terminalContent" class="terminal-content">
      <div
        v-for="(log, index) in logs"
        :key="index"
        :class="['log-line', getLogClass(log)]"
      >
        <span class="log-time">{{ formatTime(log.time) }}</span>
        <span class="log-content">{{ log.content }}</span>
      </div>
      <div v-if="logs.length === 0" class="empty-logs">
        暂无日志信息
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'

export default {
  name: 'Terminal',
  setup() {
    const logs = ref([])
    const autoScroll = ref(true)
    const terminalContent = ref(null)
    const maxLogs = 1000 // 最多保留1000条日志

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

    // 获取日志类型样式
    const getLogClass = (log) => {
      if (log.type === 'error') return 'log-error'
      if (log.type === 'warn') return 'log-warn'
      if (log.type === 'info') return 'log-info'
      if (log.type === 'success') return 'log-success'
      return 'log-default'
    }

    // 添加日志
    const addLog = (content, type = 'default') => {
      const log = {
        time: Date.now(),
        content: content.trim(),
        type: type
      }
      
      logs.value.push(log)
      
      // 限制日志数量
      if (logs.value.length > maxLogs) {
        logs.value.shift()
      }

      // 自动滚动到底部
      if (autoScroll.value) {
        nextTick(() => {
          scrollToBottom()
        })
      }
    }

    // 清空日志
    const clearLogs = () => {
      logs.value = []
    }

    // 切换自动滚动
    const toggleAutoScroll = () => {
      autoScroll.value = !autoScroll.value
      if (autoScroll.value) {
        scrollToBottom()
      }
    }

    // 滚动到底部
    const scrollToBottom = () => {
      if (terminalContent.value) {
        terminalContent.value.scrollTop = terminalContent.value.scrollHeight
      }
    }

    // 解析日志内容并添加
    const parseAndAddLog = (logText) => {
      if (!logText || !logText.trim()) return

      const lines = logText.split('\n').filter(line => line.trim())
      
      lines.forEach(line => {
        line = line.trim()
        
        // 判断日志类型
        let type = 'default'
        if (line.includes('ERROR') || line.includes('error')) {
          type = 'error'
        } else if (line.includes('WARN') || line.includes('warn')) {
          type = 'warn'
        } else if (line.includes('INFO') || line.includes('info') || line.includes('accepted')) {
          type = 'info'
        } else if (line.includes('SUCCESS') || line.includes('success')) {
          type = 'success'
        }

        addLog(line, type)
      })
    }

    // 模拟添加日志（用于测试）
    const addTestLogs = () => {
      const testLogs = [
        '2025/11/28 09:48:57.802245 from tcp:127.0.0.1:65295 accepted tcp:104.128.62.173:443 [socks >> proxy]',
        '2025/11/28 09:48:59.193372 from tcp:127.0.0.1:63631 accepted tcp:20.190.160.132:443 [socks >> proxy]',
        '+0800 2025-11-28 09:49:54 ERROR [94688961 4m8s] connection: connection download closed: raw-read tcp 61.169.7.156:20083->192.168.2.82:12635: An existing connection was forcibly closed by the remote host.',
      ]
      
      testLogs.forEach((log, index) => {
        setTimeout(() => {
          parseAndAddLog(log)
        }, index * 100)
      })
    }

    // 监听日志更新（可以从后端获取）
    // TODO: 集成后端日志流
    let logInterval = null

    onMounted(() => {
      // 测试：添加一些示例日志
      // addTestLogs()
      
      // TODO: 从后端获取日志
      // 可以定期轮询或使用 WebSocket
    })

    onUnmounted(() => {
      if (logInterval) {
        clearInterval(logInterval)
      }
    })

    // 暴露方法供父组件调用
    return {
      logs,
      autoScroll,
      terminalContent,
      formatTime,
      getLogClass,
      addLog,
      clearLogs,
      toggleAutoScroll,
      parseAndAddLog,
    }
  }
}
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #252526;
  border-bottom: 1px solid #3e3e42;
}

.terminal-title {
  font-weight: 600;
  color: #cccccc;
}

.terminal-controls {
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

.terminal-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  scroll-behavior: smooth;
}

.terminal-content::-webkit-scrollbar {
  width: 10px;
}

.terminal-content::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.terminal-content::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 5px;
}

.terminal-content::-webkit-scrollbar-thumb:hover {
  background: #4e4e4e;
}

.log-line {
  display: flex;
  margin-bottom: 4px;
  word-break: break-all;
}

.log-time {
  color: #858585;
  margin-right: 12px;
  flex-shrink: 0;
  min-width: 180px;
}

.log-content {
  flex: 1;
  white-space: pre-wrap;
}

.log-default {
  color: #d4d4d4;
}

.log-info {
  color: #4ec9b0;
}

.log-success {
  color: #6a9955;
}

.log-warn {
  color: #dcdcaa;
}

.log-error {
  color: #f48771;
}

.empty-logs {
  color: #858585;
  text-align: center;
  padding: 40px;
  font-style: italic;
}
</style>

