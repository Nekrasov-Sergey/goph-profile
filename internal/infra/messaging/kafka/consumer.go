// Package kafka реализует брокер сообщений на базе Kafka.
package kafka

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
)

type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer создаёт новый консьюмер Kafka.
func NewConsumer(ctx context.Context, logger zerolog.Logger, cfgKafka config.Kafka) (*Consumer, error) {
	consumer := &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: cfgKafka.Brokers,
			Topic:   cfgKafka.Topic,
			GroupID: cfgKafka.GroupID,
		}),
	}

	if err := consumer.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к Kafka")
	}

	logger.Info().Msg("Установлено подключение к Kafka")

	return consumer, nil
}

func (c *Consumer) ReadMessage(ctx context.Context) (*kafka.Message, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось прочитать сообщение из Kafka")
	}
	return &msg, nil
}

// Close закрывает соединение с Kafka.
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return errors.Wrap(err, "не удалось закрыть Kafka")
	}
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
