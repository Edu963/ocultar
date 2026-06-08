#!/bin/sh
set -e

# OCULTAR Enterprise Proxy Pre-Flight Validator
echo "================================================="
echo "   OCULTAR Enterprise Startup Validation      "
echo "================================================="

# 1. Master Key Validation
if [ -z "$OCU_MASTER_KEY" ]; then
    echo "[!] FATAL: OCU_MASTER_KEY must be set in the environment or .env file."
    exit 1
fi

if [ ${#OCU_MASTER_KEY} -lt 32 ]; then
    echo "[!] FATAL: OCU_MASTER_KEY must be at least 32 characters long."
    exit 1
fi
echo "[+] Validated OCU_MASTER_KEY entropy."

# 2. Vault Path Verification
VAULT_PATH=${OCU_VAULT_PATH:-/app/vault_data/vault.db}
VAULT_DIR=$(dirname "$VAULT_PATH")
if [ ! -d "$VAULT_DIR" ]; then
    echo "[!] WARNING: Directory $VAULT_DIR does not exist. The application may fail to create the SQLite duckdb file."
else
    echo "[+] Validated Vault persistence directory ($VAULT_DIR)."
fi

# 3. Model Engine (SLM) Healthcheck
SLM_URL=${SLM_HOST:-http://slm-ner:8080}
echo "[*] Polling SLM engine at $SLM_URL/health ..."
MAX_RETRIES=30
RETRY_COUNT=0
while ! wget -q --spider "$SLM_URL/health" ; do
    RETRY_COUNT=$((RETRY_COUNT+1))
    if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
        echo "[!] FATAL: OCULTAR SLM engine did not become healthy in time."
        exit 1
    fi
    echo "    - Waiting for SLM initialization (Attempt $RETRY_COUNT / $MAX_RETRIES)..."
    sleep 2
done
echo "[+] SLM engine is responsive and ready."

# 4. Connector Validation (Optional)
if [ -n "$SLACK_WORKSPACE_ID" ]; then
    if [ -z "$SLACK_TOKEN" ]; then
        echo "[!] WARNING: SLACK_WORKSPACE_ID is set but SLACK_TOKEN is missing. Slack connector will not start."
    else
        echo "[+] Slack connector configuration detected."
    fi
fi

echo "================================================="
echo "[+] All pre-flight checks passed! Starting Proxy."
echo "================================================="

# Execute the main entrypoint
exec "$@"
