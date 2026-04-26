package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"

	serviceinfo "github.com/Nekrasov-Sergey/goph-profile/pkg/service_info"
)

func New(ctx context.Context) (zerolog.Logger, func(), error) {
	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		return zerolog.Logger{}, nil, err
	}

	zerolog.ErrorStackMarshaler = func(err error) any {
		return fmt.Sprintf("%+v", err)
	}

	logger := zerolog.New(&otelWriter{
		console: newConsoleWriter(),
		logger:  loggerProvider.Logger(serviceinfo.ServiceName),
	}).
		With().
		Timestamp().
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
