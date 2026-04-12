// Package kafka реализует брокер сообщений на базе Kafka.
package kafka

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
)

// Producer реализует Producer с использованием kafka-go.
type Producer struct {
	writer *kafka.Writer
	logger zerolog.Logger
}

// NewProducer создаёт новый продюсер Kafka.
func NewProducer(ctx context.Context, logger zerolog.Logger, cfgKafka config.Kafka) (*Producer, error) {
	producer := &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfgKafka.Brokers...),
			Topic:                  cfgKafka.Topic,
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.LeastBytes{},
		},
		logger: logger,
	}

	if err := producer.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к Kafka")
	}

	logger.Info().Msg("Установлено подключение к Kafka")

	return producer, nil
}

// SendMessage отправляет сообщение в Kafka.
func (p *Producer) SendMessage(ctx context.Context, value []byte) error {
	kafkaMsg := kafka.Message{
		Value: value,
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return errors.Wrap(err, "не удалось отправить сообщение в Kafka")
	}

	return nil
}

// Close закрывает соединение с Kafka.
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return errors.Wrap(err, "не удалось закрыть Kafka")
	}
	p.logger.Info().Msg("Закрыто соединение с Kafka")
	return nil
}

// Ping проверяет доступность Kafka.
func (p *Producer) Ping(ctx context.Context) error {
	// Создаём временное соединение для проверки
	conn, err := kafka.DialLeader(ctx, "tcp", p.writer.Addr.String(), p.writer.Topic, 0)
	if err != nil {
		return err
	}
	return conn.Close()
}
