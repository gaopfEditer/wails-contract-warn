<template>
  <Transition name="alert">
    <div v-if="alert" class="alert-banner" :class="alert.type" :style="alertStyle">
      <div class="alert-content">
        <span class="alert-icon">{{ signalConfig.icon }}</span>
        <div class="alert-text">
          <div class="alert-title">{{ signalConfig.name }}</div>
          <div class="alert-details">
            {{ alertDetails }}
          </div>
        </div>
        <div class="alert-strength" v-if="alert.strength">
          <span class="strength-label">强度:</span>
          <span class="strength-value">{{ (alert.strength * 100).toFixed(0) }}%</span>
        </div>
        <button class="alert-close" @click="handleClose">×</button>
      </div>
    </div>
  </Transition>
</template>

<script>
import { computed } from 'vue'
import { getSignalConfig } from '../utils/signalTypes'

export default {
  name: 'AlertBanner',
  props: {
    alert: {
      type: Object,
      default: null,
    },
  },
  emits: ['close'],
  setup(props, { emit }) {
    const signalConfig = computed(() => {
      if (!props.alert) return null
      return getSignalConfig(props.alert.type)
    })

    const alertStyle = computed(() => {
      if (!signalConfig.value) return {}
      return {
        background: `linear-gradient(135deg, ${signalConfig.value.bgColor} 0%, ${signalConfig.value.bgColor}dd 100%)`,
        borderLeftColor: signalConfig.value.borderColor,
      }
    })

    const alertDetails = computed(() => {
      if (!props.alert) return ''
      const date = new Date(props.alert.time)
      const timeStr = `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${date.getMinutes()}`
      let details = `时间: ${timeStr} | 价格: ${props.alert.price.toFixed(2)}`
      
      if (props.alert.lowerBand) {
        details += ` | 下轨: ${props.alert.lowerBand.toFixed(2)}`
      }
      if (props.alert.upperBand) {
        details += ` | 上轨: ${props.alert.upperBand.toFixed(2)}`
      }
      
      return details
    })

    const handleClose = () => {
      emit('close')
    }

    return {
      signalConfig,
      alertStyle,
      alertDetails,
      handleClose,
    }
  },
}
</script>

<style scoped>
.alert-banner {
  border-left: 4px solid;
  padding: 12px 16px;
  margin-bottom: 16px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  animation: pulse 2s infinite;
}

.alert-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.alert-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.alert-text {
  flex: 1;
}

.alert-title {
  font-weight: bold;
  font-size: 16px;
  margin-bottom: 4px;
}

.alert-details {
  font-size: 14px;
  opacity: 0.9;
}

.alert-strength {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  font-size: 12px;
}

.strength-label {
  opacity: 0.8;
}

.strength-value {
  font-weight: bold;
}

.alert-close {
  background: transparent;
  border: none;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background 0.2s;
  opacity: 0.7;
}

.alert-close:hover {
  background: rgba(0, 0, 0, 0.1);
  opacity: 1;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.8;
  }
}

.alert-enter-active,
.alert-leave-active {
  transition: all 0.3s ease;
}

.alert-enter-from,
.alert-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
