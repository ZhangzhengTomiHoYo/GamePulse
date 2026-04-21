package logic

import (
	"bluebell/dao/pgsql"
	"bluebell/models"
	"bluebell/setting"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"
)

type postAnalysisLLMResult struct {
	SentimentLabel string   `json:"sentiment_label"`
	SentimentScore float64  `json:"sentiment_score"`
	RiskLevel      int32    `json:"risk_level"`
	Topics         []string `json:"topics"`
	Keywords       []string `json:"keywords"`
	Summary        string   `json:"summary"`
}

func createMessages(ctx context.Context, communityName, title, content string) ([]*schema.Message, error) {
	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(`你是一个“二次元游戏社区舆情分析器”。
你的任务是分析用户帖子内容，并输出严格的 JSON，不要输出任何额外文字、解释、markdown 或代码块。分析要求：
1. sentiment_label 只能是 "positive"、"neutral"、"negative" 之一。
2. sentiment_score 范围是 -1.00 到 1.00。
   - 越接近 1 表示越正向
   - 越接近 -1 表示越负向
   - 接近 0 表示中性或情绪不明显
3. risk_level 只能是 0、1、2、3、4 之一：
   - 0: 无风险
   - 1: 普通吐槽/抱怨
   - 2: 引战/对立倾向
   - 3: 公关危机苗头
   - 4: 明显违规/高风险
4. topics 输出 1~5 个主题词数组。
5. keywords 输出 3~8 个关键词数组。
6. summary 输出一句简短摘要，不超过 50 字。
7. 如果信息不足，保持保守判断：
   - sentiment_label 优先输出 "neutral"
   - sentiment_score 靠近 0
   - risk_level 不要随意拔高
8. community_name 仅作为语境参考，不要把社区名本身当作情绪。
9. 只分析文本内容，不要臆造图片内容。

输出 JSON 格式必须严格如下：
{{
  "sentiment_label": "positive|neutral|negative",
  "sentiment_score": 0.00,
  "risk_level": 0,
  "topics": ["..."],
  "keywords": ["..."],
  "summary": "..."
}}`),
		schema.UserMessage("请分析下面这篇帖子：\ncommunity_name: {community_name}\ntitle: {title}\ncontent: {content}"),
	)

	return template.Format(ctx, map[string]any{
		"community_name": communityName,
		"title":          title,
		"content":        content,
	})
}

func openAIForQwen(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	if setting.Conf == nil || setting.Conf.LLMConfig == nil {
		return nil, errors.New("llm config not initialized")
	}

	cfg := setting.Conf.LLMConfig
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("llm api key is empty")
	}

	temperature := float32(0.2)
	config := &openai.ChatModelConfig{
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		BaseURL:     cfg.BaseURL,
		Temperature: &temperature,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}
	cm, err = openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func AnalyzePostAsync(post *models.Post) error {
	if post == nil {
		return errors.New("post is nil")
	}
	if strings.TrimSpace(post.Title) == "" && strings.TrimSpace(post.Content) == "" {
		return errors.New("post title and content are empty")
	}
	if setting.Conf == nil || setting.Conf.LLMConfig == nil {
		return errors.New("llm config not initialized")
	}
	if strings.TrimSpace(setting.Conf.LLMConfig.APIKey) == "" {
		return errors.New("llm api key is empty")
	}

	postID := post.ID
	communityID := post.CommunityID
	title := post.Title
	content := post.Content

	go func() {
		if err := analyzeAndSavePost(postID, communityID, title, content); err != nil {
			zap.L().Error("analyzeAndSavePost failed",
				zap.Int64("postID", postID),
				zap.Int64("communityID", communityID),
				zap.Error(err))
			return
		}
		zap.L().Info("LLM Sentiment Analyze Success!",
			zap.Int64("post_id", postID),
		)
	}()
	zap.L().Info("LLM analyze task submitted to async goroutine",
		zap.Int64("post_id", postID),
	)
	return nil
}

