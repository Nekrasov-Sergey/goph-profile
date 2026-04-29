// Package metrics предоставляет метрики приложения на базе OpenTelemetry.
//
// Все метрики экспортируются через OTLP HTTP в OpenTelemetry Collector,
// который форвардит их в Prometheus.
package metrics

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	serviceinfo "github.com/Nekrasov-Sergey/goph-profile/pkg/service_info"
)

// Instruments содержит все метрические инструменты приложения.
// Создаётся через NewInstruments и передаётся во все слои приложения.
// Потокобезопасен — все инструменты горутинобезопасны.
type Instruments struct {
	// --- HTTP метрики ---
	HTTPRequestsTotal metric.Int64Counter
	HTTPRequestErrors metric.Int64Counter

	// --- Бизнес-метрики аватаров ---
	AvatarUploadCount        metric.Int64Counter
	AvatarUploadSizeBytes    metric.Int64Histogram
	AvatarProcessingDuration metric.Float64Histogram
	AvatarDeleteCount        metric.Int64Counter
	AvatarThumbnailCount     metric.Int64Counter
	AvatarThumbnailSizeBytes metric.Int64Histogram

	// --- Метрики БД ---
	DBOperationDuration metric.Float64Histogram
	DBOperationErrors   metric.Int64Counter

	// --- Метрики S3 ---
	S3OperationCount    metric.Int64Counter
	S3OperationDuration metric.Float64Histogram
	S3OperationErrors   metric.Int64Counter
	S3OperationSize     metric.Int64Histogram

	// --- Метрики Kafka ---
	KafkaMessageCount    metric.Int64Counter
	KafkaOperationErrors metric.Int64Counter
	KafkaOperationDuration metric.Float64Histogram
}

// NewInstruments создаёт и регистрирует все метрические инструменты.
// Использует глобальный MeterProvider, установленный через metrics.New().
func NewInstruments() (*Instruments, error) {
	meter := otel.Meter(serviceinfo.ServiceName,
		metric.WithInstrumentationVersion(serviceinfo.ServiceVersion),
	)

	instr := &Instruments{}
	var err error

	// --- HTTP метрики ---
	instr.HTTPRequestsTotal, err = meter.Int64Counter(
		"http.server.requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return nil, err
	}

	instr.HTTPRequestErrors, err = meter.Int64Counter(
		"http.server.request.errors_total",
		metric.WithDescription("Total number of HTTP request errors (status >= 400)"),
	)
	if err != nil {
		return nil, err
	}

	// --- Бизнес-метрики ---
	instr.AvatarUploadCount, err = meter.Int64Counter(
		"avatar.upload.count",
		metric.WithDescription("Total number of avatar uploads"),
	)
	if err != nil {
		return nil, err
	}

	instr.AvatarUploadSizeBytes, err = meter.Int64Histogram(
		"avatar.upload.size_bytes",
		metric.WithDescription("Size of uploaded avatars in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	instr.AvatarProcessingDuration, err = meter.Float64Histogram(
		"avatar.processing.duration",
		metric.WithDescription("Duration of avatar thumbnail processing"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	instr.AvatarDeleteCount, err = meter.Int64Counter(
		"avatar.delete.count",
		metric.WithDescription("Total number of avatar deletions"),
	)
	if err != nil {
		return nil, err
	}

	instr.AvatarThumbnailCount, err = meter.Int64Counter(
		"avatar.thumbnail.count",
		metric.WithDescription("Total number of generated thumbnails"),
	)
	if err != nil {
		return nil, err
	}

	instr.AvatarThumbnailSizeBytes, err = meter.Int64Histogram(
		"avatar.thumbnail.size_bytes",
		metric.WithDescription("Size of generated thumbnails in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// --- Метрики БД ---
	instr.DBOperationDuration, err = meter.Float64Histogram(
		"db.operation.duration",
		metric.WithDescription("Database operation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	instr.DBOperationErrors, err = meter.Int64Counter(
		"db.operation.errors_total",
		metric.WithDescription("Total number of database operation errors"),
	)
	if err != nil {
		return nil, err
	}

	// --- Метрики S3 ---
	instr.S3OperationCount, err = meter.Int64Counter(
		"s3.operation.count",
		metric.WithDescription("Total number of S3 operations"),
	)
	if err != nil {
		return nil, err
	}

	instr.S3OperationDuration, err = meter.Float64Histogram(
		"s3.operation.duration",
		metric.WithDescription("S3 operation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	instr.S3OperationErrors, err = meter.Int64Counter(
		"s3.operation.errors_total",
		metric.WithDescription("Total number of S3 operation errors"),
	)
	if err != nil {
		return nil, err
	}

	instr.S3OperationSize, err = meter.Int64Histogram(
		"s3.operation.size_bytes",
		metric.WithDescription("Size of S3 objects in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return nil, err
	}

	// --- Метрики Kafka ---
	instr.KafkaMessageCount, err = meter.Int64Counter(
		"kafka.messages.count",
		metric.WithDescription("Total number of Kafka messages"),
	)
	if err != nil {
		return nil, err
	}

	instr.KafkaOperationErrors, err = meter.Int64Counter(
		"kafka.operation.errors_total",
		metric.WithDescription("Total number of Kafka operation errors"),
	)
	if err != nil {
		return nil, err
	}

	instr.KafkaOperationDuration, err = meter.Float64Histogram(
		"kafka.operation.duration",
		metric.WithDescription("Kafka operation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return instr, nil
}
