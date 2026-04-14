package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// New создаёт и настраивает gin-роутер с middleware.
func New(logger zerolog.Logger, mode string) (*gin.Engine, error) {
	gin.SetMode(mode)

	r := gin.New()

	r.Use(gin.Recovery(), LoggerMiddleware(logger))

	// Раздача статических файлов фронтенда
	r.Static("/static", "./web/static")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/static/index.html")
	})

	return r, nil
}
