
# Table of Contents

1.  [Запуск](#org050be93)
2.  [Эндпоинты](#orgb8afd0b)
    1.  [проверка что сервер работает](#orgf67ed86)
    2.  [пользователи](#org34cf2f7)
        1.  [создание пользователя](#orgd10b934)
        2.  [получение юзера по id](#org817be62)
        3.  [получение юзера по email](#orgfb1ded8)
        4.  [проверка занят ли email](#org9008506)
        5.  [смена пароля юзера](#org5b3a56b)
        6.  [обновление юзера по id](#orge7e2896)
        7.  [обновление аватарки по id](#org6773d4d)
        8.  [удаление юзера по id](#orgd59f97b)
        9.  [логин](#orgd70ee62)
        10. [получение юзера по oauthId](#org10f5d8c)
    3.  [сайты](#org82ea584)
        1.  [создание сайта](#org123b162)
        2.  [получение сайта](#org73ff139)
        3.  [обновление сайта](#org7c05989)
        4.  [активация сайта](#orgd61c692)
        5.  [блокировка сайта](#org6c9bed4)
        6.  [удаление сайта](#orgd77a21c)
        7.  [получение настроек сайта](#org7f4c394)
        8.  [обновление настроек сайта](#org1fdb813)
    4.  [сессии](#orge2a110e)
        1.  [создать сессию](#org0b4fd36)
        2.  [получение сессии](#org5ac9eab)
        3.  [обновление риск-скора](#org972b9c4)
        4.  [блокировка сессии](#orgeeaee0e)
        5.  [разблокировка сессии](#org29bdf8c)
        6.  [деактивация сессии (удаление)](#orgce26db5)
        7.  [активные сессии по сайту](#orgc5a8be3)
        8.  [подозрительные сессии (риск более 60)](#org87bd099)
        9.  [статистика по сайту](#org960261f)



<a id="org050be93"></a>

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


<a id="orgb8afd0b"></a>

# Эндпоинты


<a id="orgf67ed86"></a>

## проверка что сервер работает

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/health
    {"status":"ok"}


<a id="org34cf2f7"></a>

## пользователи


<a id="orgd10b934"></a>

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


<a id="org817be62"></a>

### получение юзера по id

GET api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    {"id":"30508444-9b8e-4e29-ba39-89a393d0bed2","email":"test@example.com","name":"Test User","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T19:04:52.442724Z","updated_at":"2026-04-22T19:04:52.442724Z","last_login":null}


<a id="orgfb1ded8"></a>

### получение юзера по email

GET /api/users/email/{email}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/email/john@example.com
    {"id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","email":"john@example.com","name":"John Doe","avatar_url":null,"role":"user","oauth_provider":null,"created_at":"2026-04-22T20:07:02.521019Z","updated_at":"2026-04-22T20:07:02.521019Z","last_login":null}


<a id="org9008506"></a>

### проверка занят ли email

GET /api/users/exists?email=&#x2026;

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=john@example.com"
    {"exists":true}
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/users/exists?email=xxx@example.com"
    {"exists":false}


<a id="org5b3a56b"></a>

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


<a id="orge7e2896"></a>

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


<a id="org6773d4d"></a>

### обновление аватарки по id

POST /api/users/{id}/avatar

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/users/aa43b9e7-5142-49bd-8e6c-c4e6745e91f7/avatar \
      -H "Content-Type: application/json" \
      -d '{
        "avatar_url": "https://example.com/avatars/user123.jpg"
      }'


<a id="orgd59f97b"></a>

### удаление юзера по id

DELETE api/users/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X DELETE http://localhost:8080/api/users/30508444-9b8e-4e29-ba39-89a393d0bed2
    User not found


<a id="orgd70ee62"></a>

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


<a id="org10f5d8c"></a>

### получение юзера по oauthId

GET /api/users/oauth/{provider}/{oauthId}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/users/oauth/google/123456789
    {"id":"d1e415b6-3855-4714-affc-7486231d0af6","email":"googleuser@example.com","name":"Google User","avatar_url":null,"role":"user","oauth_provider":"google","created_at":"2026-04-22T19:39:20.573219Z","updated_at":"2026-04-22T19:39:20.573219Z","last_login":null}


<a id="org82ea584"></a>

## сайты


<a id="org123b162"></a>

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


<a id="org73ff139"></a>

### получение сайта

GET /api/sites/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef
    {"id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","user_id":"aa43b9e7-5142-49bd-8e6c-c4e6745e91f7","name":"My Blog","domain":"blog.example.com","origin_server":"https://origin-blog.example.com","status":"verifying","settings":{"collector":{"enabled":false,"mouse_tracking":false,"click_tracking":false,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":false,"rate_limiting":false,"pattern_analysis":false,"headless_detection":false,"thresholds":{"low":0,"medium":0,"high":0}},"reaction":{"enabled":false,"low_risk_action":"","medium_risk_action":"","high_risk_action":"","block_duration":0,"captcha_provider":""}},"created_at":"2026-04-22T21:36:02.224274Z","updated_at":"2026-04-22T21:36:02.224274Z"}


<a id="org7c05989"></a>

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


<a id="orgd61c692"></a>

### активация сайта

POST /api/sites/{id}/activate

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/activate


<a id="org6c9bed4"></a>

### блокировка сайта

POST /api/sites/{id}/suspend

    curl -X POST http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/suspend


<a id="orgd77a21c"></a>

### удаление сайта

DELETE /api/sites/{id}

    curl -X DELETE http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef


<a id="org7f4c394"></a>

### получение настроек сайта

GET /api/sites/{id}/settings

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/settings
    {"collector":{"enabled":true,"mouse_tracking":true,"click_tracking":true,"scroll_tracking":false,"keystroke_tracking":false,"fingerprint_enabled":false},"analyzer":{"enabled":true,"rate_limiting":true,"pattern_analysis":false,"headless_detection":true,"thresholds":{"low":30,"medium":60,"high":80}},"reaction":{"enabled":true,"low_risk_action":"","medium_risk_action":"captcha","high_risk_action":"block","block_duration":60,"captcha_provider":""}}


<a id="org1fdb813"></a>

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


<a id="orge2a110e"></a>

## сессии


<a id="org0b4fd36"></a>

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


<a id="org5ac9eab"></a>

### получение сессии

GET /api/sessions/{id}

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf
    {"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":0,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:31:40.920398Z"}


<a id="org972b9c4"></a>

### обновление риск-скора

PATCH /api/sessions/{id}/risk

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X PATCH http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/risk \
      -H "Content-Type: application/json" \
      -d '{"risk_score": 75}'


<a id="orgeeaee0e"></a>

### блокировка сессии

POST /api/sessions/{id}/block

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/block


<a id="org29bdf8c"></a>

### разблокировка сессии

POST /api/sessions/{id}/unblock

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl -X POST http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf/unblock


<a id="orgce26db5"></a>

### деактивация сессии (удаление)

DELETE /api/sessions/{id}

    curl -X DELETE http://localhost:8080/api/sessions/a5f668cf-d902-4f5d-a3f6-fb167512eebf


<a id="orgc5a8be3"></a>

### активные сессии по сайту

GET /api/sites/{id}/sessions

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions?limit=10"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="org87bd099"></a>

### подозрительные сессии (риск более 60)

GET /api/sites/{id}/sessions/suspicious

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl "http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/sessions/suspicious?min_risk=60"
    [{"id":"a5f668cf-d902-4f5d-a3f6-fb167512eebf","site_id":"138fba32-f0a9-43ce-8a52-188cd721c2ef","ip":"192.168.1.100","user_agent":"Mozilla/5.0","device":"desktop","location":"Moscow","is_active":true,"risk_score":75,"is_blocked":false,"captcha_shown":false,"created_at":"2026-04-22T22:01:40.920398Z","last_activity":"2026-04-22T22:01:40.920398Z","expires_at":"2026-04-22T22:33:37.691064Z"}]


<a id="org960261f"></a>

### статистика по сайту

GET /api/sites/{id}/stats

    ~/projects/HumanGuard/backend
    [serr@lap]-> curl http://localhost:8080/api/sites/138fba32-f0a9-43ce-8a52-188cd721c2ef/stats
    {"total":1,"active":1,"blocked":0,"avg_risk":75,"unique_ips":1}
