package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"time"

	"github.com/goccy/go-json"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/imageutils"
)

// CreateThumbnails создаёт миниатюры для аватара.
func (s *Service) CreateThumbnails(ctx context.Context, msg *types.AvatarMessage) error {
	// Получаем текущее состояние аватара
	avatar, err := s.repo.GetAvatar(ctx, msg.AvatarID)
	if err != nil {
		return err
	}

	// Идемпотентность: если уже обработан, пропускаем
	if avatar.ProcessingStatus == types.ProcessingStatusCompleted {
		return nil
	}

	// Обновляем статус на processing
	avatar.ProcessingStatus = types.ProcessingStatusProcessing
	avatar.UpdatedAt = time.Now()
	if err := s.repo.UpdateAvatar(ctx, avatar); err != nil {
		return errors.Wrap(err, "не удалось обновить статус на processing")
	}

	// Скачиваем оригинальное изображение
	reader, err := s.storage.Download(ctx, msg.S3Key)
	if err != nil {
		return multierr.Append(err, s.setStatusFailed(ctx, avatar))
	}
	defer multierr.AppendInvoke(&err, multierr.Close(reader))

	// Декодируем изображение
	img, format, err := image.Decode(reader)
	if err != nil {
		err := errors.Wrap(err, "не удалось декодировать изображение")
		return multierr.Append(err, s.setStatusFailed(ctx, avatar))
	}

	mimeType, err := imageutils.FormatToMimeType(types.ImageFormat(format))
	if err != nil {
		return multierr.Append(err, s.setStatusFailed(ctx, avatar))
	}

	// Создаём миниатюры
	thumbnailKeys := make(map[types.ThumbnailSize]string)
	for sizeName, dimension := range types.ThumbnailDimensions() {
		// Ресайзим изображение
		resized := resize.Thumbnail(dimension, dimension, img, resize.Lanczos3)

		data, err := imageutils.Encode(resized, mimeType)
		if err != nil {
			return multierr.Append(err, s.setStatusFailed(ctx, avatar))
		}

		// Загружаем в S3
		thumbKey := fmt.Sprintf("%s/%s.%s", msg.AvatarID, string(sizeName), format)

		if err := s.storage.Upload(ctx, thumbKey, bytes.NewReader(data), int64(len(data)), string(msg.MimeType)); err != nil {
			return multierr.Append(err, s.setStatusFailed(ctx, avatar))
		}

		thumbnailKeys[sizeName] = thumbKey
	}

	// Сериализуем ключи миниатюр
	thumbnailKeysJSON, err := json.Marshal(thumbnailKeys)
	if err != nil {
		err := errors.Wrap(err, "не удалось сериализовать ключи миниатюр")
		return multierr.Append(err, s.setStatusFailed(ctx, avatar))
	}

	// Обновляем статус на completed
	avatar.ProcessingStatus = types.ProcessingStatusCompleted
	avatar.ThumbnailS3Keys = thumbnailKeysJSON
	avatar.UpdatedAt = time.Now()
	if err := s.repo.UpdateAvatar(ctx, avatar); err != nil {
		return err
	}

	return nil
}

// setStatusFailed устанавливает статус failed для аватара.
func (s *Service) setStatusFailed(ctx context.Context, avatar *types.Avatar) error {
	avatar.ProcessingStatus = types.ProcessingStatusFailed
	avatar.UpdatedAt = time.Now()
	return s.repo.UpdateAvatar(ctx, avatar)
}
