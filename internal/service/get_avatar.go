package service

import (
	"context"
	"io"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/imageutils"
)

// GetAvatarRequest содержит параметры для получения аватара.
type GetAvatarRequest struct {
	ID     uuid.UUID
	Size   types.ThumbnailSize // "100x100", "300x300"
	Format types.ImageFormat   // "jpeg", "png", "webp"
}

// GetAvatarResponse содержит результат получения аватара.
type GetAvatarResponse struct {
	Reader   io.ReadCloser
	MimeType types.MIMEType
	Size     int64
}

// GetAvatar получает аватар по ID с поддержкой размера и формата.
func (s *Service) GetAvatar(ctx context.Context, req GetAvatarRequest) (*GetAvatarResponse, error) {
	// Получаем метаданные из БД
	avatar, err := s.repo.GetAvatar(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	s3Key, err := resolveS3Key(avatar, req.Size)
	if err != nil {
		return nil, err
	}

	reader, err := s.storage.Download(ctx, s3Key)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось скачать файл из хранилища")
	}

	mimeType, err := imageutils.ResolveMimeType(avatar.MimeType, req.Format)
	if err != nil {
		multierr.AppendInvoke(&err, multierr.Close(reader))
		return nil, err
	}

	newReader, size, err := imageutils.ChangeMimeType(reader, avatar.MimeType, mimeType)
	if err != nil {
		return nil, err
	}

	return &GetAvatarResponse{
		Reader:   newReader,
		MimeType: mimeType,
		Size:     size,
	}, nil
}

// resolveS3Key определяет S3-ключ для запрошенного размера миниатюры.
func resolveS3Key(avatar *types.Avatar, size types.ThumbnailSize) (s3Key string, err error) {
	if size == "" {
		return avatar.S3Key, nil
	}

	if size != types.ThumbnailSize100 && size != types.ThumbnailSize300 {
		return "", errcodes.ErrInvalidSize
	}

	// Проверяем, что миниатюры готовы
	if avatar.ProcessingStatus != types.ProcessingStatusCompleted {
		return "", errcodes.ErrThumbnailNotReady
	}

	// Десериализуем ключи миниатюр
	var thumbnailKeys map[types.ThumbnailSize]string
	if err := json.Unmarshal(avatar.ThumbnailS3Keys, &thumbnailKeys); err != nil {
		return "", errors.Wrap(err, "не удалось десериализовать ключи миниатюр")
	}

	// Получаем ключ миниатюры
	s3Key, ok := thumbnailKeys[size]
	if !ok {
		return "", errors.Wrap(errcodes.ErrThumbnailNotReady, "миниатюра не найдена")
	}

	return s3Key, nil
}
