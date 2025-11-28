// 稳定可用的交易所接口配置
const EXCHANGE_APIS = [
  {
    name: 'CoinGecko',
    type: 'aggregator',
    baseUrl: 'https://api.coingecko.com/api/v3',
    priority: 1, // 最高优先级
    timeout: 15000, // 增加超时时间
    enabled: true
  },
  {
    name: 'OKX',
    type: 'exchange',
    baseUrl: 'https://www.okx.com/api/v5',
    priority: 2,
    timeout: 15000,
    enabled: true
  },
  {
    name: 'Kraken',
    type: 'exchange',
    baseUrl: 'https://api.kraken.com/0/public',
    priority: 3,
    timeout: 15000,
    enabled: true
  },
  {
    name: 'Gate.io',
    type: 'exchange',
    baseUrl: 'https://api.gateio.ws/api/v4',
    priority: 4,
    timeout: 15000,
    enabled: true
  },
  {
    name: 'MEXC',
    type: 'exchange',
    baseUrl: 'https://api.mexc.com/api/v3',
    priority: 5,
    timeout: 15000,
    enabled: true
  },
  {
    name: 'Bitget',
    type: 'exchange',
    baseUrl: 'https://api.bitget.com/api/mix/v1',
    priority: 6,
    timeout: 15000,
    enabled: true
  },
  {
    name: 'Binance',
    type: 'exchange',
    baseUrl: 'https://api.binance.com/api/v3',
    priority: 7,
    timeout: 15000,
    enabled: true
  },
  {
    name: 'Bybit',
    type: 'exchange',
    baseUrl: 'https://api.bybit.com/v5',
    priority: 8,
    timeout: 15000,
    enabled: true
  }
];

// 币种映射表（支持更多币种）
const SYMBOL_MAP = {
  'bitcoin': { 
    coingecko: 'bitcoin', 
    exchange: 'BTCUSDT',
    kraken: 'XBTUSDT'
  },
  'ethereum': { 
    coingecko: 'ethereum', 
    exchange: 'ETHUSDT',
    kraken: 'ETHUSDT'
  },
  'btc': { 
    coingecko: 'bitcoin', 
    exchange: 'BTCUSDT',
    kraken: 'XBTUSDT'
  },
  'eth': { 
    coingecko: 'ethereum', 
    exchange: 'ETHUSDT',
    kraken: 'ETHUSDT'
  },
  'solana': {
    coingecko: 'solana',
    exchange: 'SOLUSDT',
    kraken: 'SOLUSDT'
  },
  'sol': {
    coingecko: 'solana',
    exchange: 'SOLUSDT',
    kraken: 'SOLUSDT'
  }
};

// 带超时的 fetch 封装（增强版，支持更详细的错误信息）
async function fetchWithTimeout(url, timeout = 10000, options = {}) {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  
  const startTime = Date.now();
  
  try {
    const fetchOptions = {
      signal: controller.signal,
      headers: {
        'Accept': 'application/json',
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
      },
      ...options
    };
    
    // 如果是 Node.js 环境，可能需要配置代理
    // 可以通过环境变量 HTTP_PROXY 或 HTTPS_PROXY 设置
    
    const response = await fetch(url, fetchOptions);
    clearTimeout(timeoutId);
    
    const responseTime = Date.now() - startTime;
    
    if (!response.ok) {
      const errorText = await response.text().catch(() => '');
      throw new Error(`HTTP ${response.status}: ${errorText.substring(0, 100)}`);
    }
    
    return await response.json();
  } catch (error) {
    clearTimeout(timeoutId);
    const responseTime = Date.now() - startTime;
    
    // 提供更详细的错误信息
    if (error.name === 'AbortError') {
      throw new Error(`请求超时 (${responseTime}ms)`);
    } else if (error.message.includes('Failed to fetch') || error.message.includes('fetch failed')) {
      // 这可能是网络连接问题、DNS 解析问题或 CORS 问题
      throw new Error(`网络连接失败: ${error.message} (可能是 DNS 解析失败、防火墙阻止或 CORS 限制)`);
    } else if (error.message.includes('NetworkError') || error.message.includes('Network request failed')) {
      throw new Error(`网络错误: ${error.message} (请检查网络连接和代理设置)`);
    } else if (error.message.includes('CORS')) {
      throw new Error(`CORS 错误: ${error.message} (浏览器跨域限制)`);
    }
    
    throw error;
  }
}

