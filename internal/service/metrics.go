package service

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// businessStatus конвертирует bool в статусную метку "success"/"failed".
func businessStatus(success bool) string {
	if success {
		return "success"
	}
	return "failed"
}

// recordUploadMetrics записывает метрики загрузки аватара.
func (s *Service) recordUploadMetrics(ctx context.Context, mimeType types.MIMEType, size int64, success bool) {
	if s.meter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		metrics.AttrMimeType(string(mimeType)),
		metrics.AttrBusinessStatus(businessStatus(success)),
	}

	s.meter.AvatarUploadCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	s.meter.AvatarUploadSizeBytes.Record(ctx, size,
		metric.WithAttributes(metrics.AttrMimeType(string(mimeType))),
	)
}

// recordDeleteMetrics записывает метрики удаления аватара.
func (s *Service) recordDeleteMetrics(ctx context.Context, op string, success bool) {
	if s.meter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		metrics.AttrAvatarOperation(op),
		metrics.AttrBusinessStatus(businessStatus(success)),
	}

	s.meter.AvatarDeleteCount.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// recordProcessingMetrics записывает метрики обработки миниатюр.
func (s *Service) recordProcessingMetrics(ctx context.Context, op string, dur time.Duration, success bool) {
	if s.meter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		metrics.AttrAvatarOperation(op),
		metrics.AttrBusinessStatus(businessStatus(success)),
	}

	s.meter.AvatarProcessingDuration.Record(ctx, dur.Seconds(), metric.WithAttributes(attrs...))
}

// recordThumbnailMetrics записывает метрики создания миниатюры.
func (s *Service) recordThumbnailMetrics(ctx context.Context, mimeType types.MIMEType, size int64, success bool) {
	if s.meter == nil {
		return
	}

	countAttrs := []attribute.KeyValue{
		metrics.AttrMimeType(string(mimeType)),
		metrics.AttrBusinessStatus(businessStatus(success)),
	}

	s.meter.AvatarThumbnailCount.Add(ctx, 1, metric.WithAttributes(countAttrs...))
	s.meter.AvatarThumbnailSizeBytes.Record(ctx, size,
		metric.WithAttributes(metrics.AttrMimeType(string(mimeType))),
	)
}
