package storage

import (
	"Crawler/crawler"
	"Crawler/utils/logger"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Manager 存储管理器
type Manager struct {
	enabled  bool
	dbType   string
	db       *sql.DB
	prepared map[string]*sql.Stmt
}

// Config 存储配置
type Config struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// NewManager 创建存储管理器
func NewManager(config Config) (*Manager, error) {
	if !config.Enabled {
		return &Manager{enabled: false}, nil
	}

	manager := &Manager{
		enabled:  true,
		dbType:   config.Type,
		prepared: make(map[string]*sql.Stmt),
	}

	// 连接数据库
	var dsn string
	switch config.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
			config.User, config.Password, config.Host, config.Port, config.Database)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	var err error
	manager.db, err = sql.Open(config.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %v", err)
	}

	// 测试连接
	if err := manager.db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 初始化数据库
	if err := manager.initDatabase(); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %v", err)
	}

	// 准备SQL语句
	if err := manager.prepareStatements(); err != nil {
		return nil, fmt.Errorf("准备SQL语句失败: %v", err)
	}

	logger.Info("数据库连接成功: %s", config.Type)
	return manager, nil
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 关闭预处理语句
	for _, stmt := range m.prepared {
		stmt.Close()
	}

	// 关闭数据库连接
	return m.db.Close()
}

