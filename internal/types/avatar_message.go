// Package types содержит доменные модели приложения.
package types

import (
	"github.com/google/uuid"
)

// AvatarOperation представляет тип операции над аватаром.
type AvatarOperation string

const (
	// OperationCreateThumbnails - создание миниатюр.
	OperationCreateThumbnails AvatarOperation = "create_thumbnails"
	// OperationDeleteFromS3 - удаление аватара.
	OperationDeleteFromS3 AvatarOperation = "delete_from_s3"
)

// AvatarMessage представляет сообщение для обработки аватара.
type AvatarMessage struct {
	AvatarID  uuid.UUID       `json:"avatar_id"`
	UserID    string          `json:"user_id"`
	S3Key     string          `json:"s3_key"`
	MimeType  MIMEType        `json:"mime_type"`
	Operation AvatarOperation `json:"operation"`
}
