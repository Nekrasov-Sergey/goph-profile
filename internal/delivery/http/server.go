package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

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
	server  *http.Server
	service Service
	logger  zerolog.Logger
}

// New создаёт новый экземпляр HTTP-сервера.
func New(handler http.Handler, addr string, service Service, logger zerolog.Logger) *Server {
	return &Server{
		server: &http.Server{
			Handler: handler,
			Addr:    addr,
		},
		service: service,
		logger:  logger,
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
