# HumanGuard

**HumanGuard** — это SaaS-решение для защиты веб-сайтов от автоматизированных атак и ботов. Система работает как обратный прокси, анализируя поведение посетителей и HTTP-запросы для выявления аномальной активности.

## Содержание

- [Общее описание](#-общее-описание)
- [Функциональное назначение](#-функциональное-назначение)
- [Алгоритм настройки и запуска](#-алгоритм-настройки-и-запуска)
- [Требования к аппаратному обеспечению](#-требования-к-аппаратному-обеспечению)
- [Описание безопасности системы](#-описание-безопасности-системы)

## Общее описание

### Назначение

- Защита веб-сайтов от ботов и автоматизированных атак
- Анализ поведенческих метрик посетителей (движения мыши, клики, скроллинг)
- Детектирование headless-браузеров и подозрительных паттернов
- Автоматическая блокировка или показ CAPTCHA при высоком уровне риска
- Предоставление API для интеграции со сторонними сервисами

### Архитектура системы

| Модуль | Назначение |
|--------|------------|
| **Сбор поведенческих данных** | Сбор клиентских метрик на стороне браузера |
| **Анализ паттернов обращений** | Анализ HTTP-запросов на серверной стороне |
| **Детекция аномалий** | Присвоение каждой сессии оценки риска (0–100) |
| **Реагирование** | Настраиваемые действия в зависимости от уровня риска |
| **Логирование и мониторинг** | Запись событий, экспорт метрик в Prometheus, визуализация в Grafana |
| **Управление настройками** | Настройка порогов срабатывания |
| **Аутентификация** | Регистрация по email с 2FA, OAuth, управление API ключами |
| **Файловое хранилище** | Загрузка файлов, расшаривание по токенам, поддержка MinIO S3 |

### Среда исполнения

| Компонент | Версии |
|-----------|--------|
| Go | 1.25+ |
| React | 18+ |
| Docker | 29+ |
| Docker Compose | 2+ |
| Nginx | alpine |
| PostgreSQL | 15 |
| MinIO | latest |

## Функциональное назначение

### API Эндпоинты

| Метод | Путь | Назначение | Auth |
|-------|------|------------|------|
| POST | `/api/users` | Регистрация пользователя | Нет |
| POST | `/api/login` | Вход с TOTP 2FA | Нет |
| GET | `/api/auth/google/login` | OAuth авторизация Google | Нет |
| GET | `/api/auth/github/login` | OAuth авторизация GitHub | Нет |
| GET | `/api/auth/keycloak/login` | OAuth авторизация Keycloak | Нет |
| POST | `/api/check` | Проверка сессии (для nginx reverse-proxy) | Нет |
| POST | `/api/behavior/{id}` | Отправка поведенческих метрик | Нет |
| GET | `/api/me` | Профиль текущего пользователя | JWT/API Key |
| GET | `/api/sites` | Список сайтов пользователя | JWT/API Key |
| POST | `/api/sites` | Создание сайта | JWT/API Key |
| GET | `/api/sites/{id}/stats` | Статистика по сайту | JWT/API Key |
| GET | `/api/sites/{id}/sessions` | Активные сессии сайта | JWT/API Key |
| GET | `/api/sites/{id}/sessions/suspicious` | Подозрительные сессии | JWT/API Key |
| PUT | `/api/sites/{id}/settings` | Обновление настроек сайта | JWT/API Key |
| POST | `/api/sites/{id}/activate` | Активация сайта | JWT/API Key |
| POST | `/api/sites/{id}/suspend` | Блокировка сайта | JWT/API Key |
| POST | `/api/keys` | Создание API ключа | JWT |
| GET | `/api/keys` | Список API ключей | JWT |
| DELETE | `/api/keys/{id}` | Отзыв API ключа | JWT |
| GET | `/api/files` | Список файлов пользователя | JWT/API Key |
| POST | `/api/files/upload` | Загрузка файла | JWT/API Key |
| GET | `/api/files/{id}` | Скачивание файла | JWT/API Key |
| POST | `/api/files/share` | Создать публичную ссылку | JWT/API Key |
| GET | `/api/files/share/{token}` | Скачать по публичной ссылке | Нет |
| POST | `/api/sessions/{id}/block` | Блокировка сессии | JWT/API Key |
| GET | `/metrics` | Метрики для Prometheus | internal |
| GET | `/health` | Health check | Нет |

### Коды ошибок

| Код | Описание |
|-----|----------|
| 200 | Успешный запрос |
| 201 | Ресурс создан |
| 400 | Неверный запрос |
| 401 | Неавторизован / неверные учётные данные |
| 403 | Доступ запрещён |
| 404 | Ресурс не найден |
| 409 | Конфликт (email уже занят / домен уже используется) |
| 413 | Файл слишком большой (превышает 5GB) |
| 415 | Неподдерживаемый тип файла |
| 429 | Слишком много запросов (превышен rate limit) |
| 500 | Внутренняя ошибка сервера |

### Возможности пользователя

1. **Регистрация и вход**
   - Регистрация по email с паролем (мин. 8 символов)
   - Двухфакторная аутентификация (TOTP) — обязательна
   - Вход через OAuth (Google, GitHub, Keycloak)

2. **Управление сайтами**
   - Создание сайта (указать domain и origin-server)
   - Активация/блокировка сайта
   - Настройка порогов риска и действий (allow/captcha/block)
   - Просмотр активных и подозрительных сессий

3. **Управление API ключами**
   - Создание API ключей для автоматизации
   - Настройка срока действия (дни)
   - Отзыв ключей

4. **Файлы**
   - Загрузка файлов (до 5GB)
   - Создание публичных ссылок с ограничением по времени
   - Скачивание по прямой ссылке (без авторизации)

5. **Профиль**
   - Смена пароля
   - Обновление аватара (через URL или загрузку изображения)

## Алгоритм настройки и запуска

### Структура директорий

    HumanGuard/
    ├── backend/                 # Go бэкенд
    │   ├── cmd/server/         # Точка входа
    │   ├── auth/               # Аутентификация
    │   ├── handlers/           # HTTP обработчики
    │   ├── storage/            # PostgreSQL и MinIO
    │   ├── detector/           # Детектор аномалий
    │   ├── reaction/           # Реакции на риск
    │   ├── metrics/            # Prometheus метрики
    │   ├── middleware/         # CSP, RateLimit, RequestID
    │   ├── migrations/         # SQL миграции
    │   └── Dockerfile
    ├── frontend/               # React фронтенд
    │   ├── src/
    │   ├── nginx.conf
    │   └── Dockerfile
    ├── infra/                  # Инфраструктура
    │   ├── nginx/
    │   ├── docker-compose.yml
    │   ├── docker-compose.release.yml
    │   ├── .env.example
    │   └── setup-scripts/
    ├── docs/                   # Документация
    ├── .github/workflows/      # CI/CD пайплайны
    └── README.md

### Быстрый старт для разработки

#### 1. Запуск базы данных

    docker run -d \
      --name humanguard-db \
      -e POSTGRES_DB=humanguard \
      -e POSTGRES_USER=postgres \
      -e POSTGRES_PASSWORD=123 \
      -p 5432:5432 \
      postgres:15

#### 2. Применение миграций

    cd backend
    docker cp migrations/001_init_up.sql humanguard-db:/tmp/init.sql
    docker exec -i humanguard-db psql -U postgres -d humanguard < migrations/001_init_up.sql

#### 3. Запуск MinIO (опционально)

    docker run -d \
      --name minio \
      -p 9000:9000 \
      -p 9001:9001 \
      -e "MINIO_ROOT_USER=minioadmin" \
      -e "MINIO_ROOT_PASSWORD=minioadmin123" \
      minio/minio server /data --console-address ":9001"

#### 4. Запуск бэкенда

    cd backend
    go run cmd/server/main.go

Сервер доступен на `http://localhost:8080`

#### 5. Запуск фронтенда

    cd frontend
    npm install
    npm run dev

Фронтенд доступен на `http://localhost:5173`

#### 6. Проверка работоспособности

    # Health check
    curl http://localhost:8080/health

    # Prometheus метрики
    curl http://localhost:8080/metrics

### Production установка на Ubuntu 24.04

#### Требования к системе

- **ОС:** Ubuntu 24.04
- **Утилиты:** wget, tar, vi, docker, docker compose

#### Шаги установки

**1. Скачать релизный архив**

    wget https://github.com/laserattack/HumanGuard/releases/download/{tag}/humanguard-{tag}.tar.gz
    tar -xf humanguard-{tag}.tar.gz
    cd humanguard-{tag}

**2. Настроить переменные окружения**

    vi .env

Пример `.env`:

    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=secure_password
    POSTGRES_DB=humanguard
    APP_DOMAIN=example.com
    JWT_SECRET=your-super-secret-key-32-chars-minimum
    STORAGE_TYPE=local

**3. Получить SSL сертификат Let's Encrypt**

    docker compose --profile certbot-init run --rm --service-ports certbot certonly --standalone -d example.com -m admin@example.com --agree-tos --no-eff-email

**4. Запустить сервис**

    docker compose up -d --remove-orphans

Сервис доступен на `https://ваш-домен`

**5. Остановить сервис**

    docker compose down

### Переменные окружения

| Имя | Описание | Значение по умолчанию |
|-----|----------|----------------------|
| `POSTGRES_USER` | Пользователь БД | postgres |
| `POSTGRES_PASSWORD` | Пароль БД | change-me |
| `POSTGRES_DB` | Название БД | humanguard |
| `APP_DOMAIN` | Домен приложения | example.com |
| `JWT_SECRET` | Секрет JWT (32+ символов) | super-secret-key |
| `STORAGE_TYPE` | Тип хранилища (local/minio) | local |
| `MINIO_ENDPOINT` | Адрес MinIO | localhost:9000 |
| `GOOGLE_CLIENT_ID` | OAuth Client ID Google | (пусто) |
| `GITHUB_CLIENT_ID` | OAuth Client ID GitHub | (пусто) |

### CI/CD пайплайн

#### GitHub Secrets для CI

| Secret | Описание |
|--------|----------|
| `SONAR_TOKEN` | Токен для SonarQube |

#### Secrets для деплоя

| Secret | Описание |
|--------|----------|
| `DEPLOY_HOST` | IP адрес сервера |
| `DEPLOY_USER` | Пользователь SSH |
| `SSH_PRIVATE_KEY` | Приватный SSH ключ |
| `SSH_PORT` | Порт SSH |
| `POSTGRES_USER` | Пользователь БД |
| `POSTGRES_PASSWORD` | Пароль БД |
| `POSTGRES_DB` | Название БД |
| `LETSENCRYPT_EMAIL` | Email для SSL сертификата |

## Требования к аппаратному обеспечению

| Компонент | Минимальные | Рекомендуемые |
|-----------|-------------|---------------|
| CPU | 2 ядра | 4 ядра |
| RAM | 2 GB | 4 GB |
| Диск | 10 GB | 30 GB SSD |

## Описание безопасности системы

### Аутентификация и авторизация

| Механизм | Реализация |
|----------|------------|
| **JWT токены** | HMAC-SHA256 подпись, срок жизни 24 часа, содержит user_id, role, session_id |
| **2FA (TOTP)** | Обязательна для всех пользователей, 6-значный код, обновляется каждые 30 секунд |
| **OAuth 2.0** | Поддержка Google, GitHub, Keycloak. Автоматическое создание пользователя |
| **API ключи** | Формат `hg_v1_{32 байта hex}`, хешируются в БД (SHA-256) |
| **Роли** | `user` — базовый доступ, `admin` — полный доступ |

### Защита интерфейсов

| Мера | Реализация |
|------|------------|
| **CSP** | Заголовок Content-Security-Policy с ограничением скриптов, стилей |
| **CORS** | Ограниченный список доверенных доменов |
| **Rate Limiting** | login: 5/мин, check: 100/мин, behavior: 300/мин. Возврат 429 |
| **Request ID** | UUIDv7 в каждом запросе, заголовок X-Request-ID |
| **CSRF** | Токены в формах (в разработке) |

### Защита данных

| Данные | Защита |
|--------|--------|
| **Пароли** | Хеширование bcrypt (cost 10) |
| **API ключи** | Хеширование SHA-256, ключ отдаётся только один раз |
| **JWT секрет** | Переменная окружения JWT_SECRET |
| **TOTP секрет** | Хранится в БД, отдаётся только при создании (QR код) |
| **Файлы** | Доступ через JWT или API ключ. Публичные ссылки с expiration |
| **Соединение с БД** | SSL/TLS (опционально) |

### Мониторинг безопасности

| Инструмент | Назначение |
|------------|------------|
| **Prometheus** | Сбор метрик: количество запросов, активные сессии, риск-скоры |
| **Grafana** | Визуализация: CPU/RAM, RPS, rate limit ошибки |
| **Node Exporter** | Метрики сервера |
| **Semgrep (CI)** | SAST статический анализ кода |
| **SonarQube (CI)** | Анализ качества кода и уязвимостей |
| **Trivy (CI)** | Сканирование Docker образов |

### Журналирование

Каждый HTTP запрос логируется в формате JSON:

    {
      "request_id": "019e2752-21c8-7457-bcec-be1665692a65",
      "method": "POST",
      "path": "/api/check",
      "duration": "1.2ms",
      "auth_method": "api_key",
      "remote_addr": "192.168.1.100"
    }

### Безопасность сессий

- **User Session** — для аутентифицированных пользователей. TTL 24 часа
- **Visitor Session** — для посетителей защищаемых сайтов. TTL 30 минут
- **Защита от подделки** — сессия привязана к site_id

### Обеспечение безопасности среды эксплуатации

Скрипты настройки сервера (`infra/setup-scripts/`):

| Скрипт | Назначение |
|--------|------------|
| `startup-1_install-git-docker.sh` | Установка Git, Docker, Docker Compose |
| `startup-2_setup-ssh-ufw.sh` | Настройка SSH, UFW, fail2ban, политика паролей |

#### Требования к серверу (hardening)

- Запрещён вход под root по SSH (используется пользователь `deploy`)
- Настроена политика сложных паролей (minlen=12, цифры, спецсимволы)
- Включена защита от перебора паролей (fail2ban)
- Настроено подключение по SSH-ключам
- Порт SSH перенесён на нестандартный (2222)
- Настроен UFW (разрешены только 80, 443, SSH порт)
- Приложение запускается в Docker от непривилегированного пользователя
- Readonly файловая система для контейнеров
- Nginx reverse-proxy с TLS (Let's Encrypt)
- HSTS заголовок (`Strict-Transport-Security: max-age=31536000; includeSubDomains; preload`)

## Документация

- [Техническое задание](docs/1_SRS.md)
- [Описание системы](docs/2_SYSTEM_SPECS.md)
- [Руководство пользователя](docs/3_USER_SPECS.md)
- [Руководство по развёртыванию](docs/4_DEPLOY.md)
- [Описание мер безопасности](docs/5_SECURITY.md)
- [API документация](docs/api/README.md)
- [OpenAPI спецификация](docs/api/swagger.yaml)