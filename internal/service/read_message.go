package service

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// GetAvatarMessage читает и десериализует сообщение об аватаре из Kafka.
func (s *Service) GetAvatarMessage(ctx context.Context) (*types.AvatarMessage, error) {
	kafkaMessage, err := s.consumer.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var avatarMessage types.AvatarMessage
	if err := json.Unmarshal(kafkaMessage.Value, &avatarMessage); err != nil {
		return nil, errors.Wrap(err, "Не удалось десериализовать сообщение")
	}

	return &avatarMessage, nil
}
