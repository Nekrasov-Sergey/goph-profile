package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

// supportedMimeTypes содержит поддерживаемые MIME-типы изображений.
var supportedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// UploadAvatarRequest содержит данные для загрузки аватара.
type UploadAvatarRequest struct {
	UserID   string
	File     multipart.File
	FileName string
	Size     int64
}

// UploadAvatarResponse содержит результат загрузки аватара.
type UploadAvatarResponse struct {
	ID        uuid.UUID
	UserID    string
	URL       string
	Status    types.ProcessingStatus
	CreatedAt time.Time
}

// UploadAvatar загружает аватар пользователя.
func (s *Service) UploadAvatar(ctx context.Context, req UploadAvatarRequest) (*UploadAvatarResponse, error) {
	// Валидация размера файла
	if req.Size > maxFileSize {
		return nil, errcodes.ErrFileTooLarge
	}

	// Валидация MIME-типа
	buffer := make([]byte, 512)
	if _, err := req.File.Read(buffer); err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "не удалось прочитать файл")
	}

	mimeType := http.DetectContentType(buffer)
	if !supportedMimeTypes[mimeType] {
		return nil, errcodes.ErrInvalidFormat
	}

	// Сбрасываем позицию в файле после чтения буфера
	if seeker, ok := req.File.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, errors.Wrap(err, "не удалось сбросить позицию в файле")
		}
	}

	// Генерируем ID и ключ для S3
	avatarID := uuid.New()
	ext := filepath.Ext(req.FileName)
	if ext == "" {
		ext = s.extensionFromMimeType(mimeType)
	}
	s3Key := fmt.Sprintf("avatars/%s/original%s", avatarID, ext)

	// Загружаем файл в S3
	if err := s.storage.Upload(ctx, s3Key, req.File, req.Size, mimeType); err != nil {
		return nil, errors.Wrap(err, "не удалось загрузить файл в хранилище")
	}

	// Создаём запись в БД
	now := time.Now()
	avatar := &types.Avatar{
		ID:               avatarID,
		UserID:           req.UserID,
		FileName:         req.FileName,
		MimeType:         mimeType,
		SizeBytes:        req.Size,
		S3Key:            s3Key,
		ProcessingStatus: types.ProcessingStatusPending,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.CreateAvatar(ctx, avatar); err != nil {
		// Пытаемся удалить файл из S3 при ошибке записи в БД
		if delErr := s.storage.Delete(ctx, s3Key); delErr != nil {
			s.logger.Error().Err(delErr).Msg("не удалось удалить файл из S3 при откате")
		}
		return nil, errors.Wrap(err, "не удалось создать запись аватара")
	}

	// todo отправить задачу на создание миниатюр в кафку
	// todo воркер при получении задачи обновляет статус в postgres на ProcessingStatusProcessing
	// todo воркер создает миниатюры, загружает их в минио, а затем обновляет в postgres статус на ProcessingStatusCompleted
	// todo обновляет ThumbnailS3Keys и UpdatedAt. В случае неудачи ставится статус ProcessingStatusFailed

	return &UploadAvatarResponse{
		ID:        avatarID,
		UserID:    req.UserID,
		URL:       s.storage.GetURL(s3Key),
		Status:    types.ProcessingStatusPending,
		CreatedAt: now,
	}, nil
}

// extensionFromMimeType возвращает расширение файла по MIME-типу.
func (s *Service) extensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
