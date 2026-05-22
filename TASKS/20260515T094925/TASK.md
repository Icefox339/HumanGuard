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

# Настройка сайта клиентом

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

# Имитация нормального пользователя

## Создание сессии

``` bash
curl -X POST http://localhost:8080/api/check \
  -H "X-Site-ID: 4c0bafa1-3de9-496a-b001-44f1352351e9" \
  -H "Content-Type: application/json"
```

Если сайт не активирован

``` json
{
  "error": "site is not active"
}
```

Если сайт активирован

``` json
{
  "action": "allow",
  "risk_score": 0,
  "session_id": "07c46dc2-9862-4bde-9868-90aeba0e545c"
}
```

Ид сессии сохраню в переменную в баш сессии

``` bash
SESSION_ID="07c46dc2-9862-4bde-9868-90aeba0e545c"
```

## Работа в рамках сессии

### Отправка поведенческих метрик

``` bash
curl -X POST http://localhost:8080/api/behavior/$SESSION_ID \
  -H "X-Site-ID: 4c0bafa1-3de9-496a-b001-44f1352351e9" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "'$SESSION_ID'",
    "metrics": {
      "counters": {
        "mouse_move": 350,
        "click": 15,
        "scroll": 40,
        "keydown": 80,
        "duration_sec": 120
      },
      "fingerprint": {
        "js_hash": "normal_user_123",
        "webgl_renderer": "ANGLE (NVIDIA Corporation)",
        "canvas_hash": "canvas_normal"
      },
      "timing": {
        "load_time_ms": 800
      }
    }
  }'
```

Ответ

```
{
  "status": "accepted"
}
```

### Проверка risk-score

``` bash
curl -X POST http://localhost:8080/api/check \
  -H "X-Site-ID: 4c0bafa1-3de9-496a-b001-44f1352351e9" \
  -H "Cookie: hg_session=$SESSION_ID"
```

Ответ

``` json
{
  "action": "allow",
  "risk_score": 0,
  "session_id": "07c46dc2-9862-4bde-9868-90aeba0e545c"
}
```

# Имитация бота

## Создание сессии

``` bash
curl -X POST http://localhost:8080/api/check \
  -H "X-Site-ID: 4c0bafa1-3de9-496a-b001-44f1352351e9" \
  -H "Content-Type: application/json"
```

Ответ

``` json
{
  "action": "allow",
  "risk_score": 0,
  "session_id": "8e9ac433-734b-427e-aa32-f80904909ae5"
}
```

``` bash
SESSION_ID_BOT="8e9ac433-734b-427e-aa32-f80904909ae5"
```

## Работа в рамках сессии

### Отправка поведенческих метрик

``` bash
curl -X POST http://localhost:8080/api/behavior/$SESSION_ID_BOT \
  -H "X-Site-ID: 4c0bafa1-3de9-496a-b001-44f1352351e9" \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "'$SESSION_ID_BOT'",
    "metrics": {
      "counters": {
        "mouse_move": 0,
        "click": 0,
        "scroll": 0,
        "keydown": 0,
        "duration_sec": 3
      },
      "fingerprint": {
        "js_hash": "",
        "webgl_renderer": "SwiftShader",
        "canvas_hash": ""
      },
      "timing": {
        "load_time_ms": 45
      }
    }
  }'
```

Ответ

``` json
{
  "status": "accepted"
}
```

### Проверка risk-score

``` bash
curl -X POST http://localhost:8080/api/check \
  -H "X-Site-ID: 4c0bafa1-3de9-496a-b001-44f1352351e9" \
  -H "Cookie: hg_session=$SESSION_ID_BOT" \
  -H "Content-Type: application/json"
```

``` json
{
  "action": "block",
  "risk_score": 100,
  "session_id": "e5722d95-c2fb-4482-8ac1-4b98055e3673"
}
```

В реальной инфраструктуре когда nginx клиента получает `"action": block` он не должен пускать к сайту

# API ключи

## Создание API ключа по JWT токену

API ключ можно использовать вместо JWT токена. JWT токен истекает
быстро (24ч фиксировано в коде), а время жизни API ключа определяется
на стороне клиента в момент создания

Например на 30 дней создаю

