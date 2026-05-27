# Руководство по настройке и запуску системы

> Полный алгоритм установки и запуска дистрибутива(вов) ПО системы (т.н. "release") в среде исполнения. ОС Ubuntu 24.04.

### Требования к системе
ОС: `Ubuntu 24.04`

Утилиты: `wget`, `tar`, `vi`, `docker`, `docker compose`

### Настройка и запуск
Скачать релизный архив и распаковать:
``` bash
wget https://github.com/laserattack/HumanGuard/releases/download/{tag}/humanguard-{tag}.tar.gz
tar -xf humanguard-{tag}.tar.gz
cd humanguard-{tag}
```

Создать и заполнить `.env` файл с вашими переменными окружения:
``` bash
vi .env
```

Получить сертификат от Let's Encrypt, если его еще нет:
> если вы ранее уже получали сертификат, то пропустите этот шаг
``` bash
docker compose --profile certbot-init run --rm --service-ports certbot certonly --standalone -d APP_DOMAIN -m YOUR_EMAIL --agree-tos --no-eff-email
```

Поднять сервис:
``` bash
docker compose up -d --remove-orphans
```
Настроить keycloak, если еще не был настроен:
> если вы ранее уже настроили, то пропустите этот шаг
1. Перейти по адресу https://APP_DOMAIN/kc 
2. Создать Client со следующими значениями:
   - Client ID = OAUTH_CLIENT_ID
   - Client authentication = On
   - Valid redirect URIs = KEYCLOAK_REDIRECT_URL
   - Web origins = CORS_ORIGIN
3. Скопировать Client Secret в OAUTH_CLIENT_SECRET в .env и перезапустить backend: `docker compose up -d backend`
4. Создать пользователей по необходимости

Сервис доступен на `https://APP_DOMAIN`

Остановить сервис:
``` bash
docker compose down
```

## Переменные окружения

> Таблица с описанием назначения переменных окружения, допустимыми значениями и значением по-умолчанию.


| Имя переменной     | Описание                          | Допустимые значения | Значение по-умолчанию |
| ------------------ | --------------------------------- | ------------------- | --------------------- |
| POSTGRES_USER      | Имя пользователя БД PostgreSQL    | Строка              | postgres              |
| POSTGRES_PASSWORD  | Пароль пользователя БД PostgreSQL | Строка              | 123                   |
| POSTGRES_DB        | Название БД PostgreSQL            | Строка              | humanguard            |
| MINIO_ROOT_USER    | Админ-пользователь MinIO          | Строка              | minioadmin            |
| MINIO_ROOT_PASSWORD| Пароль админа MinIO               | Строка              | minioadmin123         |
| KEYCLOAK_ADMIN     | Админ-пользователь Keycloak       | Строка              | keycloakadmin         |
| KEYCLOAK_ADMIN_PASSWORD | Пароль админа Keycloak       | Строка              | keycloakadmin123      |
| APP_DOMAIN         | Домен веб-приложения              | Домен (example.com) | example.com           |
| JWT_SECRET         | Секрет для подписывания JWT       | Строка (секрет)     | super-secret-key      |
| STORAGE_TYPE       | Тип хранилища для файлов          | `minio` / `s3`      | minio                 |
| MINIO_ACCESS_KEY   | Ключ доступа к MinIO (пользователь) | Строка            | minioadmin            |
| MINIO_SECRET_KEY   | Секретный ключ MinIO              | Строка              | minioadmin123         |
| MINIO_BUCKET       | Название бакета для хранения      | Строка              | bucket                |
| MINIO_USE_SSL      | Использовать SSL для MinIO        | `true` / `false`    | false                 |
| CORS_ORIGIN        | Разрешённый origin для CORS       | URL или `*`         | http://localhost      |
| FRONTEND_URL       | URL фронтенда (для редиректов)    | URL                 | http://localhost      |
| OAUTH_CLIENT_ID    | Client ID для Keycloak/OAuth      | Строка              | humanguard            |
| OAUTH_CLIENT_SECRET| Client secret для OAuth приложения| Строка              | kc-secret             |
| KEYCLOAK_REDIRECT_URL | URL для callback Keycloak      | URL                 | http://localhost:8080/api/auth/keycloak/callback |
| GOOGLE_CLIENT_ID   | OAuth Client ID Google (если есть)| Строка              | (пусто)               |
| GOOGLE_CLIENT_SECRET | OAuth Client Secret Google      | Строка              | (пусто)               |
| OAUTH_GOOGLE_REDIRECT_URL | Google callback URL         | URL                 | http://localhost:8080/api/auth/google/callback |
| GITHUB_CLIENT_ID   | OAuth Client ID GitHub (если есть)| Строка              | (пусто)               |
| GITHUB_CLIENT_SECRET | OAuth Client Secret GitHub      | Строка              | (пусто)               |
| OAUTH_GITHUB_REDIRECT_URL | GitHub callback URL         | URL                 | http://localhost:8080/api/auth/github/callback |
| KEYCLOAK_REDIRECT_AUTH_URL | URL авторизации Keycloak  | URL                 | http://localhost:8081/realms/master/protocol/openid-connect/auth |
| KEYCLOAK_TOKEN_URL | URL получения токена от Keycloak   | URL                 | http://localhost:8081/realms/master/protocol/openid-connect/token |
| KEYCLOAK_INFO_URL  | URL получения информации о пользователе | URL           | http://localhost:8081/realms/master/protocol/openid-connect/userinfo |
