package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// DouyinScraper 抖音平台爬虫实现
type DouyinScraper struct {
	client    *http.Client
	userAgent string
	cookies   string
}

// NewDouyinScraper 创建抖音爬虫实例
func NewDouyinScraper(userAgent, cookies string) *DouyinScraper {
	return &DouyinScraper{
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		userAgent: userAgent,
		cookies:   cookies,
	}
}

// Initialize 初始化爬虫
func (s *DouyinScraper) Initialize() error {
	// 验证cookies是否有效
	req, err := http.NewRequest("GET", "https://www.douyin.com/", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("invalid cookies or blocked by anti-crawler")
	}

	return nil
}

// GetUserInfo 获取用户信息
func (s *DouyinScraper) GetUserInfo(userID string) (*UserData, error) {
	// 构建API请求URL
	apiURL := fmt.Sprintf("https://www.douyin.com/aweme/v1/web/user/profile/other/?user_id=%s", url.QueryEscape(userID))

	// 发送请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Referer", "https://www.douyin.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %v", err)
	}

	// 解析JSON响应
	var result struct {
		UserInfo   UserData `json:"user_info"`
		StatusCode int      `json:"status_code"`
		StatusMsg  string   `json:"status_msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %v", err)
	}

	// 检查API响应状态
	if result.StatusCode != 0 {
		return nil, fmt.Errorf("API返回错误: %s", result.StatusMsg)
	}

	return &result.UserInfo, nil
}

// GetUserVideos 获取用户视频列表
func (s *DouyinScraper) GetUserVideos(userID string, cursor string) ([]*VideoData, string, error) {
	// 构建API请求URL
	apiURL := fmt.Sprintf("https://www.douyin.com/aweme/v1/web/aweme/post/?user_id=%s&count=20&cursor=%s",
		url.QueryEscape(userID), cursor)

	// 发送请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, "", err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Referer", "https://www.douyin.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// 解析JSON响应
	var result struct {
		AwemeList []*VideoData `json:"aweme_list"`
		HasMore   int          `json:"has_more"`
		Cursor    string       `json:"cursor"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "", err
	}

	return result.AwemeList, result.Cursor, nil
}

// GetVideoComments 获取视频评论
func (s *DouyinScraper) GetVideoComments(videoID string, cursor string) ([]*CommentData, string, error) {
	// 构建API请求URL
	apiURL := fmt.Sprintf("https://www.douyin.com/aweme/v2/web/comment/list/?aweme_id=%s&cursor=%s&count=20",
		url.QueryEscape(videoID), cursor)

	// 发送请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, "", err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Referer", "https://www.douyin.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// 解析JSON响应
	var result struct {
		Comments []*CommentData `json:"comments"`
		HasMore  int            `json:"has_more"`
		Cursor   string         `json:"cursor"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "", err
	}

	return result.Comments, result.Cursor, nil
}

// GetProductInfo 获取商品信息
func (s *DouyinScraper) GetProductInfo(productID string) (*ProductInfo, error) {
	// 构建API请求URL
	apiURL := fmt.Sprintf("https://www.douyin.com/aweme/v1/web/promotion/product/detail/?product_id=%s",
		url.QueryEscape(productID))

	// 发送请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Referer", "https://www.douyin.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var result struct {
		ProductInfo ProductInfo `json:"product_info"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.ProductInfo, nil
}
