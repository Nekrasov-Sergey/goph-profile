package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apitypes "github.com/oapi-codegen/runtime/types"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/utils"
)

// GetAvatarMetadata обрабатывает получение метаданных аватара по ID.
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
		MimeType:  string(avatar.MimeType),
		Size:      avatar.SizeBytes,
		CreatedAt: avatar.CreatedAt,
		UpdatedAt: avatar.UpdatedAt,
		Dimensions: api.ImageDimensions{
			Width:  avatar.Width,
			Height: avatar.Height,
		},
	}

	if len(avatar.ThumbnailS3Keys) > 0 {
		// Десериализуем ключи миниатюр
		var thumbnailKeys map[string]string
		if err := json.Unmarshal(avatar.ThumbnailS3Keys, &thumbnailKeys); err == nil {
			thumbnails := make([]api.Thumbnail, 0, len(thumbnailKeys))
			for size, key := range thumbnailKeys {
				thumbnails = append(thumbnails, api.Thumbnail{
					Size: size,
					Url:  s.service.GetURL(key),
				})
			}
			response.Thumbnails = utils.Ptr(thumbnails)
		}
	}

	c.JSON(http.StatusOK, response)
}
