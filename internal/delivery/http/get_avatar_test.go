package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
	"github.com/Nekrasov-Sergey/goph-profile/pkg/errcodes"
)

func TestGetAvatar(t *testing.T) {
	tests := []struct {
		name       string
		buildReq   func() *http.Request
		setup      func(env *testEnv)
		wantStatus int
		checkBody  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Успешное получение аватара",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				env.repo.GetAvatarMock.Return(avatar, nil)

				jpegData := newValidJPEG().Bytes()
				env.storage.DownloadMock.Return(nopReadCloser(jpegData), nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Body.Len() == 0 {
					t.Error("тело ответа не должно быть пустым")
				}
				contentType := w.Header().Get("Content-Type")
				if contentType != "image/jpeg" {
					t.Errorf("Content-Type: получили %s, ожидали image/jpeg", contentType)
				}
				cacheControl := w.Header().Get("Cache-Control")
				if cacheControl != "max-age=86400" {
					t.Errorf("Cache-Control: получили %s, ожидали max-age=86400", cacheControl)
				}
			},
		},
		{
			name: "Аватар не найден",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
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
			name: "Ошибка скачивания из хранилища",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				env.repo.GetAvatarMock.Return(avatar, nil)
				env.storage.DownloadMock.Return(nil, errors.New("S3 недоступно"))
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
		{
			name: "Получение с параметром size",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000?size=100x100", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				avatar.ProcessingStatus = types.ProcessingStatusCompleted
				thumbnailKeys := map[types.ThumbnailSize]string{
					types.ThumbnailSize100: "550e8400/thumbnail_100x100.jpg",
				}
				avatar.ThumbnailS3Keys, _ = json.Marshal(thumbnailKeys)
				env.repo.GetAvatarMock.Return(avatar, nil)

				jpegData := newValidJPEG().Bytes()
				env.storage.DownloadMock.Return(nopReadCloser(jpegData), nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Body.Len() == 0 {
					t.Error("тело ответа не должно быть пустым")
				}
			},
		},
		{
			name: "Получение с конвертацией формата JPEG→PNG",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000?format=png", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				env.repo.GetAvatarMock.Return(avatar, nil)

				// storage.Download возвращает оригинальный JPEG,
				// ChangeMimeType декодирует его и перекодирует в PNG
				jpegData := newValidJPEG().Bytes()
				env.storage.DownloadMock.Return(nopReadCloser(jpegData), nil)
			},
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Body.Len() == 0 {
					t.Error("тело ответа не должно быть пустым")
				}
				contentType := w.Header().Get("Content-Type")
				if contentType != "image/png" {
					t.Errorf("Content-Type: получили %s, ожидали image/png", contentType)
				}
			},
		},
		{
			name: "Миниатюра ещё не готова",
			buildReq: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000?size=100x100", nil)
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				avatar.ProcessingStatus = types.ProcessingStatusPending
				env.repo.GetAvatarMock.Return(avatar, nil)
			},
			wantStatus: http.StatusInternalServerError,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				// ErrThumbnailNotReady мапится в 500, сообщение скрыто
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
