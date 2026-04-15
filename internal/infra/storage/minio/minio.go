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

type options struct {
	minIOCfg config.MinIO
}

type Option func(*options)

func WithMinIOCfg(minIOCfg config.MinIO) Option {
	return func(o *options) {
		o.minIOCfg = minIOCfg
	}
}

// MinIO реализует хранилище файлов на базе S3.
type MinIO struct {
	client *minio.Client
	bucket string
	logger zerolog.Logger
}

// New создаёт новое подключение к S3 хранилищу.
func New(ctx context.Context, logger zerolog.Logger, opts ...Option) (*MinIO, error) {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	minIOCfg := o.minIOCfg

	client, err := minio.New(minIOCfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minIOCfg.AccessKey, minIOCfg.SecretKey, ""),
		Secure: minIOCfg.UseSSL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать S3 клиент")
	}

	// Проверяем существование бакета, создаём если не существует
	exists, err := client.BucketExists(ctx, minIOCfg.Bucket)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось проверить существование бакета")
	}

	if !exists {
		if err := client.MakeBucket(ctx, minIOCfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, errors.Wrap(err, "не удалось создать бакет")
		}
		logger.Info().Str("bucket", minIOCfg.Bucket).Msg("Создан S3 бакет")
	}

	logger.Info().Msg("Установлено подключение к S3")

	return &MinIO{
		client: client,
		bucket: minIOCfg.Bucket,
		logger: logger,
	}, nil
}

// Ping проверяет доступность S3 хранилища.
func (s *MinIO) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
