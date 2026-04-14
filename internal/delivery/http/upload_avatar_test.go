package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
)

func TestUploadAvatar(t *testing.T) {
	tests := []struct {
		name       string
		buildReq   func() *http.Request
		setup      func(env *testEnv)
		wantStatus int
		checkBody  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Успешная загрузка аватара",
			buildReq: func() *http.Request {
				jpegData := newValidJPEG().Bytes()
				return createUploadRequest("user-123", "image", jpegData, "avatar.jpg")
			},
			setup: func(env *testEnv) {
				env.storage.UploadMock.Return(nil)
				env.repo.CreateAvatarMock.Return(nil)
				env.producer.SendMessageMock.Return(nil)
				env.storage.GetURLMock.Return("http://minio:9000/bucket/avatars/avatar.jpg")
			},
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.UploadAvatarResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.UserId != "user-123" {
					t.Errorf("user_id: получили %s, ожидали user-123", resp.UserId)
				}
				if resp.Url == "" {
					t.Error("url не должен быть пустым")
				}
				if resp.Status != api.Pending {
					t.Errorf("status: получили %s, ожидали pending", resp.Status)
				}
				if resp.Id == uuid.Nil {
					t.Error("id не должен быть нулевым UUID")
				}
			},
		},
		{
			name: "Файл отсутствует в запросе",
			buildReq: func() *http.Request {
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)
				_ = writer.Close()
				req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", &buf)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("X-User-ID", "user-123")
				return req
			},
			setup:      func(env *testEnv) {},
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Error != "файл отсутствует" {
					t.Errorf("error: получили %s, ожидали 'файл отсутствует'", resp.Error)
				}
			},
		},
		{
			name: "Неверный формат файла",
			buildReq: func() *http.Request {
				textData := []byte("это не изображение, а просто текстовый файл")
				return createUploadRequest("user-123", "image", textData, "file.txt")
			},
			setup:      func(env *testEnv) {},
			wantStatus: http.StatusBadRequest,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Error != "неверный формат файла. Поддерживаемые форматы: jpeg, png, webp" {
					t.Errorf("error: получили %s", resp.Error)
				}
			},
		},
		{
			name: "Файл слишком большой",
			buildReq: func() *http.Request {
				largeData := bytes.Repeat([]byte{0xFF}, 10*1024*1024+1)
				return createUploadRequest("user-123", "image", largeData, "large.jpg")
			},
			setup:      func(env *testEnv) {},
			wantStatus: http.StatusRequestEntityTooLarge,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Error != "файл слишком большой. Максимум: 10MB" {
					t.Errorf("error: получили %s", resp.Error)
				}
			},
		},
		{
			name: "Ошибка загрузки в хранилище",
			buildReq: func() *http.Request {
				jpegData := newValidJPEG().Bytes()
				return createUploadRequest("user-123", "image", jpegData, "avatar.jpg")
			},
			setup: func(env *testEnv) {
				env.storage.UploadMock.Return(errors.New("S3 недоступно"))
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
			name: "Ошибка создания записи в БД",
			buildReq: func() *http.Request {
				jpegData := newValidJPEG().Bytes()
				return createUploadRequest("user-123", "image", jpegData, "avatar.jpg")
			},
			setup: func(env *testEnv) {
				env.storage.UploadMock.Return(nil)
				env.repo.CreateAvatarMock.Return(errors.New("БД недоступна"))
				env.storage.DeleteMock.Return(nil) // откат S3
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
			name: "Ошибка отправки в Kafka",
			buildReq: func() *http.Request {
				jpegData := newValidJPEG().Bytes()
				return createUploadRequest("user-123", "image", jpegData, "avatar.jpg")
			},
			setup: func(env *testEnv) {
				env.storage.UploadMock.Return(nil)
				env.repo.CreateAvatarMock.Return(nil)
				env.producer.SendMessageMock.Return(errors.New("брокер недоступен"))
				env.repo.UpdateAvatarMock.Return(nil) // обновление статуса на failed
				env.storage.GetURLMock.Return("http://minio:9000/bucket/avatars/avatar.jpg")
			},
			wantStatus: http.StatusCreated,
			checkBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp api.UploadAvatarResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("не удалось десериализовать ответ: %v", err)
				}
				if resp.Status != api.Failed {
					t.Errorf("status: получили %s, ожидали failed (ошибка Kafka не прерывает загрузку)", resp.Status)
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
