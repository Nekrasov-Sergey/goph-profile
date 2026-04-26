package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/tracer"
)

// respondError возвращает ошибку в формате ErrorResponse и записывает ошибку в span.
func respondError(c *gin.Context, err error, status int) {
	// Записываем ошибку в активный span трейсинга
	if span := trace.SpanFromContext(c.Request.Context()); span.IsRecording() {
		_ = tracer.SpanError(span, err)
	}

	if err == nil {
		c.AbortWithStatus(status)
		return
	}

	if status == http.StatusForbidden || status >= http.StatusInternalServerError {
		c.AbortWithStatusJSON(status, api.ErrorResponse{
			Error: http.StatusText(status),
		})
	} else {
		c.AbortWithStatusJSON(status, api.ErrorResponse{
			Error: err.Error(),
		})
	}

	_ = c.Error(err)
}