// 从 CoinGecko 获取数据
async function fetchFromCoinGecko(symbol) {
  const coinId = SYMBOL_MAP[symbol.toLowerCase()]?.coingecko || symbol.toLowerCase();
  const api = EXCHANGE_APIS.find(e => e.name === 'CoinGecko');
  const url = `${api.baseUrl}/simple/price?ids=${coinId}&vs_currencies=usd`;
  
  try {
    const data = await fetchWithTimeout(url, api.timeout);
    
    if (!data[coinId] || !data[coinId].usd) {
      throw new Error('未找到价格数据');
    }
    
    return {
      price: data[coinId].usd,
      source: 'CoinGecko',
      raw: data
    };
  } catch (error) {
    throw new Error(`CoinGecko API 错误: ${error.message}`);
  }
}

// 从 OKX 获取数据
async function fetchFromOKX(symbol) {
  // OKX 使用 BTC-USDT 格式（带连字符）
  const exchangePair = SYMBOL_MAP[symbol.toLowerCase()]?.exchange || 'BTCUSDT';
  const tradingPair = exchangePair.replace('USDT', '-USDT');
  const api = EXCHANGE_APIS.find(e => e.name === 'OKX');
  const url = `${api.baseUrl}/market/ticker?instId=${tradingPair}`;
  
  try {
    const data = await fetchWithTimeout(url, api.timeout);
    
    if (data.code !== '0' || !data.data || !data.data[0]) {
      throw new Error(data.msg || '未找到价格数据');
    }
    
    return {
      price: parseFloat(data.data[0].last),
      source: 'OKX',
      raw: data.data[0]
    };
  } catch (error) {
    throw new Error(`OKX API 错误: ${error.message}`);
  }
}

// 从 Kraken 获取数据
async function fetchFromKraken(symbol) {
  // Kraken 使用 XBT 表示 BTC，格式为 XBTUSDT
  const tradingPair = SYMBOL_MAP[symbol.toLowerCase()]?.kraken || 'XBTUSDT';
  const api = EXCHANGE_APIS.find(e => e.name === 'Kraken');
  const url = `${api.baseUrl}/Ticker?pair=${tradingPair}`;
  
  const data = await fetchWithTimeout(url, api.timeout);
  
  if (data.error && data.error.length > 0) {
    throw new Error(data.error.join(', '));
  }
  
  const tickerKey = Object.keys(data.result)[0];
  if (!tickerKey || !data.result[tickerKey] || !data.result[tickerKey].c) {
    throw new Error('未找到价格数据');
  }
  
  return {
    price: parseFloat(data.result[tickerKey].c[0]),
    source: 'Kraken',
    raw: data.result[tickerKey]
  };
}

// 从 Gate.io 获取数据
async function fetchFromGateIO(symbol) {
  // Gate.io 使用 BTC_USDT 格式（下划线）
  const exchangePair = SYMBOL_MAP[symbol.toLowerCase()]?.exchange || 'BTCUSDT';
  const tradingPair = exchangePair.replace('USDT', '_USDT');
  const api = EXCHANGE_APIS.find(e => e.name === 'Gate.io');
  const url = `${api.baseUrl}/spot/tickers?currency_pair=${tradingPair}`;
  
  const data = await fetchWithTimeout(url, api.timeout);
  
  if (!data || !Array.isArray(data) || data.length === 0 || !data[0].last) {
    throw new Error('未找到价格数据');
  }
  
  return {
    price: parseFloat(data[0].last),
    source: 'Gate.io',
    raw: data[0]
  };
}

