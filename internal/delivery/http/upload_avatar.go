package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/gen"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

func (s *Server) UploadAvatar(c *gin.Context, params api.UploadAvatarParams) {
	ctx := c.Request.Context()

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		respondError(c, errors.New("файл отсутствует"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	req := service.UploadAvatarRequest{
		UserID:   params.XUserID,
		File:     file,
		FileName: header.Filename,
		Size:     header.Size,
	}

	resp, err := s.service.UploadAvatar(ctx, req)
	if err != nil {
		if errors.Is(err, errcodes.ErrFileTooLarge) {
			respondError(c, errcodes.ErrFileTooLarge, http.StatusRequestEntityTooLarge)
			return
		}
		if errors.Is(err, errcodes.ErrInvalidFormat) {
			respondError(c, errcodes.ErrInvalidFormat, http.StatusBadRequest)
			return
		}
		respondError(c, err, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, api.UploadAvatarResponse{
		Id:        resp.ID,
		UserId:    resp.UserID,
		Url:       resp.URL,
		Status:    api.UploadAvatarResponseStatus(resp.Status),
		CreatedAt: resp.CreatedAt,
	})
}