``` bash
curl -X POST http://localhost:8080/api/keys \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test API Key",
    "expires_in_days": 30
  }'
```

Ответ

``` json
{
  "id": "cb511f77-7178-4bbf-a905-7bc8edd61c4f",
  "name": "Test API Key",
  "key": "hg_v1_2ca39aa1642340c8b18efd00852fc59c1a3769a77758bc9d1f096d891dc041e1",
  "prefix": "hg_v1_",
  "created_at": "2026-05-15T11:03:56.040079865+03:00",
  "expires_at": "2026-06-14T11:03:56.040076078+03:00",
  "revoked": false
}
```

``` bash
API_KEY="hg_v1_2ca39aa1642340c8b18efd00852fc59c1a3769a77758bc9d1f096d891dc041e1"
```

## Проверка работы API ключа вместо JWT токена

``` bash
curl -X GET http://localhost:8080/api/me \
  -H "X-API-Key: $API_KEY"
```

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
  "updated_at": "2026-05-15T10:24:44.098394Z",
  "last_login": "2026-05-15T10:24:44.098394Z"
}
```

# Работа с файлами

## Загрузка файла

``` bash
echo "This is a test file for HumanGuard" > ~/projects/temp/testhg.txt
```

``` bash
curl -X POST http://localhost:8080/api/files/upload \
  -H "X-API-Key: $API_KEY" \
  -F "file=@/home/serr/projects/temp/testhg.txt"
```

Ответ

``` json
{
  "id": "3a7122cd-172e-4e62-b865-740ab868aebb",
  "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "name": "9956daf5-e1da-468d-9ac1-8891f465ad54.txt",
  "original_name": "testhg.txt",
  "size": 35,
  "mime_type": "text/plain",
  "hash": "ef8ec0e48741322ed8c745e127177b59cacf5c2a066c3c2b4a60eb5ae10c8fb1",
  "path": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed/2026/05/15/9956daf5-e1da-468d-9ac1-8891f465ad54.txt",
  "created_at": "2026-05-15T11:25:14.963398985+03:00"
}
```

## Скачивание файла

``` bash
curl -X GET http://localhost:8080/api/files/$FILE_ID \
  -H "X-API-Key: $API_KEY" \
  --output downloaded.txt
```

Проверка

``` bash
~/projects/HumanGuard/backend
[serr@lap]-> cat downloaded.txt
This is a test file for HumanGuard
```

## Создание публичной ссылки на файл

``` bash
curl -X POST http://localhost:8080/api/files/share \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": "'$FILE_ID'",
    "expires_in_hours": 48
  }'
```

``` json
{
  "token": "769656badcc13aeb09a80ea68876a3efae9c8fff995c0d672114034a8d92deb1"
}
```

Сохраняю в переменную

``` bash
SHARE_TOKEN="769656badcc13aeb09a80ea68876a3efae9c8fff995c0d672114034a8d92deb1"
```

## Скачивание файла по публичной ссылке

``` bash
curl -X GET http://localhost:8080/api/files/share/$SHARE_TOKEN \
  --output public_downloaded.txt
```

Проверка

```
~/projects/HumanGuard/backend
[serr@lap]-> cat public_downloaded.txt
This is a test file for HumanGuard
```

## Список файлов юзера

``` bash
curl -X GET http://localhost:8080/api/files \
  -H "X-API-Key: $API_KEY"
