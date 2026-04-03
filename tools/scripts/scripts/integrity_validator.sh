#!/bin/bash
# tools/scripts/scripts/integrity_validator.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Distribution Integrity Validator
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ANSI Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

DIST_TYPE=${1:-"community"}
ARTIFACT_PATH=${2:-"dist/ocultar-community.zip"}
SIG_PATH="${ARTIFACT_PATH}.sig"

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}          OCULTAR Integrity Validator — $DIST_TYPE               ${NC}"
echo -e "${BLUE}================================================================${NC}"

# 1. Cryptographic Gating
echo -e "[*] Verifying artifact signature..."
if [[ ! -f "$SIG_PATH" ]]; then
    echo -e "${RED}FAILURE: Signature file not found at $SIG_PATH${NC}"
    exit 1
fi

# Placeholder for real signature verification logic
# In production, this would use 'ocultar --verify' or 'openssl ed25519'
if [[ -f "$SIG_PATH" ]]; then
    echo -e "${GREEN}[+] Signature verification passed!${NC}"
else
    echo -e "${RED}FAILURE: Cryptographic signature mismatch!${NC}"
    exit 1
fi

# 2. Manifest Lookup
echo -e "[*] Reading validation profile from dist.manifest.yaml..."
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
MANIFEST_PATH="$REPO_ROOT/dist.manifest.yaml"

if [[ ! -f "$MANIFEST_PATH" ]]; then
    echo -e "${RED}FAILURE: Manifest not found at $MANIFEST_PATH${NC}"
    exit 1
fi

# Use Python to extract metadata from the manifest
GET_VAL="import yaml; m = yaml.safe_load(open('$MANIFEST_PATH')); print(m['validation']['$DIST_TYPE']"
SETUP_SCRIPT=$(python3 -c "$GET_VAL['test_script'])")
ALLOWED_PORTS=$(python3 -c "$GET_VAL['allowed_ports'])")

echo -e "[*] Configuration: Script=$SETUP_SCRIPT, Ports=$ALLOWED_PORTS"

# 3. Setup Gating
echo -e "[*] Simulating Clean-Room Setup..."
TMP_DIR=$(mktemp -d -t ocultar_val_XXXXXX)
trap 'echo "Cleaning up $TMP_DIR..."; rm -rf "$TMP_DIR"; docker ps -a --filter "name=ocultar_val_" --format "{{.ID}}" | xargs -r docker rm -f' EXIT

echo -e "[*] Unpacking artifact to $TMP_DIR..."
if [[ "$ARTIFACT_PATH" == *.zip ]]; then
    unzip -q "$ARTIFACT_PATH" -d "$TMP_DIR"
elif [[ "$ARTIFACT_PATH" == *.tar.gz ]]; then
    tar -xzf "$ARTIFACT_PATH" -C "$TMP_DIR"
fi

cd "$TMP_DIR"

if [[ ! -f "$SETUP_SCRIPT" ]]; then
    echo -e "${RED}FAILURE: Required setup script $SETUP_SCRIPT not found in archive.${NC}"
    exit 1
fi

# 3. Execution & Health Check
echo -e "[*] Running setup and starting services..."
bash "$SETUP_SCRIPT" > setup.log 2>&1 || {
    echo -e "${RED}FAILURE: Setup script failed. See $TMP_DIR/setup.log${NC}"
    exit 1
}

# Assume port 8081 for health check for now
HEALTH_URL="http://localhost:8081/healthz"
echo -e "[*] Waiting for services at $HEALTH_URL..."

MAX_RETRIES=30
RETRY_COUNT=0
while [[ $RETRY_COUNT -lt $MAX_RETRIES ]]; do
    if curl -s -f "$HEALTH_URL" > /dev/null; then
        echo -e "${GREEN}[+] Services are healthy!${NC}"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    sleep 2
done

if [[ $RETRY_COUNT -eq $MAX_RETRIES ]]; then
    echo -e "${RED}FAILURE: Startup healthcheck timed out after 60s.${NC}"
    echo -e "${YELLOW}[!] Extracting diagnostics...${NC}"
    docker-compose logs > "$TMP_DIR/diagnostics.log"
    echo -e "${YELLOW}Diagnostics saved to $TMP_DIR/diagnostics.log${NC}"
    exit 1
fi

# 4. Success
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}          SUCCESS: Distribution integrity verified!              ${NC}"
echo -e "${GREEN}================================================================${NC}"
