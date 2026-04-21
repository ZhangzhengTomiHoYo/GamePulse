package milvus

import (
	"bluebell/setting"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// VectorRecord 表示一条待写入 Milvus 的帖子向量记录。
type VectorRecord struct {
	PostID         int64
	CommunityID    int64
	PostCreateTime time.Time
	ModelName      string
	ModelVersion   string
	ContentHash    string
	Embedding      []float32
}

// SearchRequest 定义一次向量相似检索所需的输入参数。
type SearchRequest struct {
	Vector      []float32
	TopK        int
	CommunityID int64
	PostID      int64
	Filter      string
}

// SearchHit 表示一次向量检索返回的单条命中结果。
type SearchHit struct {
	PostID       int64
	Score        float32
	CommunityID  int64
	PostCreateTS int64
	ModelName    string
	ModelVersion string
	ContentHash  string
}

// UpsertPostVectors 批量写入或覆盖帖子向量记录。
func UpsertPostVectors(ctx context.Context, records []*VectorRecord) error {
	if len(records) == 0 {
		return nil
	}
	if cli == nil {
		return errors.New("milvus client is not initialized")
	}
	if setting.Conf == nil || setting.Conf.MilvusConfig == nil {
		return errors.New("milvus config is not initialized")
	}

	cfg := setting.Conf.MilvusConfig
	if cfg.Dimension <= 0 {
		return errors.New("milvus dimension must be greater than 0")
	}

	ctx, cancel := withTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	postIDs := make([]int64, 0, len(records))
	communityIDs := make([]int64, 0, len(records))
	postCreateTS := make([]int64, 0, len(records))
	modelNames := make([]string, 0, len(records))
	modelVersions := make([]string, 0, len(records))
	contentHashes := make([]string, 0, len(records))
	embeddings := make([][]float32, 0, len(records))

	for i, record := range records {
		if record == nil {
			return fmt.Errorf("vector record at index %d is nil", i)
		}
		if record.PostID <= 0 {
			return fmt.Errorf("vector record at index %d has invalid post_id", i)
		}
		if len(record.Embedding) != cfg.Dimension {
			return fmt.Errorf("vector record at index %d has embedding dim %d, want %d", i, len(record.Embedding), cfg.Dimension)
		}

		modelName := strings.TrimSpace(record.ModelName)
		if modelName == "" {
			return fmt.Errorf("vector record at index %d has empty model_name", i)
		}

		createTime := record.PostCreateTime
		if createTime.IsZero() {
			createTime = time.Now()
		}

		postIDs = append(postIDs, record.PostID)
		communityIDs = append(communityIDs, record.CommunityID)
		postCreateTS = append(postCreateTS, createTime.UnixMilli())
		modelNames = append(modelNames, modelName)
		modelVersions = append(modelVersions, strings.TrimSpace(record.ModelVersion))
		contentHashes = append(contentHashes, strings.TrimSpace(record.ContentHash))
		embeddings = append(embeddings, record.Embedding)
	}

	_, err := cli.Upsert(ctx, milvusclient.NewColumnBasedInsertOption(normalizedCollectionName(cfg)).
		WithInt64Column(fieldPostID, postIDs).
		WithInt64Column(fieldCommunityID, communityIDs).
		WithInt64Column(fieldPostCreateTS, postCreateTS).
		WithVarcharColumn(fieldModelName, modelNames).
		WithVarcharColumn(fieldModelVersion, modelVersions).
		WithVarcharColumn(fieldContentHash, contentHashes).
		WithFloatVectorColumn(fieldEmbedding, cfg.Dimension, embeddings),
	)
	if err != nil {
		return fmt.Errorf("upsert milvus vectors failed: %w", err)
	}

	return nil
}

// UpsertSinglePostVector 单条写入或覆盖帖子向量记录
func UpsertSinglePostVector(ctx context.Context, record *VectorRecord) error {
	// 1. 基础校验（和批量版完全一致）
	if record == nil {
		return errors.New("vector record is nil")
	}
	if cli == nil {
		return errors.New("milvus client is not initialized")
	}
	if setting.Conf == nil || setting.Conf.MilvusConfig == nil {
		return errors.New("milvus config is not initialized")
	}

	cfg := setting.Conf.MilvusConfig
	if cfg.Dimension <= 0 {
		return errors.New("milvus dimension must be greater than 0")
	}

	// 2. 单条记录合法性校验（移除循环，直接校验）
	if record.PostID <= 0 {
		return fmt.Errorf("vector record has invalid post_id: %d", record.PostID)
	}
	if len(record.Embedding) != cfg.Dimension {
		return fmt.Errorf("vector record embedding dim %d, want %d", len(record.Embedding), cfg.Dimension)
	}

	modelName := strings.TrimSpace(record.ModelName)
	if modelName == "" {
		return errors.New("vector record has empty model_name")
	}

	// 时间零值兜底
	createTime := record.PostCreateTime
	if createTime.IsZero() {
		createTime = time.Now()
	}

	// 3. 超时控制（复用原有逻辑）
	ctx, cancel := withTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	// 4. 构造单列数据（Milvus 列插入必须传切片，单条就是长度1的切片）
	postIDs := []int64{record.PostID}
	communityIDs := []int64{record.CommunityID}
	postCreateTS := []int64{createTime.UnixMilli()}
	modelNames := []string{modelName}
	modelVersions := []string{strings.TrimSpace(record.ModelVersion)}
	contentHashes := []string{strings.TrimSpace(record.ContentHash)}
	embeddings := [][]float32{record.Embedding}

	// 5. 执行 Upsert（和批量版逻辑完全一致）
	_, err := cli.Upsert(ctx, milvusclient.NewColumnBasedInsertOption(normalizedCollectionName(cfg)).
		WithInt64Column(fieldPostID, postIDs).
		WithInt64Column(fieldCommunityID, communityIDs).
		WithInt64Column(fieldPostCreateTS, postCreateTS).
		WithVarcharColumn(fieldModelName, modelNames).
		WithVarcharColumn(fieldModelVersion, modelVersions).
		WithVarcharColumn(fieldContentHash, contentHashes).
		WithFloatVectorColumn(fieldEmbedding, cfg.Dimension, embeddings),
	)
	if err != nil {
		return fmt.Errorf("upsert single milvus vector failed: %w", err)
	}

	return nil
}

// DeletePostVectors 按帖子 ID 删除其在 Milvus 中的向量记录。
func DeletePostVectors(ctx context.Context, postID int64) error {
	if postID <= 0 {
		return errors.New("postID must be greater than 0")
	}
	if cli == nil {
		return errors.New("milvus client is not initialized")
	}
	if setting.Conf == nil || setting.Conf.MilvusConfig == nil {
		return errors.New("milvus config is not initialized")
	}

	ctx, cancel := withTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	_, err := cli.Delete(ctx, milvusclient.NewDeleteOption(normalizedCollectionName(setting.Conf.MilvusConfig)).
		WithExpr(fmt.Sprintf("%s == %d", fieldPostID, postID)),
	)
	if err != nil {
		return fmt.Errorf("delete milvus vectors by post_id failed: %w", err)
	}

	return nil
}

// SearchSimilar 执行向量相似检索，并返回命中结果及元数据。
func SearchSimilar(ctx context.Context, req *SearchRequest) ([]SearchHit, error) {
	if req == nil {
		return nil, errors.New("search request is nil")
	}
	if cli == nil {
		return nil, errors.New("milvus client is not initialized")
	}
	if setting.Conf == nil || setting.Conf.MilvusConfig == nil {
		return nil, errors.New("milvus config is not initialized")
	}

	cfg := setting.Conf.MilvusConfig
	if cfg.Dimension <= 0 {
		return nil, errors.New("milvus dimension must be greater than 0")
	}
	if len(req.Vector) != cfg.Dimension {
		return nil, fmt.Errorf("search vector dim %d, want %d", len(req.Vector), cfg.Dimension)
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	ctx, cancel := withTimeout(ctx, defaultRequestTimeout)
	defer cancel()

	option := milvusclient.NewSearchOption(
		normalizedCollectionName(cfg),
		topK,
		[]entity.Vector{entity.FloatVector(req.Vector)},
	).
		WithANNSField(fieldEmbedding).
		WithOutputFields(
			fieldCommunityID,
			fieldPostCreateTS,
			fieldModelName,
			fieldModelVersion,
			fieldContentHash,
		)

	if expr := buildFilter(req); expr != "" {
		option = option.WithFilter(expr)
	}

	if annParam := buildAnnParam(cfg.IndexType, topK); annParam != nil {
		option = option.WithAnnParam(annParam)
	}

	resultSets, err := cli.Search(ctx, option)
	if err != nil {
		return nil, fmt.Errorf("milvus search failed: %w", err)
	}
	if len(resultSets) == 0 {
		return []SearchHit{}, nil
	}

	return buildHits(resultSets[0])
}

// buildHits 将 Milvus 检索结果集转换为项目内部使用的命中结构。
func buildHits(resultSet milvusclient.ResultSet) ([]SearchHit, error) {
	hits := make([]SearchHit, 0, resultSet.Len())

	communityIDs := resultSet.GetColumn(fieldCommunityID)
	postCreateTS := resultSet.GetColumn(fieldPostCreateTS)
	modelNames := resultSet.GetColumn(fieldModelName)
	modelVersions := resultSet.GetColumn(fieldModelVersion)
	contentHashes := resultSet.GetColumn(fieldContentHash)

	for i := 0; i < resultSet.Len(); i++ {
		postID, err := resultSet.IDs.GetAsInt64(i)
		if err != nil {
			return nil, fmt.Errorf("get search hit post_id failed: %w", err)
		}

		communityID, err := communityIDs.GetAsInt64(i)
		if err != nil {
			return nil, fmt.Errorf("get search hit community_id failed: %w", err)
		}

		createTS, err := postCreateTS.GetAsInt64(i)
		if err != nil {
			return nil, fmt.Errorf("get search hit post_create_ts failed: %w", err)
		}

		modelName, err := modelNames.GetAsString(i)
		if err != nil {
			return nil, fmt.Errorf("get search hit model_name failed: %w", err)
		}

		modelVersion, err := modelVersions.GetAsString(i)
		if err != nil {
			return nil, fmt.Errorf("get search hit model_version failed: %w", err)
		}

		contentHash, err := contentHashes.GetAsString(i)
		if err != nil {
			return nil, fmt.Errorf("get search hit content_hash failed: %w", err)
		}

		score := float32(0)
		if i < len(resultSet.Scores) {
			score = resultSet.Scores[i]
		}

		hits = append(hits, SearchHit{
			PostID:       postID,
			Score:        score,
			CommunityID:  communityID,
			PostCreateTS: createTS,
			ModelName:    modelName,
			ModelVersion: modelVersion,
			ContentHash:  contentHash,
		})
	}

	return hits, nil
}

// buildFilter 组合社区、帖子和自定义条件形成 Milvus 过滤表达式。
func buildFilter(req *SearchRequest) string {
	parts := make([]string, 0, 3)

	if req.CommunityID > 0 {
		parts = append(parts, fmt.Sprintf("%s == %d", fieldCommunityID, req.CommunityID))
	}
	if req.PostID > 0 {
		parts = append(parts, fmt.Sprintf("%s == %d", fieldPostID, req.PostID))
	}
	if extra := strings.TrimSpace(req.Filter); extra != "" {
		parts = append(parts, extra)
	}

	return strings.Join(parts, " && ")
}

// buildAnnParam 根据索引类型构造对应的检索参数。
func buildAnnParam(indexTypeRaw string, topK int) index.AnnParam {
	switch normalizedIndexType(indexTypeRaw) {
	case "HNSW":
		ef := topK * 8
		if ef < 64 {
			ef = 64
		}
		return index.NewHNSWAnnParam(ef)
	default:
		return nil
	}
}
