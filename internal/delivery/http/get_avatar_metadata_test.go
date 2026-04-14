package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

func TestGetAvatarMetadata(t *testing.T) {
	tests := []struct {
		name       string
		buildReq   func() *http.Request
		setup      func(env *testEnv)
		wantStatus int
		checkBody  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Успешное получение метаданных без миниатюр",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000/metadata", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				avatar.ThumbnailS3Keys = nil
				env.repo.GetAvatarMock.Return(avatar, nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.GetAvatarMetadataResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.UserId != "user-123" {
					t.Errorf("user_id: получили %s, ожидали user-123", resp.UserId)
				}
				if resp.FileName != "avatar.jpg" {
					t.Errorf("file_name: получили %s, ожидали avatar.jpg", resp.FileName)
				}
				if resp.MimeType != "image/jpeg" {
					t.Errorf("mime_type: получили %s, ожидали image/jpeg", resp.MimeType)
				}
				if resp.Dimensions.Width != 100 || resp.Dimensions.Height != 100 {
					t.Errorf("dimensions: получили %dx%d, ожидали 100x100", resp.Dimensions.Width, resp.Dimensions.Height)
				}
				if resp.Thumbnails != nil {
					t.Error("thumbnails должен быть nil")
				}
			},
		},
		{
			name: "Аватар не найден",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000/metadata", nil)
			},
			setup: func(env *testEnv) {
				env.repo.GetAvatarMock.Return(nil, errcodes.ErrAvatarNotFound)
			},
			wantStatus: http.StatusNotFound,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Error != "аватар не найден" {
					t.Errorf("error: получили %s, ожидали 'аватар не найден'", resp.Error)
				}
			},
		},
		{
			name: "Получение метаданных с миниатурами",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000/metadata", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				thumbnailKeys := map[string]string{
					"100x100": "avatars/550e8400/thumb_100x100.jpg",
					"300x300": "avatars/550e8400/thumb_300x300.jpg",
				}
				avatar.ThumbnailS3Keys, _ = json.Marshal(thumbnailKeys)
				env.repo.GetAvatarMock.Return(avatar, nil)
				env.storage.GetURLMock.Return("http://minio:9000/bucket/avatars/thumb.jpg")
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.GetAvatarMetadataResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Thumbnails == nil || len(*resp.Thumbnails) == 0 {
					t.Fatal("thumbnails не должен быть пустым")
				}
				if len(*resp.Thumbnails) != 2 {
					t.Errorf("количество миниатюр: получили %d, ожидали 2", len(*resp.Thumbnails))
				}
			},
		},
		{
			name: "Ошибка репозитория",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000/metadata", nil)
			},
			setup: func(env *testEnv) {
				env.repo.GetAvatarMock.Return(nil, errors.New("внутренняя ошибка"))
			},
			wantStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Error != http.StatusText(http.StatusInternalServerError) {
					t.Errorf("error: получили %s, ожидали скрытое сообщение", resp.Error)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			env := setupTestServer(mc)

			tt.setup(env)

			req := tt.buildReq()
			w := doRequest(env.router, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("статус: получили %d, ожидали %d, тело: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.checkBody != nil {
				tt.checkBody(t, w)
			}
		})
	}
}
