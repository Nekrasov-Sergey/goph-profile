package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

const ServiceName = "avatar-service"

func New(ctx context.Context, component string) (zerolog.Logger, func(), error) {
	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		return zerolog.Logger{}, nil, err
	}

	zerolog.ErrorStackMarshaler = func(err error) any {
		return fmt.Sprintf("%+v", err)
	}

	logger := zerolog.New(&otelWriter{
		console: newConsoleWriter(),
		logger:  loggerProvider.Logger(ServiceName),
	}).
		With().
		Timestamp().
		Str("component", component).
		Logger().
		Hook(&traceHook{})

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := loggerProvider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	return logger, shutdown, nil
}
