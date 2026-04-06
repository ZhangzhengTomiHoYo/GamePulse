package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
)

// UploadFile 真正的上传操作，供 logic 层调用
func UploadFile(objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	ctx := context.Background()

	// 推流到 MinIO
	_, err := client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// 拼装公网访问链接返回给上层
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	// 拼接规则：http://127.0.0.1:9000/gamepulse-images/123456.jpg
	imageURL := fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, bucketName, objectName)

	return imageURL, nil
}
