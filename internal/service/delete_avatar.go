package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

// DeleteAvatarRequest содержит параметры для удаления аватара.
type DeleteAvatarRequest struct {
	ID     uuid.UUID
	UserID string
}

// DeleteAvatar удаляет аватар по ID.
func (s *Service) DeleteAvatar(ctx context.Context, req DeleteAvatarRequest) error {
	// Получаем аватар для проверки владельца и получения S3 ключа
	avatar, err := s.repo.GetAvatar(ctx, req.ID)
	if err != nil {
		return err
	}

	// Проверяем, что пользователь владеет аватаром
	if avatar.UserID != req.UserID {
		return errcodes.ErrAccessDenied
	}

	// Удаляем файл из S3
	// todo удаление из s3 аватарки и миниатюр должно происходить асинхронно через воркер
	if err := s.storage.Delete(ctx, avatar.S3Key); err != nil {
		s.logger.Error().Err(err).Msg("не удалось удалить файл из S3")
		// Продолжаем удаление из БД даже если не удалось удалить из S3
	}

	// Мягкое удаление из БД
	return s.repo.SoftDeleteAvatar(ctx, req.ID, req.UserID)
}