func analyzeAndSavePost(postID, communityID int64, title, content string) error {
	timeout := llmTimeout()
	setupCtx := context.Background()

	community, err := pgsql.GetCommunityDetailByID(communityID)
	if err != nil {
		return fmt.Errorf("get community detail failed: %w", err)
	}

	messages, err := createMessages(setupCtx, community.Name, title, content)
	if err != nil {
		return fmt.Errorf("create llm messages failed: %w", err)
	}

	cm, err := openAIForQwen(setupCtx)
	if err != nil {
		return fmt.Errorf("init llm failed: %w", err)
	}

	generateCtx, cancel := context.WithTimeout(setupCtx, timeout)
	defer cancel()

	resp, err := cm.Generate(generateCtx, messages)
	if err != nil {
		zap.L().Error("大模型API调用失败!")
		return fmt.Errorf("llm generate failed: %w", err)
	}
	if resp == nil {
		return errors.New("llm generate returned nil response")
	}
	if strings.TrimSpace(resp.Content) == "" {
		return errors.New("llm returned empty content")
	}

	result := new(postAnalysisLLMResult)
	if err := json.Unmarshal([]byte(resp.Content), result); err != nil {
		return fmt.Errorf("unmarshal llm result failed: %w", err)
	}

	analysis, err := buildPostAnalysis(postID, result)
	if err != nil {
		return err
	}

	if err := pgsql.UpsertPostAnalysis(analysis); err != nil {
		return fmt.Errorf("save post analysis failed: %w", err)
	}

	return nil
}

func buildPostAnalysis(postID int64, result *postAnalysisLLMResult) (*models.PostAnalysis, error) {
	if result == nil {
		return nil, errors.New("llm result is nil")
	}

	sentimentLabel := normalizeSentimentLabel(result.SentimentLabel)
	score := clampSentimentScore(result.SentimentScore)
	riskLevel := normalizeRiskLevel(result.RiskLevel)
	topics := sanitizeStringList(result.Topics, 5)
	keywords := sanitizeStringList(result.Keywords, 8)
	summary := strings.TrimSpace(result.Summary)

	topicsJSON, err := json.Marshal(topics)
	if err != nil {
		return nil, fmt.Errorf("marshal topics failed: %w", err)
	}
	keywordsJSON, err := json.Marshal(keywords)
	if err != nil {
		return nil, fmt.Errorf("marshal keywords failed: %w", err)
	}

	analysis := &models.PostAnalysis{
		PostID:         postID,
		SentimentLabel: sentimentLabel,
		SentimentScore: sql.NullFloat64{Float64: score, Valid: true},
		RiskLevel:      riskLevel,
		Topics:         topicsJSON,
		Keywords:       keywordsJSON,
		Summary:        sql.NullString{String: summary, Valid: summary != ""},
	}
	return analysis, nil
}

func llmTimeout() time.Duration {
	if setting.Conf == nil || setting.Conf.LLMConfig == nil || setting.Conf.LLMConfig.TimeoutSeconds <= 0 {
		return 30 * time.Second
	}
	return time.Duration(setting.Conf.LLMConfig.TimeoutSeconds) * time.Second
}

func normalizeSentimentLabel(label string) string {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "positive":
		return "positive"
	case "negative":
		return "negative"
	default:
		return "neutral"
	}
}

func normalizeRiskLevel(level int32) int32 {
	if level < 0 {
		return 0
	}
	if level > 4 {
		return 4
	}
	return level
}

func clampSentimentScore(score float64) float64 {
	if score > 1 {
		return 1
	}
	if score < -1 {
		return -1
	}
	return score
}

func sanitizeStringList(items []string, limit int) []string {
	if limit <= 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, min(len(items), limit))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
		if len(result) == limit {
			break
		}
	}

	if len(result) == 0 {
		return []string{}
	}
	return result
}
