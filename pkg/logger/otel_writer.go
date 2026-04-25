package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/log"
)

// otelWriter пишет в консоль (через zerolog.ConsoleWriter) и отправляет
// структурированный лог в OpenTelemetry со всеми пользовательскими полями.
type otelWriter struct {
	console zerolog.ConsoleWriter
	logger  log.Logger
}

func (w *otelWriter) Write(p []byte) (int, error) {
	return w.console.Write(p)
}

func (w *otelWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	n, err := w.console.Write(p)
	if err != nil {
		return n, err
	}
	w.emitToOTel(level, p)
	return n, nil
}

func (w *otelWriter) emitToOTel(level zerolog.Level, p []byte) {
	var fields map[string]any
	if err := json.Unmarshal(p, &fields); err != nil {
		return
	}

	var record log.Record

	record.SetSeverity(otelSeverity(level))
	record.SetSeverityText(level.String())

	if t, ok := fields["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			record.SetTimestamp(parsed)
		}
	}

	if msg, ok := fields["message"].(string); ok {
		record.SetBody(log.StringValue(msg))
	}

	// Все поля, кроме служебных — как атрибуты OTel
	for key, val := range fields {
		switch key {
		case "level", "time", "message":
			continue
		}
		switch v := val.(type) {
		case string:
			record.AddAttributes(log.String(key, v))
		case float64:
			if v == float64(int64(v)) {
				record.AddAttributes(log.Int64(key, int64(v)))
			} else {
				record.AddAttributes(log.Float64(key, v))
			}
		case bool:
			record.AddAttributes(log.Bool(key, v))
		default:
			record.AddAttributes(log.String(key, fmt.Sprintf("%v", v)))
		}
	}

	w.logger.Emit(context.Background(), record)
}

func otelSeverity(level zerolog.Level) log.Severity {
	switch level {
	case zerolog.DebugLevel:
		return log.SeverityDebug
	case zerolog.InfoLevel:
		return log.SeverityInfo
	case zerolog.WarnLevel:
		return log.SeverityWarn
	case zerolog.ErrorLevel:
		return log.SeverityError
	case zerolog.PanicLevel:
		return log.SeverityFatal1
	case zerolog.FatalLevel:
		return log.SeverityFatal2
	default:
		return log.SeverityUndefined
	}
}
