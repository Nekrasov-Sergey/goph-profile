package service

import (
	"context"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// GetAvatarMessage читает следующее сообщение об аватаре из брокера.
// Возвращает контекст с восстановленным trace context из заголовков сообщения.
func (s *Service) GetAvatarMessage(ctx context.Context) (context.Context, *types.AvatarMessage, error) {
	return s.consumer.ReadAvatarMessage(ctx)
}
