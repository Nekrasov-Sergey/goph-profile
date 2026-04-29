#!/bin/bash
set -euo pipefail

# test-alert.sh — отправка тестовых алертов в Alertmanager и генерация HTTP-трафика
# для проверки правил алертинга Prometheus.
#
# Использование:
#   ./scripts/test-alert.sh                         # тестовый алерт уровня warning
#   ./scripts/test-alert.sh critical                # тестовый алерт уровня critical
#   ./scripts/test-alert.sh all                     # все тестовые алерты + трафик
#   ./scripts/test-alert.sh traffic                 # генерация HTTP-трафика для Prometheus
#   ./scripts/test-alert.sh check                   # только проверка доступности
#   ./scripts/test-alert.sh status                  # статус системы алертинга

ALERTMANAGER_URL="${ALERTMANAGER_URL:-http://localhost:9093}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"
SERVER_URL="${SERVER_URL:-http://localhost:8080}"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${GREEN}[✓]${NC} $1"; }
warn()    { echo -e "${YELLOW}[!]${NC} $1"; }
error()   { echo -e "${RED}[✗]${NC} $1"; }
link_hint() { echo -e "${BLUE}  → $1${NC}"; }

check_deps() {
  if ! command -v curl &>/dev/null; then
    error "curl не установлен"
    exit 1
  fi
}

check_alertmanager() {
  local status
  status=$(curl -sf -o /dev/null -w "%{http_code}" "${ALERTMANAGER_URL}/api/v2/status" 2>/dev/null || true)
  if [ "$status" = "200" ]; then
    info "Alertmanager доступен (${ALERTMANAGER_URL})"
    return 0
  else
    warn "Alertmanager недоступен (${ALERTMANAGER_URL})"
    warn "Запустите инфраструктуру: make docker-infra"
    return 1
  fi
}

check_prometheus() {
  local status
  status=$(curl -sf -o /dev/null -w "%{http_code}" "${PROMETHEUS_URL}/-/ready" 2>/dev/null || true)
  if [ "$status" = "200" ]; then
    info "Prometheus доступен (${PROMETHEUS_URL})"
    return 0
  else
    warn "Prometheus недоступен (${PROMETHEUS_URL})"
    return 1
  fi
}

check_server() {
  local status
  status=$(curl -sf -o /dev/null -w "%{http_code}" "${SERVER_URL}" 2>/dev/null || true)
  if [ "$status" != "000" ]; then
    info "Сервер приложения доступен (${SERVER_URL})"
    return 0
  else
    warn "Сервер приложения недоступен (${SERVER_URL})"
    warn "Трафик не будет сгенерирован — Prometheus-алерты не сработают без метрик"
    warn "Запустите сервер: make docker-up или make server"
    return 1
  fi
}

show_links() {
  echo ""
  echo "=== Ссылки ==="
  link_hint "Prometheus:     ${PROMETHEUS_URL}/alerts"
  link_hint "Alertmanager:   ${ALERTMANAGER_URL}/#/alerts"
  link_hint "Grafana:        http://localhost:3000"
  echo ""
}

send_alert() {
  local severity="$1"
  local alertname="$2"
  local summary="$3"
  local description="$4"

  local payload
  payload=$(cat <<EOF
[{
  "labels": {
    "alertname": "${alertname}",
    "severity": "${severity}",
    "job": "otel-collector",
    "service": "goph-profile"
  },
  "annotations": {
    "summary": "${summary}",
    "description": "${description}"
  },
  "startsAt": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "endsAt": "$(date -u -d '+30 minutes' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+30M +%Y-%m-%dT%H:%M:%SZ)"
}]
EOF
)

  local http_code
  http_code=$(curl -sf -o /dev/null -w "%{http_code}" \
    -XPOST \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "${ALERTMANAGER_URL}/api/v2/alerts" 2>/dev/null || true)

  if [ "$http_code" = "200" ]; then
    info "Алерт «${summary}» (${severity}) отправлен в Alertmanager"
  else
    error "Ошибка отправки алерта (HTTP ${http_code})"
  fi
}

# === Тестовые алерты (напрямую в Alertmanager API) ===

test_warning() {
  send_alert "warning" "TestAlertWarning" "Тестовый алерт: повышенный уровень 4xx" \
    "Тестовое описание для проверки доставки warning-алертов. Сервис: avatar-service"
}

test_critical() {
  send_alert "critical" "TestAlertCritical" "Тестовый алерт: критическая ошибка" \
    "Тестовое описание для проверки доставки critical-алертов. Сервис: avatar-service"
}

test_all_alerts() {
  echo ""
  echo "=== Отправка тестовых алертов в Alertmanager ==="

  test_warning
  sleep 0.5
  test_critical
  sleep 0.5

  # Алёрты, похожие на реальные сценарии
  send_alert "critical" "HighHTTPErrorRate5xx" "Высокий уровень 5xx ошибок HTTP" \
    "Доля 5xx ошибок 12.5% за последние 5 минут. Сервис: avatar-service, маршрут: /api/avatar"
  sleep 0.5

  send_alert "warning" "SlowDBOperations" "Медленные операции БД (P99)" \
    "P99 время выполнения операции БД 2.5s. Операция: db_operation_duration"
}

# === Генерация HTTP-трафика для Prometheus ===

