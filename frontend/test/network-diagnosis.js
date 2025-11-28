// 网络诊断工具 - 帮助诊断 API 访问问题

// 测试 DNS 解析
async function testDNSResolution(hostname) {
  console.log(`\n🔍 测试 DNS 解析: ${hostname}`);
  
  // 在浏览器环境中，可以通过尝试连接来测试 DNS
  // 在 Node.js 环境中，可以使用 dns 模块
  try {
    const testUrl = `https://${hostname}`;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 3000);
    
    const startTime = Date.now();
    await fetch(testUrl, { 
      signal: controller.signal,
      method: 'HEAD',
      mode: 'no-cors' // 避免 CORS 问题，只测试连接
    });
    clearTimeout(timeoutId);
    
    const time = Date.now() - startTime;
    console.log(`   ✅ DNS 解析成功 (${time}ms)`);
    return true;
  } catch (error) {
    console.log(`   ❌ DNS 解析失败: ${error.message}`);
    return false;
  }
}

// 测试基本连接
async function testBasicConnection(url) {
  console.log(`\n🔗 测试基本连接: ${url}`);
  
  const startTime = Date.now();
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 10000);
    
    const response = await fetch(url, {
      signal: controller.signal,
      method: 'HEAD',
      headers: {
        'User-Agent': 'Mozilla/5.0'
      }
    });
    clearTimeout(timeoutId);
    
    const time = Date.now() - startTime;
    console.log(`   ✅ 连接成功 (${time}ms) - 状态码: ${response.status}`);
    return { success: true, time, status: response.status };
  } catch (error) {
    const time = Date.now() - startTime;
    console.log(`   ❌ 连接失败 (${time}ms): ${error.message}`);
    return { success: false, time, error: error.message };
  }
}

// 检查环境信息
function checkEnvironment() {
  console.log('\n📋 环境信息:');
  
  // 检查是否在浏览器环境
  if (typeof window !== 'undefined') {
    console.log('   环境: 浏览器');
    console.log(`   User Agent: ${navigator.userAgent}`);
    console.log(`   在线状态: ${navigator.onLine ? '在线' : '离线'}`);
    
    // 检查代理设置（浏览器中无法直接检测，但可以提示）
    console.log('   ⚠️  提示: 浏览器环境可能受 CORS 限制');
    console.log('   ⚠️  提示: 请确保 VPN 已正确配置系统代理');
  } else {
    console.log('   环境: Node.js');
    
    // 检查环境变量
    const httpProxy = process.env.HTTP_PROXY || process.env.http_proxy;
    const httpsProxy = process.env.HTTPS_PROXY || process.env.https_proxy;
    
    if (httpProxy || httpsProxy) {
      console.log(`   HTTP_PROXY: ${httpProxy || '未设置'}`);
      console.log(`   HTTPS_PROXY: ${httpsProxy || '未设置'}`);
    } else {
      console.log('   ⚠️  未检测到代理环境变量');
      console.log('   💡 提示: 如果使用 VPN，可能需要设置代理环境变量');
      console.log('   💡 例如: export HTTPS_PROXY=http://127.0.0.1:7890');
    }
  }
}

