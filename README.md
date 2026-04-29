# Goph Profile

Сервис для управления аватарками пользователей: загрузка, получение с ресайзом и конвертацией формата, удаление.

Загруженное изображение сохраняется в S3-хранилище, а в Kafka отправляется сообщение для асинхронного создания
миниатюр (100x100 и 300x300). Worker читает сообщения из Kafka, генерирует миниатюры и обновляет статус обработки в БД.

Логи, метрики и трейсы приложения экспортируются через OpenTelemetry Collector: логи — в Grafana Loki,
метрики — в Prometheus, трейсы — в Jaeger.

## Архитектура

**Server** — HTTP-API. Принимает загрузку, отдаёт файлы (с ресайзом и конвертацией на лету), удаляет аватары, шлёт
сообщения в Kafka.

**Worker** — Консьюмер Kafka. Создаёт миниатюры и удаляет файлы из S3 по команде.

**PostgreSQL** — Хранит метаданные аватаров и статус обработки.

**MinIO** — S3-совместимое хранилище оригиналов и миниатюр.

**Kafka** — Брокер сообщений между Server и Worker.

**OpenTelemetry Collector** — Принимает логи (OTLP gRPC), метрики (OTLP HTTP) и трейсы (OTLP gRPC) от приложения,
маршрутизирует в Loki, Prometheus и Jaeger.

**Grafana Loki** — Хранилище структурированных логов.

**Prometheus** — Хранилище и система оповещения метрик (скрейпит OTel Collector на порту 8889).

**Grafana** — Дашборды для просмотра логов (Loki datasource по умолчанию) и метрик (Prometheus datasource).

**Jaeger** — Визуализация распределённых трассировок.

## API

Базовый путь: `/api/v1`

| Метод    | Путь                     | Описание                                                        |
|----------|--------------------------|-----------------------------------------------------------------|
| `GET`    | `/health`                | Проверка здоровья сервиса                                       |
| `POST`   | `/avatars`               | Загрузить аватар (`multipart/form-data`, заголовок `X-User-ID`) |
| `GET`    | `/avatars/{id}`          | Получить файл аватара (параметры: `size`, `format`)             |
| `DELETE` | `/avatars/{id}`          | Удалить аватар (заголовок `X-User-ID`, только свои)             |
| `GET`    | `/avatars/{id}/metadata` | Получить метаданные аватара                                     |

**Поддерживаемые форматы:** `jpeg`, `png`, `webp` (до 10 МБ).

**Параметры запроса `GET /avatars/{id}`:**

- `size` — `100x100`, `300x300` или без параметра (оригинал)
- `format` — `jpeg`, `png`, `webp` или без параметра (исходный формат)

Спецификация: [`api/rest/swagger.yaml`](api/rest/swagger.yaml).

## Стек

- **Go 1.26**, Gin, sqlx, pgx
- **PostgreSQL 17** — хранение метаданных
- **MinIO** — S3-хранилище файлов
- **Apache Kafka 4** (KRaft, без Zookeeper) — брокер сообщений
- **golang-migrate** — миграции БД
- **oapi-codegen** — генерация HTTP-кода из OpenAPI
- **minimock** — генерация моков
- **OpenTelemetry** — OTLP-экспорт логов, метрик и трейсов
- **OpenTelemetry Collector** — приём и маршрутизация телеметрии
- **Grafana Loki 3.6** — хранилище логов
- **Prometheus** — хранилище метрик
- **Grafana 12.3** — дашборды логов и метрик
- **Jaeger** — визуализация трейсов

## Быстрый старт

### Предварительные требования

- Go ≥ 1.26 (с CGO, нужен `gcc`)
- Docker + Docker Compose
- Make

### Инициализация проекта

```bash
make setup
```

Эта команда установит инструменты, скачает зависимости, сгенерирует код и соберёт бинарники.

## Запуск

Есть два варианта запуска: полностью в Docker или только инфраструктуру в Docker, а приложение локально.

### Вариант 1: Всё в Docker

Поднимает Server, Worker и всю инфраструктуру:

