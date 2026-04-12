package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// GetAvatarMetadata получает метаданные аватара без самого файла.
func (s *Service) GetAvatarMetadata(ctx context.Context, avatarID uuid.UUID) (*types.Avatar, error) {
	return s.repo.GetAvatar(ctx, avatarID)
}