```

Ответ

``` json
[
  {
    "id": "d4ca1292-c927-454a-9d78-af39350455b3",
    "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "name": "4804d9dc-77b8-4b64-a018-0d0a4eb3c9ad.txt",
    "original_name": "testhg.txt",
    "size": 35,
    "mime_type": "text/plain",
    "hash": "ef8ec0e48741322ed8c745e127177b59cacf5c2a066c3c2b4a60eb5ae10c8fb1",
    "path": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed/2026/05/15/4804d9dc-77b8-4b64-a018-0d0a4eb3c9ad.txt",
    "created_at": "2026-05-15T11:31:07.922983Z"
  },
  {
    "id": "35e18ff8-e392-4475-92f6-b411f58b8368",
    "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "name": "edfc71ad-04c8-4d68-b96b-1ac5f6f56903.txt",
    "original_name": "testhg.txt",
    "size": 35,
    "mime_type": "text/plain",
    "hash": "ef8ec0e48741322ed8c745e127177b59cacf5c2a066c3c2b4a60eb5ae10c8fb1",
    "path": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed/2026/05/15/edfc71ad-04c8-4d68-b96b-1ac5f6f56903.txt",
    "created_at": "2026-05-15T11:28:17.060076Z"
  },
  {
    "id": "3a7122cd-172e-4e62-b865-740ab868aebb",
    "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
    "name": "9956daf5-e1da-468d-9ac1-8891f465ad54.txt",
    "original_name": "testhg.txt",
    "size": 35,
    "mime_type": "text/plain",
    "hash": "ef8ec0e48741322ed8c745e127177b59cacf5c2a066c3c2b4a60eb5ae10c8fb1",
    "path": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed/2026/05/15/9956daf5-e1da-468d-9ac1-8891f465ad54.txt",
    "created_at": "2026-05-15T11:25:14.963399Z"
  }
]
```

## Удаление файла по ид

``` bash
~/projects/HumanGuard/backend
[serr@lap]-> curl -X DELETE http://localhost:8080/api/files/$FILE_ID \
  -H "X-API-Key: $API_KEY"
```

Ответ

``` json
{
  "file_id": "d4ca1292-c927-454a-9d78-af39350455b3",
  "message": "file deleted successfully",
  "original_name": "testhg.txt"
}
```

## WebSocket прогресс загрузки

Создаю большой файл

``` bash
dd if=/dev/zero of=/home/serr/projects/temp/10mb.bin bs=1M count=10
```

В одном терминале смотрю прогрес используя *wscat* (поменял на время
код чтобы ожидалось подключение)

``` bash
npx wscat -c "ws://localhost:8080/api/files/upload/progress?upload_id=test123" \
  -H "X-API-Key: $API_KEY"
```

В другом терминале запустил

``` bash
curl -X POST "http://localhost:8080/api/files/upload?upload_id=test123" \
  -H "X-API-Key: $API_KEY" \
  -F "file=@/home/serr/projects/temp/10mb.txt"
```

В терминале с *wscat*

``` bash
Connected (press CTRL+C to quit)
< {"upload_id":"test123","bytes_done":0,"total_bytes":10485972,"percentage":0,"completed":false}

< {"upload_id":"test123","bytes_done":10485760,"total_bytes":10485958,"percentage":100,"completed":true}

Disconnected (code: 1006, reason: "")
```

В терминале где запускал скачивание

``` json
{
  "id": "47153ada-0ac9-4632-ac9a-bbd1c681f609",
  "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "name": "7e1ff277-a6b0-4f99-9034-757b8628cf87.txt",
  "original_name": "10mb.txt",
  "size": 10485760,
  "mime_type": "text/plain",
  "hash": "e5b844cc57f57094ea4585e235f36c78c1cd222262bb89d53c94dcb4d6b3e55d",
  "path": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed/2026/05/15/7e1ff277-a6b0-4f99-9034-757b8628cf87.txt",
  "created_at": "2026-05-15T12:46:23.829292118+03:00"
}
```

## MinIO

До этого все в качестве хранилища файлов использовалась просто папка,
но также можно использовать объектное хранилище *MinIO*, для этого надо
запустить его и перезапустить сервер с правильными переменными
окружения

### Запуск MinIO

``` bash
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin123" \
  minio/minio server /data --console-address ":9001"
```

Браузерный интерфейс: http://localhost:9001
Логин и пароль: minioadmin / minioadmin123

Установка переменных окружения

``` bash
export STORAGE_TYPE=minio
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=minioadmin
export MINIO_SECRET_KEY=minioadmin123
export MINIO_BUCKET=humanguard
export MINIO_USE_SSL=false
```

Далее при запуске сервера будет написано что подключено к хранилищу *MinIO*

``` bash
~/projects/HumanGuard/backend
[serr@lap]-> go run cmd/server/main.go
2026/05/15 13:02:25 Connected to database
2026/05/15 13:02:25 Database ping successful
2026/05/15 13:02:25 Created bucket: humanguard
2026/05/15 13:02:25 Connected to MinIO storage
2026/05/15 13:02:25 Server starting on http://localhost:8080
```

Загружаю файл

``` bash
curl -X POST http://localhost:8080/api/files/upload \
  -H "X-API-Key: $API_KEY" \
  -F "file=@/home/serr/projects/temp/testhg.txt"
