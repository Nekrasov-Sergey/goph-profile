# Директория для сборки
BUILD_DIR ?= build

# Первичная сборка проекта
.PHONY: first-build
first-build: gen build-server

# Собрать сервер
.PHONY: build-server
build-server:
	@echo "🔧 Сборка сервера..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/server ./cmd/server

# Собрать и запустить сервер
.PHONY: server
server: build-server
	@./$(BUILD_DIR)/server

# Собрать воркер
.PHONY: build-worker
build-worker:
	@echo "🔧 Сборка воркера..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/worker ./cmd/worker

# Собрать и запустить воркер
.PHONY: worker
worker: build-worker
	@./$(BUILD_DIR)/worker

# Генерация кода из OpenAPI спецификации
.PHONY: openapi
openapi: clean
	mkdir -p internal/delivery/http/openapi
	oapi-codegen -package api \
		-generate "models,gin-server" \
		-o internal/delivery/http/openapi/generated.go \
		api/rest/swagger.yaml

# Очистка сгенерированных файлов
.PHONY: clean
clean:
	rm -rf internal/delivery/http/openapi

# Установка зависимостей
.PHONY: deps
deps:
	go get github.com/gin-gonic/gin
	go get github.com/oapi-codegen/runtime
