package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/dbutils"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

// GetAvatar получает аватар по ID.
func (p *Postgres) GetAvatar(ctx context.Context, avatarID uuid.UUID) (*types.Avatar, error) {
	const query = `select id,
       id,
       user_id,
       file_name,
       mime_type,
       size_bytes,
       width,
       height,
       s3_key,
       thumbnail_s3_keys,
       processing_status,
       created_at,
       updated_at,
       deleted_at
from avatars
where id = :id
  and deleted_at is null
	`

	args := map[string]any{
		"id": avatarID,
	}

	avatar := &types.Avatar{}
	if err := dbutils.NamedGet(ctx, p.db, avatar, query, args); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcodes.ErrAvatarNotFound
		}
		return nil, errors.Wrap(err, "не удалось получить аватар")
	}

	return avatar, nil
}
