package minio

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// recordS3Metrics записывает метрики выполнения операции с S3.
func (s *MinIO) recordS3Metrics(ctx context.Context, operation string, size int64, err error, dur time.Duration) {
	if s.meter == nil {
		return
	}

	status := "ok"
	if err != nil {
		status = "error"
	}

	countAttrs := []attribute.KeyValue{
		metrics.AttrS3Operation(operation),
		metrics.AttrS3Status(status),
	}

	s.meter.S3OperationCount.Add(ctx, 1, metric.WithAttributes(countAttrs...))
	s.meter.S3OperationDuration.Record(ctx, dur.Seconds(),
		metric.WithAttributes(metrics.AttrS3Operation(operation)),
	)

	if size > 0 {
		s.meter.S3OperationSize.Record(ctx, size,
			metric.WithAttributes(metrics.AttrS3Operation(operation)),
		)
	}

	if err != nil {
		s.meter.S3OperationErrors.Add(ctx, 1,
			metric.WithAttributes(metrics.AttrS3Operation(operation)),
		)
	}
}
