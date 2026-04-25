package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/logger"
)

// options — параметры HTTP-сервера, настраиваемые через функциональные опции.
type options struct {
	httpHandler http.Handler
	httpAddress string
}

// Option — функциональная опция для Server.
type Option func(*options)

// WithHTTPHandler устанавливает HTTP-обработчик.
func WithHTTPHandler(httpHandler http.Handler) Option {
	return func(o *options) {
		o.httpHandler = httpHandler
	}
}

// WithHTTPAddress устанавливает адрес прослушивания.
func WithHTTPAddress(httpAddress string) Option {
	return func(o *options) {
		o.httpAddress = httpAddress
	}
}

// Service определяет интерфейс бизнес-логики для HTTP handlers.
type Service interface {
	CreateAvatar(ctx context.Context, req service.UploadAvatarRequest) (*service.UploadAvatarResponse, error)
	GetAvatar(ctx context.Context, req service.GetAvatarRequest) (*service.GetAvatarResponse, error)
	GetAvatarMetadata(ctx context.Context, avatarID uuid.UUID) (*types.Avatar, error)
	DeleteAvatarFromDB(ctx context.Context, req service.DeleteAvatarRequest) error
	HealthCheck(ctx context.Context) *service.HealthCheckResult
	GetURL(s3Key string) string
}

// Server инкапсулирует http-сервер приложения.
type Server struct {
	logger  zerolog.Logger
	server  *http.Server
	service Service
}

// New создаёт новый экземпляр HTTP-сервера.
func New(l zerolog.Logger, service Service, opts ...Option) *Server {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	wrappedHandler := otelhttp.NewHandler(
		o.httpHandler,
		logger.ServiceName,
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)

	return &Server{
		logger: l,
		server: &http.Server{
			Handler: wrappedHandler,
			Addr:    o.httpAddress,
		},
		service: service,
	}
}

// Run запускает HTTP-сервер.
func (s *Server) Run() error {
	s.logger.Info().Msgf("HTTP-сервер запущен на %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "ошибка при запуске сервера")
	}
	return nil
}

// Shutdown корректно останавливает сервер.
func (s *Server) Shutdown(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info().Msg("Запущен graceful shutdown HTTP-сервера")
		errCh <- s.server.Shutdown(ctx)
	}()

	select {
	case <-ctx.Done():
		if err := s.server.Close(); err != nil {
			return errors.Wrap(err, "ошибка при принудительной остановке HTTP-сервера")
		}
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return errors.Wrap(err, "ошибка при graceful shutdown HTTP-сервера")
		}
		s.logger.Info().Msg("HTTP-сервер корректно остановлен")
		return nil
	}
}
