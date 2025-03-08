package crawler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/url"
	"time"
)

// KuaishouScraper 快手平台爬虫实现
type KuaishouScraper struct {
	client    *http.Client
	userAgent string
	cookies   string
}

// NewKuaishouScraper 创建快手爬虫实例
func NewKuaishouScraper(userAgent, cookies string) *KuaishouScraper {
	return &KuaishouScraper{
		client: &http.Client{
			Timeout: time.Second * 30,
		},
		userAgent: userAgent,
		cookies:   cookies,
	}
}

// Initialize 初始化爬虫
func (s *KuaishouScraper) Initialize() error {
	// 验证cookies是否有效
	req, err := http.NewRequest("GET", "https://www.kuaishou.com/", nil)
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
func (s *KuaishouScraper) GetUserInfo(userID string) (*UserData, error) {
	// 构建API请求URL
	apiURL := "https://www.kuaishou.com/graphql"

	// 构建GraphQL查询
	query := fmt.Sprintf(`{
		visionProfile(userId: "%s") {
			user {
				id
				name
				followersCount
				followingCount
				description
				tags
			}
		}
	}`, userID)

	// 构建请求体
	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.kuaishou.com/")

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
		Data struct {
			VisionProfile struct {
				User UserData `json:"user"`
			} `json:"visionProfile"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Data.VisionProfile.User, nil
}

// GetUserVideos 获取用户视频列表
func (s *KuaishouScraper) GetUserVideos(userID string, cursor string) ([]*VideoData, string, error) {
	// 构建API请求URL
	apiURL := "https://www.kuaishou.com/graphql"

	// 构建GraphQL查询
	query := fmt.Sprintf(`{
		visionProfilePhotoList(userId: "%s", pcursor: "%s", page: "profile") {
			pcursor
			feeds {
				photoId
				caption
				likeCount
				commentCount
				viewCount
				tags
				productInfo {
					id
					name
					price
					category
					description
					sales
				}
			}
		}
	}`, userID, cursor)

	// 构建请求体
	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, "", err
	}

	// 发送请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, "", err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.kuaishou.com/")

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
		Data struct {
			VisionProfilePhotoList struct {
				Pcursor string       `json:"pcursor"`
				Feeds   []*VideoData `json:"feeds"`
			} `json:"visionProfilePhotoList"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "", err
	}

	return result.Data.VisionProfilePhotoList.Feeds, result.Data.VisionProfilePhotoList.Pcursor, nil
}

// GetVideoComments 获取视频评论
func (s *KuaishouScraper) GetVideoComments(videoID string, cursor string) ([]*CommentData, string, error) {
	// 构建API请求URL
	apiURL := "https://www.kuaishou.com/graphql"

	// 构建GraphQL查询
	query := fmt.Sprintf(`{
		photoCommentList(photoId: "%s", pcursor: "%s") {
			pcursor
			comments {
				id
				photoId
				authorId
				content
				likeCount
				replyCount
				createTime
			}
		}
	}`, videoID, cursor)

	// 构建请求体
	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, "", err
	}

	// 发送请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, "", err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.kuaishou.com/")

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
		Data struct {
			PhotoCommentList struct {
				Pcursor  string         `json:"pcursor"`
				Comments []*CommentData `json:"comments"`
			} `json:"photoCommentList"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, "", err
	}

	return result.Data.PhotoCommentList.Comments, result.Data.PhotoCommentList.Pcursor, nil
}

// GetProductInfo 获取商品信息
func (s *KuaishouScraper) GetProductInfo(productID string) (*ProductInfo, error) {
	// 构建API请求URL
	apiURL := "https://www.kuaishou.com/graphql"

	// 构建GraphQL查询
	query := fmt.Sprintf(`{
		productInfo(productId: "%s") {
			id
			name
			price
			category
			description
			sales
		}
	}`, productID)

	// 构建请求体
	requestBody := map[string]interface{}{
		"query": query,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.kuaishou.com/")

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
		Data struct {
			ProductInfo ProductInfo `json:"productInfo"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result.Data.ProductInfo, nil
}
