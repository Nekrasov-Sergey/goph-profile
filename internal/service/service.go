// Package service реализует бизнес-логику приложения.
package service

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// Repository определяет интерфейс хранилища данных.
//
//go:generate minimock -i Repository -o ./mocks/repo.go -n RepoMock
type Repository interface {
	WithTx(ctx context.Context, fn func(txRepo Repository) error) error
	// Close закрывает соединение с хранилищем и освобождает ресурсы
	Close() error
	// Ping проверяет доступность БД
	Ping(ctx context.Context) error
	// CreateAvatar создаёт новую запись аватара
	CreateAvatar(ctx context.Context, avatar *types.Avatar) error
	// GetAvatar получает аватар по ID
	GetAvatar(ctx context.Context, avatarID uuid.UUID) (*types.Avatar, error)
	// UpdateAvatar обновляет запись аватара
	UpdateAvatar(ctx context.Context, avatar *types.Avatar) error
	// SoftDeleteAvatar выполняет мягкое удаление аватара
	SoftDeleteAvatar(ctx context.Context, id uuid.UUID, userID string) error
}

// Storage определяет интерфейс файлового хранилища.
//
//go:generate minimock -i Storage -o ./mocks/storage.go -n StorageMock
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GetURL(key string) string
	// Ping проверяет доступность хранилища
	Ping(ctx context.Context) error
}

// Producer определяет интерфейс брокера сообщений.
//
//go:generate minimock -i Producer -o ./mocks/producer.go -n ProducerMock
type Producer interface {
	SendMessage(ctx context.Context, value []byte) error
	Ping(ctx context.Context) error
}

// Consumer определяет интерфейс консьюмера Kafka.
//
//go:generate minimock -i Consumer -o ./mocks/consumer.go -n ConsumerMock
type Consumer interface {
	ReadMessage(ctx context.Context) (*kafka.Message, error)
	// Close закрывает соединение с Kafka.
	Close() error
}

// Service реализует бизнес-логику работы с аватарами.
type Service struct {
	repo     Repository
	storage  Storage
	producer Producer
	consumer Consumer
	logger   zerolog.Logger
}

// New создаёт новый экземпляр сервиса.
func New(repo Repository, storage Storage, producer Producer, consumer Consumer, logger zerolog.Logger) *Service {
	return &Service{
		repo:     repo,
		storage:  storage,
		producer: producer,
		consumer: consumer,
		logger:   logger,
	}
}

// GetURL возвращает URL для доступа к файлу в хранилище.
func (s *Service) GetURL(s3Key string) string {
	return s.storage.GetURL(s3Key)
}
