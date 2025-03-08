package crawler

// Platform 定义支持的平台类型
type Platform string

const (
	Douyin  Platform = "douyin"
	Kuaishou Platform = "kuaishou"
)

// UserData 用户数据结构
type UserData struct {
	UserID      string   `json:"user_id"`
	Nickname    string   `json:"nickname"`
	Followers   int      `json:"followers"`
	Following   int      `json:"following"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// VideoData 视频数据结构
type VideoData struct {
	VideoID     string   `json:"video_id"`
	UserID      string   `json:"user_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Likes       int      `json:"likes"`
	Comments    int      `json:"comments"`
	Shares      int      `json:"shares"`
	Tags        []string `json:"tags"`
	ProductInfo *ProductInfo `json:"product_info,omitempty"`
}

// ProductInfo 商品信息结构
type ProductInfo struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Sales       int     `json:"sales"`
}

// CommentData 评论数据结构
type CommentData struct {
	CommentID string `json:"comment_id"`
	VideoID   string `json:"video_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Likes     int    `json:"likes"`
	Replies   int    `json:"replies"`
	Timestamp int64  `json:"timestamp"`
}

// Scraper 爬虫接口定义
type Scraper interface {
	// Initialize 初始化爬虫
	Initialize() error
	
	// GetUserInfo 获取用户信息
	GetUserInfo(userID string) (*UserData, error)
	
	// GetUserVideos 获取用户视频列表
	GetUserVideos(userID string, cursor string) ([]*VideoData, string, error)
	
	// GetVideoComments 获取视频评论
	GetVideoComments(videoID string, cursor string) ([]*CommentData, string, error)
	
	// GetProductInfo 获取商品信息
	GetProductInfo(productID string) (*ProductInfo, error)
}