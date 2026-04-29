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
	$(DOCKER_COMPOSE) up -d db minio kafka kafka-ui otel-collector loki grafana prometheus node_exporter alertmanager

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

# === Логи ===

# Логи сервера
.PHONY: logs-server
logs-server:
	$(DOCKER_COMPOSE) logs -f server

# Логи воркера
.PHONY: logs-worker
logs-worker:
	$(DOCKER_COMPOSE) logs -f worker

# Логи всех сервисов
.PHONY: logs-all
logs-all:
	$(DOCKER_COMPOSE) logs -f

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

# === Алертинг ===

# Проверка синтаксиса правил и конфигов алертинга
.PHONY: alert-check
alert-check:
	@echo "🔍 Проверка правил алертинга..."
	@docker run --rm --entrypoint promtool -v $(PWD)/deploy/prometheus:/rules prom/prometheus:latest \
		check rules /rules/alert-rules.yml && \
		echo "  ✓ Правила алертинга (Prometheus) корректны" || \
		(echo "  ✗ Ошибка в правилах алертинга"; exit 1)
	@docker run --rm --entrypoint amtool -v $(PWD)/deploy/prometheus:/config prom/alertmanager:latest \
		check-config /config/alertmanager.yml && \
		echo "  ✓ Конфигурация Alertmanager корректна" || \
		(echo "  ✗ Ошибка в конфигурации Alertmanager"; exit 1)
	@echo "✅ Все проверки пройдены"

# Отправка тестовых алертов в Alertmanager (+ генерация трафика для Prometheus)
# SEVERITY=all — алерты + трафик + статус
# SEVERITY=traffic — только генерация HTTP-трафика для проверки Prometheus-правил
.PHONY: alert-test
alert-test:
	@ALERTMANAGER_URL=$${ALERTMANAGER_URL:-http://localhost:9093} \
		PROMETHEUS_URL=$${PROMETHEUS_URL:-http://localhost:9090} \
		SERVER_URL=$${SERVER_URL:-http://localhost:8080} \
		./scripts/test-alert.sh ${SEVERITY}

# Показать статус системы алертинга
.PHONY: alert-status
alert-status:
	@echo "📊 Статус системы алертинга..."
	@ALERTMANAGER_URL=$${ALERTMANAGER_URL:-http://localhost:9093} \
		PROMETHEUS_URL=$${PROMETHEUS_URL:-http://localhost:9090} \
		./scripts/test-alert.sh status

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
	@echo "  Логи:"
	@echo "    make logs-server   - Логи сервера"
	@echo "    make logs-worker   - Логи воркера"
	@echo "    make logs-all      - Логи всех сервисов"
	@echo ""
	@echo "  Алертинг:"
	@echo "    make alert-check   - Проверить синтаксис правил и конфигов алертинга"
	@echo "    make alert-test    - Отправить тестовый алерт (SEVERITY=warning|critical|all|traffic)"
	@echo "    make alert-status  - Показать статус системы алертинга"
	@echo ""
	@echo "  Разработка:"
	@echo "    make gen           - Генерация кода"
	@echo "    make test          - Запустить тесты"
	@echo "    make test-cover    - Тесты с покрытием"
	@echo "    make lint          - Запустить линтер"
	@echo "    make lint-fix      - Запустить линтер с исправлениями"
	@echo "    make fmt           - Форматировать код"
	@echo "    make deps          - Скачать зависимости"
