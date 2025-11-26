# 数据库使用示例

## 完整使用流程

### 1. 初始化数据库

```javascript
// 在 Vue 组件中
async function initDatabase() {
  const dsn = 'root:password@tcp(localhost:3306)/contract_warn?charset=utf8mb4&parseTime=True&loc=Local'
  try {
    const result = await window.go.main.App.InitDatabase(dsn)
    console.log(result) // "数据库初始化成功"
  } catch (error) {
    console.error('数据库初始化失败:', error)
  }
}
```

### 2. 首次同步历史数据

```javascript
async function syncInitialData() {
  try {
    // 同步BTC/USDT最近7天的数据
    const result = await window.go.main.App.SyncKlineDataInitial('BTCUSDT', 7)
    console.log(result) // "初始同步成功"
  } catch (error) {
    console.error('同步失败:', error)
  }
}
```

### 3. 启动自动同步

```javascript
async function startAutoSync() {
  try {
    // 启动自动同步，每60秒同步一次
    const result = await window.go.main.App.StartAutoSync('BTCUSDT', 60)
    console.log(result) // "自动同步已启动"
  } catch (error) {
    console.error('启动自动同步失败:', error)
  }
}
```

### 4. 获取多周期K线数据

```javascript
async function getKlineData() {
  try {
    // 获取30分钟K线（系统会自动从1分钟数据聚合）
    const dataStr = await window.go.main.App.GetMarketData('BTCUSDT', '30m')
    const klines = JSON.parse(dataStr)
    console.log('30分钟K线数据:', klines)
  } catch (error) {
    console.error('获取数据失败:', error)
  }
}
```

### 5. 在 App.vue 中集成

```vue
<template>
  <div>
    <button @click="initDB">初始化数据库</button>
    <button @click="syncData">同步数据</button>
    <button @click="startSync">启动自动同步</button>
  </div>
</template>

<script>
import { ref } from 'vue'

export default {
  setup() {
    const dbInitialized = ref(false)

    const initDB = async () => {
      const dsn = 'root:password@tcp(localhost:3306)/contract_warn?charset=utf8mb4&parseTime=True&loc=Local'
      try {
        await window.go.main.App.InitDatabase(dsn)
        dbInitialized.value = true
        alert('数据库初始化成功')
      } catch (error) {
        alert('数据库初始化失败: ' + error)
      }
    }

    const syncData = async () => {
      try {
        await window.go.main.App.SyncKlineDataInitial('BTCUSDT', 7)
        alert('数据同步成功')
      } catch (error) {
        alert('数据同步失败: ' + error)
      }
    }

    const startSync = async () => {
      try {
        await window.go.main.App.StartAutoSync('BTCUSDT', 60)
        alert('自动同步已启动')
      } catch (error) {
        alert('启动自动同步失败: ' + error)
      }
    }

    return {
      initDB,
      syncData,
      startSync,
    }
  },
}
</script>
```

## 支持的周期

系统支持以下周期（通过聚合1分钟数据生成）：

- `1m` - 1分钟
- `5m` - 5分钟
- `15m` - 15分钟
- `30m` - 30分钟
- `1h` - 1小时
- `4h` - 4小时
- `1d` - 1天

## 性能说明

- **聚合速度**: 10,000根1分钟K线聚合为2,000根5分钟K线，耗时 < 5ms
- **存储空间**: 5年BTC数据约400MB
- **API配额**: 增量同步每次只拉取新数据，配额消耗极低

## 注意事项

1. **首次同步**: 建议至少同步7天历史数据，确保有足够数据计算技术指标
2. **自动同步**: 建议每60秒同步一次，避免过于频繁触发API限制
3. **错误处理**: 网络错误时，系统会使用本地缓存数据，确保应用可用
4. **数据库连接**: 确保MySQL服务运行，且用户有足够权限

