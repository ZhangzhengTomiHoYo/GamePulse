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

const (
	fieldPostID       = "post_id"
	fieldCommunityID  = "community_id"
	fieldPostCreateTS = "post_create_ts"
	fieldModelName    = "model_name"
	fieldModelVersion = "model_version"
	fieldContentHash  = "content_hash"
	fieldEmbedding    = "embedding"

	vectorIndexName      = "idx_embedding"
	communityIDIndexName = "idx_community_id"
	postCreateIndexName  = "idx_post_create_ts"

	modelNameMaxLength    int64 = 64
	modelVersionMaxLength int64 = 32
	contentHashMaxLength  int64 = 64
)

// EnsureCollection 负责幂等地创建集合、校验结构并加载集合。
func EnsureCollection(cfg *setting.MilvusConfig) error {
	if cli == nil {
		return errors.New("milvus client is not initialized")
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := ensureDatabase(ctx, cli, normalizedDBName(cfg)); err != nil {
		return err
	}

	return ensureCollection(ctx, cli, cfg)
}

// ensureDatabase 确保目标数据库存在，并切换到对应数据库。
func ensureDatabase(ctx context.Context, currentClient *milvusclient.Client, dbName string) error {
	if dbName == defaultDBName {
		return nil
	}

	dbs, err := currentClient.ListDatabase(ctx, milvusclient.NewListDatabaseOption())
	if err != nil {
		return fmt.Errorf("list milvus databases failed: %w", err)
	}

	if !containsString(dbs, dbName) {
		if err := currentClient.CreateDatabase(ctx, milvusclient.NewCreateDatabaseOption(dbName)); err != nil {
			return fmt.Errorf("create milvus database %q failed: %w", dbName, err)
		}
	}

	if err := currentClient.UseDatabase(ctx, milvusclient.NewUseDatabaseOption(dbName)); err != nil {
		return fmt.Errorf("use milvus database %q failed: %w", dbName, err)
	}

	return nil
}

// ensureCollection 确保目标集合存在、结构正确且已加载到内存。
func ensureCollection(ctx context.Context, currentClient *milvusclient.Client, cfg *setting.MilvusConfig) error {
	collectionName := normalizedCollectionName(cfg)

	has, err := currentClient.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("check milvus collection %q failed: %w", collectionName, err)
	}

	if !has {
		if err := createCollection(ctx, currentClient, cfg); err != nil {
			has, checkErr := currentClient.HasCollection(ctx, milvusclient.NewHasCollectionOption(collectionName))
			if checkErr != nil || !has {
				return err
			}
		}
	}

	collection, err := currentClient.DescribeCollection(ctx, milvusclient.NewDescribeCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("describe milvus collection %q failed: %w", collectionName, err)
	}
	if err := validateCollectionSchema(collection, cfg); err != nil {
		return err
	}
	if err := ensureIndexes(ctx, currentClient, cfg); err != nil {
		return err
	}

	loadTask, err := currentClient.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collectionName))
	if err != nil {
		return fmt.Errorf("load milvus collection %q failed: %w", collectionName, err)
	}
	if err := loadTask.Await(ctx); err != nil {
		return fmt.Errorf("await milvus collection %q load failed: %w", collectionName, err)
	}

	return nil
}

// createCollection 按照项目约定的 schema 创建帖子向量集合。
func createCollection(ctx context.Context, currentClient *milvusclient.Client, cfg *setting.MilvusConfig) error {
	collectionName := normalizedCollectionName(cfg)
	vectorIndex, err := buildVectorIndex(cfg.MetricType, cfg.IndexType)
	if err != nil {
		return err
	}

	schema := entity.NewSchema().
		WithName(collectionName).
		WithDescription("GamePulse 帖子向量集合").
		WithAutoID(false).
		WithDynamicFieldEnabled(false).
		WithField(entity.NewField().
			WithName(fieldPostID).
			WithDataType(entity.FieldTypeInt64).
			WithIsPrimaryKey(true)).
		WithField(entity.NewField().
			WithName(fieldCommunityID).
			WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().
			WithName(fieldPostCreateTS).
			WithDataType(entity.FieldTypeInt64)).
		WithField(entity.NewField().
			WithName(fieldModelName).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(modelNameMaxLength)).
		WithField(entity.NewField().
			WithName(fieldModelVersion).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(modelVersionMaxLength)).
		WithField(entity.NewField().
			WithName(fieldContentHash).
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(contentHashMaxLength)).
		WithField(entity.NewField().
			WithName(fieldEmbedding).
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(int64(cfg.Dimension)))

	indexOptions := []milvusclient.CreateIndexOption{
		milvusclient.NewCreateIndexOption(collectionName, fieldEmbedding, vectorIndex).WithIndexName(vectorIndexName),
		milvusclient.NewCreateIndexOption(collectionName, fieldCommunityID, index.NewSortedIndex()).WithIndexName(communityIDIndexName),
		milvusclient.NewCreateIndexOption(collectionName, fieldPostCreateTS, index.NewSortedIndex()).WithIndexName(postCreateIndexName),
	}

	if err := currentClient.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collectionName, schema).WithIndexOptions(indexOptions...)); err != nil {
		return fmt.Errorf("create milvus collection %q failed: %w", collectionName, err)
	}

	return nil
}

