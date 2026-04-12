// Package minio реализует хранилище файлов на базе S3.
package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
)

// MinIO реализует хранилище файлов на базе S3.
type MinIO struct {
	client *minio.Client
	bucket string
	logger zerolog.Logger
}

// New создаёт новое подключение к S3 хранилищу.
func New(ctx context.Context, cfg config.MinIO, logger zerolog.Logger) (*MinIO, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать S3 клиент")
	}

	// Проверяем существование бакета, создаём если не существует
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось проверить существование бакета")
	}

	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, errors.Wrap(err, "не удалось создать бакет")
		}
		logger.Info().Str("bucket", cfg.Bucket).Msg("Создан S3 бакет")
	}

	logger.Info().Msg("Установлено подключение к S3")

	return &MinIO{
		client: client,
		bucket: cfg.Bucket,
		logger: logger,
	}, nil
}

// Ping проверяет доступность S3 хранилища.
func (s *MinIO) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
