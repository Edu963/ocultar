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

# 3. Post-Build Compliance
echo -e "${BLUE}Running parity checks...${NC}"
bash "$SCRIPT_DIR/check_parity.sh" "$DIST_DIR"

# 4. Final Packaging
echo -e "${BLUE}Finalizing archives...${NC}"
(cd "$DIST_DIR" && zip -q -r ocultar-community.zip community/)
(cd "$DIST_DIR" && tar -czf ocultar-enterprise.tar.gz enterprise/)

echo -e "${GREEN}🚀 Release artifacts generated and verified in $DIST_DIR/${NC}"
ls -lh "$DIST_DIR"/*.zip "$DIST_DIR"/*.tar.gz