// 从 MEXC 获取数据
async function fetchFromMEXC(symbol) {
  const tradingPair = SYMBOL_MAP[symbol.toLowerCase()]?.exchange || 'BTCUSDT';
  const api = EXCHANGE_APIS.find(e => e.name === 'MEXC');
  const url = `${api.baseUrl}/ticker/price?symbol=${tradingPair}`;
  
  const data = await fetchWithTimeout(url, api.timeout);
  
  if (!data.price) {
    throw new Error('未找到价格数据');
  }
  
  return {
    price: parseFloat(data.price),
    source: 'MEXC',
    raw: data
  };
}

// 从 Bitget 获取数据
async function fetchFromBitget(symbol) {
  const tradingPair = SYMBOL_MAP[symbol.toLowerCase()]?.exchange || 'BTCUSDT';
  // 使用现货 API
  const api = EXCHANGE_APIS.find(e => e.name === 'Bitget');
  const url = `https://api.bitget.com/api/spot/v1/market/ticker?symbol=${tradingPair}`;
  
  const data = await fetchWithTimeout(url, api.timeout);
  
  if (data.code !== '00000' || !data.data || !data.data.close) {
    throw new Error(data.msg || '未找到价格数据');
  }
  
  return {
    price: parseFloat(data.data.close),
    source: 'Bitget',
    raw: data.data
  };
}

// 从 Binance 获取数据
async function fetchFromBinance(symbol) {
  const tradingPair = SYMBOL_MAP[symbol.toLowerCase()]?.exchange || 'BTCUSDT';
  const api = EXCHANGE_APIS.find(e => e.name === 'Binance');
  const url = `${api.baseUrl}/ticker/price?symbol=${tradingPair}`;
  
  const data = await fetchWithTimeout(url, api.timeout);
  
  if (!data.price) {
    throw new Error('未找到价格数据');
  }
  
  return {
    price: parseFloat(data.price),
    source: 'Binance',
    raw: data
  };
}

// 从 Bybit 获取数据
async function fetchFromBybit(symbol) {
  const tradingPair = SYMBOL_MAP[symbol.toLowerCase()]?.exchange || 'BTCUSDT';
  const api = EXCHANGE_APIS.find(e => e.name === 'Bybit');
  const url = `${api.baseUrl}/market/tickers?category=spot&symbol=${tradingPair}`;
  
  const data = await fetchWithTimeout(url, api.timeout);
  
  if (data.retCode !== 0 || !data.result || !data.result.list || data.result.list.length === 0) {
    throw new Error(data.retMsg || '未找到价格数据');
  }
  
  return {
    price: parseFloat(data.result.list[0].lastPrice),
    source: 'Bybit',
    raw: data.result.list[0]
  };
}

// 交易所获取函数映射
const EXCHANGE_FETCHERS = {
  'CoinGecko': fetchFromCoinGecko,
  'OKX': fetchFromOKX,
  'Kraken': fetchFromKraken,
  'Gate.io': fetchFromGateIO,
  'MEXC': fetchFromMEXC,
  'Bitget': fetchFromBitget,
  'Binance': fetchFromBinance,
  'Bybit': fetchFromBybit
};

