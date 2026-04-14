package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
	"github.com/Nekrasov-Sergey/goph-profile/internal/infra/db/postgres"
	"github.com/Nekrasov-Sergey/goph-profile/internal/infra/messaging/kafka"
	"github.com/Nekrasov-Sergey/goph-profile/internal/infra/storage/minio"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/internal/worker"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/logger"
)

// main — точка входа worker.
func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Worker завершился с ошибкой")
	}
}

// run инициализирует зависимости и запускает worker.
func run() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	l := logger.New()

	cfg, err := config.NewWorkerConfig(l)
	if err != nil {
		return err
	}

	psql, err := postgres.New(cfg.DatabaseDSN, l)
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(psql))

	minIO, err := minio.New(ctx, cfg.MinIO, l)
	if err != nil {
		return err
	}

	consumer, err := kafka.NewConsumer(ctx, l, cfg.Kafka)
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(consumer))

	svc := service.New(psql, minIO, nil, consumer, l)

	w := worker.New(svc, l)

	startWorker(ctx, w, l)

	return nil
}

// startWorker запускает worker и ожидает сигнал завершения.
func startWorker(ctx context.Context, w *worker.Worker, l zerolog.Logger) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	<-ctx.Done()
	l.Info().Msg("Получен сигнал завершения")

	wg.Wait()
	l.Info().Msg("Воркер остановлен")
}