```

Ответ

``` json
{
  "id": "4074c594-465c-4335-90b7-f8c7ebb91395",
  "user_id": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed",
  "name": "1a913627-a0c1-4a4d-9eef-93fb580ef889.txt",
  "original_name": "testhg.txt",
  "size": 35,
  "mime_type": "text/plain",
  "hash": "ef8ec0e48741322ed8c745e127177b59cacf5c2a066c3c2b4a60eb5ae10c8fb1",
  "path": "283a3b1a-fa2e-4ba7-8c1c-af67a76e29ed/2026/05/15/1a913627-a0c1-4a4d-9eef-93fb580ef889.txt",
  "created_at": "2026-05-15T13:03:53.006641023+03:00"
}
```

Далее через интерфейс в браузере можно скачать либо через API

``` bash
FILE_ID="4074c594-465c-4335-90b7-f8c7ebb91395"
curl -X GET http://localhost:8080/api/files/$FILE_ID \
  -H "X-API-Key: $API_KEY" \
  --output downloaded-from-minio.txt
```

Проверка

``` bash
~/projects/HumanGuard/backend
[serr@lap]-> cat downloaded-from-minio.txt
This is a test file for HumanGuard
```

ну и все методы работают как и работали для работы с файлами

# KeyCloak

Запуск

``` bash
docker run -d \
  --name keycloak \
  -p 8081:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:24.0.1 \
  start-dev
```

Далее тут http://localhost:8081

- Логин: admin
- Пароль: admin

по этой ссылке в браузере http://localhost:8080/api/auth/keycloak/login

```
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzk1MzUyNzksImlhdCI6MTc3OTQ0ODg3OSwicm9sZSI6InVzZXIiLCJzaWQiOiIyMmJkNzAyMy1jMjgyLTRlMzAtYmQ3NC1jZGFkOThmOWI2MTUiLCJ1c2VyX2lkIjoiNDhjOTFmNDctMmQ4Yy00OWE3LTg4MjItMDdjZGJjZjhmNWNmIn0.0ymu15bhDyOoJZkYLO8hDNekL4IWUAEIEnTELSP-TJg",
  "user": {
    "id": "48c91f47-2d8c-49a7-8822-07cdbcf8f5cf",
    "email": "",
    "name": "",
    "avatar_url": null,
    "role": "user",
    "is_verified": false,
    "oauth_provider": "keycloak",
    "created_at": "2026-05-22T14:21:19.069064191+03:00",
    "updated_at": "2026-05-22T14:21:19.069064191+03:00",
    "last_login": null
  }
}
```

проверка эндпоинта с использованием токена, полученного от keycloak

```
~/projects/HumanGuard/backend
[serr@lap]-> TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzk1MzUyNzksImlhdCI6MTc3OTQ0ODg3OSwicm9sZSI6InVzZXIiLCJzaWQiOiIyMmJkNzAyMy1jMjgyLTRlMzAtYmQ3NC1jZGFkOThmOWI2MTUiLCJ1c2VyX2lkIjoiNDhjOTFmNDctMmQ4Yy00OWE3LTg4MjItMDdjZGJjZjhmNWNmIn0.0ymu15bhDyOoJZkYLO8hDNekL4IWUAEIEnTELSP-TJg"
~/projects/HumanGuard/backend
[serr@lap]-> curl -X GET http://localhost:8080/api/me \
  -H "Authorization: Bearer $TOKEN" | jq '.'
  % Total    % Received % Xferd  Average Speed  Time    Time    Time   Current
                                 Dload  Upload  Total   Spent   Left   Speed
100    251 100    251   0      0 101.1k      0                              0
{
  "id": "48c91f47-2d8c-49a7-8822-07cdbcf8f5cf",
  "email": "",
  "name": "",
  "avatar_url": null,
  "role": "user",
  "is_verified": false,
  "oauth_provider": "keycloak",
  "created_at": "2026-05-22T14:21:19.069064Z",
  "updated_at": "2026-05-22T14:21:19.069064Z",
  "last_login": null
}
```
