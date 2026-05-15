- NAME: Тестирование бэкенда
- PRIORITY: 10
- TAGS: test, backend, guide
- STATUS: opened

# Развертывание сервера

``` bash
docker run -d \
  --name humanguard-db \
  -e POSTGRES_DB=humanguard \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=123 \
  -p 5432:5432 \
  postgres:15
```

``` bash
cd backend
docker cp migrations/001_init_up.sql humanguard-db:/tmp/init.sql
docker exec -i humanguard-db psql -U postgres -d humanguard < migrations/001_init_up.sql
```

``` bash
cd backend
go mod download
go run cmd/server/main.go
```

# Проверка того что сервер поднят

``` bash
~/projects/HumanGuard/backend
[serr@lap]-> curl http://localhost:8080/health
{"status":"ok"}
```

# Регистрация пользователя и логин

## Регистрация

``` bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

Ответ:

``` json
{
  "qr_code_url": "otpauth://totp/HumanGuard:test@example.com?secret=BFMZRR3EKU4TM3EORXEIV6VCVPHCTAR7&issuer=HumanGuard",
  "totp_secret": "BFMZRR3EKU4TM3EORXEIV6VCVPHCTAR7",
  "user": {
    "id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "email": "test@example.com",
    "name": "Test User",
    "avatar_url": null,
    "role": "user",
    "is_verified": false,
    "oauth_provider": null,
    "created_at": "2026-05-15T09:52:18.114952333+03:00",
    "updated_at": "2026-05-15T09:52:18.114952333+03:00",
    "last_login": null
  }
}
```

## Логин

totp_secret полученный при регистрации (у меня это
`BFMZRR3EKU4TM3EORXEIV6VCVPHCTAR7`) вот например на этом сайте
https://totp.danhersam.com/ использую для получения одноразовых кодов

``` bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"test@example.com\",
    \"password\": \"password123\",
    \"totp_code\": \"307049\"
  }"
```

Ответ:

``` json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzg5MTQ3NzksImlhdCI6MTc3ODgyODM3OSwicm9sZSI6InVzZXIiLCJzaWQiOiI4NTE0ZjI2Mi00NzBlLTQxNzUtOWIxNS1hYmY3Y2U3MWRmMzkiLCJ1c2VyX2lkIjoiMjgzYTNiMWEtZmEyZS00YmE3LThjMWMtYWY2N2E3NmUyOWVkIn0.pW9lsJdtu3tRixKyQqNLiGtcabU313NFOuPOZCqKUAY",
  "user": {
    "id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "email": "test@example.com",
    "name": "Test User",
    "avatar_url": null,
    "role": "user",
    "is_verified": false,
    "oauth_provider": null,
    "created_at": "2026-05-15T09:52:18.114952Z",
    "updated_at": "2026-05-15T09:52:18.114952Z",
    "last_login": null
  }
}
```

Токен сразу в переменную сохраню в bash сессии

``` bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzg5MTQ3NzksImlhdCI6MTc3ODgyODM3OSwicm9sZSI6InVzZXIiLCJzaWQiOiI4NTE0ZjI2Mi00NzBlLTQxNzUtOWIxNS1hYmY3Y2U3MWRmMzkiLCJ1c2VyX2lkIjoiMjgzYTNiMWEtZmEyZS00YmE3LThjMWMtYWY2N2E3NmUyOWVkIn0.pW9lsJdtu3tRixKyQqNLiGtcabU313NFOuPOZCqKUAY"
```

# Получение текущего пользователя

``` bash
curl -X GET http://localhost:8080/api/me \
  -H "Authorization: Bearer $TOKEN"
```

Ответ:

``` json
{
  "id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "email": "test@example.com",
  "name": "Test User",
  "avatar_url": null,
  "role": "user",
  "is_verified": false,
  "oauth_provider": null,
  "created_at": "2026-05-15T09:52:18.114952Z",
  "updated_at": "2026-05-15T09:59:39.260223Z",
  "last_login": "2026-05-15T09:59:39.260223Z"
}
```

# Настройка сайта пользователем

## Создание сайта

``` bash
curl -X POST http://localhost:8080/api/sites \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "name": "My Test Site",
    "domain": "test-site.example.com",
    "origin_server": "http://localhost:3000"
  }'
