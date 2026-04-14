package worker

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// Service определяет интерфейс бизнес-логики для worker.
type Service interface {
	GetAvatarMessage(ctx context.Context) (*types.AvatarMessage, error)
	CreateThumbnails(ctx context.Context, msg *types.AvatarMessage) error
	DeleteAvatarFromS3(ctx context.Context, msg *types.AvatarMessage) error
}

// Worker обрабатывает сообщения из Kafka и выполняет операции над аватарами.
type Worker struct {
	service Service
	logger  zerolog.Logger
}

// New создаёт новый экземпляр worker.
func New(service Service, logger zerolog.Logger) *Worker {
	return &Worker{service: service, logger: logger}
}

// Run запускает цикл обработки сообщений из Kafka.
func (w *Worker) Run(ctx context.Context) {
	w.logger.Info().Msg("Worker запущен")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := w.service.GetAvatarMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				w.logger.Error().Err(err).Send()
				continue
			}

			switch msg.Operation {
			case types.OperationCreateThumbnails:
				if err := w.service.CreateThumbnails(ctx, msg); err != nil {
					w.logger.Error().Err(err).
						Str("avatar_id", msg.AvatarID.String()).
						Msg("Не удалось создать миниатюры аватарки")
					continue
				}
				w.logger.Info().
					Str("avatar_id", msg.AvatarID.String()).
					Msg("Миниатюры аватарки успешно созданы")

			case types.OperationDeleteFromS3:
				if err := w.service.DeleteAvatarFromS3(ctx, msg); err != nil {
					w.logger.Error().Err(err).
						Str("avatar_id", msg.AvatarID.String()).
						Msg("Не удалось удалить аватарку из S3")
					continue
				}
				w.logger.Info().
					Str("avatar_id", msg.AvatarID.String()).
					Msg("Аватарка и её миниатюры успешно удалены из S3")

			default:
				w.logger.Error().Msgf("Неизвестная операция: %s", msg.Operation)
				continue
			}
		}
	}
}
