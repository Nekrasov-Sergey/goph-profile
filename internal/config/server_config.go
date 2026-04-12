// Package config содержит конфигурацию клиента и сервера.
package config

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// ServerConfig содержит конфигурацию сервера.
type ServerConfig struct {
	HTTPAddr    string
	DatabaseDSN string
	MinIO       MinIO
	Kafka       Kafka
}

// Kafka содержит конфигурацию Kafka.
type Kafka struct {
	Brokers []string
	Topic   string
	GroupID string
}

// MinIO содержит конфигурацию S3 хранилища.
type MinIO struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// GetServerConfigPath возвращает путь к файлу конфигурации сервера.
func GetServerConfigPath() string {
	c := os.Getenv("SERVER_CONFIG_PATH")
	if c == "" {
		c = "./config/server.yml"
	}
	return c
}

// NewServerConfig загружает конфигурацию сервера из файла.
func NewServerConfig(logger zerolog.Logger) (*ServerConfig, error) {
	viper.SetConfigFile(GetServerConfigPath())

	cfg := ServerConfig{
		HTTPAddr: ":8080",
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "не удалось прочитать конфигурацию из файла")
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить конфигурацию в структуру")
	}

	logger.Info().
		Str("HTTPAddr", cfg.HTTPAddr).
		Str("DatabaseDSN", cfg.DatabaseDSN).
		Str("MinIO.Endpoint", cfg.MinIO.Endpoint).
		Str("MinIO.Bucket", cfg.MinIO.Bucket).
		Bool("MinIO.UseSSL", cfg.MinIO.UseSSL).
		Strs("Kafka.Brokers", cfg.Kafka.Brokers).
		Str("Kafka.Topic", cfg.Kafka.Topic).
		Msg("Загружена конфигурация сервера")

	return &cfg, nil
}