```

Ответ:

``` json
{
  "id": "4c0bafa1-3de9-496a-b001-44f1352351e9",
  "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "name": "My Test Site",
  "domain": "test-site.example.com",
  "origin_server": "http://localhost:3000",
  "status": "verifying",
  "settings": null,
  "created_at": "2026-05-15T10:04:59.428617697+03:00",
  "updated_at": "2026-05-15T10:04:59.428617697+03:00"
}
```

## Активация сайта

``` bash
curl -X POST http://localhost:8080/api/sites/4c0bafa1-3de9-496a-b001-44f1352351e9/activate \
  -H "Authorization: Bearer $TOKEN"
```

Возвращает сайт

``` json
{
  "id": "4c0bafa1-3de9-496a-b001-44f1352351e9",
  "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "name": "My Test Site",
  "domain": "test-site.example.com",
  "origin_server": "http://localhost:3000",
  "status": "active",
  "settings": {
    "collector": {
      "enabled": false,
      "mouse_tracking": false,
      "click_tracking": false,
      "scroll_tracking": false,
      "keystroke_tracking": false,
      "fingerprint_enabled": false
    },
    "analyzer": {
      "enabled": false,
      "rate_limiting": false,
      "pattern_analysis": false,
      "headless_detection": false,
      "thresholds": {
        "low": 0,
        "medium": 0,
        "high": 0
      },
      "weights": {
        "ip_reputation": 0,
        "headless": 0,
        "rate_limit": 0,
        "behavior_anomaly": 0,
        "fingerprint_change": 0
      }
    },
    "reaction": {
      "enabled": false,
      "low_risk_action": "",
      "medium_risk_action": "",
      "high_risk_action": "",
      "block_duration": 0,
      "captcha_provider": ""
    }
  },
  "created_at": "2026-05-15T10:04:59.428618Z",
  "updated_at": "2026-05-15T10:16:38.255811Z"
}
```

## Деактивация сайта

``` bash
curl -X POST http://localhost:8080/api/sites/4c0bafa1-3de9-496a-b001-44f1352351e9/suspend \
  -H "Authorization: Bearer $TOKEN"
```

``` json
{
  "id": "4c0bafa1-3de9-496a-b001-44f1352351e9",
  "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "name": "My Test Site",
  "domain": "test-site.example.com",
  "origin_server": "http://localhost:3000",
  "status": "suspended",
  "settings": {
    "collector": {
      "enabled": false,
      "mouse_tracking": false,
      "click_tracking": false,
      "scroll_tracking": false,
      "keystroke_tracking": false,
      "fingerprint_enabled": false
    },
    "analyzer": {
      "enabled": false,
      "rate_limiting": false,
      "pattern_analysis": false,
      "headless_detection": false,
      "thresholds": {
        "low": 0,
        "medium": 0,
        "high": 0
      },
      "weights": {
        "ip_reputation": 0,
        "headless": 0,
        "rate_limit": 0,
        "behavior_anomaly": 0,
        "fingerprint_change": 0
      }
    },
    "reaction": {
      "enabled": false,
      "low_risk_action": "",
      "medium_risk_action": "",
      "high_risk_action": "",
      "block_duration": 0,
      "captcha_provider": ""
    }
  },
  "created_at": "2026-05-15T10:04:59.428618Z",
  "updated_at": "2026-05-15T10:20:44.643166Z"
}
```

## Получение всех сайтов юзера

``` bash
curl -X GET http://localhost:8080/api/sites \
  -H "Authorization: Bearer $TOKEN"
```

Ответ:

``` json
[
  {
    "id": "4c0bafa1-3de9-496a-b001-44f1352351e9",
    "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "name": "My Test Site",
    "domain": "test-site.example.com",
    "origin_server": "http://localhost:3000",
    "status": "suspended",
    "settings": {
      "collector": {
        "enabled": false,
        "mouse_tracking": false,
        "click_tracking": false,
        "scroll_tracking": false,
        "keystroke_tracking": false,
        "fingerprint_enabled": false
      },
      "analyzer": {
        "enabled": false,
        "rate_limiting": false,
        "pattern_analysis": false,
        "headless_detection": false,
        "thresholds": {
          "low": 0,
          "medium": 0,
          "high": 0
        },
        "weights": {
          "ip_reputation": 0,
          "headless": 0,
          "rate_limit": 0,
          "behavior_anomaly": 0,
          "fingerprint_change": 0
        }
      },
      "reaction": {
        "enabled": false,
        "low_risk_action": "",
        "medium_risk_action": "",
        "high_risk_action": "",
        "block_duration": 0,
        "captcha_provider": ""
      }
    },
    "created_at": "2026-05-15T10:04:59.428618Z",
    "updated_at": "2026-05-15T10:20:44.643166Z"
  }
]
```
