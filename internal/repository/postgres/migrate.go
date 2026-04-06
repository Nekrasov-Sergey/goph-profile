// Package postgres реализует хранилище данных на базе PostgreSQL.
package postgres

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// migrateDB применяет миграции к базе данных.
func migrateDB(databaseDSN string, logger zerolog.Logger) error {
	logger.Info().Msg("Запуск миграций базы данных...")

	m, err := migrate.New("file://migrations", databaseDSN)
	if err != nil {
		return errors.Wrap(err, "не удалось инициализировать миграции")
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info().Msg("Миграции не требуются — база данных уже актуальна")
		} else {
			return errors.Wrap(err, "не удалось применить миграции")
		}
	} else {
		logger.Info().Msg("Миграции успешно применены")
	}

	return nil
}