```bash
make docker-build   # собрать образ
make docker-up      # запустить
```

API будет доступен на `http://localhost:8080`.

Остальные сервисы:

| Сервис        | URL                                                 |
|---------------|-----------------------------------------------------|
| MinIO Console | http://localhost:9001 (`minioadmin` / `minioadmin`) |
| Kafka UI      | http://localhost:8088                               |
| Prometheus    | http://localhost:9090                               |
| Grafana       | http://localhost:3000                               |
| Jaeger        | http://localhost:16686                              |

### Вариант 2: Только инфраструктура в Docker, приложение локально

Поднимает PostgreSQL, MinIO, Kafka, Loki, Grafana, OTel Collector, Prometheus и node_exporter — без Server и Worker:

```bash
make docker-infra
```

Затем запускаешь сервер и worker в отдельных терминалах:

```bash
# Терминал 1 — сервер
make server

# Терминал 2 — worker
make worker
```

Они подключатся к инфраструктуре через `localhost` с портами из `config/server.yaml` и `config/worker.yaml`.

> Конфиги для Docker (`*-docker.yaml`) используют внутренние имена сервисов (`db`, `minio`, `kafka`), а локальные —
`localhost` с проброшенными портами.

## Остановка

```bash
# Остановить контейнеры (данные сохранятся)
make docker-down

# Остановить и удалить тома (полная очистка)
make docker-clean
```

## Основные команды Make

| Команда                 | Описание                                                          |
|-------------------------|-------------------------------------------------------------------|
| `make setup`            | Полная инициализация: инструменты, зависимости, генерация, сборка |
| `make build`            | Собрать бинарники server и worker                                 |
| `make server`           | Собрать и запустить сервер                                        |
| `make worker`           | Собрать и запустить worker                                        |
| `make docker-build`     | Собрать Docker-образ                                              |
| `make docker-up`        | Запустить всё в Docker                                            |
| `make docker-infra`     | Запустить инфраструктуру (БД, MinIO, Kafka, OTel, Loki, Grafana, Prometheus) |
| `make docker-down`      | Остановить контейнеры                                             |
| `make docker-clean`     | Остановить контейнеры и удалить тома                              |
| `make logs-server`      | Логи сервера                                                      |
| `make logs-worker`      | Логи воркера                                                      |
| `make test`             | Запустить тесты                                                   |
| `make test-cover`       | Тесты с покрытием (отчёт в `coverage.html`)                       |
| `make lint`             | Запустить golangci-lint                                           |
| `make gen`              | Сгенерировать код из OpenAPI + моки                               |
| `make deps`             | Скачать зависимости и выполнить `go mod tidy`                     |
| `make alert-check`      | Проверить синтаксис правил алертинга (Prometheus + Alertmanager)   |
| `make alert-test`       | Отправить тестовый алерт (SEVERITY=warning\|critical\|all\|traffic) |
| `make alert-status`     | Показать статус системы алертинга (правила + активные алерты)      |

Полный список: `make help`.

## Порты

| Сервис                  | Порт  | Назначение                   |
|-------------------------|-------|------------------------------|
| Server                  | 8080  | HTTP API                     |
| PostgreSQL              | 5555  | База данных                  |
| MinIO API               | 9000  | S3-совместимый API           |
| MinIO Console           | 9001  | Веб-интерфейс MinIO          |
| Kafka                   | 29092 | Внешний listener             |
| Kafka UI                | 8088  | Веб-интерфейс Kafka          |
| OTel Collector (gRPC)   | 4317  | OTLP gRPC (логи, трейсы)     |
| OTel Collector (HTTP)   | 4318  | OTLP HTTP (метрики)          |
| OTel Collector (метрики)| 8889  | Prometheus-эндпоинт (скрейп) |
| Prometheus              | 9090  | Веб-интерфейс Prometheus     |
| Loki                    | 3100  | HTTP API + OTLP              |
| Grafana                 | 3000  | Веб-интерфейс Grafana        |
| Jaeger                  | 16686 | Веб-интерфейс Jaeger         |
| node_exporter           | 9100  | Метрики хоста                |
| Alertmanager            | 9093  | Управление алертами          |

