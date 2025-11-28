package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"wails-contract-warn/logger"
)

// ProxyClient HTTP 代理客户端
type ProxyClient struct {
	client *http.Client
}

// NewProxyClient 创建代理客户端
func NewProxyClient() *ProxyClient {
	// 配置 HTTP 客户端，支持代理和自定义 DNS
	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:      90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
	}

	// 如果设置了代理环境变量，会自动使用
	// 也可以通过代码设置：transport.Proxy = http.ProxyURL(proxyURL)

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second, // 30秒超时
	}

	return &ProxyClient{
		client: client,
	}
}

// FetchAPI 代理获取 API 数据
func (p *ProxyClient) FetchAPI(url string, headers map[string]string) (map[string]interface{}, error) {
	logger.Debugf("代理请求: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	// 添加自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		logger.Errorf("请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("API 返回非200状态码: %d, 响应: %s", resp.StatusCode, string(body[:min(200, len(body))]))
		return nil, fmt.Errorf("API 返回错误: %s (状态码: %d)", string(body[:min(200, len(body))]), resp.StatusCode)
	}

	// 解析 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		// 如果不是 JSON，返回原始文本
		return map[string]interface{}{
			"raw": string(body),
		}, nil
	}

	logger.Debugf("代理请求成功: %s (状态码: %d)", url, resp.StatusCode)
	return result, nil
}

// FetchAPIRaw 获取原始响应（不解析 JSON）
func (p *ProxyClient) FetchAPIRaw(url string, headers map[string]string) ([]byte, error) {
	logger.Debugf("代理请求(原始): %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	// 添加自定义请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		logger.Errorf("请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("API 返回非200状态码: %d", resp.StatusCode)
		return nil, fmt.Errorf("API 返回错误 (状态码: %d)", resp.StatusCode)
	}

	logger.Debugf("代理请求成功: %s (状态码: %d, 大小: %d bytes)", url, resp.StatusCode, len(body))
	return body, nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

