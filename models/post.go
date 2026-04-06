package models

import (
	"time"
)

// ParamCreatePost 接收创建帖子请求的参数 (专门用于接收前端 JSON)
type ParamCreatePost struct {
	Title       string   `json:"title" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	CommunityID int64    `json:"community_id,string" binding:"required"`
	ImageURLs   []string `json:"image_urls"` // 新增：接收前端传来的图片链接数组
}

// Post 数据库模型 记住内存对齐概念
type Post struct {
	ID          int64     `json:"id,string" db:"post_id"`
	AuthorID    int64     `json:"author_id,string" db:"author_id"`
	CommunityID int64     `json:"community_id,string" db:"community_id" binding:"required"`
	Status      int32     `json:"status" db:"status"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Content     string    `json:"content" db:"content" binding:"required"`
	ImageURL    string    `json:"image_url" db:"image_url"` // 新增：对应 PG 里的 VARCHAR(1024)，存 JSON 字符串
	CreateTime  time.Time `json:"create_time" db:"create_time"`
}

// ApiPostDetail 帖子详情接口的结构体 (方便前端渲染)
type ApiPostDetail struct {
	AuthorName       string   `json:"author_name"`
	VoteNum          int64    `json:"votes"`
	ImageURLs        []string `json:"image_urls"` // 新增：给前端直接返回反序列化好的数组，方便 v-for 渲染图片
	*Post                                         // 嵌入帖子结构体
	*CommunityDetail `json:"community"`           // 嵌入社区信息
}
