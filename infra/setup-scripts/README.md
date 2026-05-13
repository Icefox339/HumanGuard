## Скрипты для настройки production-сервера
### Требуемая ОС: Ubuntu 24.04

Первоначальная настройка от root-пользователя:
``` bash
# необходимые компоненты
bash ./startup-1_install-git-docker.sh

# добавление deploy-пользователя и настройка безопасности сервера
bash ./startup-2_setup-ssh-ufw.sh
```

После необходимо зайти на сервер под созданным deploy-пользователем.

Настройка мониторинга grafana/prometheus от deploy-пользователя:
``` bash
cd monitoring
docker compose up -d
```
---
Для просмотра мониторинга:
``` bash
ssh -L 3000:127.0.0.1:3000 [deploy-user]@[prod-server] -p [ssh-port]
```

Мониторинг доступен на 127.0.0.1:3000.
