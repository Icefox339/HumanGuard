
# Table of Contents

1.  [Запуск](#orgf52788a)
2.  [Эндпоинты](#org62841f1)
    1.  [проверка что сервер работает](#org8a1cc2c)
    2.  [пользователи](#org5891821)
        1.  [создание пользователя](#org2f85bb2)
        2.  [получение юзера по id](#org695e420)
        3.  [получение юзера по email](#org72a16a7)
        4.  [проверка занят ли email](#org7cdf376)
        5.  [смена пароля юзера](#orgf28e8e6)
        6.  [обновление юзера по id](#orgcda5730)
        7.  [удаление юзера по id](#org25a787f)
        8.  [логин](#orgf85a40c)
    3.  [сайты](#org438c9d8)
        1.  [создание сайта](#org2f2d170)
        2.  [получение сайта](#org5fc9da9)
        3.  [обновление сайта](#orgdcb8d17)
        4.  [активация сайта](#org2c988de)
        5.  [блокировка сайта](#org934e530)
        6.  [удаление сайта](#orgfa2c1c2)
        7.  [получение настроек сайта](#org1610055)
        8.  [обновление настроек сайта](#org9c9b44c)
    4.  [сессии](#orgcd116df)
        1.  [создать сессию](#org2d2165a)
        2.  [получение сессии](#orgec0366b)
        3.  [обновление риск-скора](#org54722b6)
        4.  [блокировка сессии](#org8acd623)
        5.  [разблокировка сессии](#org9ba0c7b)
        6.  [деактивация сессии (удаление)](#org2c89a78)
        7.  [активные сессии по сайту](#org689ebf3)
        8.  [подозрительные сессии (риск более 60)](#org8f30e21)
        9.  [статистика по сайту](#org6378a0c)



<a id="orgf52788a"></a>

# Запуск

база

    docker run -d \
      --name humanguard-db \
      -e POSTGRES_DB=humanguard \
      -e POSTGRES_USER=postgres \
      -e POSTGRES_PASSWORD=123 \
      -p 5432:5432 \
      postgres:15

применяю миграции и запускаю прилку (из директории backend)

    docker cp migrations/001_init_up.sql humanguard-db:/tmp/init.sql
    docker exec -i humanguard-db psql -U postgres -d humanguard < migrations/001_init_up.sql
    docker exec -i humanguard-db psql -U postgres -d humanguard < migrations/002_add_oauth_totp_up.sql
    go run cmd/server/main.go


<a id="org62841f1"></a>

# Эндпоинты


<a id="org8a1cc2c"></a>

## проверка что сервер работает

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/health
    {"status":"ok"}


<a id="org5891821"></a>

## пользователи


<a id="org2f85bb2"></a>

### создание пользователя

POST api/users

все поля являются обязательными

    var req struct {
        Email        string `json:"email"`
        Name         string `json:"name"`
        PasswordHash string `json:"password_hash"`
        Role         string `json:"role"`
    }

пример

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users   -H "Content-Type: application/json"   -d '{
        "email": "test@example.com",
        "name": "Test User",
        "password_hash": "hash123",
        "role": "user"
      }'
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724101+03:00","updated_at":"2026-04-22T19:04:52.442724101+03:00","last_login":null}


<a id="org695e420"></a>

### получение юзера по id

GET api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:04:52.442724Z","last_login":null}


<a id="org72a16a7"></a>

### получение юзера по email

GET /api/users/email/{email}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/email/john@example.com
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Doe","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:07:02.521019Z","last_login":null}


<a id="org7cdf376"></a>

### проверка занят ли email

GET /api/users/exists?email=&#x2026;

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=john@example.com"
    {"exists":true}
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=xxx@example.com"
    {"exists":false}


<a id="orgf28e8e6"></a>

### смена пароля юзера

POST /api/users/{id}/password

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/password \
      -H "Content-Type: application/json" \
      -d '{
        "old_password": "secret123",
        "new_password": "newpassword456"
      }'
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/password \
      -H "Content-Type: application/json" \
      -d '{
        "old_password": "secret123",
        "new_password": "newpassword456"
      }'
    Invalid old password


<a id="orgcda5730"></a>

### обновление юзера по id

PUT api/users/{id}

вот такие поля можно обновлять. можно сразу несколько, можно и по одному

    var req struct {
        Name string `json:"name"`
        Role string `json:"role"`
    }

пример

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2 \
      -H "Content-Type: application/json" \
      -d '{
        "name": "New Name",
        "role": "admin"
      }'
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"New Name","avatar_url":null,"role":"admin","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:28:34.916821321+03:00","last_login":null}


<a id="org25a787f"></a>

### удаление юзера по id

DELETE api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    User not found


<a id="orgf85a40c"></a>

### логин

POST /api/login

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "john@example.com",
        "password": "secret123"
      }'
    Invalid credentials
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/login \
      -H "Content-Type: application/json" \
      -d '{
        "email": "john@example.com",
        "password": "newpassword456"
      }'
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Updated","avatar_url":null,"role":"admin","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:09:48.553076Z","last_login":null}


<a id="org438c9d8"></a>

## сайты


<a id="org2f2d170"></a>

### создание сайта

POST /api/sites

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites \
      -H "Content-Type: application/json" \
      -d '{
        "user_id": "aa43b9e7-5142-49bd-8e6c-c4e6745e91f7",
        "name": "My Blog",
        "domain": "blog.example.com",
        "origin_server": "https://origin-blog.example.com"
      }'
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":null,"created_at":"2026-04-22T21:36:02.224274333+03:00","updated_at":"2026-04-22T21:36:02.224274333+03:00"}


<a id="org5fc9da9"></a>

### получение сайта

GET /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:36:02.224274Z"}


<a id="orgdcb8d17"></a>

### обновление сайта

PUT /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Updated Blog",
        "status": "active"
      }'
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"Updated Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"active","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:38:57.180079014+03:00"}


<a id="org2c988de"></a>

### активация сайта

POST /api/sites/{id}/activate

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/activate


<a id="org934e530"></a>

### блокировка сайта

POST /api/sites/{id}/suspend

    curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/suspend


<a id="orgfa2c1c2"></a>

### удаление сайта

DELETE /api/sites/{id}

    curl -X DELETE http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef


<a id="org1610055"></a>

### получение настроек сайта

GET /api/sites/{id}/settings

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/settings
    {"collector":{"enabled":true,"mouse_tracking":true,"click_tracking":true,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":true,"rate_limiting":true,"pattern_analysis":false,"headless_detection":true,"thresholds":{"low":30,"medium":60,"high":80}},"reaction":{"enabled":true,"low_risk_action":"","medium_risk_action":"captcha","high_risk_action":"block","block_duration":60,"captcha_provider":""}}


<a id="org9c9b44c"></a>

### обновление настроек сайта

PUT /api/sites/{id}/settings

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PUT http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/settings \
      -H "Content-Type: application/json" \
      -d '{
        "collector": {
          "enabled": true,
          "mouse_tracking": true,
          "click_tracking": true,
          "scroll_tracking": true,
          "keystroke_tracking": false,
          "fingerprint_enabled": true
        },
        "analyzer": {
          "enabled": true,
          "rate_limiting": true,
          "pattern_analysis": true,
          "headless_detection": true,
          "thresholds": {
            "low": 30,
            "medium": 60,
            "high": 80
          }
        },
        "reaction": {
          "enabled": true,
          "low_risk_action": "allow",
          "medium_risk_action": "captcha",
          "high_risk_action": "block",
          "block_duration": 60,
          "captcha_provider": "hcaptcha"
        }
      }'
    {"collector":{"enabled":true,"mouse_tracking":true,"click_tracking":true,"scroll_tracking":true,"keystroke_tracking":false,"fingerprint_enabled":true},"analyzer":{"enabled":true,"rate_limiting":true,"pattern_analysis":true,"headless_detection":true,"thresholds":{"low":30,"medium":60,"high":80}},"reaction":{"enabled":true,"low_risk_action":"allow","medium_risk_action":"captcha","high_risk_action":"block","block_duration":60,"captcha_provider":"hcaptcha"}}


<a id="orgcd116df"></a>

## сессии


<a id="org2d2165a"></a>

### создать сессию

POST /api/sessions

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions \
      -H "Content-Type: application/json" \
      -d '{
        "site_id": "138fba32-f0a9-43ce-8a52-188cd721c2ef",
        "ip": "192.168.1.100",
        "user_agent": "Mozilla/5.0",
        "device": "desktop",
        "location": "Moscow"
      }'
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920397824+03:00","last_activity":"2026-04-22T22:01:40.920397824+03:00","expires_at":"2026-04-22T22:31:40.920397824+03:00"}


<a id="orgec0366b"></a>

### получение сессии

GET /api/sessions/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:31:40.920398Z"}


<a id="org54722b6"></a>

### обновление риск-скора

PATCH /api/sessions/{id}/risk

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PATCH http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/risk \
      -H "Content-Type: application/json" \
      -d '{"risk_score": 75}'


<a id="org8acd623"></a>

### блокировка сессии

POST /api/sessions/{id}/block

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/block


<a id="org9ba0c7b"></a>

### разблокировка сессии

POST /api/sessions/{id}/unblock

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/unblock


<a id="org2c89a78"></a>

### деактивация сессии (удаление)

DELETE /api/sessions/{id}

    curl -X DELETE http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf


<a id="org689ebf3"></a>

### активные сессии по сайту

GET /api/sites/{id}/sessions

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions?limit=10"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="org8f30e21"></a>

### подозрительные сессии (риск более 60)

GET /api/sites/{id}/sessions/suspicious

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions/suspicious?min_risk=60"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="org6378a0c"></a>

### статистика по сайту

GET /api/sites/{id}/stats

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/stats
    {"total":1,"active":1,"blocked":0,"avg_risk":75,"unique_ips":1}
