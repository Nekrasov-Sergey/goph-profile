package logger

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	sdklog "go.opentelemetry.io/otel/sdk/log"

	"go.opentelemetry.io/otel/sdk/resource"
)

// newLoggerProvider создаёт OTel LoggerProvider
func newLoggerProvider(ctx context.Context) (*sdklog.LoggerProvider, error) {
	exporter, err := otlploggrpc.New(ctx, otlploggrpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать log exporter")
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать resource")
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	), nil
}
