package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// ChatStreamRequest 处理后的前端请求
type ChatStreamRequest struct {
	ConversationID  int64  `json:"conversation_id,string"`
	ParentMessageID int64  `json:"parent_message_id,string"`
	Query           string `json:"query"`
}

// ChatPostBackend 智能体 查询后端帖子 调用工具所用到的参数
type ChatPostBackend struct {
	CommunityName  string `json:"community_name"`
	SentimentLabel string `json:"sentiment_label"`
	LastTime       int    `json:"last_time"`
}

// ChatPostBackend 智能体 查询后端帖子 后得到的结果 需要返回给大模型进行分析，由于是两张表的结果 因此需要新建数据模型
type ChatPostBackendResult struct {
	Title          string          `json:"title" db:"title" binding:"required"`
	Content        string          `json:"content" db:"content" binding:"required"`
	SentimentLabel string          `json:"sentiment_label" db:"sentiment_label"`
	SentimentScore sql.NullFloat64 `json:"-" db:"sentiment_score"`
	RiskLevel      int32           `json:"risk_level" db:"risk_level"`
	Topics         json.RawMessage `json:"topics,omitempty" db:"topics"`
	Keywords       json.RawMessage `json:"keywords,omitempty" db:"keywords"`
	Summary        sql.NullString  `json:"-" db:"summary"`
}

// Conversation 一次对话的元数据 不包含具体的消息
type Conversation struct {
	ConversationID int64     `json:"conversation_id"`
	UserID         int64     `json:"user_id"`
	AgentCode      string    `json:"agent_code"`
	Title          string    `json:"title"`
	MessageCount   int       `json:"message_count"`
	LastMessageAt  time.Time `json:"last_message_at"`
	Status         int       `json:"status"`
	CreateTime     time.Time `json:"create_time"`
	UpdateTime     time.Time `json:"update_time"`
}

// Message 一条消息的所有数据
type Message struct {
	MessageID      int64     `json:"message_id"`
	ConversationID int64     `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	Seq            int64     `json:"seq"`
	MetaJSON       []byte    `json:"meta_json"`
	CreateTime     time.Time `json:"create_time"`
}
