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

# 1. Clean and Prepare Dist Directory
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR/community"
mkdir -p "$DIST_DIR/enterprise"

# 2. Build Binaries and Web Dashboard
echo -e "${BLUE}Compiling binaries...${NC}"
(cd "$REPO_ROOT" && go build -o "$DIST_DIR/community/ocultar" ./services/engine/cmd/main.go)
(cd "$REPO_ROOT" && go build -o "$DIST_DIR/enterprise/ocultar-enterprise" ./enterprise/engine-extensions/cmd/ocultar-enterprise/main.go) || echo "Enterprise main.go might be elsewhere"

echo -e "${BLUE}Building Web Dashboard...${NC}"
(cd "$REPO_ROOT/apps/web" && npm install && npm run build)

# 3. Prepare Community Edition
echo -e "${BLUE}Preparing Community Edition...${NC}"
cp "$REPO_ROOT/docker-compose.proxy.yml" "$DIST_DIR/community/docker-compose.yml"
mkdir -p "$DIST_DIR/community/scripts"
cp "$SCRIPT_DIR/setup-community.sh" "$DIST_DIR/community/scripts/"
cp "$SCRIPT_DIR/setup-community.ps1" "$DIST_DIR/community/scripts/"
# Copy web dashboard if it exists
if [ -d "$REPO_ROOT/apps/web/dist" ]; then
    cp -r "$REPO_ROOT/apps/web/dist" "$DIST_DIR/community/dashboard"
fi

# 4. Prepare Enterprise Edition
echo -e "${BLUE}Preparing Enterprise Edition...${NC}"
cp "$REPO_ROOT/configs/config.yaml" "$DIST_DIR/enterprise/" 2>/dev/null || touch "$DIST_DIR/enterprise/config.yaml"
# (Add other enterprise assets here)

# 5. Package Artifacts
echo -e "${BLUE}Packaging artifacts...${NC}"
(cd "$DIST_DIR" && zip -r ocultar-community.zip community/)
(cd "$DIST_DIR" && tar -czf ocultar-enterprise.tar.gz enterprise/)

echo -e "${GREEN}🚀 Release artifacts generated in $DIST_DIR/${NC}"
ls -lh "$DIST_DIR"/*.zip "$DIST_DIR"/*.tar.gz
