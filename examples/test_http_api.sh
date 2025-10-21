#!/bin/bash
# Пример скрипта для тестирования HTTP API mlogtail

BASE_URL="http://localhost:8080"

echo "=== Проверка здоровья сервиса ==="
curl -s "${BASE_URL}/health" | jq .
echo ""

echo "=== Получение всех статистик ==="
curl -s "${BASE_URL}/stats" | jq .
echo ""

echo "=== Получение конкретного счетчика (received) ==="
curl -s "${BASE_URL}/counter/received" | jq .
echo ""

echo "=== Получение конкретного счетчика (delivered) ==="
curl -s "${BASE_URL}/counter/delivered" | jq .
echo ""

echo "=== Получение конкретного счетчика (rejected) ==="
curl -s "${BASE_URL}/counter/rejected" | jq .
echo ""

echo "=== Попытка получить несуществующий счетчик ==="
curl -s "${BASE_URL}/counter/nonexistent" | jq .
echo ""

echo "=== Красиво отформатированная статистика ==="
curl -s "${BASE_URL}/stats" | jq '{
  "Получено писем": .received,
  "Доставлено писем": .delivered,
  "Отклонено писем": .rejected,
  "Получено байт": .bytes_received,
  "Доставлено байт": .bytes_delivered,
  "Процент доставки": (if .received > 0 then ((.delivered / .received * 100) | round) else 0 end)
}'
echo ""

# Раскомментируйте для сброса счетчиков
# echo "=== Сброс счетчиков ==="
# curl -s -X POST "${BASE_URL}/reset" | jq .
# echo ""

# Раскомментируйте для получения и сброса счетчиков
# echo "=== Получение и сброс счетчиков ==="
# curl -s -X POST "${BASE_URL}/stats_reset" | jq .
# echo ""
