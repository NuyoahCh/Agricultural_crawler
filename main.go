package main

import (
	"Crawler/crawler"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Config 存储爬虫配置信息
type Config struct {
	Concurrency int
	UserAgent   string
	Timeout     int
	Retries     int
	Cookies     string
	Platform    string
	OutputDir   string
}

// Crawler 爬虫主结构体
type Crawler struct {
	config     Config
	wg         sync.WaitGroup
	mutex      sync.Mutex
	urlChannel chan string
	scraper    crawler.Scraper
}

// NewCrawler 创建新的爬虫实例
func NewCrawler(config Config) *Crawler {
	return &Crawler{
		config:     config,
		urlChannel: make(chan string, config.Concurrency),
	}
}

// Initialize 初始化爬虫
func (c *Crawler) Initialize() error {
	// 根据平台选择对应的爬虫实现
	switch crawler.Platform(c.config.Platform) {
	case crawler.Douyin:
		c.scraper = crawler.NewDouyinScraper(c.config.UserAgent, c.config.Cookies)
	case crawler.Kuaishou:
		c.scraper = crawler.NewKuaishouScraper(c.config.UserAgent, c.config.Cookies)
	default:
		return fmt.Errorf("不支持的平台: %s", c.config.Platform)
	}

	// 初始化爬虫
	if err := c.scraper.Initialize(); err != nil {
		return fmt.Errorf("爬虫初始化失败: %v", err)
	}

	// 创建输出目录
	if err := os.MkdirAll(c.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	return nil
}

// Start 启动爬虫
func (c *Crawler) Start(userIDs []string) {
	// 启动工作协程
	for i := 0; i < c.config.Concurrency; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	// 将用户ID加入队列
	for _, userID := range userIDs {
		c.urlChannel <- userID
	}

	// 关闭通道
	close(c.urlChannel)

	// 等待所有工作协程完成
	c.wg.Wait()
	log.Println("爬虫任务完成")
}

// worker 工作协程
func (c *Crawler) worker() {
	defer c.wg.Done()

	for userID := range c.urlChannel {
		// 获取用户信息
		userData, err := c.scraper.GetUserInfo(userID)
		if err != nil {
			log.Printf("获取用户 %s 信息失败: %v", userID, err)
			continue
		}

		// 保存用户信息
		c.saveUserData(userData)

		// 获取用户视频列表
		c.crawlUserVideos(userID)

		// 休眠一段时间，避免请求过于频繁
		time.Sleep(time.Duration(c.config.Timeout) * time.Second)
	}
}

// crawlUserVideos 爬取用户视频列表
func (c *Crawler) crawlUserVideos(userID string) {
	cursor := ""
	retryCount := 0

	for {
		// 获取用户视频列表
		videos, nextCursor, err := c.scraper.GetUserVideos(userID, cursor)
		if err != nil {
			retryCount++
			if retryCount > c.config.Retries {
				log.Printf("获取用户 %s 视频列表失败，已达到最大重试次数", userID)
				break
			}
			log.Printf("获取用户 %s 视频列表失败: %v，正在重试...", userID, err)
			time.Sleep(time.Duration(c.config.Timeout) * time.Second)
			continue
		}

		// 重置重试计数
		retryCount = 0

		// 保存视频数据
		for _, video := range videos {
			c.saveVideoData(video)

			// 获取视频评论
			c.crawlVideoComments(video.VideoID)

			// 如果有商品信息，获取商品详情
			if video.ProductInfo != nil && video.ProductInfo.ProductID != "" {
				c.crawlProductInfo(video.ProductInfo.ProductID)
			}

			// 休眠一段时间，避免请求过于频繁
			time.Sleep(time.Second * 2)
		}

		// 如果没有更多数据，或者下一页游标与当前游标相同，则退出循环
		if nextCursor == "" || nextCursor == cursor {
			break
		}

		// 更新游标
		cursor = nextCursor

		// 休眠一段时间，避免请求过于频繁
		time.Sleep(time.Second * 5)
	}
}

// crawlVideoComments 爬取视频评论
func (c *Crawler) crawlVideoComments(videoID string) {
	cursor := ""
	retryCount := 0

	for {
		// 获取视频评论
		comments, nextCursor, err := c.scraper.GetVideoComments(videoID, cursor)
		if err != nil {
			retryCount++
			if retryCount > c.config.Retries {
				log.Printf("获取视频 %s 评论失败，已达到最大重试次数", videoID)
				break
			}
			log.Printf("获取视频 %s 评论失败: %v，正在重试...", videoID, err)
			time.Sleep(time.Duration(c.config.Timeout) * time.Second)
			continue
		}

		// 重置重试计数
		retryCount = 0

		// 保存评论数据
		for _, comment := range comments {
			c.saveCommentData(comment)
		}

		// 如果没有更多数据，或者下一页游标与当前游标相同，则退出循环
		if nextCursor == "" || nextCursor == cursor {
			break
		}

		// 更新游标
		cursor = nextCursor

		// 休眠一段时间，避免请求过于频繁
		time.Sleep(time.Second * 3)
	}
}

// crawlProductInfo 爬取商品信息
func (c *Crawler) crawlProductInfo(productID string) {
	// 获取商品信息
	productInfo, err := c.scraper.GetProductInfo(productID)
	if err != nil {
		log.Printf("获取商品 %s 信息失败: %v", productID, err)
		return
	}

	// 保存商品信息
	c.saveProductInfo(productInfo)
}

// saveUserData 保存用户数据
func (c *Crawler) saveUserData(userData *crawler.UserData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 将用户数据转换为JSON
	jsonData, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		log.Printf("序列化用户数据失败: %v", err)
		return
	}

	// 保存到文件
	filePath := fmt.Sprintf("%s/user_%s.json", c.config.OutputDir, userData.UserID)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("保存用户数据到文件失败: %v", err)
		return
	}

	log.Printf("已保存用户 %s 的数据到 %s", userData.UserID, filePath)
}

