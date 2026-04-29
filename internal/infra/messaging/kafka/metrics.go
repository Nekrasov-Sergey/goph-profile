package kafka

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// recordKafkaProducerMetrics записывает метрики отправки сообщения в Kafka.
func (p *Producer) recordKafkaProducerMetrics(ctx context.Context, err error, dur time.Duration) {
	if p.meter == nil {
		return
	}

	p.meter.KafkaMessageCount.Add(ctx, 1,
		metric.WithAttributes(
			metrics.AttrMessagingDirection("produced"),
			attribute.String("messaging.destination.name", p.writer.Topic),
		),
	)

	p.meter.KafkaOperationDuration.Record(ctx, dur.Seconds(),
		metric.WithAttributes(metrics.AttrMessagingOp("send")),
	)

	if err != nil {
		p.meter.KafkaOperationErrors.Add(ctx, 1,
			metric.WithAttributes(metrics.AttrMessagingOp("send")),
		)
	}
}

// recordKafkaConsumerMetrics записывает метрики чтения сообщения из Kafka.
func (c *Consumer) recordKafkaConsumerMetrics(ctx context.Context, err error, dur time.Duration) {
	if c.meter == nil {
		return
	}

	c.meter.KafkaMessageCount.Add(ctx, 1,
		metric.WithAttributes(
			metrics.AttrMessagingDirection("consumed"),
			attribute.String("messaging.destination.name", c.reader.Config().Topic),
		),
	)

	c.meter.KafkaOperationDuration.Record(ctx, dur.Seconds(),
		metric.WithAttributes(metrics.AttrMessagingOp("receive")),
	)

	if err != nil {
		c.meter.KafkaOperationErrors.Add(ctx, 1,
			metric.WithAttributes(metrics.AttrMessagingOp("receive")),
		)
	}
}
