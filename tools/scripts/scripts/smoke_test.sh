#!/usr/bin/env bash
# tools/scripts/scripts/smoke_test.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Proxy Smoke Test
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

PROXY_PORT=${OCU_PROXY_PORT:-8081}
PROXY_URL="http://localhost:${PROXY_PORT}"
HEALTH_URL="${PROXY_URL}/healthz"
COMPLETIONS_URL="${PROXY_URL}/v1/chat/completions"

# ANSI Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# 1. Health Check
if ! curl -s -f "$HEALTH_URL" > /dev/null; then
    # Some older community versions might not have /healthz on 8081, 
    # instead checking if port is open
    if ! curl -s "http://localhost:${PROXY_PORT}/" > /dev/null; then
        echo -e "${RED}[!] Proxy is NOT responding on port ${PROXY_PORT}. Ensure it is running.${NC}"
        exit 1
    fi
fi
echo -e "${GREEN}[+] Proxy is healthy!${NC}"

# 2. Leaky Payload Test
echo "[*] Running smoke test with leaky payload..."

RESPONSE=$(curl -s -X POST "$COMPLETIONS_URL" \
  -H "Content-Type: application/json" \
  -H "Ocultar-Target: http://echo-upstream:8080" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Email me at leaky@example.com"}]
  }')

if echo "$RESPONSE" | grep -q "\[EMAIL_"; then
    echo -e "${GREEN}[+] SUCCESS: PII successfully intercepted and redacted!${NC}"
else
    echo -e "${RED}[!] FAILURE: PII was NOT redacted. Response received:${NC}"
    echo "$RESPONSE"
    exit 1
fi
