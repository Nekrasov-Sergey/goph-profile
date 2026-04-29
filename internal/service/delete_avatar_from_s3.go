package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/imageutils"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

// DeleteAvatarFromS3 удаляет аватар и его миниатюры из S3.
func (s *Service) DeleteAvatarFromS3(ctx context.Context, msg *types.AvatarMessage) (err error) {
	ctx, span := s.tracer.Start(ctx, "service.DeleteAvatarFromS3",
		trace.WithAttributes(
			attribute.String("avatar.id", msg.AvatarID.String()),
		),
	)
	defer span.End()

	s3Success := false
	defer func() {
		s.recordDeleteMetrics(ctx, "s3_delete", s3Success)
	}()

	// Удаляем оригинальный файл
	if err := s.storage.Delete(ctx, msg.S3Key); err != nil {
		return tracer.SpanError(span, errors.Wrap(err, "не удалось удалить оригинальный файл"))
	}

	// Определяем формат из MIME-типа
	format, err := imageutils.MimeTypeToFormat(msg.MimeType)
	if err != nil {
		return tracer.SpanError(span, err)
	}

	// Удаляем миниатюры
	for size := range types.ThumbnailDimensions() {
		thumbKey := fmt.Sprintf("%s/%s.%s", msg.AvatarID, string(size), format)
		if err := s.storage.Delete(ctx, thumbKey); err != nil {
			return tracer.SpanError(span, errors.Wrap(err, "не удалось удалить миниатюру"))
		}
	}

	s3Success = true
	return nil
}
