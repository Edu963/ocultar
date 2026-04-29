#!/bin/bash
set -e

export OCU_VAULT_PATH="${OCU_VAULT_PATH:-/data/vault.db}"
export OCU_AUDIT_LOG_PATH="${OCU_AUDIT_LOG_PATH:-/data/audit.log}"
export OCU_PROXY_PORT="${OCU_PROXY_PORT:-8081}"

mkdir -p /data

cd /app

/app/ocultar-refinery --serve 8080 &
REFINERY_PID=$!

/app/ocultar-proxy &
PROXY_PID=$!

_term() {
    kill "$REFINERY_PID" "$PROXY_PID" 2>/dev/null
    wait "$REFINERY_PID" "$PROXY_PID" 2>/dev/null
    exit 0
}
trap _term TERM INT

# Exit the container if either process dies
wait -n "$REFINERY_PID" "$PROXY_PID"
STATUS=$?
kill "$REFINERY_PID" "$PROXY_PID" 2>/dev/null || true
exit $STATUS
