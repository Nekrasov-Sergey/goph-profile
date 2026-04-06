package service

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// GetAvatarRequest содержит параметры для получения аватара.
type GetAvatarRequest struct {
	ID     uuid.UUID
	Size   string // "original", "100x100", "300x300"
	Format string // "jpeg", "png", "webp" (пока не используется)
}

// GetAvatarResponse содержит результат получения аватара.
type GetAvatarResponse struct {
	Reader   io.ReadCloser
	MimeType string
	Size     int64
}

// GetAvatar получает аватар по ID.
func (s *Service) GetAvatar(ctx context.Context, req GetAvatarRequest) (*GetAvatarResponse, error) {
	// Получаем метаданные из БД
	avatar, err := s.repo.GetAvatar(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Определяем S3 ключ
	s3Key := avatar.S3Key
	mimeType := avatar.MimeType

	// todo: когда будут миниатюры, выбирать по size
	// if req.Size != "" && req.Size != "original" {
	//     if thumbnailKey, ok := getThumbnailKey(avatar.ThumbnailS3Keys, req.Size); ok {
	//         s3Key = thumbnailKey
	//     }
	// }

	// Скачиваем файл из S3
	reader, err := s.storage.Download(ctx, s3Key)
	if err != nil {
		return nil, err
	}

	return &GetAvatarResponse{
		Reader:   reader,
		MimeType: mimeType,
		Size:     avatar.SizeBytes,
	}, nil
}
