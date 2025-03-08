package checkpoint

import (
	"Crawler/utils/logger"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Manager 断点管理器
type Manager struct {
	enabled  bool
	interval int
	file     string
	data     *Data
	mutex    sync.Mutex
	lastSave time.Time
}

// Data 断点数据
type Data struct {
	Platform        string            `json:"platform"`
	StartTime       time.Time         `json:"start_time"`
	LastUpdateTime  time.Time         `json:"last_update_time"`
	ProcessedUsers  map[string]bool   `json:"processed_users"`
	ProcessedVideos map[string]bool   `json:"processed_videos"`
	UserCursors     map[string]string `json:"user_cursors"`
	CommentCursors  map[string]string `json:"comment_cursors"`
	Stats           Stats             `json:"stats"`
}

// Stats 统计数据
type Stats struct {
	TotalUsers    int `json:"total_users"`
	TotalVideos   int `json:"total_videos"`
	TotalComments int `json:"total_comments"`
	TotalProducts int `json:"total_products"`
}

// Config 断点配置
type Config struct {
	Enabled  bool   `json:"enabled"`
	Interval int    `json:"interval"`
	File     string `json:"file"`
}

// NewManager 创建断点管理器
func NewManager(config Config) *Manager {
	return &Manager{
		enabled:  config.Enabled,
		interval: config.Interval,
		file:     config.File,
		data: &Data{
			StartTime:       time.Now(),
			LastUpdateTime:  time.Now(),
			ProcessedUsers:  make(map[string]bool),
			ProcessedVideos: make(map[string]bool),
			UserCursors:     make(map[string]string),
			CommentCursors:  make(map[string]string),
		},
		lastSave: time.Now(),
	}
}

// Load 加载断点数据
func (m *Manager) Load() error {
	if !m.enabled {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(m.file); os.IsNotExist(err) {
		logger.Info("断点文件不存在，将创建新的断点")
		return nil
	}

	// 读取文件内容
	data, err := os.ReadFile(m.file)
	if err != nil {
		return fmt.Errorf("读取断点文件失败: %v", err)
	}

	// 解析JSON
	var checkpointData Data
	if err := json.Unmarshal(data, &checkpointData); err != nil {
		return fmt.Errorf("解析断点数据失败: %v", err)
	}

	m.data = &checkpointData
	logger.Info("成功加载断点数据，上次更新时间: %v", m.data.LastUpdateTime)

	return nil
}

// Save 保存断点数据
func (m *Manager) Save() error {
	if !m.enabled {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 更新时间
	m.data.LastUpdateTime = time.Now()

	// 序列化数据
	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化断点数据失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(m.file, data, 0644); err != nil {
		return fmt.Errorf("写入断点文件失败: %v", err)
	}

	m.lastSave = time.Now()
	logger.Debug("已保存断点数据")

	return nil
}

// CheckSave 检查是否需要保存断点
func (m *Manager) CheckSave() error {
	if !m.enabled || m.interval <= 0 {
		return nil
	}

	if time.Since(m.lastSave).Seconds() >= float64(m.interval) {
		return m.Save()
	}

	return nil
}

// SetPlatform 设置平台
func (m *Manager) SetPlatform(platform string) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.Platform = platform
}

// MarkUserProcessed 标记用户已处理
func (m *Manager) MarkUserProcessed(userID string) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.ProcessedUsers[userID] = true
	m.data.Stats.TotalUsers++

	m.CheckSave()
}

// IsUserProcessed 检查用户是否已处理
func (m *Manager) IsUserProcessed(userID string) bool {
	if !m.enabled {
		return false
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.data.ProcessedUsers[userID]
}

// MarkVideoProcessed 标记视频已处理
func (m *Manager) MarkVideoProcessed(videoID string) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.ProcessedVideos[videoID] = true
	m.data.Stats.TotalVideos++

	m.CheckSave()
}

// IsVideoProcessed 检查视频是否已处理
func (m *Manager) IsVideoProcessed(videoID string) bool {
	if !m.enabled {
		return false
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.data.ProcessedVideos[videoID]
}

// SetUserCursor 设置用户游标
func (m *Manager) SetUserCursor(userID string, cursor string) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.UserCursors[userID] = cursor
}

// GetUserCursor 获取用户游标
func (m *Manager) GetUserCursor(userID string) string {
	if !m.enabled {
		return ""
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.data.UserCursors[userID]
}

// SetCommentCursor 设置评论游标
func (m *Manager) SetCommentCursor(videoID string, cursor string) {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.CommentCursors[videoID] = cursor
}

// GetCommentCursor 获取评论游标
func (m *Manager) GetCommentCursor(videoID string) string {
	if !m.enabled {
		return ""
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.data.CommentCursors[videoID]
}

// IncrementCommentCount 增加评论计数
func (m *Manager) IncrementCommentCount() {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.Stats.TotalComments++
}

// IncrementProductCount 增加商品计数
func (m *Manager) IncrementProductCount() {
	if !m.enabled {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data.Stats.TotalProducts++
}

// GetStats 获取统计数据
func (m *Manager) GetStats() Stats {
	if !m.enabled {
		return Stats{}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.data.Stats
}
