package minio

import (
	"bluebell/setting"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

var (
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
)

// Init 初始化 MinIO 客户端
func Init(cfg *setting.MinIOConfig) error {
	var err error
	endpoint = cfg.Endpoint
	bucketName = cfg.BucketName
	useSSL = cfg.UseSSL

	// 1. 建立连接
	client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("minio connect failed: %w", err)
	}

	// 2. 检查 Bucket 是否存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket exists failed: %w", err)
	}

	// 3. 如果不存在，自动创建并赋予公共读权限
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("make bucket failed: %w", err)
		}
		zap.L().Info("created new minio bucket", zap.String("bucketName", bucketName))

		// 设置桶策略为公开读（必须设置，否则前端拿到了 URL 也无法显示图片）
		policy := fmt.Sprintf(`{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"]}]}`, bucketName)
		err = client.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			// 策略设置失败不影响主流程，打个 Warn 即可
			zap.L().Warn("set bucket policy failed", zap.Error(err))
		}
	}

	zap.L().Info("minio init success")
	return nil
}