gen_traffic() {
  echo ""
  echo "=== Генерация HTTP-трафика для Prometheus ==="
  echo "  (метрики появятся в Prometheus через ~15-30 секунд)"
  echo "  (правила алертинга проверяются каждые 30 секунд)"
  echo ""

  check_server || return 1

  local api_base="${SERVER_URL}/api/v1"

  # --- 2xx: корректные запросы ---
  echo "--- Корректные запросы (2xx) ---"

  for i in $(seq 1 10); do
    curl -sf -o /dev/null -w "  [%{http_code}] GET  /health\n" \
      "${SERVER_URL}/health" 2>/dev/null || true
  done

  # --- 4xx: запросы с ошибками ---
  echo "--- Ошибочные запросы (4xx) ---"

  # POST /api/v1/avatars без файла — 400 BadRequest
  for i in $(seq 1 25); do
    curl -sf -o /dev/null -w "  [%{http_code}] POST /api/v1/avatars (no file)\n" \
      -XPOST -H "X-User-Id: user-test-${i}" \
      "${api_base}/avatars" 2>/dev/null || true
  done

  # GET несуществующий эндпоинт — 404
  for i in $(seq 1 10); do
    curl -sf -o /dev/null -w "  [%{http_code}] GET  /api/v1/nonexistent\n" \
      "${api_base}/nonexistent" 2>/dev/null || true
  done

  # POST без X-User-Id — вероятно 4xx
  for i in $(seq 1 10); do
    curl -sf -o /dev/null -w "  [%{http_code}] POST /api/v1/avatars (no user-id)\n" \
      -XPOST "${api_base}/avatars" 2>/dev/null || true
  done

  # --- Нестандартные методы ---
  echo "--- Нестандартные методы (4xx/5xx) ---"
  for method in PUT DELETE PATCH OPTIONS; do
    curl -sf -o /dev/null -w "  [%{http_code}] ${method} /api/v1/avatars\n" \
      -X"${method}" "${api_base}/avatars" 2>/dev/null || true
  done

  echo ""
  info "Первая волна трафика сгенерирована."

  # Повторный всплеск через 10 секунд (чтобы Prometheus увидел rate)
  echo ""
  echo "  Ожидание 10 секунд перед повторным всплеском..."
  sleep 10

  echo "--- Повторный всплеск (4xx) ---"
  for i in $(seq 1 30); do
    curl -sf -o /dev/null \
      -XPOST -H "X-User-Id: user-test-burst-${i}" \
      "${api_base}/avatars" 2>/dev/null || true
  done
  for i in $(seq 1 10); do
    curl -sf -o /dev/null \
      -XPOST "${api_base}/avatars" 2>/dev/null || true
  done

  echo ""
  info "Трафик сгенерирован (2 волны). Prometheus заберёт метрики при следующем скрейпе."
  echo ""
  echo "  ⏳ Через 30-60 секунд проверьте Prometheus:"
  link_hint "${PROMETHEUS_URL}/alerts"
  echo ""
  echo "  Ожидаемый алерт: ElevatedHTTPErrorRate4xx"
  echo "  (должен перейти в состояние PENDING → FIRING через 5 минут)"
  echo ""
  echo "  Пример запроса для проверки метрик:"
  echo '    rate(http_server_request_errors_total{http_response_status_code=~"4.."}[5m])'
  link_hint "${PROMETHEUS_URL}/graph"
}

# === Статус ===

show_status() {
  echo ""
  echo "=== Статус системы алертинга ==="
  check_prometheus
  check_alertmanager
  check_server || true

  echo ""
  echo "--- Правила алертинга (Prometheus) ---"
  curl -sf "${PROMETHEUS_URL}/api/v1/rules" 2>/dev/null | \
    python3 -c "
import json, sys
data = json.load(sys.stdin)
groups = data.get('data', {}).get('groups', [])
for g in groups:
    print(f\"  Группа: {g['name']} (интервал: {g.get('interval', '?')})\")
    for r in g.get('rules', []):
        state = r.get('state', 'unknown')
        name = r.get('name', r.get('alert', 'unknown'))
        icon = '🚨' if 'alert' in r else '📊'
        print(f\"    {icon} {name}: {state}\")
" 2>/dev/null || warn "Не удалось получить правила из Prometheus"

  echo ""
  echo "--- Активные алерты (Alertmanager) ---"
  local alerts
  alerts=$(curl -sf "${ALERTMANAGER_URL}/api/v2/alerts" 2>/dev/null | \
    python3 -c "
import json, sys
data = json.load(sys.stdin)
if not data:
    print('    Нет активных алертов')
else:
    for a in data:
        labels = a.get('labels', {})
        status = a.get('status', {}).get('state', '?')
        name = labels.get('alertname', '?')
        sev = labels.get('severity', '?')
        print(f'    [{status}] {name} ({sev})')
" 2>/dev/null || true)

  if [ -z "$alerts" ]; then
    warn "Не удалось получить алерты. Возможно, Alertmanager недоступен."
  else
    echo "$alerts"
  fi

  show_links
}

# === main ===

check_deps

case "${1:-warning}" in
  check)
    check_alertmanager || true
    check_prometheus || true
    check_server || true
    show_links
    ;;
  warning)
    check_alertmanager || exit 1
    test_warning
    show_links
    ;;
  critical)
    check_alertmanager || exit 1
    test_critical
    show_links
    ;;
  traffic)
    gen_traffic
    show_links
    ;;
  all)
    check_alertmanager || exit 1
    test_all_alerts
    echo ""
    gen_traffic || true
    echo ""
    show_status
    ;;
  status)
    show_status
    ;;
  *)
    echo "Использование: $0 {warning|critical|all|traffic|check|status}"
    echo ""
    echo "  warning   — отправить тестовый алерт уровня warning (по умолчанию)"
    echo "  critical  — отправить тестовый алерт уровня critical"
    echo "  all       — тестовые алерты + генерация трафика + статус"
    echo "  traffic   — сгенерировать HTTP-трафик для проверки Prometheus-правил"
    echo "  check     — проверить доступность сервисов"
    echo "  status    — показать статус системы алертинга"
    exit 1
    ;;
esac
