package minio

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

// Upload загружает файл в S3 хранилище.
func (s *MinIO) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (err error) {
	ctx, span := s.tracer.Start(ctx, "s3.Upload",
		trace.WithAttributes(
			attribute.String("s3.bucket", s.bucket),
			attribute.String("s3.key", key),
			attribute.Int64("s3.size_bytes", size),
		),
	)
	defer span.End()

	start := time.Now()
	defer func() {
		s.recordS3Metrics(ctx, "upload", size, err, time.Since(start))
	}()

	putObjectOptions := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if _, err := s.client.PutObject(ctx, s.bucket, key, reader, size, putObjectOptions); err != nil {
		return tracer.SpanError(span, errors.Wrap(err, "не удалось загрузить файл в S3"))
	}
	return nil
}
