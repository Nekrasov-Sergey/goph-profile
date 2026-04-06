package postgres

import (
	"context"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/dbutils"
)

// UpdateAvatar обновляет аватар.
func (p *Postgres) UpdateAvatar(ctx context.Context, avatar *types.Avatar) error {
	// todo перенести на уровень сервиса когда воркер будет обновлять миниатюры
	//if len(avatar.ThumbnailS3Keys) > 0 {
	//	thumbnailKeysJSON, err := json.Marshal(avatar.ThumbnailS3Keys)
	//	if err != nil {
	//		return errors.Wrap(err, "не удалось сериализовать ключи миниатюр")
	//	}
	//	avatar.ThumbnailS3Keys = thumbnailKeysJSON
	//}

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
	return dbutils.NamedExec(ctx, p.db, query, avatar)
}
