package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
)

// HealthCheck обрабатывает проверку здоровья сервиса.
func (s *Server) HealthCheck(c *gin.Context) {
	ctx, span := s.tracer.Start(c.Request.Context(), "handler.HealthCheck",
		trace.WithAttributes(
			attribute.String("component", "healthcheck"),
		),
	)
	defer span.End()
	c.Request = c.Request.WithContext(ctx)

	result := s.service.HealthCheck(ctx)

	response := api.HealthResponse{
		Status: api.HealthResponseStatus(result.Status),
	}
	response.Components.Database = api.ComponentHealth{
		Status: api.ComponentHealthStatus(result.Components.Database),
	}
	response.Components.Storage = api.ComponentHealth{
		Status: api.ComponentHealthStatus(result.Components.Storage),
	}
	response.Components.Kafka = api.ComponentHealth{
		Status: api.ComponentHealthStatus(result.Components.Kafka),
	}

	statusCode := http.StatusOK
	if result.Status != service.StatusHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}
