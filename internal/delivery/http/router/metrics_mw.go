package router

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/Nekrasov-Sergey/goph-profile/pkg/metrics"
)

// MetricsMiddleware возвращает Gin-middleware для внедрения http.route в метрики otelhttp
// и записи дополнительных HTTP-метрик.
//
// otelhttp уже записывает http.server.request.duration и прочие метрики,
// но не знает о маршрутах Gin. Middleware добавляет http.route через Labeler
// и записывает счётчики запросов и ошибок.
func MetricsMiddleware(meter *metrics.Instruments) gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		ctx := c.Request.Context()
		if labeler, ok := otelhttp.LabelerFromContext(ctx); ok {
			labeler.Add(attribute.String("http.route", route))
		}

		c.Next()

		if meter == nil {
			return
		}

		status := c.Writer.Status()

		attrs := metric.WithAttributes(
			metrics.AttrHTTPMethod(c.Request.Method),
			metrics.AttrHTTPRoute(route),
			metrics.AttrHTTPStatus(status),
		)

		meter.HTTPRequestsTotal.Add(c.Request.Context(), 1, attrs)

		if status >= 400 {
			meter.HTTPRequestErrors.Add(c.Request.Context(), 1, attrs)
		}
	}
}
