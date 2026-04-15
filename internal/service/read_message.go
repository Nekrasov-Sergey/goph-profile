package service

import (
	"context"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// GetAvatarMessage читает следующее сообщение об аватаре из брокера.
func (s *Service) GetAvatarMessage(ctx context.Context) (*types.AvatarMessage, error) {
	return s.consumer.ReadAvatarMessage(ctx)
}
