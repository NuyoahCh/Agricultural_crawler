package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// ProxyInfo 代理信息结构
type ProxyInfo struct {
	IP        string    `json:"ip"`
	Port      string    `json:"port"`
	Protocol  string    `json:"protocol"` // http, https, socks5
	Username  string    `json:"username,omitempty"`
	Password  string    `json:"password,omitempty"`
	LastUsed  time.Time `json:"last_used"`
	FailCount int       `json:"fail_count"`
	IsValid   bool      `json:"is_valid"`
}

// ProxyString 返回代理字符串
func (p *ProxyInfo) ProxyString() string {
	if p.Username != "" && p.Password != "" {
		return fmt.Sprintf("%s://%s:%s@%s:%s", p.Protocol, p.Username, p.Password, p.IP, p.Port)
	}
	return fmt.Sprintf("%s://%s:%s", p.Protocol, p.IP, p.Port)
}

// ProxyPool 代理IP池
type ProxyPool struct {
	proxies     []*ProxyInfo
	mutex       sync.RWMutex
	proxyFile   string
	testURL     string
	maxFailures int
}

// NewProxyPool 创建新的代理池
func NewProxyPool(proxyFile, testURL string, maxFailures int) *ProxyPool {
	pool := &ProxyPool{
		proxies:     make([]*ProxyInfo, 0),
		proxyFile:   proxyFile,
		testURL:     testURL,
		maxFailures: maxFailures,
	}

	// 加载代理列表
	pool.LoadProxies()

	// 启动代理验证协程
	go pool.validateProxiesRoutine()

	return pool
}

// LoadProxies 从文件加载代理列表
func (p *ProxyPool) LoadProxies() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(p.proxyFile); os.IsNotExist(err) {
		// 文件不存在，创建空文件
		p.proxies = make([]*ProxyInfo, 0)
		return p.SaveProxies()
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(p.proxyFile)
	if err != nil {
		return err
	}

	// 解析JSON
	if len(data) > 0 {
		if err := json.Unmarshal(data, &p.proxies); err != nil {
			return err
		}
	}

	return nil
}

// SaveProxies 保存代理列表到文件
func (p *ProxyPool) SaveProxies() error {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 将代理列表转换为JSON
	data, err := json.MarshalIndent(p.proxies, "", "  ")
	if err != nil {
		return err
	}

	// 保存到文件
	return ioutil.WriteFile(p.proxyFile, data, 0644)
}

// AddProxy 添加代理
func (p *ProxyPool) AddProxy(proxy *ProxyInfo) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 检查代理是否已存在
	for _, existingProxy := range p.proxies {
		if existingProxy.IP == proxy.IP && existingProxy.Port == proxy.Port {
			// 代理已存在，更新信息
			existingProxy.Protocol = proxy.Protocol
			existingProxy.Username = proxy.Username
			existingProxy.Password = proxy.Password
			existingProxy.IsValid = proxy.IsValid
			return
		}
	}

	// 添加新代理
	p.proxies = append(p.proxies, proxy)

	// 保存代理列表
	go p.SaveProxies()
}

// RemoveProxy 移除代理
func (p *ProxyPool) RemoveProxy(ip, port string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i, proxy := range p.proxies {
		if proxy.IP == ip && proxy.Port == port {
			// 移除代理
			p.proxies = append(p.proxies[:i], p.proxies[i+1:]...)
			break
		}
	}

	// 保存代理列表
	go p.SaveProxies()
}

// GetProxy 获取一个可用代理
func (p *ProxyPool) GetProxy() *ProxyInfo {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 筛选有效代理
	validProxies := make([]*ProxyInfo, 0)
	for _, proxy := range p.proxies {
		if proxy.IsValid && proxy.FailCount < p.maxFailures {
			validProxies = append(validProxies, proxy)
		}
	}

	// 如果没有有效代理，返回nil
	if len(validProxies) == 0 {
		return nil
	}

	// 随机选择一个代理
	rand.Seed(time.Now().UnixNano())
	selectedProxy := validProxies[rand.Intn(len(validProxies))]

	// 更新最后使用时间
	selectedProxy.LastUsed = time.Now()

	return selectedProxy
}

// ValidateProxy 验证代理是否可用
func (p *ProxyPool) ValidateProxy(proxy *ProxyInfo) bool {
	// 创建代理URL
	proxyURL, err := url.Parse(proxy.ProxyString())
	if err != nil {
		return false
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	// 发送请求
	resp, err := client.Get(p.testURL)
	if err != nil {
		proxy.FailCount++
		proxy.IsValid = false
		return false
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		proxy.FailCount++
		proxy.IsValid = false
		return false
	}

	// 代理有效
	proxy.FailCount = 0
	proxy.IsValid = true
	return true
}

// validateProxiesRoutine 定期验证代理
func (p *ProxyPool) validateProxiesRoutine() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.ValidateAllProxies()
	}
}

// ValidateAllProxies 验证所有代理
func (p *ProxyPool) ValidateAllProxies() {
	p.mutex.Lock()
	proxies := make([]*ProxyInfo, len(p.proxies))
	copy(proxies, p.proxies)
	p.mutex.Unlock()

	var wg sync.WaitGroup
	for _, proxy := range proxies {
		wg.Add(1)
		go func(proxy *ProxyInfo) {
			defer wg.Done()
			p.ValidateProxy(proxy)
		}(proxy)
	}

	wg.Wait()
	p.SaveProxies()
}

// ImportProxiesFromAPI 从API导入代理
func (p *ProxyPool) ImportProxiesFromAPI(apiURL string) error {
	// 发送请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 解析JSON
	var apiProxies []struct {
		IP       string `json:"ip"`
		Port     string `json:"port"`
		Protocol string `json:"protocol"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}

	if err := json.Unmarshal(body, &apiProxies); err != nil {
		return err
	}

	// 添加代理
	for _, apiProxy := range apiProxies {
		proxy := &ProxyInfo{
			IP:       apiProxy.IP,
			Port:     apiProxy.Port,
			Protocol: apiProxy.Protocol,
			Username: apiProxy.Username,
			Password: apiProxy.Password,
			IsValid:  true,
		}
		p.AddProxy(proxy)
	}

	return nil
}

// GetProxyCount 获取代理数量
func (p *ProxyPool) GetProxyCount() (total, valid int) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	total = len(p.proxies)
	for _, proxy := range p.proxies {
		if proxy.IsValid && proxy.FailCount < p.maxFailures {
			valid++
		}
	}

	return
}
