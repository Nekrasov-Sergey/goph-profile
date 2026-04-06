// Package minio реализует хранилище файлов на базе S3.
package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

// Download скачивает файл из S3 хранилища.
func (s *MinIO) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "не удалось скачать файл из S3")
	}
	return obj, nil
}
