package service

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

// DeleteAvatarRequest содержит параметры для удаления аватара.
type DeleteAvatarRequest struct {
	AvatarID uuid.UUID
	UserID   string
}

// DeleteAvatarFromDB удаляет аватар по ID.
func (s *Service) DeleteAvatarFromDB(ctx context.Context, req DeleteAvatarRequest) (err error) {
	ctx, span := s.tracer.Start(ctx, "service.DeleteAvatarFromDB",
		trace.WithAttributes(
			attribute.String("avatar.id", req.AvatarID.String()),
			attribute.String("user.id", req.UserID),
		),
	)
	defer span.End()

	dbSuccess := false
	defer func() {
		s.recordDeleteMetrics(ctx, "soft_delete", dbSuccess)
	}()

	// Получаем аватар для проверки владельца и получения S3 ключа
	avatar, err := s.repo.GetAvatar(ctx, req.AvatarID)
	if err != nil {
		return tracer.SpanError(span, err)
	}

	// Проверяем, что пользователь владеет аватаром
	if avatar.UserID != req.UserID {
		return tracer.SpanError(span, errcodes.ErrAccessDenied)
	}

	// Мягкое удаление из БД
	if err := s.repo.SoftDeleteAvatar(ctx, req.AvatarID, req.UserID); err != nil {
		return tracer.SpanError(span, err)
	}

	dbSuccess = true

	// Отправляем сообщение в Kafka для удаления аватарки
	msg := types.AvatarMessage{
		AvatarID:  avatar.ID,
		UserID:    avatar.UserID,
		S3Key:     avatar.S3Key,
		MimeType:  avatar.MimeType,
		Operation: types.OperationDeleteFromS3,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return tracer.SpanError(span, errors.Wrap(err, "не удалось сериализовать сообщение"))
	}

	if err := s.producer.SendMessage(ctx, msgBytes); err != nil {
		s.logger.Error().Err(err).Send()
	}

	return nil
}
