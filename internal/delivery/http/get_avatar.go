package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apitypes "github.com/oapi-codegen/runtime/types"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/gen"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/utils"
)

func (s *Server) GetAvatar(c *gin.Context, avatarId apitypes.UUID, params api.GetAvatarParams) {
	ctx := c.Request.Context()

	req := service.GetAvatarRequest{
		ID: avatarId,
	}

	if params.Size != nil {
		req.Size = string(utils.Deref(params.Size))
	}
	if params.Format != nil {
		req.Format = string(utils.Deref(params.Format))
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
	defer resp.Reader.Close()

	// Устанавливаем заголовки
	c.Header("Content-Type", resp.MimeType)
	c.Header("Cache-Control", "max-age=86400")

	// Stream'им файл в ответ
	c.DataFromReader(http.StatusOK, resp.Size, resp.MimeType, resp.Reader, nil)
}
