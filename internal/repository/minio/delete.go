// Package minio реализует хранилище файлов на базе S3.
package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
)

// Delete удаляет файл из S3 хранилища.
func (s *MinIO) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return errors.Wrap(err, "не удалось удалить файл из S3")
	}
	return nil
}
