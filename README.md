<<<<<<< HEAD
# Agricultural_crawler
农产品爬虫项目
=======
# 社交媒体爬虫

这是一个用Go语言编写的社交媒体爬虫项目，目前支持抖音和快手平台的数据采集。

## 功能特点

- 支持多平台：抖音、快手
- 并发采集：可配置并发数量
- 数据类型：用户信息、视频列表、视频评论、商品信息
- 自动重试：遇到错误时自动重试
- 数据持久化：将采集到的数据保存为JSON文件

## 安装

确保已安装Go 1.22或更高版本，然后克隆本仓库：

```bash
git clone https://github.com/yourusername/Crawler.git
cd Crawler
```

安装依赖：

```bash
go mod tidy
```

## 使用方法

### 编译

```bash
go build -o crawler
```

### 运行

```bash
./crawler -platform=douyin -cookies="your_cookies" -users="user_id1,user_id2"
```

### 命令行参数

- `-platform`: 爬虫平台，可选值：`douyin`（抖音）或 `kuaishou`（快手），默认为 `douyin`
- `-concurrency`: 并发数，默认为 5
- `-timeout`: 超时时间（秒），默认为 30
- `-retries`: 重试次数，默认为 3
- `-cookies`: Cookie字符串，**必填**
- `-output`: 输出目录，默认为 `output`
- `-users`: 用户ID列表，以逗号分隔，**必填**

### 示例

抖音平台：

```bash
./crawler -platform=douyin -cookies="your_cookies" -users="123456789"
```

快手平台：

```bash
./crawler -platform=kuaishou -cookies="your_cookies" -users="987654321"
```

## 数据输出

所有数据将保存在指定的输出目录中（默认为 `output`），格式为JSON文件：

- 用户数据：`user_{user_id}.json`
- 视频数据：`video_{video_id}.json`
- 评论数据：`comment_{comment_id}.json`
- 商品数据：`product_{product_id}.json`

## 注意事项

1. 需要提供有效的Cookie才能正常采集数据
2. 请遵守相关平台的使用条款和政策
3. 过于频繁的请求可能会导致IP被封禁
4. 建议适当调整并发数和超时时间，避免请求过于频繁

## 许可证

MIT 
>>>>>>> 5cb46aa (Initial commit - create the project)
