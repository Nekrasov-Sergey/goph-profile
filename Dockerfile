# Этап сборки
FROM golang:1.26.1-alpine3.23 AS builder

# Установка зависимостей для сборки (включая libwebp для CGO)
RUN apk add --no-cache git ca-certificates tzdata build-base libwebp-dev

WORKDIR /app

# Копирование go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка бинарников с CGO (требуется для github.com/chai2010/webp)
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /worker ./cmd/worker

# Финальный этап
FROM alpine:3.23

# Установка необходимых пакетов
RUN apk add --no-cache ca-certificates tzdata libwebp

# Создание пользователя для безопасности
RUN adduser -D -g '' appuser

WORKDIR /app

# Копирование бинарников из этапа сборки
COPY --from=builder /server .
COPY --from=builder /worker .

# Копирование конфигурации
COPY config/ config/

# Копирование миграций
COPY migrations/ migrations/

# Копирование статики
COPY web/ web/

# Установка прав
RUN chown -R appuser:appuser /app

USER appuser

# Порт по умолчанию для server
EXPOSE 8080
