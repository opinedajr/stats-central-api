#!/usr/bin/env bash
set -euo pipefail

APP_NAME="stats-central-api"
APP_DIR="/opt/${APP_NAME}"
APP_USER="www-data"
SERVICE_NAME="${APP_NAME}.service"
BIN_PATH="${APP_DIR}/${APP_NAME}"

log() { printf '\033[1;34m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }
err() { printf '\033[1;31m[ERROR]\033[0m %s\n' "$*" >&2; exit 1; }

[[ "$(uname)" == "Linux" ]] || err "Este script roda apenas em Linux (VPS)."
[[ $EUID -eq 0 ]] || err "Rode com sudo: sudo $0"
command -v systemctl >/dev/null || err "systemd não encontrado."

[[ -f "${BIN_PATH}" ]] || err "Binário não encontrado em ${BIN_PATH}. Copie a nova versão antes:
  GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o bin/${APP_NAME} cmd/api/main.go
  scp bin/${APP_NAME} ${BIN_PATH}"

[[ -f "${APP_DIR}/.env" ]] || err ".env ausente em ${APP_DIR}/.env — rode install.sh primeiro."

log "Ajustando permissões do binário"
chmod 0755 "${BIN_PATH}"
chown "${APP_USER}:${APP_USER}" "${BIN_PATH}"

log "Reiniciando ${SERVICE_NAME}"
systemctl restart "${SERVICE_NAME}"

log "Aguardando serviço subir..."
sleep 2

if systemctl is-active --quiet "${SERVICE_NAME}"; then
  log "OK — ${SERVICE_NAME} ativo"
  systemctl --no-pager --full status "${SERVICE_NAME}" || true
else
  err "Serviço não subiu. Diagnóstico:
  systemctl status ${SERVICE_NAME}
  journalctl -u ${SERVICE_NAME} -n 50 --no-pager"
fi
