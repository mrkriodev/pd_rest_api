# Примеры curl команд для Startup Roulette

## Предположения
- Сервер запущен на `localhost:8080`
- Базовый URL: `http://localhost:8080/api`

## 1. Получить конфигурацию startup roulette (id=1)

```bash
curl -X GET "http://localhost:8080/api/roulette/get?id=1" \
  -H "Content-Type: application/json"
```

**Ответ:**
```json
{
  "id": 1,
  "type": "on_start",
  "event_id": null,
  "max_spins": 3,
  "is_active": true,
  "created_at": 1234567890000,
  "updated_at": 1234567890000
}
```

## 2. Получить preauth token (для неавторизованных пользователей)

```bash
curl -X GET "http://localhost:8080/api/roulette/get_preauth_token" \
  -H "Content-Type: application/json" \
  -H "X-SESSION-ID: my-session-id-12345"
```

**Ответ:**
```json
{
  "preauth_token": "abc123def456..."
}
```

**Примечание:** Если у вас нет X-SESSION-ID, можно использовать cookie:
```bash
curl -X GET "http://localhost:8080/api/roulette/get_preauth_token" \
  -H "Content-Type: application/json" \
  -H "Cookie: X-SESSION-ID=my-session-id-12345"
```

## 3. Выполнить спин (spin)

```bash
curl -X POST "http://localhost:8080/api/roulette/spin" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: YOUR_PREAUTH_TOKEN_HERE" \
  -d '{
    "roulette_id": 1
  }'
```

**Ответ:**
```json
{
  "result": {
    "segmentId": "2",
    "label": "100 USDT"
  },
  "spinsLeft": 2,
  "reward": {
    "type": "usdt",
    "amount": 100
  }
}
```

## 4. Полный пример: от получения токена до спина

```bash
# Шаг 1: Получить preauth token
PREAUTH_TOKEN=$(curl -s -X GET "http://localhost:8080/api/roulette/get_preauth_token" \
  -H "X-SESSION-ID: my-session-$(date +%s)" | jq -r '.preauth_token')

echo "Preauth Token: $PREAUTH_TOKEN"

# Шаг 2: Выполнить спин
curl -X POST "http://localhost:8080/api/roulette/spin" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: $PREAUTH_TOKEN" \
  -d '{
    "roulette_id": 1
  }'
```

## 5. Получить статус roulette

```bash
curl -X GET "http://localhost:8080/api/roulette/status?preauth_token=YOUR_PREAUTH_TOKEN_HERE" \
  -H "Content-Type: application/json"
```

Или через header:
```bash
curl -X GET "http://localhost:8080/api/roulette/status" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: YOUR_PREAUTH_TOKEN_HERE"
```

## 6. Забрать приз (после завершения всех спинов)

```bash
curl -X POST "http://localhost:8080/api/roulette/take-prize" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: YOUR_PREAUTH_TOKEN_HERE" \
  -H "X-SESSION-ID: my-session-id-12345" \
  -d '{
    "roulette_id": 1
  }'
```

**Ответ:**
```json
{
  "success": true,
  "prize": "100 USDT",
  "message": "Prize taken successfully",
  "preauth_token": "abc123..." // Возвращается только для неавторизованных пользователей
}
```

## Примечания

1. **X-SESSION-ID**: Может быть передан через header или cookie
2. **X-Preauth-Token**: Обязателен для `/spin` и `/take-prize`, передается только через header
3. **roulette_id**: Всегда должен быть `1` для startup roulette
4. **max_spins**: Для startup roulette установлено `3` спинов
5. **Authorization**: Не требуется для `on_start` типа roulette (startup roulette)

## Пример последовательности спинов

```bash
# Получить токен
TOKEN=$(curl -s -X GET "http://localhost:8080/api/roulette/get_preauth_token" \
  -H "X-SESSION-ID: session-123" | jq -r '.preauth_token')

# Спин 1
curl -X POST "http://localhost:8080/api/roulette/spin" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: $TOKEN" \
  -d '{"roulette_id": 1}'

# Спин 2
curl -X POST "http://localhost:8080/api/roulette/spin" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: $TOKEN" \
  -d '{"roulette_id": 1}'

# Спин 3
curl -X POST "http://localhost:8080/api/roulette/spin" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: $TOKEN" \
  -d '{"roulette_id": 1}'

# Забрать приз
curl -X POST "http://localhost:8080/api/roulette/take-prize" \
  -H "Content-Type: application/json" \
  -H "X-Preauth-Token: $TOKEN" \
  -H "X-SESSION-ID: session-123" \
  -d '{"roulette_id": 1}'
```

