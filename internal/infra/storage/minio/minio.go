// Package minio реализует хранилище файлов на базе S3.
package minio

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

const tracerName = "avatar-service/minio"

// options — параметры подключения к S3, настраиваемые через функциональные опции.
type options struct {
	minIOCfg config.MinIO
	meter    *metrics.Instruments
}

// Option — функциональная опция для MinIO.
type Option func(*options)

// WithMinIOCfg устанавливает конфигурацию S3 хранилища.
func WithMinIOCfg(minIOCfg config.MinIO) Option {
	return func(o *options) {
		o.minIOCfg = minIOCfg
	}
}

// WithMeter задаёт метрические инструменты.
func WithMeter(meter *metrics.Instruments) Option {
	return func(o *options) {
		o.meter = meter
	}
}

// MinIO реализует хранилище файлов на базе S3.
type MinIO struct {
	client *minio.Client
	bucket string
	logger zerolog.Logger
	tracer trace.Tracer
	meter  *metrics.Instruments
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
		tracer: otel.Tracer(tracerName),
		meter:  o.meter,
	}, nil
}

// Ping проверяет доступность S3 хранилища.
func (s *MinIO) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
