package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apitypes "github.com/oapi-codegen/runtime/types"
	"go.uber.org/multierr"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/utils"
)

// GetAvatar обрабатывает получение файла аватара по ID с опциональным размером и форматом.
func (s *Server) GetAvatar(c *gin.Context, avatarId apitypes.UUID, params api.GetAvatarParams) {
	ctx := c.Request.Context()

	req := service.GetAvatarRequest{
		ID: avatarId,
	}

	if params.Size != nil {
		req.Size = types.ThumbnailSize(utils.Deref(params.Size))
	}
	if params.Format != nil {
		req.Format = types.ImageFormat(utils.Deref(params.Format))
	}

	resp, err := s.service.GetAvatar(ctx, req)
	if err != nil {
		if errors.Is(err, errcodes.ErrAvatarNotFound) {
			respondError(c, errcodes.ErrAvatarNotFound, http.StatusNotFound)
			return
		}
		respondError(c, err, http.StatusInternalServerError)
		return
	}
	defer multierr.AppendInvoke(&err, multierr.Close(resp.Reader))

	// Устанавливаем заголовки
	c.Header("Content-Type", string(resp.MimeType))
	c.Header("Cache-Control", "max-age=86400")

	// Stream'им файл в ответ
	c.DataFromReader(http.StatusOK, resp.Size, string(resp.MimeType), resp.Reader, nil)
}
