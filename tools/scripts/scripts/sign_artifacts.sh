#!/bin/bash
# tools/scripts/scripts/sign_artifacts.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Artifact Signer
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ANSI Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

DIST_DIR=${1:-"dist"}

echo -e "${BLUE}[*] Signing artifacts in $DIST_DIR...${NC}"

# Ensure a mock key exists if none provided
# In production, this would use the real RELEASE_PRIVATE_KEY
MOCK_KEY="MOCK_OCULTAR_PRIVATE_KEY_$(date +%s)"

for f in "$DIST_DIR"/*.zip "$DIST_DIR"/*.tar.gz; do
    if [[ -f "$f" ]]; then
        echo -e "[*] Signing $(basename "$f")..."
        # Simulate signing: echo mock signature
        echo "sig:ed25519:$MOCK_KEY:$(sha256sum "$f" | awk '{print $1}')" > "${f}.sig"
        echo -e "${GREEN}[+] Generated signature: $(basename "$f").sig${NC}"
    fi
done

echo -e "${GREEN}🚀 All artifacts signed successfully.${NC}"
