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

// Download скачивает файл из S3 хранилища.
func (s *MinIO) Download(ctx context.Context, key string) (obj io.ReadCloser, err error) {
	ctx, span := s.tracer.Start(ctx, "s3.Download",
		trace.WithAttributes(
			attribute.String("s3.bucket", s.bucket),
			attribute.String("s3.key", key),
		),
	)
	defer span.End()

	start := time.Now()
	defer func() {
		s.recordS3Metrics(ctx, "download", 0, err, time.Since(start))
	}()

	obj, err = s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, tracer.SpanError(span, errors.Wrap(err, "не удалось скачать файл из S3"))
	}
	return obj, nil
}
