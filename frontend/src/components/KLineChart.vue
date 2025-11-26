<template>
  <div class="kline-chart-container">
    <div ref="chartContainer" class="chart-container"></div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'

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
    symbol: {
      type: String,
      default: 'BTCUSDT',
    },
  },
  setup(props) {
    const chartContainer = ref(null)
    let chartInstance = null

    const initChart = () => {
      if (!chartContainer.value) return

      if (chartInstance) {
        chartInstance.dispose()
      }

      chartInstance = echarts.init(chartContainer.value, 'dark')
      updateChart()
    }

    const updateChart = () => {
      if (!chartInstance || !props.klineData || props.klineData.length === 0) {
        return
      }

      const data = props.klineData
      const times = data.map(item => item.time)
      const values = data.map(item => [item.open, item.close, item.low, item.high])
      const volumes = data.map(item => item.volume)

      const option = {
        backgroundColor: 'transparent',
        animation: false,
        legend: {
          top: 10,
          left: 'center',
          data: ['K线', 'MA5', 'MA10', 'MA20', 'MACD', 'Signal', 'Hist'],
          textStyle: {
            color: '#ffffff',
          },
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross',
          },
          backgroundColor: 'rgba(50, 50, 50, 0.9)',
          borderColor: '#777',
          borderWidth: 1,
          textStyle: {
            color: '#fff',
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
                const date = new Date(value)
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
            splitNumber: 2,
            axisLabel: { show: false },
            axisLine: { show: false },
            axisTick: { show: false },
            splitLine: { show: false },
          },
        ],
        dataZoom: [
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
        ],
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
              label: {
                formatter: (param) => {
                  return param != null ? Math.round(param.value) + '' : ''
                },
              },
              data: [
                {
                  name: 'Mark',
                  coord: ['2013/5/31', 2300],
                  value: 2300,
                  itemStyle: {
                    color: 'rgb(41,60,85)',
                  },
                },
              ],
              tooltip: {
                formatter: (param) => {
                  return param.name + '<br>' + (param.data.coord || '')
                },
              },
            },
          },
          {
            name: 'MA5',
            type: 'line',
            data: props.indicators.MA5 || [],
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#fbbf24',
            },
            symbol: 'none',
          },
          {
            name: 'MA10',
            type: 'line',
            data: props.indicators.MA10 || [],
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#3b82f6',
            },
            symbol: 'none',
          },
          {
            name: 'MA20',
            type: 'line',
            data: props.indicators.MA20 || [],
            smooth: true,
            lineStyle: {
              width: 1,
              color: '#8b5cf6',
            },
            symbol: 'none',
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
          },
          {
            name: 'MACD',
            type: 'line',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: props.indicators.MACD || [],
            lineStyle: {
              width: 1,
              color: '#3b82f6',
            },
            symbol: 'none',
          },
          {
            name: 'Signal',
            type: 'line',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: props.indicators.Signal || [],
            lineStyle: {
              width: 1,
              color: '#f59e0b',
            },
            symbol: 'none',
          },
          {
            name: 'Hist',
            type: 'bar',
            xAxisIndex: 1,
            yAxisIndex: 1,
            data: props.indicators.Hist || [],
            itemStyle: {
              color: (params) => {
                return params.value >= 0 ? '#26a69a' : '#ef5350'
              },
            },
          },
        ],
      }

      chartInstance.setOption(option, true)
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
      () => [props.klineData, props.indicators],
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

