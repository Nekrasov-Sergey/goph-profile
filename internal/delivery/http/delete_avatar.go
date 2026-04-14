package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apitypes "github.com/oapi-codegen/runtime/types"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

// DeleteAvatar обрабатывает удаление аватара по ID.
func (s *Server) DeleteAvatar(c *gin.Context, avatarId apitypes.UUID, params api.DeleteAvatarParams) {
	ctx := c.Request.Context()

	req := service.DeleteAvatarRequest{
		AvatarID: avatarId,
		UserID:   params.XUserID,
	}

	err := s.service.DeleteAvatarFromDB(ctx, req)
	if err != nil {
		if errors.Is(err, errcodes.ErrAvatarNotFound) {
			respondError(c, errcodes.ErrAvatarNotFound, http.StatusNotFound)
			return
		}
		if errors.Is(err, errcodes.ErrAccessDenied) {
			respondError(c, errcodes.ErrAccessDenied, http.StatusForbidden)
			return
		}
		respondError(c, err, http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}
