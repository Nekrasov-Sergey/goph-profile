// Package minio реализует хранилище файлов на базе S3.
package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

// Upload загружает файл в S3 хранилище.
func (s *MinIO) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	putObjectOptions := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if _, err := s.client.PutObject(ctx, s.bucket, key, reader, size, putObjectOptions); err != nil {
		return errors.Wrap(err, "не удалось загрузить файл в S3")
	}
	return nil
}
