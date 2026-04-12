package postgres

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/dbutils"
)

// UpdateAvatar обновляет аватар.
func (p *Postgres) UpdateAvatar(ctx context.Context, avatar *types.Avatar) error {
	const query = `update avatars
set user_id           = :user_id,
    file_name         = :file_name,
    mime_type         = :mime_type,
    size_bytes        = :size_bytes,
    s3_key            = :s3_key,
    thumbnail_s3_keys = :thumbnail_s3_keys,
    processing_status = :processing_status,
    updated_at        = :updated_at
where id = :id
  AND deleted_at is null
	`

	if err := dbutils.NamedExec(ctx, p.db, query, avatar); err != nil {
		return errors.Wrap(err, "не удалось обновить аватар")
	}
	return nil
}
