package worker

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

const tracerName = "avatar-service/worker"

// Service определяет интерфейс бизнес-логики для worker.
type Service interface {
	GetAvatarMessage(ctx context.Context) (context.Context, *types.AvatarMessage, error)
	CreateThumbnails(ctx context.Context, msg *types.AvatarMessage) error
	DeleteAvatarFromS3(ctx context.Context, msg *types.AvatarMessage) error
}

// Worker обрабатывает сообщения из Kafka и выполняет операции над аватарами.
type Worker struct {
	logger  zerolog.Logger
	tracer  trace.Tracer
	service Service
}

// New создаёт новый экземпляр worker.
func New(logger zerolog.Logger, service Service) *Worker {
	return &Worker{
		logger:  logger.With().Str("component", "worker").Logger(),
		tracer:  otel.Tracer(tracerName),
		service: service,
	}
}

// Run запускает цикл обработки сообщений из Kafka.
func (w *Worker) Run(ctx context.Context) {
	w.logger.Info().Msg("Worker запущен")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgCtx, msg, err := w.service.GetAvatarMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				w.logger.Error().Err(err).Send()
				continue
			}

			// Создаём дочерний span от восстановленного trace context.
			// SpanKindConsumer указывает Jaeger, что это асинхронный consumer span,
			// и предупреждение о clock skew не показывается.
			msgCtx, span := w.tracer.Start(msgCtx, fmt.Sprintf("process_%s", msg.Operation),
				trace.WithSpanKind(trace.SpanKindConsumer),
			)
			err = w.processMessage(msgCtx, msg)
			span.End()

			if err != nil {
				w.logger.Error().Err(err).
					Str("avatar_id", msg.AvatarID.String()).
					Str("operation", string(msg.Operation)).
					Msg("Не удалось обработать сообщение")
				continue
			}

			w.logger.Info().
				Str("avatar_id", msg.AvatarID.String()).
				Str("operation", string(msg.Operation)).
				Msg("Сообщение успешно обработано")
		}
	}
}

// processMessage выполняет операцию над аватаром.
func (w *Worker) processMessage(ctx context.Context, msg *types.AvatarMessage) error {
	switch msg.Operation {
	case types.OperationCreateThumbnails:
		return w.service.CreateThumbnails(ctx, msg)
	case types.OperationDeleteFromS3:
		return w.service.DeleteAvatarFromS3(ctx, msg)
	default:
		return errors.Errorf("неизвестная операция: %s", msg.Operation)
	}
}
