#!/bin/bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

if [[ $EUID -ne 0 ]]; then
    error "Этот скрипт должен запускаться от root (sudo)."
fi

info "Настройка безопасности сервера..."

info "Установка необходимых пакетов..."
apt-get update
apt-get install -y openssh-server ufw fail2ban libpam-pwquality

read -p "Введите имя нового пользователя для удаленного доступа (не root): " NEW_USER
if [[ -z "$NEW_USER" ]]; then
    error "Имя пользователя не может быть пустым."
fi

read -p "Введите новый порт для SSH (например, 2222): " SSH_PORT
if [[ ! "$SSH_PORT" =~ ^[0-9]+$ ]] || [[ "$SSH_PORT" -lt 1024 ]] || [[ "$SSH_PORT" -gt 65535 ]]; then
    error "Порт должен быть числом от 1024 до 65535."
fi

# создание нового пользователя
if id "$NEW_USER" &>/dev/null; then
    warn "Пользователь $NEW_USER уже существует. Пропускаем создание."
else
    info "Создание пользователя $NEW_USER..."
    useradd -m -s /bin/bash "$NEW_USER"
    # Добавление в группу sudo
    usermod -aG sudo "$NEW_USER"
    # Установка пароля
    echo "Задайте пароль для пользователя $NEW_USER (требования: минимум 12 символов, хотя бы 1 цифра, 1 спецсимвол, заглавная и строчная буквы):"
    while true; do
        passwd "$NEW_USER" && break
        echo "Пароль не удовлетворяет политике. Повторите попытку."
    done
fi

# Добавление пользователя в группу docker
if getent group docker > /dev/null; then
    usermod -aG docker "$NEW_USER"
    info "Пользователь $NEW_USER добавлен в группу docker."
else
    groupadd docker
    usermod -aG docker "$NEW_USER"
    info "Группа docker создана, пользователь добавлен."
fi

# настройка политики сложных паролей (PAM и pwquality)
info "Настройка политики паролей..."
cat > /etc/security/pwquality.conf <<EOF
# Минимальная длина пароля
minlen = 12
# Требовать хотя бы одну цифру
dcredit = -1
# Хотя бы одну заглавную букву
ucredit = -1
# Хотя бы одну строчную букву
lcredit = -1
# Хотя бы один спецсимвол
ocredit = -1
# Запрет повторяющихся символов
maxrepeat = 3
# Запрет включения имени пользователя
usercheck = 1
# Словарная проверка (простые пароли)
dictcheck = 1
EOF

# включение pwquality в PAM для паролей
sed -i 's/^password.*pam_unix.so.*/password requisite pam_pwquality.so retry=3\npassword sufficient pam_unix.so obscure use_authtok try_first_pass yescrypt/' /etc/pam.d/common-password

# настройка SSH
info "Настройка SSH..."
cat >> "/etc/ssh/sshd_config" <<EOF
Port $SSH_PORT
PermitRootLogin no
PasswordAuthentication no
ChallengeResponseAuthentication no
PubkeyAuthentication yes
PermitEmptyPasswords no
UsePAM yes
MaxAuthTries 3
LoginGraceTime 30
AuthenticationMethods publickey
AllowUsers $NEW_USER
EOF

# добавление SSH-ключа для нового пользователя
info "Настройка SSH-ключа для пользователя $NEW_USER..."
USER_SSH_DIR="/home/$NEW_USER/.ssh"
mkdir -p "$USER_SSH_DIR"
chmod 700 "$USER_SSH_DIR"
AUTH_KEYS_FILE="$USER_SSH_DIR/authorized_keys"

echo "Введите открытый ключ (публичный SSH-ключ) для пользователя $NEW_USER (строка начинается с ssh-rsa, ssh-ed25519 и т.д.):"
read -r PUBLIC_KEY
if [[ -n "$PUBLIC_KEY" ]]; then
    echo "$PUBLIC_KEY" >> "$AUTH_KEYS_FILE"
    chmod 600 "$AUTH_KEYS_FILE"
    chown -R "$NEW_USER":"$NEW_USER" "$USER_SSH_DIR"
    info "Ключ добавлен."
else
    warn "Ключ не добавлен. Пользователь не сможет войти по SSH. Пожалуйста, добавьте ключ вручную."
fi

# настройка fail2ban для защиты от перебора (на новом порту)
info "Настройка fail2ban..."
cat > /etc/fail2ban/jail.local <<EOF
[DEFAULT]
bantime = 1h
findtime = 10m
maxretry = 3

[sshd]
enabled = true
port = $SSH_PORT
logpath = %(sshd_log)s
backend = %(sshd_backend)s
EOF

systemctl restart fail2ban
systemctl enable fail2ban

# настройка ufw (межсетевой экран)
info "Настройка ufw..."
# Сброс правил
ufw --force disable
ufw default deny incoming
ufw default allow outgoing

# Разрешаем новый SSH-порт
ufw allow "$SSH_PORT"/tcp comment 'SSH new port'
# Если нужно открыть порты приложения (например, HTTP/HTTPS)
read -p "Открыть порты 80 (HTTP) и 443 (HTTPS) для приложения? (y/n): " OPEN_WEB
if [[ "$OPEN_WEB" =~ ^[Yy]$ ]]; then
    ufw allow 80/tcp comment 'HTTP'
    ufw allow 443/tcp comment 'HTTPS'
fi

# Включаем ufw
ufw --force enable

info "Перезапуск SSH для применения изменений..."
systemctl restart ssh


info "Рекомендации по запуску приложений:"
echo "  - Запускайте сервисы (например, Docker контейнеры) от непривилегированного пользователя, а не от root."
echo "  - Используйте read-only файловые системы для контейнеров (флаг --read-only)."


echo ""
warn "SSH работает на порту $SSH_PORT. Убедитесь, что этот порт разрешен."
warn "root вход запрещён. Используйте пользователя $NEW_USER и SSH-ключ."
warn "Не закрывайте текущую сессию root, пока не проверите новое подключение:"
echo "  ssh -p $SSH_PORT $NEW_USER@$(curl -s ifconfig.me)"
read -p "Убедитесь, что новое подключение работает, и нажмите Enter для завершения скрипта..."

info "Настройка завершена."
