package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/goph-profile/internal/config"
	"github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http"
	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/router"
	"github.com/Nekrasov-Sergey/goph-profile/internal/infra/db/postgres"
	"github.com/Nekrasov-Sergey/goph-profile/internal/infra/messaging/kafka"
	"github.com/Nekrasov-Sergey/goph-profile/internal/infra/storage/minio"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/logger"
)

// main — точка входа HTTP-сервера.
func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Сервер завершился с ошибкой")
	}
}

// run инициализирует зависимости и запускает HTTP-сервер.
func run() (err error) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	l := logger.New()

	cfg, err := config.NewServerConfig(l)
	if err != nil {
		return err
	}

	r, err := router.New(l, gin.ReleaseMode)
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

	producer, err := kafka.NewProducer(ctx, l, cfg.Kafka)
	if err != nil {
		return err
	}
	defer multierr.AppendInvoke(&err, multierr.Close(producer))

	svc := service.New(l, psql, minIO, producer, nil)

	httpSrv := http.New(l, svc, http.WithHTTPHandler(r), http.WithHTTPAddress(cfg.HTTPAddr))

	// Регистрация маршрутов
	registerHandlers(r, httpSrv)

	// Запуск сервера
	return startServer(ctx, l, httpSrv)
}

// registerHandlers регистрирует все HTTP обработчики.
func registerHandlers(r *gin.Engine, httpSrv *http.Server) {
	// Healthcheck на /health (без префикса /api/v1)
	r.GET("/health", httpSrv.HealthCheck)

	// API endpoints с префиксом /api/v1
	api.RegisterHandlersWithOptions(r, httpSrv, api.GinServerOptions{
		BaseURL: "/api/v1",
	})
}

// startServer запускает HTTP сервер и обрабатывает graceful shutdown.
func startServer(ctx context.Context, l zerolog.Logger, httpSrv *http.Server) error {
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpSrv.Run(); err != nil {
			errCh <- err
		}
	}()

	var runErr error

	select {
	case <-ctx.Done():
		l.Info().Msg("Получен сигнал завершения")
	case err := <-errCh:
		runErr = err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	errChShutdown := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			errChShutdown <- err
		}
	}()

	wg.Wait()
	close(errChShutdown)

	for e := range errChShutdown {
		runErr = multierr.Append(runErr, e)
	}

	return runErr
}
