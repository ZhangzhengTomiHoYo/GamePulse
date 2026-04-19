package models

import (
	"database/sql"
	"encoding/json"
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
	ImageURL    string    `json:"image_url" db:"image_url"` // 当前业务里存的是图片 URL 数组的 JSON 字符串
	CreateTime  time.Time `json:"create_time" db:"create_time"`
	UpdateTime  time.Time `json:"update_time" db:"update_time"`
}

// ApiPostDetail 帖子详情接口的结构体 (方便前端渲染)
type ApiPostDetail struct {
	AuthorName       string             `json:"author_name"`
	VoteNum          int64              `json:"votes"`
	ImageURLs        []string           `json:"image_urls"` // 新增：给前端直接返回反序列化好的数组，方便 v-for 渲染图片
	SentimentLabel   string             `json:"sentiment_label"`
	*Post                               // 嵌入帖子结构体
	*CommunityDetail `json:"community"` // 嵌入社区信息
}

// PostAnalysis 对应 post_analysis 表，是帖子主表的 1:1 派生分析结果。
type PostAnalysis struct {
	PostID         int64           `json:"post_id,string" db:"post_id"`
	SentimentLabel string          `json:"sentiment_label" db:"sentiment_label"`
	SentimentScore sql.NullFloat64 `json:"-" db:"sentiment_score"`
	RiskLevel      int32           `json:"risk_level" db:"risk_level"`
	Topics         json.RawMessage `json:"topics,omitempty" db:"topics"`
	Keywords       json.RawMessage `json:"keywords,omitempty" db:"keywords"`
	Summary        sql.NullString  `json:"-" db:"summary"`
	AnalyzedAt     time.Time       `json:"analyzed_at" db:"analyzed_at"`
}

// PostEmbedding 对应 post_embeddings 表。
// 目前 embedding 先按数据库返回的原始文本格式保存，后续接入专用 pgvector codec 再细化。
type PostEmbedding struct {
	ID             int64          `json:"id" db:"id"`
	PostID         int64          `json:"post_id,string" db:"post_id"`
	ChunkIndex     int32          `json:"chunk_index" db:"chunk_index"`
	ChunkText      sql.NullString `json:"-" db:"chunk_text"`
	CommunityID    int64          `json:"community_id,string" db:"community_id"`
	PostCreateTime time.Time      `json:"post_create_time" db:"post_create_time"`
	ModelName      string         `json:"model_name" db:"model_name"`
	ModelVersion   string         `json:"model_version" db:"model_version"`
	ContentHash    string         `json:"content_hash" db:"content_hash"`
	Embedding      sql.NullString `json:"-" db:"embedding"`
	Status         string         `json:"status" db:"status"`
	ErrorMsg       sql.NullString `json:"-" db:"error_msg"`
	CreateTime     time.Time      `json:"create_time" db:"create_time"`
	UpdateTime     time.Time      `json:"update_time" db:"update_time"`
}