// 诊断所有交易所 API
async function diagnoseAllAPIs() {
  console.log('🏥 开始网络诊断...\n');
  console.log('='.repeat(60));
  
  checkEnvironment();
  
  const apis = [
    { name: 'CoinGecko', url: 'https://api.coingecko.com/api/v3/ping', hostname: 'api.coingecko.com' },
    { name: 'OKX', url: 'https://www.okx.com/api/v5/market/ticker?instId=BTC-USDT', hostname: 'www.okx.com' },
    { name: 'Kraken', url: 'https://api.kraken.com/0/public/Time', hostname: 'api.kraken.com' },
    { name: 'Gate.io', url: 'https://api.gateio.ws/api/v4/spot/currencies', hostname: 'api.gateio.ws' },
    { name: 'MEXC', url: 'https://api.mexc.com/api/v3/ping', hostname: 'api.mexc.com' },
    { name: 'Bitget', url: 'https://api.bitget.com/api/spot/v1/market/ticker?symbol=BTCUSDT', hostname: 'api.bitget.com' },
    { name: 'Binance', url: 'https://api.binance.com/api/v3/ping', hostname: 'api.binance.com' },
    { name: 'Bybit', url: 'https://api.bybit.com/v5/market/time', hostname: 'api.bybit.com' }
  ];
  
  console.log('\n' + '='.repeat(60));
  console.log('🔍 DNS 解析测试:');
  console.log('='.repeat(60));
  
  const dnsResults = [];
  for (const api of apis) {
    const result = await testDNSResolution(api.hostname);
    dnsResults.push({ name: api.name, hostname: api.hostname, success: result });
    // 避免请求过快
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  
  console.log('\n' + '='.repeat(60));
  console.log('🔗 连接测试:');
  console.log('='.repeat(60));
  
  const connectionResults = [];
  for (const api of apis) {
    const result = await testBasicConnection(api.url);
    connectionResults.push({ name: api.name, ...result });
    // 避免请求过快
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  console.log('\n' + '='.repeat(60));
  console.log('📊 诊断结果汇总:');
  console.log('='.repeat(60));
  
  console.log('\nDNS 解析:');
  const dnsSuccess = dnsResults.filter(r => r.success).length;
  console.log(`   ✅ 成功: ${dnsSuccess}/${dnsResults.length}`);
  dnsResults.filter(r => r.success).forEach(r => {
    console.log(`      ${r.name} (${r.hostname})`);
  });
  const dnsFailed = dnsResults.filter(r => !r.success);
  if (dnsFailed.length > 0) {
    console.log(`   ❌ 失败: ${dnsFailed.length}/${dnsResults.length}`);
    dnsFailed.forEach(r => {
      console.log(`      ${r.name} (${r.hostname})`);
    });
    console.log('\n   💡 DNS 解析失败的可能原因:');
    console.log('      1. DNS 服务器配置问题');
    console.log('      2. 网络连接问题');
    console.log('      3. 防火墙阻止 DNS 查询');
    console.log('      4. VPN 未正确配置 DNS');
  }
  
  console.log('\n连接测试:');
  const connSuccess = connectionResults.filter(r => r.success).length;
  console.log(`   ✅ 成功: ${connSuccess}/${connectionResults.length}`);
  connectionResults.filter(r => r.success).forEach(r => {
    console.log(`      ${r.name} - ${r.time}ms (状态码: ${r.status})`);
  });
  const connFailed = connectionResults.filter(r => !r.success);
  if (connFailed.length > 0) {
    console.log(`   ❌ 失败: ${connFailed.length}/${connectionResults.length}`);
    connFailed.forEach(r => {
      console.log(`      ${r.name} - ${r.error}`);
    });
    console.log('\n   💡 连接失败的可能原因:');
    console.log('      1. 防火墙或安全软件阻止连接');
    console.log('      2. VPN 代理未正确配置');
    console.log('      3. 某些 API 可能被地区限制');
    console.log('      4. 网络质量问题（超时）');
    console.log('      5. CORS 限制（浏览器环境）');
  }
  
  console.log('\n' + '='.repeat(60));
  console.log('💡 建议解决方案:');
  console.log('='.repeat(60));
  
  if (dnsFailed.length > 0 || connFailed.length > 0) {
    console.log('\n1. 检查 VPN 配置:');
    console.log('   - 确保 VPN 已正确连接');
    console.log('   - 检查 VPN 是否配置了系统代理');
    console.log('   - 如果使用 Node.js，设置代理环境变量:');
    console.log('     export HTTPS_PROXY=http://127.0.0.1:你的代理端口');
    
    console.log('\n2. 检查 DNS 设置:');
    console.log('   - 尝试更换 DNS 服务器（如 8.8.8.8 或 1.1.1.1）');
    console.log('   - 检查 hosts 文件是否有相关配置');
    
    console.log('\n3. 检查防火墙:');
    console.log('   - 临时关闭防火墙测试');
    console.log('   - 检查安全软件是否阻止了连接');
    
    console.log('\n4. 如果是在浏览器环境:');
    console.log('   - 某些 API 可能有 CORS 限制');
    console.log('   - 考虑使用后端代理来访问这些 API');
    
    console.log('\n5. 增加超时时间:');
    console.log('   - 某些 API 响应较慢，可以增加超时时间');
  } else {
    console.log('\n✅ 所有测试通过！如果仍然无法获取数据，可能是 API 本身的问题。');
  }
  
  console.log('\n' + '='.repeat(60));
  
  return {
    dns: dnsResults,
    connection: connectionResults
  };
}

// 如果在 Node.js 环境中运行
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    diagnoseAllAPIs,
    testDNSResolution,
    testBasicConnection,
    checkEnvironment
  };
}

// 如果在浏览器环境中运行
if (typeof window !== 'undefined') {
  window.networkDiagnosis = {
    diagnoseAllAPIs,
    testDNSResolution,
    testBasicConnection,
    checkEnvironment
  };
}