## Метрики

Приложение записывает три категории метрик через OpenTelemetry SDK и экспортирует их по OTLP HTTP
в OpenTelemetry Collector, который форвардит их в Prometheus.

### HTTP-метрики

| Метрика в Prometheus          | Тип        | Атрибуты                                                         |
|-------------------------------|------------|------------------------------------------------------------------|
| `http_server_requests_total`  | Counter    | `http.request.method`, `http.route`, `http.response.status_code` |
| `http_server_request_errors_total` | Counter | `http.request.method`, `http.route`, `http.response.status_code` |

### Бизнес-метрики

| Метрика в Prometheus                   | Тип        | Атрибуты                         |
|----------------------------------------|------------|----------------------------------|
| `avatar_upload_count_total`            | Counter    | `avatar.mime_type`, `avatar.status` |
| `avatar_upload_size_bytes_*`           | Histogram  | `avatar.mime_type`               |
| `avatar_processing_duration_*`         | Histogram  | `avatar.operation`, `avatar.status` |
| `avatar_delete_count_total`            | Counter    | `avatar.operation`, `avatar.status` |
| `avatar_thumbnail_count_total`         | Counter    | `avatar.mime_type`, `avatar.status` |
| `avatar_thumbnail_size_bytes_*`        | Histogram  | `avatar.mime_type`               |

### Инфраструктурные метрики

| Метрика в Prometheus                   | Тип        | Атрибуты                         |
|----------------------------------------|------------|----------------------------------|
| `db_operation_duration_*`              | Histogram  | `db.operation`, `db.status`      |
| `db_operation_errors_total`            | Counter    | `db.operation`                   |
| `s3_operation_count_total`             | Counter    | `s3.operation`, `s3.status`      |
| `s3_operation_duration_*`              | Histogram  | `s3.operation`                   |
| `s3_operation_errors_total`            | Counter    | `s3.operation`                   |
| `s3_operation_size_bytes_*`            | Histogram  | `s3.operation`                   |
| `kafka_messages_count_total`           | Counter    | `messaging.direction`, `messaging.destination.name` |
| `kafka_operation_errors_total`         | Counter    | `messaging.operation`            |
| `kafka_operation_duration_*`           | Histogram  | `messaging.operation`            |

> Для гистограмм Prometheus добавляет суффиксы `_bucket`, `_count`, `_sum`.

### Примеры PromQL-запросов

```promql
# RPS по всем маршрутам
rate(http_server_requests_total[5m])

# Ошибки HTTP в минуту
rate(http_server_request_errors_total[1m])

# Всего загружено аватаров
avatar_upload_count_total

# Загрузок в секунду по типу
rate(avatar_upload_count_total[5m])

# P50 длительности операций БД
histogram_quantile(0.5, rate(db_operation_duration_bucket[5m]))

# P99 длительности S3 upload
histogram_quantile(0.99, rate(s3_operation_duration_bucket{s3_operation="upload"}[5m]))

# Ошибки БД по типу операции
sum by (db_operation) (rate(db_operation_errors_total[5m]))

# Средний размер загружаемых аватаров
rate(avatar_upload_size_bytes_sum[5m]) / rate(avatar_upload_size_bytes_count[5m])

# Kafka сообщений в минуту
rate(kafka_messages_count_total[1m])
```

### Как посмотреть метрики

1. **Prometheus UI:** `http://localhost:9090` — перейти на вкладку **Graph**, ввести PromQL-запрос
2. **Grafana:** `http://localhost:3000` — добавить Prometheus datasource (`http://prometheus:9090`),
   затем строить дашборды или использовать **Explore**
3. **Сырые метрики:** `http://localhost:8889/metrics` — Prometheus-эндпоинт OTel Collector

## Алерты

Prometheus Alertmanager настроен на отслеживание критических показателей и автоматическое срабатывание при превышении порогов. Правила хранятся в [`deploy/prometheus/alert-rules.yml`](deploy/prometheus/alert-rules.yml) и оцениваются каждые 30 секунд.

