// Package dbutils содержит утилиты для работы с базой данных.
package dbutils

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

const tracerName = "avatar-service/pkg/dbutils"

// NamedGet выполняет SQL-запрос с именованными параметрами и загружает одну запись в dest.
func NamedGet(ctx context.Context, db sqlx.ExtContext, dest any, q string, arg any) error {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "db.NamedGet",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", truncateSQL(q)),
		),
	)
	defer span.End()

	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedGet)")
	}

	return runWithRetries(ctx, func() error {
		if err := sqlx.GetContext(ctx, db, dest, nq, args...); err != nil {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedGet)")
		}
		return nil
	}, span)
}

// NamedSelect выполняет SQL-запрос с именованными параметрами и загружает результат в слайс dest.
func NamedSelect(ctx context.Context, db sqlx.ExtContext, dest any, q string, arg any) error {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "db.NamedSelect",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", truncateSQL(q)),
		),
	)
	defer span.End()

	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedSelect)")
	}

	return runWithRetries(ctx, func() error {
		if err := sqlx.SelectContext(ctx, db, dest, nq, args...); err != nil {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedSelect)")
		}
		return nil
	}, span)
}

// NamedExec выполняет SQL-запрос с именованными параметрами без возврата данных.
func NamedExec(ctx context.Context, db sqlx.ExtContext, q string, arg any) error {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "db.NamedExec",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", truncateSQL(q)),
		),
	)
	defer span.End()

	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedExec)")
	}

	return runWithRetries(ctx, func() error {
		if _, err := db.ExecContext(ctx, nq, args...); err != nil {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedExec)")
		}
		return nil
	}, span)
}

// runWithRetries выполняет функцию с повторными попытками при ошибках соединения.
func runWithRetries(ctx context.Context, fn func() error, span trace.Span) (err error) {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second, 0}
	for i, delay := range delays {
		if err = fn(); err == nil {
			return nil
		}

		if !isConnectionError(err) {
			return tracer.SpanError(span, err)
		}

		log.Error().Err(err).Msgf("Ошибка соединения c PostgreSQL, попытка №%d", i+1)
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Error().Msg("Запрос отменён контекстом во время ожидания")
				return tracer.SpanError(span, ctx.Err())
			case <-timer.C:
			}
		}
	}

	log.Error().Msg("Все попытки подключения исчерпаны")
	return tracer.SpanError(span, err)
}

// isConnectionError проверяет, является ли ошибка ошибкой соединения с PostgreSQL.
func isConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist, pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection, pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown, pgerrcode.ProtocolViolation, pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected:
			return true
		}
	}
	return false
}

// DB описывает минимальный интерфейс базы данных, необходимый для работы с транзакциями.
type DB interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// TxFunc описывает функцию, выполняемую внутри транзакции.
type TxFunc func(tx *sqlx.Tx) error

// truncateSQL обрезает SQL-запрос для атрибута span, оставляя первые 500 символов.
func truncateSQL(q string) string {
	if len(q) > 500 {
		return q[:500] + "..."
	}
	return q
}

// WrapTxx выполняет функцию внутри транзакции.
func WrapTxx(ctx context.Context, db DB, opts *sql.TxOptions, f TxFunc) (err error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "db.Transaction",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)
	defer func() {
		_ = tracer.SpanError(span, err)
		span.End()
	}()

	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "не удалось начать транзакцию")
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = multierr.Append(err, errors.Wrapf(rbErr, "не удалось выполнить Rollback"))
			}
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			err = errors.Wrap(commitErr, "не удалось зафиксировать транзакцию")
		}
	}()

	return f(tx)
}
