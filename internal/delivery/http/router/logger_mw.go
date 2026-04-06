package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// LoggerMiddleware добавляет zerolog в gin
func LoggerMiddleware(baseLogger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		l := baseLogger.With().
			Str("method", c.Request.Method).
			Str("url", c.Request.URL.Path).
			Str("req_id", uuid.NewString()[:8]).
			Logger()

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		l = l.With().
			Int("status", status).
			Str("duration", time.Since(start).String()).
			Int("size", c.Writer.Size()).
			Logger()

		if len(c.Errors) > 0 || status >= 400 {
			errLog := l.Error()
			if status >= 500 {
				errLog = errLog.Stack()
			}
			for _, e := range c.Errors {
				errLog.Err(e.Err)
			}
			errLog.Msg("Ошибка выполнения http запроса")
			return
		}

		l.Info().Msg("HTTP запрос успешно выполнен")
	}
}