### Правила алертов (7 critical, 6 warning)

| Алерт | Уровень | Условие | Длительность |
|-------|---------|---------|-------------|
| `HighHTTPErrorRate5xx` | critical | >5% 5xx ошибок за 5 мин | 3 мин |
| `ElevatedHTTPErrorRate4xx` | warning | >10% 4xx ошибок за 5 мин | 5 мин |
| `HighHTTPResponseTime` | critical | P99 >5 сек | 2 мин |
| `HighAvatarUploadErrorRate` | warning | >10% ошибок загрузки за 10 мин | 5 мин |
| `AvatarProcessingErrors` | warning | Любые ошибки обработки за 10 мин | 5 мин |
| `HighDBErrorRate` | critical | Любые ошибки БД за 5 мин | 2 мин |
| `SlowDBOperations` | warning | P99 >1 сек | 3 мин |
| `DeadDBCalls` | critical | P99 >10 сек | 1 мин |
| `HighS3ErrorRate` | critical | Любые ошибки S3 за 5 мин | 2 мин |
| `SlowS3Operations` | warning | P99 >5 сек | 3 мин |
| `HighKafkaErrorRate` | critical | Любые ошибки Kafka за 5 мин | 2 мин |
| `SlowKafkaOperations` | warning | P99 >10 сек | 3 мин |
| `NoMetrics` | critical | OTel Collector недоступен | 1 мин |

### Как посмотреть алерты

1. **Prometheus Alerts:** `http://localhost:9090/alerts` — все алерты с текущим статусом (Inactive/Pending/Firing)
2. **Alertmanager UI:** `http://localhost:9093` — сработавшие алерты, управление подавлением (silences)
3. **Grafana:** `http://localhost:3000` — добавить Alertmanager datasource (`http://alertmanager:9093`) для дашбордов

### Как тестировать алерты

Для проверки системы алертинга есть три make-команды:

```bash
# 1. Проверка синтаксиса правил и конфига Alertmanager
make alert-check

# 2. Отправка тестовых алертов напрямую в Alertmanager
make alert-test                # тестовый warning-алерт
make alert-test SEVERITY=all   # набор алертов + генерация трафика + статус

# 3. Генерация HTTP-трафика для проверки Prometheus-правил
make alert-test SEVERITY=traffic
```

**Как это работает:**

- `make alert-check` использует Docker-образы `prom/prometheus` и `prom/alertmanager` для валидации синтаксиса — не требует установки `promtool`/`amtool` локально.
- `make alert-test` без параметров отправляет тестовый алерт уровня warning в Alertmanager API (`POST /api/v2/alerts`). Алерт сразу появляется в Alertmanager UI (`http://localhost:9093`).
- `make alert-test SEVERITY=traffic` генерирует HTTP-запросы к серверу приложения (4xx, 2xx) — Prometheus собирает метрики и через 5–10 минут переводит правило `ElevatedHTTPErrorRate4xx` в статус `FIRING`.
- `make alert-test SEVERITY=all` выполняет всё вместе: отправляет тестовые алерты, генерирует трафик и показывает итоговый статус системы.

Тестовые алерты, отправленные через `make alert-test`, видны в Alertmanager UI сразу. Для появления алертов на странице Prometheus `/alerts` нужно, чтобы метрики превышали порог правила в течение заданного времени (`for:`).

## Логи

Логи приложения пишутся в консоль и одновременно экспортируются по OTLP gRPC в OpenTelemetry Collector, который
перенаправляет их в Loki. Каждая запись лога содержит `service.name`, `trace_id` и `span_id` для корреляции
с трассировками.

Логи можно фильтровать по уровню (`{service_name="avatar-service"} | severity_text="error"`), сервису и другим атрибутам.

### Примеры LogQL-запросов

