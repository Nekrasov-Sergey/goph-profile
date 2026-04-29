// Package postgres реализует хранилище данных на базе PostgreSQL.
package postgres

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/dbutils"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// options — параметры подключения к БД, настраиваемые через функциональные опции.
type options struct {
	databaseDSN string
	meter       *metrics.Instruments
}

// Option — функциональная опция для Postgres.
type Option func(*options)

// WithDatabaseDSN устанавливает строку подключения.
func WithDatabaseDSN(databaseDSN string) Option {
	return func(o *options) {
		o.databaseDSN = databaseDSN
	}
}

// WithMeter задаёт метрические инструменты.
func WithMeter(meter *metrics.Instruments) Option {
	return func(o *options) {
		o.meter = meter
	}
}

// Postgres реализует хранилище данных на базе PostgreSQL.
type Postgres struct {
	db     sqlx.ExtContext
	rawDB  *sqlx.DB
	logger zerolog.Logger
	meter  *metrics.Instruments
}

// New создаёт новое подключение к базе данных и применяет миграции.
func New(logger zerolog.Logger, opts ...Option) (*Postgres, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if err := migrateDB(o.databaseDSN, logger); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("pgx", o.databaseDSN)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к БД")
	}

	logger.Info().Msg("Установлено подключение к БД")

	return &Postgres{
		db:     db,
		rawDB:  db,
		logger: logger,
		meter:  o.meter,
	}, nil
}

// Close закрывает соединение с базой данных.
func (p *Postgres) Close() error {
	if err := p.rawDB.Close(); err != nil {
		return errors.Wrap(err, "не удалось закрыть соединения с БД")
	}
	p.logger.Info().Msg("Закрыто соединение с БД")
	return nil
}

// WithTx выполняет функцию в рамках транзакции.
func (p *Postgres) WithTx(ctx context.Context, fn func(txRepo service.Repository) error) error {
	return dbutils.WrapTxx(ctx, p.rawDB, nil, func(tx *sqlx.Tx) error {
		txRepo := &Postgres{
			db:     tx,
			rawDB:  nil,
			logger: p.logger.With().Str("scope", "tx").Logger(),
		meter:  p.meter,
		}
		return fn(txRepo)
	})
}

// Ping проверяет доступность базы данных.
func (p *Postgres) Ping(ctx context.Context) (err error) {
	start := time.Now()
	defer func() {
		p.recordDBMetrics(ctx, "ping", err, time.Since(start))
	}()
	return p.rawDB.PingContext(ctx)
}
