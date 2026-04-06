package postgres

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/dbutils"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

// CreateAvatar создаёт новую запись аватара в базе данных.
func (p *Postgres) CreateAvatar(ctx context.Context, avatar *types.Avatar) error {
	const q = `insert into avatars (id, user_id, file_name, mime_type, size_bytes, s3_key,
                     thumbnail_s3_keys, processing_status, created_at, updated_at)
values (:id, :user_id, :file_name, :mime_type, :size_bytes, :s3_key, :thumbnail_s3_keys,
        :processing_status, :created_at, :updated_at)
        `

	if err := dbutils.NamedExec(ctx, p.db, q, avatar); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return errcodes.ErrAvatarIDAlreadyExists
		}
		return errors.Wrap(err, "не удалось создать запись аватара")
	}

	return nil
}