```logql
# Все логи сервиса
{service_name="avatar-service"}

# Только ошибки
{service_name="avatar-service"} | severity_text="error"

# Ошибки в секунду
rate({service_name="avatar-service"} | severity_text="error" [5m])

# Все логи, кроме debug
{service_name="avatar-service"} | severity_text!="debug"

# Поиск по trace_id (корреляция с трейсами)
{service_name="avatar-service"} |= "trace_id"

# Сообщения в секунду по уровню
sum by (severity_text) (rate({service_name="avatar-service"} [5m]))

# Поиск по атрибуту (JSON-поля лога)
{service_name="avatar-service"} | json | avatar_id="123"

# Фильтр по регулярному выражению (операции с аватаром)
{service_name="avatar-service"} |~ "(upload|process)"
```

### Как посмотреть логи

Открой Grafana (`http://localhost:3000`), перейди в **Explore** и выбери datasource **Loki**.

## Трассировки

Трейсы экспортируются по OTLP gRPC в OpenTelemetry Collector, который форвардит их в Jaeger.
Для просмотра открой Jaeger UI (`http://localhost:16686`) и выбери сервис `avatar-service` или `avatar-worker`.

## Структура проекта

```
cmd/
  server/                  # Точка входа HTTP-сервера
  worker/                  # Точка входа worker-консьюмера
deploy/
  grafana/                 # Конфиг datasource (Loki)
  loki/                    # Конфиг Loki
  otel/                    # Конфиг OpenTelemetry Collector
  prometheus/              # Конфиг Prometheus (скрейп OTel Collector)
internal/
  config/                  # Загрузка конфигурации (YAML + env)
  delivery/http/           # HTTP-обработчики, роутер, middleware
    openapi/               # Сгенерированный из Swagger код
  infra/
    db/postgres/           # Репозиторий PostgreSQL, миграции, метрики БД
    messaging/kafka/       # Продюсер и консьюмер Kafka, метрики Kafka
    storage/minio/         # S3: загрузка, скачивание, удаление, метрики S3
  service/                 # Бизнес-логика и бизнес-метрики
  types/                   # Доменные модели
  worker/                  # Цикл обработки сообщений worker
pkg/
  dbutils/                 # Утилиты БД: NamedExec, транзакции, retry
  errcodes/                # Доменные ошибки
  imageutils/              # Ресайз, конвертация изображений
  logger/                  # Zerolog + OTel-провайдер (OTLP-экспорт в Loki)
  metrics/                 # Метрические инструменты и атрибуты (OTLP-экспорт в Prometheus)
  service_info/            # Константы имени и версии сервиса
  tracer/                  # Tracer-провайдер (OTLP-экспорт в Jaeger)
  utils/                   # Дженерик-утилиты (Ptr, Deref)
api/rest/swagger.yaml      # OpenAPI-спецификация
config/                    # YAML-конфиги (локальные и Docker)
migrations/                # SQL-миграции
```

## Переменные окружения

Задаются в файле `.env` (скопируй из `.env.example` — команда `make init` сделает это автоматически):

```env
POSTGRES_USER=profile
POSTGRES_PASSWORD=profile
POSTGRES_DB=profile

MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

Пути к конфигам можно переопределить через переменные окружения:

- `SERVER_CONFIG_PATH` — путь к конфигу сервера (по умолчанию `./config/server.yaml`)
- `WORKER_CONFIG_PATH` — путь к конфигу worker (по умолчанию `./config/worker.yaml`)

## Как работает обработка аватара

1. Клиент загружает изображение через `POST /avatars`.
2. Server валидирует файл, загружает оригинал в MinIO, создаёт запись в БД со статусом `pending` и отправляет сообщение
   в Kafka.
3. Worker читает сообщение, скачивает оригинал, генерирует миниатюры (100x100 и 300x300), загружает их в MinIO,
   обновляет статус на `completed`.
4. Клиент получает аватар через `GET /avatars/{id}` с опциональными параметрами `size` и `format` — конвертация
   происходит на лету.
5. При удалении (`DELETE /avatars/{id}`) Server помечает запись как удалённую и отправляет сообщение в Kafka — Worker
   удаляет файлы из MinIO.
