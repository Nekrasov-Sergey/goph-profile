package postgres

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// recordDBMetrics записывает метрики выполнения операции с БД.
func (p *Postgres) recordDBMetrics(ctx context.Context, operation string, err error, dur time.Duration) {
	if p.meter == nil {
		return
	}

	status := "ok"
	if err != nil {
		status = "error"
	}

	attrs := []attribute.KeyValue{
		metrics.AttrDBOperation(operation),
		metrics.AttrDBStatus(status),
	}

	p.meter.DBOperationDuration.Record(ctx, dur.Seconds(), metric.WithAttributes(attrs...))

	if err != nil {
		p.meter.DBOperationErrors.Add(ctx, 1,
			metric.WithAttributes(metrics.AttrDBOperation(operation)),
		)
	}
}
