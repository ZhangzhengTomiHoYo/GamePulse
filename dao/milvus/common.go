package milvus

import (
	"context"
	"time"

	"go.uber.org/zap"
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
