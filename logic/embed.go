package logic

import (
	"bluebell/dao/milvus"
	"bluebell/models"
	"bluebell/setting"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"go.uber.org/zap"
)

func EmbedPostAsync(post *models.Post) error {
	if post == nil {
		return errors.New("post is nil")
	}
	if strings.TrimSpace(post.Title) == "" && strings.TrimSpace(post.Content) == "" {
		return errors.New("post title and content are empty")
	}
	if setting.Conf == nil || setting.Conf.EmbeddingConfig == nil {
		return errors.New("embedding config not initialized")
	}
	if strings.TrimSpace(setting.Conf.EmbeddingConfig.APIKey) == "" {
		return errors.New("embedding api key is empty")
	}

	postID := post.ID
	communityID := post.CommunityID
	title := post.Title
	content := post.Content

	go func() {
		if err := embedAndSavePost(postID, communityID, title, content); err != nil {
			zap.L().Error("embedAndSavePost failed",
				zap.Int64("postID", postID),
				zap.Int64("communityID", communityID),
				zap.Error(err))
		}
	}()

	return nil
}

func embedAndSavePost(postID int64, communityID int64, title string, content string) error {
	ctx := context.Background()
	embedder, err := qwenEmbedding(ctx)
	if err != nil {
		return err
	}

	vectors, err := embedder.EmbedStrings(ctx, []string{title + content})
	if err != nil {
		return fmt.Errorf("embed strings failed: %w", err)
	}
	if len(vectors) == 0 {
		return errors.New("embedding returned empty vectors")
	}

	// 2. 核心：类型转换！[][]float64 → [][]float32
	// Milvus 向量字段只接收 []float32，必须手动转换
	var dataRecord *milvus.VectorRecord
	// 循环处理标题、内容两个向量
	for _, vec := range vectors {
		// float64 → float32 转换
		float32Vec := make([]float32, len(vec))
		for i, f64 := range vec {
			float32Vec[i] = float32(f64)
		}

		// 3. 生成内容哈希（用于去重，可选但推荐）
		hash := md5.Sum([]byte(title + content))
		contentHash := hex.EncodeToString(hash[:])

		// 4. 组装Milvus实体数据
		data := &milvus.VectorRecord{
			PostID:         postID,
			CommunityID:    communityID,
			PostCreateTime: time.Now(), // 传入真实创建时间，不要空值
			ModelName:      setting.Conf.EmbeddingConfig.Model,
			ModelVersion:   setting.Conf.EmbeddingConfig.Model,
			ContentHash:    contentHash,
			Embedding:      float32Vec, // 赋值转换后的float32向量
		}
		dataRecord = data
	}

	// 5. 批量插入/更新Milvus（Upsert：存在则更新，不存在则插入）
	// 修复：传入组装好的实体切片 dataRecords
	if err = milvus.UpsertSinglePostVector(ctx, dataRecord); err != nil {
		zap.L().Error("UpsertPostVectors to Milvus failed", zap.Error(err), zap.Int64("post_id", postID))
		return err
	}

	zap.L().Info("Embed and Upsert to Milvus success", zap.Int64("post_id", postID))
	return nil
}

func qwenEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	if setting.Conf == nil || setting.Conf.EmbeddingConfig == nil {
		return nil, errors.New("embedding config not initialized")
	}

	cfg := setting.Conf.EmbeddingConfig
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("embedding api key is empty")
	}

	model := cfg.Model
	apiKey := cfg.APIKey
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	dim := 1024
	embedder, err := dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		Model:      model,
		APIKey:     apiKey,
		Timeout:    timeout,
		Dimensions: &dim,
	})
	if err != nil {
		zap.L().Error("new embedder error: %v\n", zap.Error(err))
		return nil, err
	}
	return embedder, nil
}
