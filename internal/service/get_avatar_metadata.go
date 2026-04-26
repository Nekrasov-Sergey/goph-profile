package service

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

// GetAvatarMetadata получает метаданные аватара без самого файла.
func (s *Service) GetAvatarMetadata(ctx context.Context, avatarID uuid.UUID) (*types.Avatar, error) {
	ctx, span := s.tracer.Start(ctx, "service.GetAvatarMetadata",
		trace.WithAttributes(
			attribute.String("avatar.id", avatarID.String()),
		),
	)
	defer span.End()

	avatar, err := s.repo.GetAvatar(ctx, avatarID)
	if err != nil {
		return nil, tracer.SpanError(span, err)
	}
	return avatar, nil
}
