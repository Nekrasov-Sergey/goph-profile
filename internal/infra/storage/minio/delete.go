// Package minio реализует хранилище файлов на базе S3.
package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

// Delete удаляет файл из S3 хранилища.
func (s *MinIO) Delete(ctx context.Context, key string) error {
	ctx, span := s.tracer.Start(ctx, "s3.Delete",
		trace.WithAttributes(
			attribute.String("s3.bucket", s.bucket),
			attribute.String("s3.key", key),
		),
	)
	defer span.End()

	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return tracer.SpanError(span, errors.Wrap(err, "не удалось удалить файл из S3"))
	}
	return nil
}
