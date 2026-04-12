// Package domain содержит доменные модели приложения.
package types

import (
	"time"

	"github.com/google/uuid"
)

// ProcessingStatus представляет статус обработки аватара.
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// Avatar представляет модель аватара.
type Avatar struct {
	ID               uuid.UUID        `db:"id"`
	UserID           string           `db:"user_id"`
	FileName         string           `db:"file_name"`
	MimeType         MIMEType         `db:"mime_type"`
	SizeBytes        int64            `db:"size_bytes"`
	Width            int              `db:"width"`
	Height           int              `db:"height"`
	S3Key            string           `db:"s3_key"`
	ThumbnailS3Keys  []byte           `db:"thumbnail_s3_keys"` // map[size]s3_key
	ProcessingStatus ProcessingStatus `db:"processing_status"`
	CreatedAt        time.Time        `db:"created_at"`
	UpdatedAt        time.Time        `db:"updated_at"`
	DeletedAt        *time.Time       `db:"deleted_at"`
}

// ThumbnailInfo содержит информацию о миниатюре.
type ThumbnailInfo struct {
	Size string
	URL  string
}
