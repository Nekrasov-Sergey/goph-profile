// Package logger содержит настройку логирования приложения.
package logger

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// newConsoleWriter задаёт консольный вывод
func newConsoleWriter() zerolog.ConsoleWriter {
	keys := []string{"method", "url", "req_id", "status", "duration", "size", "stack"}

	return zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
		FormatExtra: func(m map[string]any, b *bytes.Buffer) error {
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
}
