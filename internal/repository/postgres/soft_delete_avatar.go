package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/dbutils"
)

// SoftDeleteAvatar выполняет мягкое удаление аватара.
func (p *Postgres) SoftDeleteAvatar(ctx context.Context, id uuid.UUID, userID string) error {
	const query = `update avatars
set deleted_at = now(),
    updated_at = now()
where id = :id
  and user_id = :user_id
  and deleted_at is null
	`

	args := map[string]any{
		"id":      id,
		"user_id": userID,
	}

	if err := dbutils.NamedExec(ctx, p.db, query, args); err != nil {
		return errors.Wrap(err, "не удалось пометить удалённым аватар")
	}

	return nil
}
