// Package kafka реализует брокер сообщений на базе Kafka.
package kafka

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

type options struct {
	kafkaCfg config.Kafka
}

type Option func(*options)

func WithKafkaCfg(kafkaCfg config.Kafka) Option {
	return func(o *options) {
		o.kafkaCfg = kafkaCfg
	}
}

// Consumer реализует консьюмер сообщений Kafka.
type Consumer struct {
	reader *kafka.Reader
	logger zerolog.Logger
}

// NewConsumer создаёт новый консьюмер Kafka.
func NewConsumer(ctx context.Context, logger zerolog.Logger, opts ...Option) (*Consumer, error) {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	kafkaCfg := o.kafkaCfg

	consumer := &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: kafkaCfg.Brokers,
			Topic:   kafkaCfg.Topic,
			GroupID: kafkaCfg.GroupID,
		}),
		logger: logger,
	}

	if err := consumer.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к Kafka")
	}

	logger.Info().Msg("Установлено подключение к Kafka")

	return consumer, nil
}

// ReadAvatarMessage читает и десериализует следующее сообщение об аватаре из Kafka.
func (c *Consumer) ReadAvatarMessage(ctx context.Context) (*types.AvatarMessage, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось прочитать сообщение из Kafka")
	}

	var avatarMessage types.AvatarMessage
	if err := json.Unmarshal(msg.Value, &avatarMessage); err != nil {
		return nil, errors.Wrap(err, "не удалось десериализовать сообщение")
	}

	return &avatarMessage, nil
}

// Close закрывает соединение с Kafka.
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return errors.Wrap(err, "не удалось закрыть Kafka")
	}
	c.logger.Info().Msg("Закрыто соединение с Kafka")
	return nil
}

// Ping проверяет доступность Kafka.
func (c *Consumer) Ping(ctx context.Context) error {
	// Создаём временное соединение для проверки
	conn, err := kafka.DialLeader(ctx, "tcp", c.reader.Config().Brokers[0], c.reader.Config().Topic, 0)
	if err != nil {
		return err
	}
	return conn.Close()
}