// 获取市场数据（自动切换数据源）
const getMarketData = async (symbol = 'bitcoin', options = {}) => {
  const {
    maxRetries = 3, // 每个交易所最多重试次数
    timeout = 5000, // 超时时间
    enabledExchanges = null // 指定要使用的交易所，null 表示使用所有启用的
  } = options;
  
  // 获取启用的交易所，按优先级排序
  const enabledApis = EXCHANGE_APIS
    .filter(api => api.enabled && (!enabledExchanges || enabledExchanges.includes(api.name)))
    .sort((a, b) => a.priority - b.priority);
  
  if (enabledApis.length === 0) {
    throw new Error('没有可用的交易所接口');
  }
  
  const attempts = [];
  
  // 遍历所有启用的交易所
  for (let i = 0; i < enabledApis.length; i++) {
    const api = enabledApis[i];
    const fetcher = EXCHANGE_FETCHERS[api.name];
    
    if (!fetcher) {
      console.log(`⚠️  ${api.name} 没有对应的获取函数，跳过`);
      continue;
    }
    
    // 每个交易所最多重试 maxRetries 次
    for (let retry = 0; retry < maxRetries; retry++) {
      const startTime = Date.now();
      
      try {
        if (retry > 0) {
          console.log(`   🔄 ${api.name} 重试 ${retry + 1}/${maxRetries}...`);
        } else {
          console.log(`🔄 尝试数据源 ${i + 1}/${enabledApis.length}: ${api.name} (${api.type})...`);
        }
        
        const result = await fetcher(symbol);
        const responseTime = Date.now() - startTime;
        
        attempts.push({
          source: api.name,
          status: 'SUCCESS',
          responseTime: `${responseTime}ms`,
          retryCount: retry,
          data: result
        });
        
        return {
          success: true,
          source: api.name,
          type: api.type,
          statusCode: 200,
          responseTime: `${responseTime}ms`,
          price: result.price,
          data: result.raw,
          attempts: attempts
        };
      } catch (error) {
        const responseTime = Date.now() - startTime;
        
        attempts.push({
          source: api.name,
          status: 'FAILED',
          responseTime: `${responseTime}ms`,
          retryCount: retry,
          error: error.message
        });
        
        if (retry < maxRetries - 1) {
          // 等待一小段时间后重试
          await new Promise(resolve => setTimeout(resolve, 500));
        } else {
          console.log(`   ❌ ${api.name} 失败: ${error.message}`);
          
          // 如果不是最后一个交易所，继续尝试下一个
          if (i < enabledApis.length - 1) {
            console.log(`   ⏭️  切换到下一个数据源...\n`);
          }
        }
      }
    }
  }
  
  // 所有数据源都失败
  return {
    success: false,
    source: null,
    statusCode: null,
    responseTime: null,
    price: null,
    data: null,
    attempts: attempts,
    error: '所有数据源均失败'
  };
};

// 获取所有数据源的数据（并行获取）
const getAllMarketData = async (symbol = 'bitcoin', options = {}) => {
  const {
    timeout = 5000,
    enabledExchanges = null
  } = options;
  
  // 获取启用的交易所
  const enabledApis = EXCHANGE_APIS
    .filter(api => api.enabled && (!enabledExchanges || enabledExchanges.includes(api.name)))
    .sort((a, b) => a.priority - b.priority);
  
  if (enabledApis.length === 0) {
    throw new Error('没有可用的交易所接口');
  }
  
  console.log(`🔄 并行获取所有 ${enabledApis.length} 个数据源的数据...\n`);
  
  // 并行获取所有数据源
  const promises = enabledApis.map(async (api) => {
    const fetcher = EXCHANGE_FETCHERS[api.name];
    if (!fetcher) {
      return {
        source: api.name,
        type: api.type,
        priority: api.priority,
        success: false,
        error: '没有对应的获取函数'
      };
    }
    
    const startTime = Date.now();
    try {
      const result = await fetcher(symbol);
      const responseTime = Date.now() - startTime;
      
      return {
        source: api.name,
        type: api.type,
        priority: api.priority,
        success: true,
        price: result.price,
        responseTime: `${responseTime}ms`,
        data: result.raw,
        error: null
      };
    } catch (error) {
      const responseTime = Date.now() - startTime;
      return {
        source: api.name,
        type: api.type,
        priority: api.priority,
        success: false,
        price: null,
        responseTime: `${responseTime}ms`,
        data: null,
        error: error.message
      };
    }
  });
  
  const results = await Promise.all(promises);
  
  // 按优先级排序，成功的排在前面
  results.sort((a, b) => {
    if (a.success !== b.success) {
      return a.success ? -1 : 1;
    }
    return a.priority - b.priority;
  });
  
  const successful = results.filter(r => r.success);
  const failed = results.filter(r => !r.success);
  
  return {
    total: results.length,
    successful: successful.length,
    failed: failed.length,
    results: results,
    prices: successful.map(r => ({
      source: r.source,
      price: r.price
    })),
    // 计算平均价格
    averagePrice: successful.length > 0 
      ? successful.reduce((sum, r) => sum + r.price, 0) / successful.length 
      : null,
    // 计算价格范围
    priceRange: successful.length > 0
      ? {
          min: Math.min(...successful.map(r => r.price)),
          max: Math.max(...successful.map(r => r.price)),
          minSource: successful.reduce((min, r) => r.price < min.price ? r : min, successful[0]).source,
          maxSource: successful.reduce((max, r) => r.price > max.price ? r : max, successful[0]).source
        }
      : null
  };
};

