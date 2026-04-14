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

func TestDeleteAvatar(t *testing.T) {
	tests := []struct {
		name       string
		buildReq   func() *http.Request
		setup      func(env *testEnv)
		wantStatus int
		checkBody  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Успешное удаление аватара",
			buildReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
				req.Header.Set("X-User-ID", "user-123")
				return req
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				env.repo.GetAvatarMock.Return(avatar, nil)
				env.repo.SoftDeleteAvatarMock.Return(nil)
				env.producer.SendMessageMock.Return(nil)
			},
			wantStatus: http.StatusNoContent,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Body.Len() != 0 {
					t.Errorf("тело ответа должно быть пустым при 204, получено: %s", w.Body.String())
				}
			},
		},
		{
			name: "Аватар не найден",
			buildReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
				req.Header.Set("X-User-ID", "user-123")
				return req
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
			name: "Доступ запрещён",
			buildReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
				req.Header.Set("X-User-ID", "user-123")
				return req
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				avatar.UserID = "other-user" // чужой аватар
				env.repo.GetAvatarMock.Return(avatar, nil)
			},
			wantStatus: http.StatusForbidden,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				// Для 403 сообщение скрывается
				if resp.Error != http.StatusText(http.StatusForbidden) {
					t.Errorf("error: получили %s, ожидали '%s'", resp.Error, http.StatusText(http.StatusForbidden))
				}
			},
		},
		{
			name: "Ошибка мягкого удаления",
			buildReq: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/api/v1/avatars/550e8400-e29b-41d4-a716-446655440000", nil)
				req.Header.Set("X-User-ID", "user-123")
				return req
			},
			setup: func(env *testEnv) {
				avatar := newTestAvatar()
				env.repo.GetAvatarMock.Return(avatar, nil)
				env.repo.SoftDeleteAvatarMock.Return(errors.New("ошибка БД"))
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
