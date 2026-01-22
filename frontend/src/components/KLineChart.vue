<template>
  <div class="kline-chart-container">
    <div ref="chartContainer" class="chart-container"></div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getSignalConfig } from '../utils/signalTypes'

export default {
  name: 'KLineChart',
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
  setup(props) {
    const chartContainer = ref(null)
    let chartInstance = null
    let savedDataZoom = null // 保存用户的dataZoom状态
    let isUserViewingLatest = true // 标记用户是否在查看最新数据
    let isFirstInit = true // 标记是否是第一次初始化

    const initChart = () => {
      if (!chartContainer.value) return

      if (chartInstance) {
        chartInstance.dispose()
      }

      chartInstance = echarts.init(chartContainer.value, 'dark')
      updateChart()
      
      // 监听dataZoom事件，更新当前可见区域的最大值和最小值
      chartInstance.on('dataZoom', (params) => {
        // 更新用户是否在查看最新数据的标记
        if (params && params.end !== undefined) {
          isUserViewingLatest = params.end >= 99.5
          if (isUserViewingLatest) {
            // 用户在查看最新数据，清除保存的状态
            savedDataZoom = null
          }
        }
        updateVisibleRangeMinMax()
      })
      
      // 初始化 markPoint 点击事件监听（只绑定一次）
      initMarkPointClickHandler()
    }
    
    // 初始化 markPoint 点击事件处理
    const initMarkPointClickHandler = () => {
      if (!chartInstance) return
      
      // 创建自定义tooltip元素（全局复用）
      let signalTooltip = null
      const createSignalTooltip = () => {
        if (signalTooltip) return signalTooltip
        
        signalTooltip = document.createElement('div')
        signalTooltip.style.cssText = `
          position: fixed;
          background: rgba(50, 50, 50, 0.95);
          border: 1px solid #4a5568;
          border-radius: 4px;
          padding: 12px;
          color: #ffffff;
          font-size: 12px;
          z-index: 10001;
          pointer-events: none;
          display: none;
          max-width: 300px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
        `
        document.body.appendChild(signalTooltip)
        return signalTooltip
      }
      
      // 格式化周期显示名称
      const getPeriodName = (period) => {
        const periodMap = {
          '1m': '1分钟',
          '5m': '5分钟',
          '15m': '15分钟',
          '30m': '30分钟',
          '1h': '1小时',
          '2h': '2小时',
          '3h': '3小时',
          '4h': '4小时',
          '1d': '日线',
          '1w': '周线',
          '1M': '月线',
        }
        return periodMap[period] || period
      }
      
      // 格式化时间函数
      const formatTimeForClick = (timestamp) => {
        if (!timestamp) return ''
        const time = typeof timestamp === 'string' ? parseInt(timestamp) : timestamp
        const msTimestamp = time < 1e12 ? time * 1000 : time
        const date = new Date(msTimestamp)
        if (isNaN(date.getTime())) return ''
        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        const hours = String(date.getHours()).padStart(2, '0')
        const minutes = String(date.getMinutes()).padStart(2, '0')
        const seconds = String(date.getSeconds()).padStart(2, '0')
        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
      }
      
      // 监听整个图表的点击事件，通过坐标判断是否点击了 markPoint
      chartInstance.getZr().on('click', (e) => {
        console.log('图表点击事件触发:', e.offsetX, e.offsetY)
        
        if (!props.alertSignals || props.alertSignals.length === 0) {
          console.log('没有信号数据')
          return
        }
        
        // 获取点击位置的像素坐标
        const clickPixel = [e.offsetX, e.offsetY]
        
        // 查找最近的信号点（使用像素距离）
        let closestSignal = null
        let minDistance = Infinity
        const maxPixelDistance = 20 // 最大允许的像素距离（20像素）
        
        props.alertSignals.forEach(signal => {
          try {
            // 将信号点的数据坐标转换为像素坐标
            const signalCoord = [signal.index, signal.price]
            const signalPixel = chartInstance.convertToPixel({ seriesIndex: 0 }, signalCoord)
            
            if (signalPixel && signalPixel.length >= 2) {
              // 计算像素距离
              const dx = signalPixel[0] - clickPixel[0]
              const dy = signalPixel[1] - clickPixel[1]
              const distance = Math.sqrt(dx * dx + dy * dy)
              
              if (distance < maxPixelDistance && distance < minDistance) {
                minDistance = distance
                closestSignal = signal
              }
            }
          } catch (err) {
            // 忽略转换错误
            console.warn('坐标转换失败:', err)
          }
        })
        
        if (closestSignal) {
          console.log('检测到信号点点击:', closestSignal, '距离:', minDistance.toFixed(2), '像素')
          
          const config = getSignalConfig(closestSignal.type)
          const tooltip = createSignalTooltip()
          
          // 创建tooltip内容
          let tooltipContent = `
            <div style="margin-bottom: 8px;">
              <strong style="font-size: 14px; color: ${config.color};">${config.icon} ${config.name}</strong>
            </div>
            <div style="margin-bottom: 6px; color: #a0aec0;">${config.description}</div>
            <div style="margin-bottom: 4px;"><strong>周期:</strong> ${getPeriodName(props.period)}</div>
            <div style="margin-bottom: 4px;"><strong>交易对:</strong> ${props.symbol}</div>
            <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #4a5568;">
              <div style="margin-bottom: 4px;"><strong>时间:</strong> ${formatTimeForClick(closestSignal.time)}</div>
              <div style="margin-bottom: 4px;"><strong>价格:</strong> ${closestSignal.price.toFixed(2)}</div>
              <div style="margin-bottom: 4px;"><strong>收盘价:</strong> ${closestSignal.close.toFixed(2)}</div>
          `
          
          if (closestSignal.lowerBand !== undefined) {
            tooltipContent += `<div style="margin-bottom: 4px;"><strong>下轨:</strong> ${closestSignal.lowerBand.toFixed(2)}</div>`
          }
          if (closestSignal.upperBand !== undefined) {
            tooltipContent += `<div style="margin-bottom: 4px;"><strong>上轨:</strong> ${closestSignal.upperBand.toFixed(2)}</div>`
          }
          if (closestSignal.strength !== undefined) {
            tooltipContent += `<div style="margin-bottom: 4px;"><strong>强度:</strong> ${(closestSignal.strength * 100).toFixed(1)}%</div>`
          }
          
          tooltipContent += `</div>`
          tooltip.innerHTML = tooltipContent
          tooltip.style.display = 'block'
          tooltip.style.zIndex = '10001'
          
          // 设置tooltip位置
          const chartRect = chartInstance.getDom().getBoundingClientRect()
          tooltip.style.left = `${chartRect.left + e.offsetX + 15}px`
          tooltip.style.top = `${chartRect.top + e.offsetY - 50}px`
          
          // 点击其他地方时隐藏tooltip
          const hideTooltip = (e) => {
            if (tooltip && !tooltip.contains(e.target)) {
              tooltip.style.display = 'none'
              document.removeEventListener('click', hideTooltip)
            }
          }
          setTimeout(() => {
            document.addEventListener('click', hideTooltip)
          }, 100)
        }
      })
    }
    
    // 计算并显示当前可见区域的最大值和最小值
    const updateVisibleRangeMinMax = () => {
      if (!chartInstance || !props.klineData || props.klineData.length === 0) {
        return
      }
      
      // 获取当前dataZoom的范围
      let option
      try {
        option = chartInstance.getOption()
      } catch (e) {
        console.warn('getOption() 失败:', e)
        return
      }
      
      // getOption() 可能返回 undefined、null、数组或对象，需要安全处理
      if (!option) {
        return
      }
      
      const optionObj = Array.isArray(option) ? option[0] : option
      if (!optionObj || !optionObj.dataZoom || !Array.isArray(optionObj.dataZoom) || optionObj.dataZoom.length === 0) {
        return
      }
      const dataZoom = optionObj.dataZoom[0]
      if (!dataZoom) return
      
      const start = dataZoom.start || 0
      const end = dataZoom.end || 100
      
      // 计算可见区域的数据索引范围
      const dataLength = props.klineData.length
      const startIndex = Math.floor((start / 100) * dataLength)
      const endIndex = Math.ceil((end / 100) * dataLength)
      
      // 获取可见区域的数据
      const visibleData = props.klineData.slice(startIndex, endIndex)
      if (visibleData.length === 0) return
      
      // 计算最大值和最小值
      let maxValue = -Infinity
      let minValue = Infinity
      let maxIndex = -1
      let minIndex = -1
      
      visibleData.forEach((item, idx) => {
        if (item.high > maxValue) {
          maxValue = item.high
          maxIndex = startIndex + idx
        }
        if (item.low < minValue) {
          minValue = item.low
          minIndex = startIndex + idx
        }
      })
      
      if (maxIndex === -1 || minIndex === -1) return
      
      // 获取最大值和最小值对应的时间
      const maxTime = props.klineData[maxIndex].time
      const minTime = props.klineData[minIndex].time
      
      // 更新markLine显示最大值和最小值
      // ECharts markLine格式：使用坐标数组 [x, y] 或 {coord: [x, y]}
      const markLineData = [
        [
          {
            name: '最大值',
            coord: [maxTime, maxValue],
            label: {
              formatter: `最大值: ${maxValue.toFixed(2)}`,
              position: 'end',
              color: '#10b981',
              fontSize: 12,
              fontWeight: 'bold',
            },
            lineStyle: {
              color: '#10b981',
              type: 'dashed',
              width: 2,
            },
          },
          {
            coord: [maxTime, maxValue],
          },
        ],
        [
          {
            name: '最小值',
            coord: [minTime, minValue],
            label: {
              formatter: `最小值: ${minValue.toFixed(2)}`,
              position: 'end',
              color: '#ef4444',
              fontSize: 12,
              fontWeight: 'bold',
            },
            lineStyle: {
              color: '#ef4444',
              type: 'dashed',
              width: 2,
            },
          },
          {
            coord: [minTime, minValue],
          },
        ],
      ]
      
      // 更新图表配置，添加markLine
      let currentOption
      try {
        currentOption = chartInstance.getOption()
      } catch (e) {
        console.warn('getOption() 失败:', e)
        return
      }
      // getOption() 可能返回 undefined、null、数组或对象，需要安全处理
      if (!currentOption) {
        return
      }
      const currentOptionObj = Array.isArray(currentOption) ? currentOption[0] : currentOption
      if (currentOptionObj && currentOptionObj.series && Array.isArray(currentOptionObj.series) && currentOptionObj.series[0]) {
        currentOptionObj.series[0].markLine = {
          data: markLineData,
          label: {
            show: true,
            position: 'end',
          },
          lineStyle: {
            type: 'dashed',
            width: 2,
          },
          symbol: ['none', 'none'], // 不显示起点和终点标记
        }
        chartInstance.setOption(currentOptionObj, { notMerge: false })
      }
    }

    const updateChart = () => {
      if (!chartInstance || !props.klineData || props.klineData.length === 0) {
        return
      }

      // 在更新数据前，保存当前的dataZoom状态
      let currentOption
      try {
        currentOption = chartInstance.getOption()
      } catch (e) {
        console.warn('getOption() 失败:', e)
        currentOption = null
      }
      // getOption() 可能返回 undefined、null、数组或对象，需要安全处理
      if (!currentOption) {
        // 如果获取失败，跳过保存 dataZoom 状态
        currentOption = null
      }
      const currentOptionObj = currentOption ? (Array.isArray(currentOption) ? currentOption[0] : currentOption) : null
      const currentDataZoom = currentOptionObj && currentOptionObj.dataZoom && Array.isArray(currentOptionObj.dataZoom) && currentOptionObj.dataZoom.length > 0
        ? currentOptionObj.dataZoom[0]
        : null
      
      // 检查用户是否在查看最新数据（end >= 99.5 表示接近末尾）
      if (currentDataZoom && currentDataZoom.end !== undefined && currentDataZoom.start !== undefined) {
        isUserViewingLatest = currentDataZoom.end >= 99.5
        if (!isUserViewingLatest) {
          // 用户不在查看最新数据，保存当前的dataZoom状态
          savedDataZoom = {
            start: currentDataZoom.start,
            end: currentDataZoom.end,
          }
          console.log('保存dataZoom状态:', savedDataZoom)
        } else {
          // 用户在查看最新数据，清除保存的状态
          savedDataZoom = null
          console.log('用户在查看最新数据，清除保存的状态')
        }
      } else if (!savedDataZoom) {
        // 如果没有当前状态且没有保存的状态，使用默认值
        savedDataZoom = null
      }

      const data = props.klineData
      // 确保时间戳是数字类型（数据库返回的是毫秒时间戳）
      const times = data.map(item => {
        const time = typeof item.time === 'string' ? parseInt(item.time) : item.time
        // 如果是秒时间戳（10位），转换为毫秒时间戳
        return time < 1e12 ? time * 1000 : time
      })
      const values = data.map(item => [item.open, item.close, item.low, item.high])
      const volumes = data.map(item => item.volume)

      // 准备布林带数据（确保数据长度与K线数据一致）
      const bbUpper = (props.indicators.bbUpper || props.indicators.BBUpper || []).slice(0, data.length)
      const bbMiddle = (props.indicators.bbMiddle || props.indicators.BBMiddle || []).slice(0, data.length)
      const bbLower = (props.indicators.bbLower || props.indicators.BBLower || []).slice(0, data.length)
      
      // 准备MACD数据（确保数据长度与K线数据一致，过滤null/undefined/NaN）
      const macdData = (props.indicators.macd || props.indicators.MACD || [])
        .slice(0, data.length)
        .map(v => (v !== null && v !== undefined && !isNaN(v)) ? v : null)
      const signalData = (props.indicators.signal || props.indicators.Signal || [])
        .slice(0, data.length)
        .map(v => (v !== null && v !== undefined && !isNaN(v)) ? v : null)
      const histData = (props.indicators.hist || props.indicators.Hist || [])
        .slice(0, data.length)
        .map(v => (v !== null && v !== undefined && !isNaN(v)) ? v : null)
      
      // 调试：检查MACD数据
      const macdValidCount = macdData.filter(v => v !== null && v !== undefined && !isNaN(v)).length
      if (macdValidCount > 0) {
        console.log(`MACD数据: 总数=${macdData.length}, 有效=${macdValidCount}, 示例值=`, macdData.slice(0, 5))
      } else {
        console.warn('MACD数据为空或无效，请检查指标计算')
      }

      // 准备预警信号标记点（根据信号类型显示不同图标和颜色）
      // 收集所有出现的信号类型，用于图例
      const signalTypesInChart = new Set()
      // 存储信号类型到信号的映射，用于图例点击显示详情
      const signalTypeMap = new Map()
      
      // 格式化周期显示名称
      const getPeriodName = (period) => {
        const periodMap = {
          '1m': '1分钟',
          '5m': '5分钟',
          '15m': '15分钟',
          '30m': '30分钟',
          '1h': '1小时',
          '2h': '2小时',
          '3h': '3小时',
          '4h': '4小时',
          '1d': '日线',
          '1w': '周线',
          '1M': '月线',
        }
        return periodMap[period] || period
      }
      
      // 格式化时间：年月日时分秒
      const formatTime = (timestamp) => {
        const time = typeof timestamp === 'string' ? parseInt(timestamp) : timestamp
        const msTimestamp = time < 1e12 ? time * 1000 : time
        const date = new Date(msTimestamp)
        if (isNaN(date.getTime())) return ''
        const year = date.getFullYear()
        const month = String(date.getMonth() + 1).padStart(2, '0')
        const day = String(date.getDate()).padStart(2, '0')
        const hours = String(date.getHours()).padStart(2, '0')
        const minutes = String(date.getMinutes()).padStart(2, '0')
        const seconds = String(date.getSeconds()).padStart(2, '0')
        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
      }
      
      const markPoints = props.alertSignals.map((signal, signalIndex) => {
        const config = getSignalConfig(signal.type)
        signalTypesInChart.add(signal.type) // 记录信号类型
        
        // 存储信号信息到映射中
        if (!signalTypeMap.has(signal.type)) {
          signalTypeMap.set(signal.type, [])
        }
        signalTypeMap.get(signal.type).push(signal)
        
        return {
          name: `${config.icon} ${config.name}`, // 在图例中显示图标和名称
          coord: [signal.index, signal.price],
          value: signal.price,
          symbol: 'diamond',
          symbolSize: 3 + (signal.strength || 0) * 1, // 缩小到1/10：原来30+10*strength，现在3+1*strength
          itemStyle: {
            color: config.color,
            borderColor: '#ffffff',
            borderWidth: 1, // 边框也缩小
          },
          label: {
            show: true,
            formatter: config.icon,
            fontSize: 10, // 字体缩小
            color: 'rgba(255, 255, 255, 0.8)', // 50%透明度
            fontWeight: 'bold',
            position: [-5, -5], // 偏移到左上5个像素
          },
          tooltip: {
            formatter: () => {
              // 如果是该信号类型的第一个信号，显示汇总信息
              const signals = signalTypeMap.get(signal.type) || []
              const isFirstSignal = signals.length > 0 && signals[0].index === signal.index && signals[0].price === signal.price
              
              if (isFirstSignal && signals.length > 1) {
                // 显示汇总信息
                let tooltip = `
                  <div style="margin-bottom: 8px;">
                    <strong style="font-size: 14px; color: ${config.color};">${config.icon} ${config.name}</strong>
                  </div>
                  <div style="margin-bottom: 6px; color: #a0aec0;">${config.description}</div>
                  <div style="margin-bottom: 4px;"><strong>周期:</strong> ${getPeriodName(props.period)}</div>
                  <div style="margin-bottom: 4px;"><strong>交易对:</strong> ${props.symbol}</div>
                  <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #4a5568;">
                    <div style="margin-bottom: 4px;"><strong>信号数量:</strong> ${signals.length}</div>
                `
                
                // 显示前3个信号的时间
                if (signals.length > 0) {
                  tooltip += `<div style="margin-top: 6px; font-size: 11px; color: #a0aec0;">最近信号时间:</div>`
                  signals.slice(0, 3).forEach(s => {
                    const timeStr = formatTime(s.time)
                    if (timeStr) {
                      tooltip += `<div style="font-size: 11px; color: #a0aec0; padding-left: 8px;">• ${timeStr}</div>`
                    }
                  })
                  if (signals.length > 3) {
                    tooltip += `<div style="font-size: 11px; color: #a0aec0; padding-left: 8px;">... 还有 ${signals.length - 3} 个信号</div>`
                  }
                }
                
                tooltip += `</div>`
                return tooltip
              } else {
                // 显示单个信号信息
                const timeStr = formatTime(signal.time)
                let tooltip = `${config.name}<br/>时间: ${timeStr}<br/>价格: ${signal.price.toFixed(2)}`
                if (signal.lowerBand) {
                  tooltip += `<br/>下轨: ${signal.lowerBand.toFixed(2)}`
                }
                if (signal.upperBand) {
                  tooltip += `<br/>上轨: ${signal.upperBand.toFixed(2)}`
                }
                if (signal.strength) {
                  tooltip += `<br/>强度: ${(signal.strength * 100).toFixed(2)}%`
                }
                tooltip += `<br/>${config.description}`
                return tooltip
              }
            },
          },
          // 存储信号索引，用于定位
          signalIndex: signalIndex,
        }
      })

      const option = {
        backgroundColor: 'transparent',
        animation: false,
        legend: [
          {
            // 主图图例（K线、均线、布林带、信号）- 显示在图表顶部居中
            show: true,
            top: 10,
            left: 'center',
            data: ['K线', 'MA144', 'MA10', 'MA20', 'BB上轨', 'BB中轨', 'BB下轨', '信号'],
            textStyle: {
              color: '#ffffff',
              fontSize: 12,
            },
            itemGap: 20,
            itemWidth: 25,
            itemHeight: 14,
            backgroundColor: 'rgba(0, 0, 0, 0.3)',
            borderColor: '#4a5568',
            borderWidth: 1,
            borderRadius: 4,
            padding: [8, 12],
            // 信号图例项不可点击（不控制显示/隐藏）
            selected: {
              '信号': true, // 信号始终显示，不可切换
            },
          },
          {
            // 副图图例（MACD、成交量）- 显示在图表顶部右侧
            show: true,
            top: 10,
            right: 20,
            data: ['MACD', 'Signal', 'Hist', 'Volume'],
            textStyle: {
              color: '#ffffff',
              fontSize: 12,
            },
            itemGap: 15,
            itemWidth: 25,
            itemHeight: 14,
            backgroundColor: 'rgba(0, 0, 0, 0.3)',
            borderColor: '#4a5568',
            borderWidth: 1,
            borderRadius: 4,
            padding: [8, 12],
            selected: {
              'Volume': false, // 默认隐藏Volume
            },
          },
          {
            // 信号图例（动态显示当前图表中出现的信号类型）
            show: signalTypesInChart.size > 0, // 只有存在信号时才显示
            top: 50,
            left: 'center',
            data: Array.from(signalTypesInChart).map(type => {
              const config = getSignalConfig(type)
              return `${config.icon} ${config.name}`
            }),
            textStyle: {
              color: '#ffffff',
              fontSize: 11,
            },
            itemGap: 12,
            itemWidth: 20,
            itemHeight: 12,
            backgroundColor: 'rgba(0, 0, 0, 0.3)',
            borderColor: '#4a5568',
            borderWidth: 1,
            borderRadius: 4,
            padding: [6, 10],
          },
        ],
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross',
            label: {
              formatter: (params) => {
                // 只格式化横轴（x轴）的时间戳，纵轴保留两位小数
                // axisDimension为'x'表示横轴，'y'表示纵轴
                if (params.axisDimension === 'x') {
                  const timestamp = typeof params.value === 'string' ? parseInt(params.value) : params.value
                  // 如果值看起来像时间戳（大于1000），才进行格式化
                  if (timestamp && timestamp > 1000) {
                    const msTimestamp = timestamp < 1e12 ? timestamp * 1000 : timestamp
                    const date = new Date(msTimestamp)
                    if (!isNaN(date.getTime())) {
                      const year = date.getFullYear()
                      const month = String(date.getMonth() + 1).padStart(2, '0')
                      const day = String(date.getDate()).padStart(2, '0')
                      const hours = String(date.getHours()).padStart(2, '0')
                      const minutes = String(date.getMinutes()).padStart(2, '0')
                      const seconds = String(date.getSeconds()).padStart(2, '0')
                      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
                    }
                  }
                }
                // 纵轴值保留两位小数（价格等数值）
                if (params.axisDimension === 'y') {
                  const value = typeof params.value === 'string' ? parseFloat(params.value) : params.value
                  if (value !== null && value !== undefined && !isNaN(value)) {
                    return Number(value).toFixed(2)
                  }
                }
                return params.value
              },
            },
          },
          backgroundColor: 'rgba(50, 50, 50, 0.9)',
          borderColor: '#777',
          borderWidth: 1,
          textStyle: {
            color: '#fff',
          },
          formatter: (params) => {
            if (!params || !Array.isArray(params) || params.length === 0) {
              return ''
            }
            
            // 格式化时间：年月日时分秒
            const formatTime = (timestamp) => {
              const time = typeof timestamp === 'string' ? parseInt(timestamp) : timestamp
              const msTimestamp = time < 1e12 ? time * 1000 : time
              const date = new Date(msTimestamp)
              if (isNaN(date.getTime())) return ''
              const year = date.getFullYear()
              const month = String(date.getMonth() + 1).padStart(2, '0')
              const day = String(date.getDate()).padStart(2, '0')
              const hours = String(date.getHours()).padStart(2, '0')
              const minutes = String(date.getMinutes()).padStart(2, '0')
              const seconds = String(date.getSeconds()).padStart(2, '0')
              return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
            }
            
            // 格式化数值：保留两位小数
            const formatValue = (value) => {
              if (value === null || value === undefined || isNaN(value)) return '-'
              return Number(value).toFixed(2)
            }
            
            // 获取时间（从第一个参数获取）
            const timeValue = params[0].axisValue
            const timeStr = formatTime(timeValue)
            
            // 构建tooltip内容
            let tooltip = `<div style="margin-bottom: 4px;"><strong>${timeStr}</strong></div>`
            
            params.forEach((item) => {
              if (item.seriesName === 'K线') {
                // K线数据：显示开高低收
                // ECharts candlestick数据格式: [open, close, low, high]
                const value = item.value
                if (Array.isArray(value) && value.length >= 4) {
                  tooltip += `<div>${item.marker} ${item.seriesName}</div>`
                  tooltip += `<div style="padding-left: 10px;">开: ${formatValue(value[0])}</div>`
                  tooltip += `<div style="padding-left: 10px;">收: ${formatValue(value[1])}</div>`
                  tooltip += `<div style="padding-left: 10px;">低: ${formatValue(value[2])}</div>`
                  tooltip += `<div style="padding-left: 10px;">高: ${formatValue(value[3])}</div>`
                }
              } else if (item.seriesName === 'Volume') {
                // 成交量：显示原始值（可能很大，不强制两位小数）
                tooltip += `<div>${item.marker} ${item.seriesName}: ${formatValue(item.value)}</div>`
              } else {
                // 其他指标：保留两位小数
                tooltip += `<div>${item.marker} ${item.seriesName}: ${formatValue(item.value)}</div>`
              }
            })
            
            return tooltip
          },
        },
        grid: [
          {
            left: '10%',
            right: '8%',
            top: '15%',
            height: '50%',
          },
          {
            left: '10%',
            right: '8%',
            top: '70%',
            height: '16%',
          },
        ],
        xAxis: [
          {
            type: 'category',
            data: times,
            scale: true,
            boundaryGap: false,
            axisLine: { onZero: false },
            splitLine: { show: false },
            min: 'dataMin',
            max: 'dataMax',
            axisLabel: {
              formatter: (value) => {
                // 确保时间戳是数字类型
                const timestamp = typeof value === 'string' ? parseInt(value) : value
                // 如果是秒时间戳（10位），转换为毫秒时间戳
                const msTimestamp = timestamp < 1e12 ? timestamp * 1000 : timestamp
                const date = new Date(msTimestamp)
                // 检查日期是否有效
                if (isNaN(date.getTime())) {
                  return ''
                }
                return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${date.getMinutes()}`
              },
              color: '#9ca3af',
            },
          },
          {
            type: 'category',
            gridIndex: 1,
            data: times,
            scale: true,
            boundaryGap: false,
            axisLine: { onZero: false },
            axisTick: { show: false },
            splitLine: { show: false },
            axisLabel: { show: false },
            min: 'dataMin',
            max: 'dataMax',
          },
        ],
        yAxis: [
          {
            scale: true,
            splitArea: {
              show: true,
              areaStyle: {
                color: ['rgba(250,250,250,0.05)', 'rgba(200,200,200,0.02)'],
              },
            },
            axisLabel: {
              formatter: (value) => value.toFixed(2),
              color: '#9ca3af',
            },
            splitLine: {
              show: true,
              lineStyle: {
                color: '#2d3748',
              },
            },
          },
          {
            scale: true,
            gridIndex: 1,
            splitNumber: 4,
            axisLabel: {
              show: true,
              formatter: (value) => value.toFixed(4),
              color: '#9ca3af',
              fontSize: 10,
            },
            axisLine: { show: false },
            axisTick: { show: false },
            splitLine: {
              show: true,
              lineStyle: {
                color: '#2d3748',
                type: 'dashed',
              },
            },
          },
        ],
        // dataZoom 配置：始终提供，避免 ECharts 内部访问 undefined
        dataZoom: isFirstInit ? [
          {
            type: 'inside',
            xAxisIndex: [0, 1],
            start: 80,
            end: 100,
          },
          {
            show: true,
            xAxisIndex: [0, 1],
            type: 'slider',
            top: '90%',
            start: 80,
            end: 100,
            textStyle: {
              color: '#9ca3af',
            },
            borderColor: '#4a5568',
            fillerColor: 'rgba(99, 179, 237, 0.2)',
            handleStyle: {
              color: '#63b3ed',
            },
          },
        ] : (() => {
          // 非首次初始化时，尝试从当前配置获取 dataZoom，如果不存在则使用默认值
          try {
            const currentOption = chartInstance.getOption()
            if (currentOption) {
              const currentOptionObj = Array.isArray(currentOption) ? currentOption[0] : currentOption
              if (currentOptionObj && currentOptionObj.dataZoom && Array.isArray(currentOptionObj.dataZoom) && currentOptionObj.dataZoom.length > 0) {
                // 返回当前的 dataZoom 配置，保持用户的位置
                return currentOptionObj.dataZoom
              }
            }
          } catch (e) {
            console.warn('获取当前 dataZoom 失败:', e)
          }
          // 如果获取失败或不存在，返回默认配置（确保始终返回数组）
          return [
            {
              type: 'inside',
              xAxisIndex: [0, 1],
              start: 80,
              end: 100,
            },
            {
              show: true,
              xAxisIndex: [0, 1],
              type: 'slider',
              top: '90%',
              start: 80,
              end: 100,
              textStyle: {
                color: '#9ca3af',
              },
              borderColor: '#4a5568',
              fillerColor: 'rgba(99, 179, 237, 0.2)',
              handleStyle: {
                color: '#63b3ed',
              },
            },
          ]
        })(),
        series: [
          {
            name: 'K线',
            type: 'candlestick',
            data: values,
            itemStyle: {
              color: '#26a69a',
              color0: '#ef5350',
              borderColor: '#26a69a',
              borderColor0: '#ef5350',
            },
            markPoint: {
              data: markPoints,
              animation: false,
            },
          },
          {
            name: 'MA144',
            type: 'line',
            data: (props.indicators.ma144 || props.indicators.MA144 || []).slice(0, data.length),
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#fbbf24',
            },
            symbol: 'none',
          },
          // {
          //   name: 'MA10',
          //   type: 'line',
          //   data: (props.indicators.ma10 || props.indicators.MA10 || []).slice(0, data.length),
          //   smooth: true,
          //   lineStyle: {
          //     width: 1,
          //     color: '#3b82f6',
          //   },
          //   symbol: 'none',
          // },
          // {
          //   name: 'MA20',
          //   type: 'line',
          //   data: (props.indicators.ma20 || props.indicators.MA20 || []).slice(0, data.length),
          //   smooth: true,
          //   lineStyle: {
          //     width: 1,
          //     color: '#8b5cf6',
          //   },
          //   symbol: 'none',
          // },
          {
            name: 'BB上轨',
            type: 'line',
            data: bbUpper.map(v => v || null),
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#f97316', // 橙色
              type: 'solid', // 实线
            },
            symbol: 'none',
            itemStyle: {
              opacity: 1,
            },
          },
          {
            name: 'BB中轨',
            type: 'line',
            data: bbMiddle.map(v => v || null),
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#3b82f6', // 蓝色
              type: 'solid', // 实线
            },
            symbol: 'none',
            itemStyle: {
              opacity: 1,
            },
          },
          {
            name: 'BB下轨',
            type: 'line',
            data: bbLower.map(v => v || null),
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#a855f7', // 紫色
              type: 'solid', // 实线
            },
            symbol: 'none',
            itemStyle: {
              opacity: 1,
            },
          },
          {
            name: 'MACD',
            type: 'line',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: macdData,
            lineStyle: {
              width: 1,
              color: '#3b82f6',
            },
            symbol: 'none',
            z: 10, // 提高层级，确保显示在上层
          },
          {
            name: 'Signal',
            type: 'line',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: signalData,
            lineStyle: {
              width: 1,
              color: '#f59e0b',
            },
            symbol: 'none',
            z: 10,
          },
          {
            name: 'Hist',
            type: 'bar',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: histData,
            itemStyle: {
              color: (params) => {
                return params.value >= 0 ? '#26a69a' : '#ef5350'
              },
            },
            z: 5,
          },
          {
            name: 'Volume',
            type: 'bar',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: volumes,
            itemStyle: {
              color: (params) => {
                const dataIndex = params.dataIndex
                if (dataIndex === 0) return '#9ca3af'
                const current = data[dataIndex]
                const prev = data[dataIndex - 1]
                return current.close >= prev.close ? '#26a69a' : '#ef5350'
              },
            },
            silent: true, // 默认隐藏，可通过图例切换显示
            z: 1,
          },
        ],
      }

      // 设置配置：使用 notMerge: false 来合并配置，保留用户的 dataZoom 状态
      const setOptionConfig = { 
        notMerge: false, // 合并模式，会保留现有的 dataZoom 配置
        lazyUpdate: false 
      }
      
      // 标记已完成首次初始化
      if (isFirstInit) {
        isFirstInit = false
      }
      
      chartInstance.setOption(option, setOptionConfig)
      
      // 如果用户不在查看最新数据，强制恢复保存的视图位置
      if (savedDataZoom && !isUserViewingLatest) {
        nextTick(() => {
          // 使用 dispatchAction 恢复视图位置
          chartInstance.dispatchAction({
            type: 'dataZoom',
            dataZoomIndex: 0,
            start: savedDataZoom.start,
            end: savedDataZoom.end,
          })
          chartInstance.dispatchAction({
            type: 'dataZoom',
            dataZoomIndex: 1,
            start: savedDataZoom.start,
            end: savedDataZoom.end,
          })
        })
      }
      
      // 更新可见区域的最大值和最小值
      nextTick(() => {
        updateVisibleRangeMinMax()
      })
      
      // 添加信号图例点击事件，显示信号详情tooltip
      if (signalTypesInChart.size > 0) {
        // 创建自定义tooltip元素
        let signalTooltip = null
        const createSignalTooltip = () => {
          if (signalTooltip) return signalTooltip
          
          signalTooltip = document.createElement('div')
          signalTooltip.style.cssText = `
            position: fixed;
            background: rgba(50, 50, 50, 0.95);
            border: 1px solid #4a5568;
            border-radius: 4px;
            padding: 12px;
            color: #ffffff;
            font-size: 12px;
            z-index: 10001;
            pointer-events: none;
            display: none;
            max-width: 300px;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
          `
          document.body.appendChild(signalTooltip)
          return signalTooltip
        }
        
        // 格式化周期显示名称
        const getPeriodName = (period) => {
          const periodMap = {
            '1m': '1分钟',
            '5m': '5分钟',
            '15m': '15分钟',
            '30m': '30分钟',
            '1h': '1小时',
            '2h': '2小时',
            '3h': '3小时',
            '4h': '4小时',
            '1d': '日线',
            '1w': '周线',
            '1M': '月线',
          }
          return periodMap[period] || period
        }
        
        // 临时隐藏ECharts的tooltip
        const hideEChartsTooltip = () => {
          const tooltipEl = chartInstance.getDom().querySelector('.echarts-tooltip')
          if (tooltipEl) {
            tooltipEl.style.display = 'none'
          }
          // 也尝试通过dispatchAction隐藏
          chartInstance.dispatchAction({
            type: 'hideTip'
          })
        }
        
        // 格式化时间函数（用于点击事件）
        const formatTimeForClick = (timestamp) => {
          if (!timestamp) return ''
          const time = typeof timestamp === 'string' ? parseInt(timestamp) : timestamp
          const msTimestamp = time < 1e12 ? time * 1000 : time
          const date = new Date(msTimestamp)
          if (isNaN(date.getTime())) return ''
          const year = date.getFullYear()
          const month = String(date.getMonth() + 1).padStart(2, '0')
          const day = String(date.getDate()).padStart(2, '0')
          const hours = String(date.getHours()).padStart(2, '0')
          const minutes = String(date.getMinutes()).padStart(2, '0')
          const seconds = String(date.getSeconds()).padStart(2, '0')
          return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
        }
        
        // 注意：ECharts 的 markPoint 点击事件可能不支持，已移到 initMarkPointClickHandler 中处理
        
        // 监听图例点击事件（使用legendclick而不是legendselectchanged）
        chartInstance.on('legendclick', (params) => {
          // 检查是否点击了信号图例项
          const clickedName = params.name
          if (!clickedName) return
          
          // 查找对应的信号类型
          let clickedSignalType = null
          Array.from(signalTypesInChart).forEach(type => {
            const config = getSignalConfig(type)
            const legendName = `${config.icon} ${config.name}`
            if (legendName === clickedName) {
              clickedSignalType = type
            }
          })
          
          // 如果是信号图例项，阻止默认行为并显示markPoint的tooltip
          if (clickedSignalType && signalTypeMap.has(clickedSignalType)) {
            // 阻止图例切换行为
            params.event && params.event.event && params.event.event.stopPropagation()
            
            // 隐藏ECharts的tooltip
            hideEChartsTooltip()
            
            // 找到该信号类型对应的第一个markPoint
            const signals = signalTypeMap.get(clickedSignalType)
            if (signals && signals.length > 0) {
              const firstSignal = signals[0]
              const config = getSignalConfig(clickedSignalType)
              
              // 找到该markPoint在markPoints数组中的索引
              const markPointIndex = props.alertSignals.findIndex(s => 
                s.type === clickedSignalType && 
                s.index === firstSignal.index && 
                s.price === firstSignal.price
              )
              
              if (markPointIndex >= 0) {
                // 获取markPoint的数据坐标
                const markPointCoord = markPoints[markPointIndex].coord
                
                // 将数据坐标转换为像素坐标
                try {
                  const pixelCoord = chartInstance.convertToPixel({ seriesIndex: 0 }, markPointCoord)
                  
                  // 创建并显示自定义tooltip
                  const tooltip = createSignalTooltip()
                  
                  // 创建tooltip内容
                  let tooltipContent = `
                    <div style="margin-bottom: 8px;">
                      <strong style="font-size: 14px; color: ${config.color};">${config.icon} ${config.name}</strong>
                    </div>
                    <div style="margin-bottom: 6px; color: #a0aec0;">${config.description}</div>
                    <div style="margin-bottom: 4px;"><strong>周期:</strong> ${getPeriodName(props.period)}</div>
                    <div style="margin-bottom: 4px;"><strong>交易对:</strong> ${props.symbol}</div>
                    <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #4a5568;">
                      <div style="margin-bottom: 4px;"><strong>信号数量:</strong> ${signals.length}</div>
                  `
                  
                  // 显示前3个信号的时间
                  if (signals.length > 0) {
                    tooltipContent += `<div style="margin-top: 6px; font-size: 11px; color: #a0aec0;">最近信号时间:</div>`
                    signals.slice(0, 3).forEach(signal => {
                      const timeStr = formatTime(signal.time)
                      if (timeStr) {
                        tooltipContent += `<div style="font-size: 11px; color: #a0aec0; padding-left: 8px;">• ${timeStr}</div>`
                      }
                    })
                    if (signals.length > 3) {
                      tooltipContent += `<div style="font-size: 11px; color: #a0aec0; padding-left: 8px;">... 还有 ${signals.length - 3} 个信号</div>`
                    }
                  }
                  
                  tooltipContent += `</div>`
                  tooltip.innerHTML = tooltipContent
                  tooltip.style.display = 'block'
                  tooltip.style.zIndex = '10001'
                  
                  // 设置tooltip位置
                  if (pixelCoord && pixelCoord.length >= 2) {
                    const chartRect = chartInstance.getDom().getBoundingClientRect()
                    tooltip.style.left = `${chartRect.left + pixelCoord[0] + 15}px`
                    tooltip.style.top = `${chartRect.top + pixelCoord[1] - 10}px`
                  } else if (params.event && params.event.event) {
                    const event = params.event.event
                    tooltip.style.left = `${event.clientX + 15}px`
                    tooltip.style.top = `${event.clientY + 15}px`
                  }
                  
                  // 3秒后自动隐藏
                  setTimeout(() => {
                    if (tooltip) {
                      tooltip.style.display = 'none'
                    }
                  }, 3000)
                  
                  // 点击其他地方时隐藏
                  const hideTooltip = (e) => {
                    if (tooltip && !tooltip.contains(e.target)) {
                      tooltip.style.display = 'none'
                      document.removeEventListener('click', hideTooltip)
                    }
                  }
                  setTimeout(() => {
                    document.addEventListener('click', hideTooltip)
                  }, 100)
                } catch (e) {
                  console.warn('无法转换markPoint坐标:', e)
                }
              }
            }
            
            return false // 阻止默认行为
          }
        })
        
        // 同时监听legendselectchanged，确保信号图例保持选中状态
        chartInstance.on('legendselectchanged', (params) => {
          const clickedName = params.name
          if (!clickedName) return
          
          // 恢复图例选中状态（信号图例不可切换）
          const legendData = Array.from(signalTypesInChart).map(type => {
            const config = getSignalConfig(type)
            return `${config.icon} ${config.name}`
          })
          if (legendData.includes(clickedName)) {
            // 如果是信号图例项，保持选中状态
            chartInstance.dispatchAction({
              type: 'legendSelect',
              name: clickedName,
              selected: true,
            })
          }
        })
        
        // 监听鼠标移动，更新tooltip位置（如果tooltip正在显示）
        chartInstance.getZr().on('mousemove', (e) => {
          if (signalTooltip && signalTooltip.style.display === 'block') {
            signalTooltip.style.left = `${e.offsetX + 15}px`
            signalTooltip.style.top = `${e.offsetY + 15}px`
          }
        })
      }
    }

    onMounted(() => {
      nextTick(() => {
        initChart()
        window.addEventListener('resize', () => {
          if (chartInstance) {
            chartInstance.resize()
          }
        })
      })
    })

    onUnmounted(() => {
      if (chartInstance) {
        chartInstance.dispose()
        chartInstance = null
      }
      window.removeEventListener('resize', () => {})
    })

    watch(
      () => [props.klineData, props.indicators, props.alertSignals],
      () => {
        updateChart()
      },
      { deep: true }
    )

    return {
      chartContainer,
    }
  },
}
</script>

<style scoped>
.kline-chart-container {
  width: 100%;
  height: 100%;
  background: #1b2636;
}

.chart-container {
  width: 100%;
  height: 100%;
}
</style>
