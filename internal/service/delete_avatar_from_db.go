package service

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

// DeleteAvatarRequest содержит параметры для удаления аватара.
type DeleteAvatarRequest struct {
	AvatarID uuid.UUID
	UserID   string
}

// DeleteAvatarFromDB удаляет аватар по ID.
func (s *Service) DeleteAvatarFromDB(ctx context.Context, req DeleteAvatarRequest) error {
	// Получаем аватар для проверки владельца и получения S3 ключа
	avatar, err := s.repo.GetAvatar(ctx, req.AvatarID)
	if err != nil {
		return err
	}

	// Проверяем, что пользователь владеет аватаром
	if avatar.UserID != req.UserID {
		return errcodes.ErrAccessDenied
	}

	// Мягкое удаление из БД
	if err := s.repo.SoftDeleteAvatar(ctx, req.AvatarID, req.UserID); err != nil {
		return err
	}

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
		return errors.Wrap(err, "не удалось сериализовать сообщение")
	}

	if err := s.producer.SendMessage(ctx, msgBytes); err != nil {
		s.logger.Error().Err(err).Send()
	}

	return nil
}
