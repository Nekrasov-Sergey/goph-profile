package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojuno/minimock/v3"

	api "github.com/Nekrasov-Sergey/goph-profile/internal/delivery/http/openapi"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(env *testEnv)
		wantStatus int
		wantBody   api.HealthResponse
	}{
		{
			name: "Успешная проверка здоровья",
			setup: func(env *testEnv) {
				env.repo.PingMock.Return(nil)
				env.storage.PingMock.Return(nil)
				env.producer.PingMock.Return(nil)
			},
			wantStatus: http.StatusOK,
			wantBody: api.HealthResponse{
				Status: api.HealthResponseStatusHealthy,
				Components: struct {
					Database api.ComponentHealth `json:"database"`
					Kafka    api.ComponentHealth `json:"kafka"`
					Storage  api.ComponentHealth `json:"storage"`
				}{
					Database: api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
					Kafka:    api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
					Storage:  api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
				},
			},
		},
		{
			name: "База данных недоступна",
			setup: func(env *testEnv) {
				env.repo.PingMock.Return(errors.New("соединение разорвано"))
				env.storage.PingMock.Return(nil)
				env.producer.PingMock.Return(nil)
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: api.HealthResponse{
				Status: api.HealthResponseStatusUnhealthy,
				Components: struct {
					Database api.ComponentHealth `json:"database"`
					Kafka    api.ComponentHealth `json:"kafka"`
					Storage  api.ComponentHealth `json:"storage"`
				}{
					Database: api.ComponentHealth{Status: api.ComponentHealthStatusUnhealthy},
					Kafka:    api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
					Storage:  api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
				},
			},
		},
		{
			name: "Хранилище недоступно",
			setup: func(env *testEnv) {
				env.repo.PingMock.Return(nil)
				env.storage.PingMock.Return(errors.New("минио упал"))
				env.producer.PingMock.Return(nil)
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: api.HealthResponse{
				Status: api.HealthResponseStatusUnhealthy,
				Components: struct {
					Database api.ComponentHealth `json:"database"`
					Kafka    api.ComponentHealth `json:"kafka"`
					Storage  api.ComponentHealth `json:"storage"`
				}{
					Database: api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
					Kafka:    api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
					Storage:  api.ComponentHealth{Status: api.ComponentHealthStatusUnhealthy},
				},
			},
		},
		{
			name: "Kafka недоступна",
			setup: func(env *testEnv) {
				env.repo.PingMock.Return(nil)
				env.storage.PingMock.Return(nil)
				env.producer.PingMock.Return(errors.New("брокер недоступен"))
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody: api.HealthResponse{
				Status: api.HealthResponseStatusUnhealthy,
				Components: struct {
					Database api.ComponentHealth `json:"database"`
					Kafka    api.ComponentHealth `json:"kafka"`
					Storage  api.ComponentHealth `json:"storage"`
				}{
					Database: api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
					Kafka:    api.ComponentHealth{Status: api.ComponentHealthStatusUnhealthy},
					Storage:  api.ComponentHealth{Status: api.ComponentHealthStatusHealthy},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := minimock.NewController(t)
			env := setupTestServer(mc)

			tt.setup(env)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := doRequest(env.router, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("статус: получили %d, ожидали %d", w.Code, tt.wantStatus)
			}

			var resp api.HealthResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("не удалось десериализовать ответ: %v", err)
			}

			if resp.Status != tt.wantBody.Status {
				t.Errorf("статус: получили %s, ожидали %s", resp.Status, tt.wantBody.Status)
			}
			if resp.Components.Database.Status != tt.wantBody.Components.Database.Status {
				t.Errorf("database: получили %s, ожидали %s", resp.Components.Database.Status, tt.wantBody.Components.Database.Status)
			}
			if resp.Components.Storage.Status != tt.wantBody.Components.Storage.Status {
				t.Errorf("storage: получили %s, ожидали %s", resp.Components.Storage.Status, tt.wantBody.Components.Storage.Status)
			}
			if resp.Components.Kafka.Status != tt.wantBody.Components.Kafka.Status {
				t.Errorf("kafka: получили %s, ожидали %s", resp.Components.Kafka.Status, tt.wantBody.Components.Kafka.Status)
			}
		})
	}
}