// ensureIndexes 确保集合所需的向量索引和标量索引都已创建。
func ensureIndexes(ctx context.Context, currentClient *milvusclient.Client, cfg *setting.MilvusConfig) error {
	vectorIndex, err := buildVectorIndex(cfg.MetricType, cfg.IndexType)
	if err != nil {
		return err
	}

	tasks := []struct {
		fieldName string
		indexName string
		indexDef  index.Index
	}{
		{fieldName: fieldEmbedding, indexName: vectorIndexName, indexDef: vectorIndex},
		{fieldName: fieldCommunityID, indexName: communityIDIndexName, indexDef: index.NewSortedIndex()},
		{fieldName: fieldPostCreateTS, indexName: postCreateIndexName, indexDef: index.NewSortedIndex()},
	}

	for _, task := range tasks {
		indexes, err := currentClient.ListIndexes(ctx, milvusclient.NewListIndexOption(normalizedCollectionName(cfg)).WithFieldName(task.fieldName))
		if err != nil {
			return fmt.Errorf("list indexes for field %q failed: %w", task.fieldName, err)
		}
		if len(indexes) > 0 {
			continue
		}

		createTask, err := currentClient.CreateIndex(ctx, milvusclient.NewCreateIndexOption(normalizedCollectionName(cfg), task.fieldName, task.indexDef).WithIndexName(task.indexName))
		if err != nil {
			return fmt.Errorf("create index %q failed: %w", task.indexName, err)
		}
		if err := createTask.Await(ctx); err != nil {
			return fmt.Errorf("await index %q failed: %w", task.indexName, err)
		}
	}

	return nil
}

// validateCollectionSchema 校验现有集合结构是否与项目要求一致。
func validateCollectionSchema(collection *entity.Collection, cfg *setting.MilvusConfig) error {
	if collection == nil || collection.Schema == nil {
		return errors.New("milvus collection schema is nil")
	}

	requiredFields := map[string]entity.FieldType{
		fieldPostID:       entity.FieldTypeInt64,
		fieldCommunityID:  entity.FieldTypeInt64,
		fieldPostCreateTS: entity.FieldTypeInt64,
		fieldModelName:    entity.FieldTypeVarChar,
		fieldModelVersion: entity.FieldTypeVarChar,
		fieldContentHash:  entity.FieldTypeVarChar,
	}

	for name, dataType := range requiredFields {
		field := findField(collection.Schema, name)
		if field == nil {
			return fmt.Errorf("milvus collection %q missing field %q", collection.Name, name)
		}
		if field.DataType != dataType {
			return fmt.Errorf("milvus collection %q field %q type is %v, want %v", collection.Name, name, field.DataType, dataType)
		}
	}

	postIDField := findField(collection.Schema, fieldPostID)
	if postIDField == nil || !postIDField.PrimaryKey {
		return fmt.Errorf("milvus collection %q field %q is not primary key", collection.Name, fieldPostID)
	}

	vectorField := findField(collection.Schema, fieldEmbedding)
	if vectorField == nil {
		return fmt.Errorf("milvus collection %q missing vector field %q", collection.Name, fieldEmbedding)
	}
	if vectorField.DataType != entity.FieldTypeFloatVector {
		return fmt.Errorf("milvus collection %q field %q type is %v, want %v", collection.Name, fieldEmbedding, vectorField.DataType, entity.FieldTypeFloatVector)
	}

	dim, err := vectorField.GetDim()
	if err != nil {
		return fmt.Errorf("get milvus vector dim failed: %w", err)
	}
	if int(dim) != cfg.Dimension {
		return fmt.Errorf("milvus collection %q vector dim is %d, want %d", collection.Name, dim, cfg.Dimension)
	}

	return nil
}

// buildVectorIndex 根据配置构造向量索引定义。
func buildVectorIndex(metricTypeRaw, indexTypeRaw string) (index.Index, error) {
	metricType, err := parseMetricType(metricTypeRaw)
	if err != nil {
		return nil, err
	}

	switch normalizedIndexType(indexTypeRaw) {
	case "HNSW":
		return index.NewHNSWIndex(metricType, 16, 200), nil
	case "FLAT":
		return index.NewFlatIndex(metricType), nil
	case "AUTO", "AUTOINDEX":
		return index.NewAutoIndex(metricType), nil
	case "DISKANN":
		return index.NewDiskANNIndex(metricType), nil
	default:
		return nil, fmt.Errorf("unsupported milvus index_type %q", indexTypeRaw)
	}
}

// parseMetricType 将配置中的度量类型转换为 Milvus SDK 枚举。
func parseMetricType(raw string) (entity.MetricType, error) {
	switch normalizedMetricType(raw) {
	case "COSINE":
		return entity.COSINE, nil
	case "L2":
		return entity.L2, nil
	case "IP":
		return entity.IP, nil
	default:
		return entity.COSINE, fmt.Errorf("unsupported milvus metric_type %q", raw)
	}
}

// normalizedDBName 返回规范化后的数据库名称。
func normalizedDBName(cfg *setting.MilvusConfig) string {
	if cfg == nil || strings.TrimSpace(cfg.DBName) == "" {
		return defaultDBName
	}
	return strings.TrimSpace(cfg.DBName)
}

// normalizedCollectionName 返回规范化后的集合名称。
func normalizedCollectionName(cfg *setting.MilvusConfig) string {
	if cfg == nil || strings.TrimSpace(cfg.CollectionName) == "" {
		return defaultCollectionName
	}
	return strings.TrimSpace(cfg.CollectionName)
}

// normalizedMetricType 返回规范化后的向量度量类型。
func normalizedMetricType(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "COSINE"
	}
	return strings.ToUpper(strings.TrimSpace(raw))
}

// normalizedIndexType 返回规范化后的索引类型。
func normalizedIndexType(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "HNSW"
	}
	return strings.ToUpper(strings.TrimSpace(raw))
}

// findField 按字段名从集合 schema 中查找字段定义。
func findField(schema *entity.Schema, fieldName string) *entity.Field {
	for _, field := range schema.Fields {
		if field.Name == fieldName {
			return field
		}
	}
	return nil
}

// containsString 判断字符串切片中是否包含目标值。
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
