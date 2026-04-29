package metrics

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	serviceinfo "github.com/Nekrasov-Sergey/goph-profile/pkg/service_info"
)

func New(ctx context.Context) (func(), error) {
	exporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать metric exporter")
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithOS(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceinfo.ServiceName),
			semconv.ServiceVersionKey.String(serviceinfo.ServiceVersion),
			attribute.String("environment", os.Getenv("GO_ENV")),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать resource")
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
				metric.WithInterval(2*time.Second),
			),
		),
	)
	otel.SetMeterProvider(meterProvider)

	return func() {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := meterProvider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}, nil
}
