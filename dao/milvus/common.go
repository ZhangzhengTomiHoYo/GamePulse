package milvus

import (
	"context"
	"errors"
	"gamepulse/setting"
	"strings"
	"time"
)

const (
	defaultDBName         = "default"
	defaultCollectionName = "post_vectors"
	defaultInitTimeout    = 20 * time.Second
	defaultRequestTimeout = 10 * time.Second
)

// Ready 返回 Milvus 客户端是否已经完成初始化。
func Ready() bool {
	return cli != nil
}

// validateConfig 校验 Milvus 配置是否合法。
func validateConfig(cfg *setting.MilvusConfig) error {
	if cfg == nil {
		return errors.New("milvus config is nil")
	}
	if strings.TrimSpace(cfg.Address) == "" {
		return errors.New("milvus address is empty")
	}
	if cfg.Dimension <= 0 {
		return errors.New("milvus dimension must be greater than 0")
	}
	if _, err := parseMetricType(cfg.MetricType); err != nil {
		return err
	}
	if _, err := buildVectorIndex(cfg.MetricType, cfg.IndexType); err != nil {
		return err
	}
	return nil
}

// withTimeout 为没有截止时间的上下文补上默认超时。
func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), timeout)
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}
