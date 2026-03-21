#!/usr/bin/env bash
# tools/scripts/scripts/build_release.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Release Builder
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
DIST_DIR="$REPO_ROOT/dist"

# ANSI Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${BLUE}Building OCULTAR Release Artifacts...${NC}"

# 1. Manifest Validation
if [[ ! -f "$REPO_ROOT/dist.manifest.yaml" ]]; then
    echo -e "${RED}Error: dist.manifest.yaml not found. Aborting.${NC}"
    exit 1
fi

# 2. Manifest-Driven Build Execution
echo -e "${BLUE}Executing Manifest-Driven Build...${NC}"
python3 "$SCRIPT_DIR/manifest_executor.py"

# 3. Cryptographic Signing
echo -e "${BLUE}Signing artifacts...${NC}"
bash "$SCRIPT_DIR/sign_artifacts.sh" "$DIST_DIR"

# 4. Distribution Integrity Validation (Gated)
echo -e "${BLUE}Running Integrity Validator (Smoke Test)...${NC}"
bash "$SCRIPT_DIR/integrity_validator.sh" "community" "$DIST_DIR/ocultar-community.zip"

echo -e "${GREEN}🚀 Release artifacts generated, signed, and verified in $DIST_DIR/${NC}"
ls -lh "$DIST_DIR"/*.zip "$DIST_DIR"/*.tar.gz "$DIST_DIR"/*.sig
