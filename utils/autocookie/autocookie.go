package autocookie

import (
	"Crawler/utils/logger"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Manager Cookie管理器
type Manager struct {
	enabled     bool
	browserType string
	headless    bool
}

// Config Cookie配置
type Config struct {
	Enabled     bool   `json:"enabled"`
	BrowserType string `json:"browser_type"`
	Headless    bool   `json:"headless"`
}

// NewManager 创建Cookie管理器
func NewManager(config Config) *Manager {
	return &Manager{
		enabled:     config.Enabled,
		browserType: config.BrowserType,
		headless:    config.Headless,
	}
}

// GetCookies 获取指定网站的Cookie
func (m *Manager) GetCookies(platform string) (string, error) {
	if !m.enabled {
		return "", fmt.Errorf("自动获取Cookie功能未启用")
	}

	var url string
	switch platform {
	case "douyin":
		url = "https://www.douyin.com/"
	case "kuaishou":
		url = "https://www.kuaishou.com/"
	default:
		return "", fmt.Errorf("不支持的平台: %s", platform)
	}

	logger.Info("正在自动获取 %s 的Cookie...", platform)

	// 创建Chrome实例的选项
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
	}

	// 根据配置决定是否使用无头模式
	if m.headless {
		opts = append(opts, chromedp.Headless)
	}

	// 创建Chrome实例
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建新的Chrome实例
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// 访问目标网站
	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		return "", fmt.Errorf("访问网站失败: %v", err)
	}

	// 等待页面加载完成
	if err := chromedp.Run(ctx, chromedp.Sleep(5*time.Second)); err != nil {
		return "", fmt.Errorf("等待页面加载失败: %v", err)
	}

	logger.Info("页面已加载，请在浏览器中手动登录...")

	// 等待用户登录
	if !m.headless {
		logger.Info("请在打开的浏览器中登录，登录成功后按回车继续...")
		fmt.Scanln()
	} else {
		// 在无头模式下，等待一段时间
		if err := chromedp.Run(ctx, chromedp.Sleep(30*time.Second)); err != nil {
			return "", fmt.Errorf("等待登录超时: %v", err)
		}
	}

	// 获取Cookie
	var cookies []*struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := chromedp.Run(ctx, chromedp.Evaluate(`
		(function() {
			return document.cookie.split(';').map(function(cookie) {
				const parts = cookie.trim().split('=');
				return {name: parts[0], value: parts[1]};
			});
		})()
	`, &cookies)); err != nil {
		return "", fmt.Errorf("获取Cookie失败: %v", err)
	}

	// 格式化Cookie
	var cookieParts []string
	for _, cookie := range cookies {
		cookieParts = append(cookieParts, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	cookieStr := strings.Join(cookieParts, "; ")

	logger.Info("成功获取Cookie: %s", cookieStr[:min(len(cookieStr), 30)]+"...")
	return cookieStr, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
