package proxy

import (
	"Crawler/utils/logger"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Manager 代理管理器
type Manager struct {
	enabled          bool
	proxyType        string
	proxyList        []string
	currentIndex     int
	rotationInterval int
	lastRotation     time.Time
	mutex            sync.Mutex
}

// Config 代理配置
type Config struct {
	Enabled          bool     `json:"enabled"`
	Type             string   `json:"type"`
	List             []string `json:"list"`
	RotationInterval int      `json:"rotation_interval"`
}

// NewManager 创建代理管理器
func NewManager(config Config) *Manager {
	return &Manager{
		enabled:          config.Enabled,
		proxyType:        config.Type,
		proxyList:        config.List,
		currentIndex:     0,
		rotationInterval: config.RotationInterval,
		lastRotation:     time.Now(),
	}
}

// GetTransport 获取配置了代理的HTTP传输
func (m *Manager) GetTransport() *http.Transport {
	if !m.enabled || len(m.proxyList) == 0 {
		return &http.Transport{}
	}

	proxyURL := m.GetProxy()
	if proxyURL == "" {
		return &http.Transport{}
	}

	proxy, err := url.Parse(proxyURL)
	if err != nil {
		logger.Error("解析代理URL失败: %v", err)
		return &http.Transport{}
	}

	return &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
}

// GetProxy 获取当前代理
func (m *Manager) GetProxy() string {
	if !m.enabled || len(m.proxyList) == 0 {
		return ""
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否需要轮换代理
	if m.rotationInterval > 0 && time.Since(m.lastRotation).Seconds() > float64(m.rotationInterval) {
		m.currentIndex = (m.currentIndex + 1) % len(m.proxyList)
		m.lastRotation = time.Now()
		logger.Info("轮换代理，当前使用: %s", m.proxyList[m.currentIndex])
	}

	return m.proxyList[m.currentIndex]
}

// RotateProxy 手动轮换代理
func (m *Manager) RotateProxy() string {
	if !m.enabled || len(m.proxyList) == 0 {
		return ""
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.currentIndex = (m.currentIndex + 1) % len(m.proxyList)
	m.lastRotation = time.Now()
	logger.Info("手动轮换代理，当前使用: %s", m.proxyList[m.currentIndex])

	return m.proxyList[m.currentIndex]
}

// IsEnabled 检查代理是否启用
func (m *Manager) IsEnabled() bool {
	return m.enabled
}
