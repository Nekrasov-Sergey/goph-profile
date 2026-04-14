package http_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	pkghttp "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http"
	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service"
	"github.com/Nekrasov-Sergey/goph-profile/internal/service/mocks"
	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// testEnv содержит окружение для тестов.
type testEnv struct {
	repo     *mocks.RepoMock
	storage  *mocks.StorageMock
	producer *mocks.ProducerMock
	router   *gin.Engine
}

// setupTestServer создаёт тестовое окружение с моками инфраструктуры.
func setupTestServer(t minimock.Tester) *testEnv {
	gin.SetMode(gin.TestMode)

	repoMock := mocks.NewRepoMock(t)
	storageMock := mocks.NewStorageMock(t)
	producerMock := mocks.NewProducerMock(t)

	svc := service.New(repoMock, storageMock, producerMock, nil, zerolog.Nop())

	r := gin.New()
	httpSrv := pkghttp.New(r, ":0", svc, zerolog.Nop())

	r.GET("/health", httpSrv.HealthCheck)
	api.RegisterHandlersWithOptions(r, httpSrv, api.GinServerOptions{
		BaseURL: "/api/v1",
	})

	return &testEnv{
		repo:     repoMock,
		storage:  storageMock,
		producer: producerMock,
		router:   r,
	}
}

// newValidJPEG создаёт минимальное валидное JPEG-изображение 1x1 пиксель.
func newValidJPEG() *bytes.Buffer {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	_ = jpeg.Encode(&buf, img, nil)
	return &buf
}

// newTestAvatar создаёт тестовый аватар с заполненными полями.
func newTestAvatar() *types.Avatar {
	now := time.Now()
	return &types.Avatar{
		ID:               uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		UserID:           "user-123",
		FileName:         "avatar.jpg",
		MimeType:         types.MIMETypeJPEG,
		SizeBytes:        1024,
		Width:            100,
		Height:           100,
		S3Key:            "550e8400-e29b-41d4-a716-446655440000/original.jpg",
		ProcessingStatus: types.ProcessingStatusCompleted,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// createUploadRequest создаёт multipart/form-data POST-запрос для загрузки аватара.
func createUploadRequest(userID, fieldName string, fileContent []byte, fileName string) *http.Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		panic(err)
	}
	_, _ = part.Write(fileContent)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/avatars", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", userID)

	return req
}

// doRequest выполняет HTTP-запрос и возвращает ответ.
func doRequest(r *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// readCloser обёртка для io.Reader, реализующая io.ReadCloser.
type readCloser struct {
	reader io.Reader
}

func (rc readCloser) Read(p []byte) (n int, err error) { return rc.reader.Read(p) }
func (rc readCloser) Close() error                     { return nil }

// nopReadCloser возвращает io.ReadCloser, который читает из bytes.Reader.
func nopReadCloser(data []byte) io.ReadCloser {
	return readCloser{reader: bytes.NewReader(data)}
}
