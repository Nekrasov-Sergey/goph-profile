package service

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/imageutils"
)

// DeleteAvatarFromS3 удаляет аватар и его миниатюры из S3.
func (s *Service) DeleteAvatarFromS3(ctx context.Context, msg *types.AvatarMessage) error {
	// Удаляем оригинальный файл
	if err := s.storage.Delete(ctx, msg.S3Key); err != nil {
		return errors.Wrap(err, "не удалось удалить оригинальный файл")
	}

	// Определяем формат из MIME-типа
	format, err := imageutils.MimeTypeToFormat(msg.MimeType)
	if err != nil {
		return err
	}

	// Удаляем миниатюры
	for size := range types.ThumbnailDimensions() {
		thumbKey := fmt.Sprintf("%s/%s.%s", msg.AvatarID, string(size), format)
		if err := s.storage.Delete(ctx, thumbKey); err != nil {
			return errors.Wrap(err, "не удалось удалить миниатюру")
		}
	}

	return nil
}
