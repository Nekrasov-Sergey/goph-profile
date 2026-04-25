package logger

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

// traceHook извлекает span context из Go-контекста события и добавляет
// trace_id / span_id в поля JSON, чтобы они попали в OTel как атрибуты.
type traceHook struct{}

func (h *traceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.HasTraceID() {
		e.Str("trace_id", spanCtx.TraceID().String())
	}
	if spanCtx.HasSpanID() {
		e.Str("span_id", spanCtx.SpanID().String())
	}
}
