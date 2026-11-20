#!/usr/bin/env bash
set -euo pipefail

APP_NAME="stats-central-api"
APP_DIR="/opt/${APP_NAME}"
APP_USER="www-data"
SERVICE_NAME="${APP_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"

log() { printf '\033[1;34m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }
err() { printf '\033[1;31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

[[ "$(uname)" == "Linux" ]] || err "Este script roda apenas em Linux (VPS)."

[[ $EUID -eq 0 ]] || err "Rode com sudo: sudo $0"

command -v systemctl >/dev/null || err "systemd não encontrado."

log "Criando diretório ${APP_DIR}"
install -d -o "${APP_USER}" -g "${APP_USER}" -m 0755 "${APP_DIR}"

if [[ ! -f "${APP_DIR}/${APP_NAME}" ]]; then
  err "Binário não encontrado em ${APP_DIR}/${APP_NAME}. Copie o binário (GOOS=linux GOARCH=amd64) antes de rodar este script:
    GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o bin/${APP_NAME} cmd/api/main.go
    scp bin/${APP_NAME} ${APP_DIR}/${APP_NAME}"
fi

chmod 0755 "${APP_DIR}/${APP_NAME}"
chown "${APP_USER}:${APP_USER}" "${APP_DIR}/${APP_NAME}"

if [[ ! -f "${APP_DIR}/.env" ]]; then
  err ".env não encontrado em ${APP_DIR}/.env. Crie com chmod 600 antes:
    install -m 600 -o ${APP_USER} -g ${APP_USER} /dev/null ${APP_DIR}/.env
    # edite com SERVER_PORT, DB_DRIVER=mysql, DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, LOG_LEVEL"
fi
chmod 0600 "${APP_DIR}/.env"
chown "${APP_USER}:${APP_USER}" "${APP_DIR}/.env"

log "Escrevendo unit ${SERVICE_NAME}"
cat > "${SERVICE_PATH}" <<EOF
[Unit]
Description=Stats Central API
After=network.target mysql.service

[Service]
Type=simple
User=${APP_USER}
Group=${APP_USER}
WorkingDirectory=${APP_DIR}
EnvironmentFile=${APP_DIR}/.env
ExecStart=${APP_DIR}/${APP_NAME}
Restart=always
RestartSec=5

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${APP_DIR}
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
LockPersonality=true
RestrictRealtime=true
RestrictSUIDSGID=true

[Install]
WantedBy=multi-user.target
EOF

log "Habilitando e iniciando ${SERVICE_NAME}"
systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"
systemctl restart "${SERVICE_NAME}"

log "Aguardando serviço subir..."
sleep 2

if systemctl is-active --quiet "${SERVICE_NAME}"; then
  log "OK — ${SERVICE_NAME} está ativo"
  systemctl --no-pager --full status "${SERVICE_NAME}" || true
else
  err "Serviço não subiu. Diagnóstico:
  systemctl status ${SERVICE_NAME}
  journalctl -u ${SERVICE_NAME} -n 50 --no-pager"
fi
