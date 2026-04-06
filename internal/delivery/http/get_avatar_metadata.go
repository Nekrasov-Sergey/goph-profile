package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apitypes "github.com/oapi-codegen/runtime/types"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/gen"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

func (s *Server) GetAvatarMetadata(c *gin.Context, avatarId apitypes.UUID) {
	ctx := c.Request.Context()

	avatar, err := s.service.GetAvatarMetadata(ctx, avatarId)
	if err != nil {
		if errors.Is(err, errcodes.ErrAvatarNotFound) {
			respondError(c, errcodes.ErrAvatarNotFound, http.StatusNotFound)
			return
		}
		respondError(c, err, http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	response := api.GetAvatarMetadataResponse{
		Id:        avatar.ID,
		UserId:    avatar.UserID,
		FileName:  avatar.FileName,
		MimeType:  avatar.MimeType,
		Size:      avatar.SizeBytes,
		CreatedAt: avatar.CreatedAt,
		UpdatedAt: avatar.UpdatedAt,
		Dimensions: api.ImageDimensions{
			Width:  0, // todo: получать реальный размер при обработке worker'ом
			Height: 0,
		},
	}

	// todo: добавить thumbnails когда worker будет готов
	// if len(avatar.ThumbnailS3Keys) > 0 {
	//     thumbnails := make([]api.Thumbnail, 0)
	//     response.Thumbnails = &thumbnails
	// }

	c.JSON(http.StatusOK, response)
}
