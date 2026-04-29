// Package kafka реализует брокер сообщений на базе Kafka.
package kafka

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// producerOptions — параметры подключения к Kafka, настраиваемые через функциональные опции.
type producerOptions struct {
	kafkaCfg config.Kafka
	meter    *metrics.Instruments
}

// ProducerOption — функциональная опция для Producer.
type ProducerOption func(*producerOptions)

// WithProducerKafkaCfg устанавливает конфигурацию Kafka.
func WithProducerKafkaCfg(kafkaCfg config.Kafka) ProducerOption {
	return func(o *producerOptions) {
		o.kafkaCfg = kafkaCfg
	}
}

// WithProducerMeter задаёт метрические инструменты.
func WithProducerMeter(meter *metrics.Instruments) ProducerOption {
	return func(o *producerOptions) {
		o.meter = meter
	}
}

// Producer реализует Producer с использованием kafka-go.
type Producer struct {
	writer *kafka.Writer
	logger zerolog.Logger
	meter  *metrics.Instruments
}

// NewProducer создаёт новый продюсер Kafka.
func NewProducer(ctx context.Context, logger zerolog.Logger, opts ...ProducerOption) (*Producer, error) {
	o := &producerOptions{}
	for _, opt := range opts {
		opt(o)
	}

	kafkaCfg := o.kafkaCfg

	producer := &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(kafkaCfg.Brokers...),
			Topic:                  kafkaCfg.Topic,
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.LeastBytes{},
		},
		logger: logger,
		meter:  o.meter,
	}

	if err := producer.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к Kafka")
	}

	logger.Info().Msg("Установлено подключение к Kafka")

	return producer, nil
}

// SendMessage отправляет сообщение в Kafka с проброшенным trace context.
func (p *Producer) SendMessage(ctx context.Context, value []byte) (err error) {
	start := time.Now()
	defer func() {
		p.recordKafkaProducerMetrics(ctx, err, time.Since(start))
	}()

	var headers kafkaHeadersCarrier
	otel.GetTextMapPropagator().Inject(ctx, &headers)

	kafkaMsg := kafka.Message{
		Value:   value,
		Headers: headers,
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
