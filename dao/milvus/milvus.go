package milvus

import (
	"bluebell/setting"
	"errors"
	"fmt"
	"strings"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"go.uber.org/zap"
)

var cli *milvusclient.Client

// Init 初始化 Milvus 客户端连接。
func Init(cfg *setting.MilvusConfig) (err error) {
	if err = validateConfig(cfg); err != nil {
		return err
	}

	if cli != nil {
		Close()
	}

	ctx, cancel := withTimeout(nil, defaultInitTimeout)
	defer cancel()

	cli, err = milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address:  strings.TrimSpace(cfg.Address),
		Username: strings.TrimSpace(cfg.Username),
		Password: cfg.Password,
	})
	if err != nil {
		return fmt.Errorf("connect milvus failed: %w", err)
	}

	zap.L().Info("milvus init success")
	return nil
}

// Close 关闭当前 Milvus 客户端连接。
func Close() {
	if cli == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cli.Close(ctx); err != nil {
		zap.L().Warn("close milvus client failed", zap.Error(err))
	}

	cli = nil
}