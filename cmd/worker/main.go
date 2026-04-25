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

	l, otelShutdown, err := logger.New(ctx, "worker")
	if err != nil {
		return err
	}
	defer otelShutdown()

	cfg, err := config.NewWorkerConfig(l)
	if err != nil {
		return err
	}

	psql, err := postgres.New(l, postgres.WithDatabaseDSN(cfg.DatabaseDSN))
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(psql))

	minIO, err := minio.New(ctx, l, minio.WithMinIOCfg(cfg.MinIO))
	if err != nil {
		return err
	}

	consumer, err := kafka.NewConsumer(ctx, l, kafka.WithKafkaCfg(cfg.Kafka))
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(consumer))

	svc := service.New(l, psql, minIO, service.WithConsumer(consumer))

	w := worker.New(l, svc)

	startWorker(ctx, l, w)

	return nil
}

// startWorker запускает worker и ожидает сигнал завершения.
func startWorker(ctx context.Context, l zerolog.Logger, w *worker.Worker) {
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
