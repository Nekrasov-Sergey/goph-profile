// Package config содержит конфигурацию клиента и сервера.
package config

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// WorkerConfig содержит конфигурацию worker.
type WorkerConfig struct {
	DatabaseDSN string
	MinIO       MinIO
	Kafka       Kafka
}

// GetWorkerConfigPath возвращает путь к файлу конфигурации worker.
func GetWorkerConfigPath() string {
	c := os.Getenv("WORKER_CONFIG_PATH")
	if c == "" {
		c = "./config/worker.yml"
	}
	return c
}

// NewWorkerConfig загружает конфигурацию worker из файла.
func NewWorkerConfig(logger zerolog.Logger) (*WorkerConfig, error) {
	viper.SetConfigFile(GetWorkerConfigPath())

	cfg := WorkerConfig{}

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "не удалось прочитать конфигурацию из файла")
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить конфигурацию в структуру")
	}

	logger.Info().
		Str("DatabaseDSN", cfg.DatabaseDSN).
		Str("MinIO.Endpoint", cfg.MinIO.Endpoint).
		Str("MinIO.Bucket", cfg.MinIO.Bucket).
		Bool("MinIO.UseSSL", cfg.MinIO.UseSSL).
		Strs("Kafka.Brokers", cfg.Kafka.Brokers).
		Str("Kafka.Topic", cfg.Kafka.Topic).
		Str("Kafka.GroupID", cfg.Kafka.GroupID).
		Msg("Загружена конфигурация воркера")

	return &cfg, nil
}
