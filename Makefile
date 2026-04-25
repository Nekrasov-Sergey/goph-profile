# === Переменные ===
BUILD_DIR ?= build
LDFLAGS ?= -s -w
DOCKER_COMPOSE ?= docker compose

# === Инициализация ===

# Первичная настройка проекта (копирование конфигов)
.PHONY: init
init:
	@echo "📝 Копирование конфигурационных файлов..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "  ✓ Создан .env"; \
	else \
		echo "  ⚠ .env уже существует, пропускаем"; \
	fi
	@if [ ! -f config/server.yaml ]; then \
		cp config/server.example.yaml config/server.yaml; \
		echo "  ✓ Создан config/server.yaml"; \
	else \
		echo "  ⚠ config/server.yaml уже существует, пропускаем"; \
	fi
	@if [ ! -f config/worker.yaml ]; then \
		cp config/worker.example.yaml config/worker.yaml; \
		echo "  ✓ Создан config/worker.yaml"; \
	else \
		echo "  ⚠ config/worker.yaml уже существует, пропускаем"; \
	fi
	@echo "✅ Инициализация завершена"

# Полная инициализация для нового разработчика
.PHONY: setup
setup: init tools gen deps build
	@echo "✅ Проект готов к работе"

# === Сборка ===

# Собрать все бинарники
.PHONY: build
build: gen build-server build-worker

# Собрать сервер
.PHONY: build-server
build-server:
	@echo "🔧 Сборка сервера..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/server ./cmd/server
	@echo "  ✓ $(BUILD_DIR)/server"

# Собрать воркер
.PHONY: build-worker
build-worker:
	@echo "🔧 Сборка воркера..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/worker ./cmd/worker
	@echo "  ✓ $(BUILD_DIR)/worker"

# === Запуск локально ===

# Запустить сервер
.PHONY: server
server: build-server
	@./$(BUILD_DIR)/server

# Запустить воркер
.PHONY: worker
worker: build-worker
	@./$(BUILD_DIR)/worker

# === Docker ===

# Собрать Docker-образ
.PHONY: docker-build
docker-build:
	@echo "🐳 Сборка Docker-образа..."
	$(DOCKER_COMPOSE) build

# Запустить все сервисы
.PHONY: docker-up
docker-up:
	@echo "🚀 Запуск Docker-контейнеров..."
	$(DOCKER_COMPOSE) up -d

# Запустить только инфраструктуру (без приложения)
.PHONY: docker-infra
docker-infra:
	@echo "🚀 Запуск инфраструктуры..."
	$(DOCKER_COMPOSE) up -d db minio kafka kafka-ui otel-collector loki grafana

# Остановить все контейнеры
.PHONY: docker-down
docker-down:
	@echo "🛑 Остановка Docker-контейнеров..."
	$(DOCKER_COMPOSE) down

# Остановить и удалить volumes
.PHONY: docker-clean
docker-clean:
	@echo "🧹 Очистка Docker (контейнеры + volumes)..."
	$(DOCKER_COMPOSE) down -v

# Перезапуск сервисов
.PHONY: docker-restart
docker-restart:
	$(DOCKER_COMPOSE) restart server worker

# Статус контейнеров
.PHONY: docker-ps
docker-ps:
	$(DOCKER_COMPOSE) ps

# === Генерация кода ===

# Генерация кода
.PHONY: gen
gen: gen-openapi gen-mocks

# Генерация кода из OpenAPI спецификации
.PHONY: gen-openapi
gen-openapi:
	@echo "⚙️  Генерация кода из OpenAPI..."
	@rm -rf internal/delivery/http/openapi
	@mkdir -p internal/delivery/http/openapi
	@oapi-codegen -package api \
		-generate "models,gin-server" \
		-o internal/delivery/http/openapi/generated.go \
		api/rest/swagger.yaml
	@echo "  ✓ internal/delivery/http/openapi/generated.go"

# Генерация моков для тестов
.PHONY: gen-mocks
gen-mocks:
	@echo "🎭  Генерация моков..."
	@rm -rf internal/service/mocks
	@mkdir -p internal/service/mocks
	@go generate ./...
	@echo "  ✓ internal/service/mocks"

# === Тестирование и линтеры ===

# Запустить тесты
.PHONY: test
test:
	@echo "🧪 Запуск тестов..."
	@go test -v -race ./...

# Запустить тесты с покрытием
.PHONY: test-cover
test-cover:
	@echo "🧪 Запуск тестов с покрытием..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "  ✓ coverage.html"

# Запустить линтер
.PHONY: lint
lint:
	@echo "🔍 Запуск линтера..."
	@golangci-lint run ./...

# Запустить линтер с исправлениями
.PHONY: lint-fix
lint-fix:
	@echo "🔍 Запуск линтера с исправлениями..."
	@golangci-lint run --fix ./...

# === Зависимости ===

# Установка зависимостей
.PHONY: deps
deps:
	@echo "📦 Установка зависимостей..."
	@go mod download
	@go mod tidy

# Обновление зависимостей
.PHONY: deps-update
deps-update:
	@echo "📦 Обновление зависимостей..."
	@go get -u ./...
	@go mod tidy

# Установка инструментов разработки
.PHONY: tools
tools:
	@echo "🔧 Установка инструментов разработки..."
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# === Утилиты ===

# Форматирование кода
.PHONY: fmt
fmt:
	@echo "📝 Форматирование кода..."
	@go fmt ./...

# === Справка ===

.PHONY: help
help:
	@echo "Доступные команды:"
	@echo ""
	@echo "  Инициализация:"
	@echo "    make init          - Копировать конфиги из example файлов"
	@echo "    make setup         - Полная инициализация (init + tools + gen + deps + build)"
	@echo ""
	@echo "  Сборка:"
	@echo "    make build         - Собрать все бинарники"
	@echo "    make server        - Собрать и запустить сервер"
	@echo "    make worker        - Собрать и запустить воркер"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-up     - Запустить все контейнеры"
	@echo "    make docker-down   - Остановить все контейнеры"
	@echo "    make docker-infra  - Запустить только инфраструктуру"
	@echo "    make docker-build  - Собрать Docker-образ"
	@echo "    make docker-clean  - Остановить и удалить volumes"
	@echo ""
	@echo "  Разработка:"
	@echo "    make gen           - Генерация кода"
	@echo "    make test          - Запустить тесты"
	@echo "    make lint          - Запустить линтер"
	@echo "    make fmt           - Форматировать код"