// initDatabase 初始化数据库
func (m *Manager) initDatabase() error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 创建用户表
	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(64) PRIMARY KEY,
			nickname VARCHAR(255) NOT NULL,
			followers INT NOT NULL,
			following INT NOT NULL,
			description TEXT,
			tags TEXT,
			platform VARCHAR(32) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建视频表
	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS videos (
			video_id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NOT NULL,
			title VARCHAR(255),
			description TEXT,
			likes INT NOT NULL,
			comments INT NOT NULL,
			shares INT NOT NULL,
			tags TEXT,
			platform VARCHAR(32) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX (user_id)
		)
	`)
	if err != nil {
		return err
	}

	// 创建评论表
	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			comment_id VARCHAR(64) PRIMARY KEY,
			video_id VARCHAR(64) NOT NULL,
			user_id VARCHAR(64) NOT NULL,
			content TEXT NOT NULL,
			likes INT NOT NULL,
			replies INT NOT NULL,
			timestamp BIGINT NOT NULL,
			platform VARCHAR(32) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX (video_id),
			INDEX (user_id)
		)
	`)
	if err != nil {
		return err
	}

	// 创建商品表
	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			product_id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			category VARCHAR(128),
			description TEXT,
			sales INT NOT NULL,
			platform VARCHAR(32) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// 创建视频商品关联表
	_, err = m.db.Exec(`
		CREATE TABLE IF NOT EXISTS video_products (
			video_id VARCHAR(64) NOT NULL,
			product_id VARCHAR(64) NOT NULL,
			platform VARCHAR(32) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (video_id, product_id),
			INDEX (product_id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// prepareStatements 准备SQL语句
func (m *Manager) prepareStatements() error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 插入用户
	insertUser, err := m.db.Prepare(`
		INSERT INTO users (user_id, nickname, followers, following, description, tags, platform)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			nickname = VALUES(nickname),
			followers = VALUES(followers),
			following = VALUES(following),
			description = VALUES(description),
			tags = VALUES(tags),
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	m.prepared["insertUser"] = insertUser

	// 插入视频
	insertVideo, err := m.db.Prepare(`
		INSERT INTO videos (video_id, user_id, title, description, likes, comments, shares, tags, platform)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			description = VALUES(description),
			likes = VALUES(likes),
			comments = VALUES(comments),
			shares = VALUES(shares),
			tags = VALUES(tags),
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	m.prepared["insertVideo"] = insertVideo

	// 插入评论
	insertComment, err := m.db.Prepare(`
		INSERT INTO comments (comment_id, video_id, user_id, content, likes, replies, timestamp, platform)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			content = VALUES(content),
			likes = VALUES(likes),
			replies = VALUES(replies),
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	m.prepared["insertComment"] = insertComment

	// 插入商品
	insertProduct, err := m.db.Prepare(`
		INSERT INTO products (product_id, name, price, category, description, sales, platform)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			price = VALUES(price),
			category = VALUES(category),
			description = VALUES(description),
			sales = VALUES(sales),
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	m.prepared["insertProduct"] = insertProduct

	// 插入视频商品关联
	insertVideoProduct, err := m.db.Prepare(`
		INSERT INTO video_products (video_id, product_id, platform)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			created_at = created_at
	`)
	if err != nil {
		return err
	}
	m.prepared["insertVideoProduct"] = insertVideoProduct

	return nil
}

// SaveUser 保存用户数据
func (m *Manager) SaveUser(userData *crawler.UserData, platform string) error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 将标签数组转换为字符串
	tags := ""
	if len(userData.Tags) > 0 {
		for i, tag := range userData.Tags {
			if i > 0 {
				tags += ","
			}
			tags += tag
		}
	}

	// 执行插入
	_, err := m.prepared["insertUser"].Exec(
		userData.UserID,
		userData.Nickname,
		userData.Followers,
		userData.Following,
		userData.Description,
		tags,
		platform,
	)
	if err != nil {
		return fmt.Errorf("保存用户数据到数据库失败: %v", err)
	}

	logger.Debug("已保存用户 %s 的数据到数据库", userData.UserID)
	return nil
}

// SaveVideo 保存视频数据
func (m *Manager) SaveVideo(videoData *crawler.VideoData, platform string) error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 将标签数组转换为字符串
	tags := ""
	if len(videoData.Tags) > 0 {
		for i, tag := range videoData.Tags {
			if i > 0 {
				tags += ","
			}
			tags += tag
		}
	}

	// 执行插入
	_, err := m.prepared["insertVideo"].Exec(
		videoData.VideoID,
		videoData.UserID,
		videoData.Title,
		videoData.Description,
		videoData.Likes,
		videoData.Comments,
		videoData.Shares,
		tags,
		platform,
	)
	if err != nil {
		return fmt.Errorf("保存视频数据到数据库失败: %v", err)
	}

	// 如果有商品信息，保存商品关联
	if videoData.ProductInfo != nil && videoData.ProductInfo.ProductID != "" {
		// 保存商品信息
		if err := m.SaveProduct(videoData.ProductInfo, platform); err != nil {
			return err
		}

		// 保存视频商品关联
		_, err := m.prepared["insertVideoProduct"].Exec(
			videoData.VideoID,
			videoData.ProductInfo.ProductID,
			platform,
		)
		if err != nil {
			return fmt.Errorf("保存视频商品关联到数据库失败: %v", err)
		}
	}

	logger.Debug("已保存视频 %s 的数据到数据库", videoData.VideoID)
	return nil
}

// SaveComment 保存评论数据
func (m *Manager) SaveComment(commentData *crawler.CommentData, platform string) error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 执行插入
	_, err := m.prepared["insertComment"].Exec(
		commentData.CommentID,
		commentData.VideoID,
		commentData.UserID,
		commentData.Content,
		commentData.Likes,
		commentData.Replies,
		commentData.Timestamp,
		platform,
	)
	if err != nil {
		return fmt.Errorf("保存评论数据到数据库失败: %v", err)
	}

	logger.Debug("已保存评论 %s 的数据到数据库", commentData.CommentID)
	return nil
}

// SaveProduct 保存商品数据
func (m *Manager) SaveProduct(productInfo *crawler.ProductInfo, platform string) error {
	if !m.enabled || m.db == nil {
		return nil
	}

	// 执行插入
	_, err := m.prepared["insertProduct"].Exec(
		productInfo.ProductID,
		productInfo.Name,
		productInfo.Price,
		productInfo.Category,
		productInfo.Description,
		productInfo.Sales,
		platform,
	)
	if err != nil {
		return fmt.Errorf("保存商品数据到数据库失败: %v", err)
	}

	logger.Debug("已保存商品 %s 的数据到数据库", productInfo.ProductID)
	return nil
}
