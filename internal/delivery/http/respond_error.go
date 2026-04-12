package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
)

// respondError возвращает ошибку в формате ErrorResponse
func respondError(c *gin.Context, err error, status int) {
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
