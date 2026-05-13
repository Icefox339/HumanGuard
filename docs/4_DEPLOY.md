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
docker compose --profile certbot-init run --rm --service-ports certbot certonly --standalone -d ВАШ_ДОМЕН -m ВАШ_EMAIL --agree-tos --no-eff-email
```

Поднять сервис:
``` bash
docker compose up -d --remove-orphans
```

Остановить сервис:
``` bash
docker compose down
```

## Переменные окружения

> Таблица с описанием назначения переменных окружения, допустимыми значениями и значением по-умолчанию.

| Имя переменной     | Описание                          | Допустимые значения | Значение по-умолчанию |
| ------------------ | --------------------------------- | ------------------- | --------------------- |
| POSTGRES_USER      | Имя пользователя БД PostgreSQL    | Строка              | postgres              |
| POSTGRES_PASSWORD  | Пароль пользователя БД PostgreSQL | Строка              | change-me             |
| POSTGRES_DB        | Название БД PostgreSQL            | Строка              | humanguard            |
| APP_DOMAIN         | Домен веб-приложения              | Строка              | example.com           |
