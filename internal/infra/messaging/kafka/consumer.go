package kafka

import (
	"context"
	"time"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// consumerOptions — параметры подключения к Kafka, настраиваемые через функциональные опции.
type consumerOptions struct {
	kafkaCfg config.Kafka
	meter    *metrics.Instruments
}

// ConsumerOption — функциональная опция для Consumer.
type ConsumerOption func(*consumerOptions)

// WithConsumerKafkaCfg устанавливает конфигурацию Kafka.
func WithConsumerKafkaCfg(kafkaCfg config.Kafka) ConsumerOption {
	return func(o *consumerOptions) {
		o.kafkaCfg = kafkaCfg
	}
}

// WithConsumerMeter задаёт метрические инструменты.
func WithConsumerMeter(meter *metrics.Instruments) ConsumerOption {
	return func(o *consumerOptions) {
		o.meter = meter
	}
}

// Consumer реализует консьюмер сообщений Kafka.
type Consumer struct {
	reader *kafka.Reader
	logger zerolog.Logger
	meter  *metrics.Instruments
}

// NewConsumer создаёт новый консьюмер Kafka.
func NewConsumer(ctx context.Context, logger zerolog.Logger, opts ...ConsumerOption) (*Consumer, error) {
	o := &consumerOptions{}
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
		meter:  o.meter,
	}

	if err := consumer.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к Kafka")
	}

	logger.Info().Msg("Установлено подключение к Kafka")

	return consumer, nil
}

// ReadAvatarMessage читает и десериализует следующее сообщение об аватаре из Kafka.
// Возвращает контекст с восстановленным trace context из заголовков сообщения.
func (c *Consumer) ReadAvatarMessage(ctx context.Context) (retCtx context.Context, msg *types.AvatarMessage, err error) {
	start := time.Now()
	defer func() {
		c.recordKafkaConsumerMetrics(ctx, err, time.Since(start))
	}()

	kafkaMsg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return ctx, nil, errors.Wrap(err, "не удалось прочитать сообщение из Kafka")
	}

	// Восстанавливаем trace context из заголовков Kafka-сообщения
	retCtx = otel.GetTextMapPropagator().Extract(ctx, new(kafkaHeadersCarrier(kafkaMsg.Headers)))

	var avatarMessage types.AvatarMessage
	if err := json.Unmarshal(kafkaMsg.Value, &avatarMessage); err != nil {
		return retCtx, nil, errors.Wrap(err, "не удалось десериализовать сообщение")
	}

	return retCtx, &avatarMessage, nil
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