// 批量测试所有交易所的健康状态
async function testAllExchanges(symbol = 'bitcoin') {
  console.log('🏥 测试所有交易所健康状态...\n');
  
  const results = [];
  
  for (const api of EXCHANGE_APIS.filter(e => e.enabled)) {
    const fetcher = EXCHANGE_FETCHERS[api.name];
    if (!fetcher) continue;
    
    const startTime = Date.now();
    try {
      const result = await fetcher(symbol);
      const responseTime = Date.now() - startTime;
      
      results.push({
        name: api.name,
        type: api.type,
        status: '✅ 可用',
        responseTime: `${responseTime}ms`,
        price: result.price,
        error: null
      });
    } catch (error) {
      const responseTime = Date.now() - startTime;
      results.push({
        name: api.name,
        type: api.type,
        status: '❌ 不可用',
        responseTime: `${responseTime}ms`,
        price: null,
        error: error.message
      });
    }
  }
  
  return results;
}

// 测试市场数据获取（自动切换模式）
async function testMarketData() {
  console.log('🔍 测试市场数据接口（多数据源自动切换）...\n');
  console.log('📊 可用交易所列表:');
  EXCHANGE_APIS.filter(e => e.enabled).forEach((api, index) => {
    console.log(`   ${index + 1}. ${api.name} (${api.type}) - 优先级: ${api.priority} - ${api.baseUrl}`);
  });
  console.log('');
  
  const result = await getMarketData('bitcoin', {
    maxRetries: 2,
    timeout: 5000
  });
  
  console.log('\n' + '='.repeat(60));
  if (result.success) {
    console.log(`✅ 成功获取数据！`);
    console.log(`   数据来源: ${result.source} (${result.type})`);
    console.log(`   响应时间: ${result.responseTime}`);
    console.log(`   价格: $${result.price?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`);
    console.log(`\n   完整数据:`, JSON.stringify(result.data, null, 2));
  } else {
    console.log(`❌ 所有数据源均失败`);
    console.log(`   错误: ${result.error}`);
  }
  
  console.log(`\n📋 尝试记录 (共 ${result.attempts.length} 次):`);
  const groupedAttempts = {};
  result.attempts.forEach(attempt => {
    if (!groupedAttempts[attempt.source]) {
      groupedAttempts[attempt.source] = [];
    }
    groupedAttempts[attempt.source].push(attempt);
  });
  
  Object.entries(groupedAttempts).forEach(([source, attempts]) => {
    const lastAttempt = attempts[attempts.length - 1];
    const icon = lastAttempt.status === 'SUCCESS' ? '✅' : '❌';
    const retryInfo = attempts.length > 1 ? ` (重试 ${attempts.length - 1} 次)` : '';
    console.log(`   ${icon} ${source}${retryInfo} - ${lastAttempt.status} (${lastAttempt.responseTime})`);
    if (lastAttempt.error) {
      console.log(`      错误: ${lastAttempt.error}`);
    }
  });
  console.log('='.repeat(60));
}

