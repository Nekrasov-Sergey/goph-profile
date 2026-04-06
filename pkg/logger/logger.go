// Package logger содержит настройку логирования приложения.
package logger

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// New настраивает и возвращает логгер с форматированным выводом.
func New() zerolog.Logger {
	keys := []string{"method", "url", "req_id", "status", "duration", "size", "stack"}
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		FormatExtra: func(m map[string]interface{}, b *bytes.Buffer) error {
			for _, key := range keys {
				if val, ok := m[key]; ok {
					if _, err := fmt.Fprintf(b, " \033[36m%s\033[0m=%v", key, val); err != nil {
						return err
					}
				}
			}
			return nil
		},
		FieldsExclude: keys,
	}

	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		return fmt.Sprintf("%+v", err)
	}

	logger := zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger()

	log.Logger = logger
	return logger
}
