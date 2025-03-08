package ratelimit

import (
	"Crawler/utils/logger"
	"sync"
	"time"
)

// Limiter 速率限制器
type Limiter struct {
	enabled           bool
	requestsPerMinute int
	interval          time.Duration
	tokens            int
	lastRefill        time.Time
	mutex             sync.Mutex
}

// Config 速率限制配置
type Config struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
}

// NewLimiter 创建速率限制器
func NewLimiter(config Config) *Limiter {
	requestsPerMinute := config.RequestsPerMinute
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60 // 默认每分钟60个请求
	}

	interval := time.Minute / time.Duration(requestsPerMinute)

	return &Limiter{
		enabled:           config.Enabled,
		requestsPerMinute: requestsPerMinute,
		interval:          interval,
		tokens:            requestsPerMinute,
		lastRefill:        time.Now(),
	}
}

// Wait 等待获取令牌
func (l *Limiter) Wait() {
	if !l.enabled {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 计算自上次补充以来经过的时间
	elapsed := time.Since(l.lastRefill)

	// 计算应该补充的令牌数量
	tokensToAdd := int(elapsed / l.interval)

	if tokensToAdd > 0 {
		// 补充令牌
		l.tokens = min(l.requestsPerMinute, l.tokens+tokensToAdd)
		l.lastRefill = l.lastRefill.Add(time.Duration(tokensToAdd) * l.interval)
	}

	// 如果没有可用的令牌，则等待
	if l.tokens <= 0 {
		waitTime := l.interval - time.Since(l.lastRefill)
		if waitTime > 0 {
			l.mutex.Unlock()
			logger.Debug("速率限制，等待 %v", waitTime)
			time.Sleep(waitTime)
			l.mutex.Lock()
		}

		// 补充一个令牌
		l.tokens = 1
		l.lastRefill = time.Now()
	}

	// 消耗一个令牌
	l.tokens--
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