// 测试获取所有数据源
async function testAllMarketData() {
  console.log('🔍 测试获取所有数据源的数据...\n');
  console.log('📊 可用交易所列表:');
  EXCHANGE_APIS.filter(e => e.enabled).forEach((api, index) => {
    console.log(`   ${index + 1}. ${api.name} (${api.type}) - 优先级: ${api.priority}`);
  });
  console.log('');
  
  const result = await getAllMarketData('bitcoin', {
    timeout: 5000
  });
  
  console.log('\n' + '='.repeat(60));
  console.log(`📊 获取结果汇总:`);
  console.log(`   总数: ${result.total} 个数据源`);
  console.log(`   ✅ 成功: ${result.successful} 个`);
  console.log(`   ❌ 失败: ${result.failed} 个`);
  
  if (result.successful > 0) {
    console.log(`\n💰 价格信息:`);
    if (result.averagePrice) {
      console.log(`   平均价格: $${result.averagePrice.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`);
    }
    if (result.priceRange) {
      console.log(`   最低价格: $${result.priceRange.min.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} (${result.priceRange.minSource})`);
      console.log(`   最高价格: $${result.priceRange.max.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} (${result.priceRange.maxSource})`);
      console.log(`   价格差: $${(result.priceRange.max - result.priceRange.min).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`);
    }
    
    console.log(`\n✅ 成功的数据源 (${result.successful}):`);
    result.results
      .filter(r => r.success)
      .forEach((r, index) => {
        console.log(`   ${index + 1}. ${r.source} (${r.type}) - $${r.price?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} - ${r.responseTime}`);
      });
  }
  
  if (result.failed > 0) {
    console.log(`\n❌ 失败的数据源 (${result.failed}):`);
    result.results
      .filter(r => !r.success)
      .forEach((r, index) => {
        console.log(`   ${index + 1}. ${r.source} (${r.type}) - ${r.responseTime} - 错误: ${r.error}`);
      });
  }
  
  console.log('\n📋 详细数据:');
  result.results.forEach((r, index) => {
    const icon = r.success ? '✅' : '❌';
    console.log(`\n   ${index + 1}. ${icon} ${r.source} (${r.type})`);
    if (r.success) {
      console.log(`      价格: $${r.price?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`);
      console.log(`      响应时间: ${r.responseTime}`);
    } else {
      console.log(`      错误: ${r.error}`);
      console.log(`      响应时间: ${r.responseTime}`);
    }
  });
  
  console.log('='.repeat(60));
  
  return result;
}

// 测试所有交易所健康状态
async function testHealthCheck() {
  console.log('🏥 测试所有交易所健康状态...\n');
  
  const results = await testAllExchanges('bitcoin');
  
  console.log('\n' + '='.repeat(60));
  console.log('📊 健康检查结果:\n');
  
  const available = results.filter(r => r.status.includes('✅'));
  const unavailable = results.filter(r => r.status.includes('❌'));
  
  console.log(`✅ 可用: ${available.length}/${results.length}`);
  available.forEach(r => {
    console.log(`   ${r.name} (${r.type}) - ${r.responseTime} - 价格: $${r.price?.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`);
  });
  
  if (unavailable.length > 0) {
    console.log(`\n❌ 不可用: ${unavailable.length}/${results.length}`);
    unavailable.forEach(r => {
      console.log(`   ${r.name} (${r.type}) - ${r.responseTime} - 错误: ${r.error}`);
    });
  }
  
  console.log('='.repeat(60));
  
  return results;
}

// 主测试函数
async function main() {
  const args = process.argv.slice(2);
  const command = args[0] || 'all';
  
  if (command === 'health' || command === 'h') {
    await testHealthCheck();
  } else if (command === 'test' || command === 't') {
    await testMarketData();
  } else if (command === 'all' || command === 'a') {
    await testAllMarketData();
  } else if (command === 'diagnose' || command === 'd') {
    // 运行网络诊断
    const { diagnoseAllAPIs } = require('./network-diagnosis.js');
    await diagnoseAllAPIs();
  } else {
    console.log('用法:');
    console.log('  node test/api.js all      - 获取所有数据源的数据（默认）');
    console.log('  node test/api.js test    - 测试自动切换获取数据（第一个成功即返回）');
    console.log('  node test/api.js health  - 测试所有交易所健康状态');
    console.log('  node test/api.js diagnose - 运行网络诊断（推荐先运行此命令）');
    console.log('\n默认执行获取所有数据源模式...\n');
    await testAllMarketData();
  }
}

main();