// saveVideoData 保存视频数据
func (c *Crawler) saveVideoData(videoData *crawler.VideoData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 将视频数据转换为JSON
	jsonData, err := json.MarshalIndent(videoData, "", "  ")
	if err != nil {
		log.Printf("序列化视频数据失败: %v", err)
		return
	}

	// 保存到文件
	filePath := fmt.Sprintf("%s/video_%s.json", c.config.OutputDir, videoData.VideoID)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("保存视频数据到文件失败: %v", err)
		return
	}

	log.Printf("已保存视频 %s 的数据到 %s", videoData.VideoID, filePath)
}

// saveCommentData 保存评论数据
func (c *Crawler) saveCommentData(commentData *crawler.CommentData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 将评论数据转换为JSON
	jsonData, err := json.MarshalIndent(commentData, "", "  ")
	if err != nil {
		log.Printf("序列化评论数据失败: %v", err)
		return
	}

	// 保存到文件
	filePath := fmt.Sprintf("%s/comment_%s.json", c.config.OutputDir, commentData.CommentID)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("保存评论数据到文件失败: %v", err)
		return
	}

	log.Printf("已保存评论 %s 的数据到 %s", commentData.CommentID, filePath)
}

// saveProductInfo 保存商品信息
func (c *Crawler) saveProductInfo(productInfo *crawler.ProductInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 将商品信息转换为JSON
	jsonData, err := json.MarshalIndent(productInfo, "", "  ")
	if err != nil {
		log.Printf("序列化商品数据失败: %v", err)
		return
	}

	// 保存到文件
	filePath := fmt.Sprintf("%s/product_%s.json", c.config.OutputDir, productInfo.ProductID)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Printf("保存商品数据到文件失败: %v", err)
		return
	}

	log.Printf("已保存商品 %s 的数据到 %s", productInfo.ProductID, filePath)
}

func main() {
	// 解析命令行参数
	platform := flag.String("platform", "douyin", "爬虫平台 (douyin 或 kuaishou)")
	concurrency := flag.Int("concurrency", 5, "并发数")
	timeout := flag.Int("timeout", 30, "超时时间（秒）")
	retries := flag.Int("retries", 3, "重试次数")
	cookies := flag.String("cookies", "", "Cookie字符串")
	outputDir := flag.String("output", "output", "输出目录")
	userIDs := flag.String("users", "", "用户ID列表，以逗号分隔")
	flag.Parse()

	// 检查必要参数
	if *cookies == "" {
		log.Fatal("必须提供Cookie参数")
	}

	if *userIDs == "" {
		log.Fatal("必须提供至少一个用户ID")
	}

	// 初始化爬虫配置
	config := Config{
		Concurrency: *concurrency,
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		Timeout:     *timeout,
		Retries:     *retries,
		Cookies:     *cookies,
		Platform:    *platform,
		OutputDir:   *outputDir,
	}

	// 创建爬虫实例
	crawler := NewCrawler(config)

	// 初始化爬虫
	if err := crawler.Initialize(); err != nil {
		log.Fatalf("爬虫初始化失败: %v", err)
	}

	// 解析用户ID列表
	userIDList := []string{}
	for _, id := range flag.Args() {
		if id != "" {
			userIDList = append(userIDList, id)
		}
	}

	// 如果命令行参数中没有用户ID，则使用-users参数中的用户ID
	if len(userIDList) == 0 && *userIDs != "" {
		userIDList = []string{*userIDs}
	}

	// 启动爬虫
	log.Println("爬虫程序初始化完成")
	log.Println("准备开始数据采集...")
	crawler.Start(userIDList)
}
