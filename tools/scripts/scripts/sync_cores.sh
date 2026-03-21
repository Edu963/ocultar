#!/usr/bin/env bash
# scripts/sync_cores.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Cross-Version Core Synchronizer
#
# SOURCE OF TRUTH: /home/edu/dev/ocultar (DEV folder)
#
# Propagates the validated 'pkg/engine' and core dependencies from the
# Enterprise development folder to the Sombra Lab and Community distributions.
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
TARGET_LAB="/home/edu/ocultar-lab/ocultar"
TARGET_COMMUNITY="$REPO_ROOT/dist/community"

# ANSI Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE} OCULTAR Core Version Synchronizer         ${NC}"
echo -e "${BLUE}=============================================${NC}"

sync_target() {
    local target_path=$1
    local name=$2
    
    echo -e "\n${BLUE}Propagating core to $name...${NC}"
    
    if [[ ! -d "$target_path" ]]; then
        echo -e "${RED}[ERROR] Target path $target_path not found. Skipping.${NC}"
        return
    fi

    # Core Packages to sync
    local packages=("engine" "vault" "config")
    
    for pkg in "${packages[@]}"; do
        echo "  - Syncing pkg/$pkg..."
        mkdir -p "$target_path/pkg/$pkg"
        rsync -av --delete "$REPO_ROOT/services/engine/pkg/$pkg/" "$target_path/pkg/$pkg/"
    done

    # Sync essential config templates for boot/tests
    echo "  - Syncing configs/ (templates)..."
    mkdir -p "$target_path/configs"
    rsync -av "$REPO_ROOT/configs/protected_entities.json" "$target_path/configs/" 2>/dev/null || true

    echo -e "${GREEN}✅ Sync complete for $name.${NC}"
}

# 1. Sync to Sombra Lab (Optional/Local)
# sync_target "$TARGET_LAB" "Sombra Lab"

# 2. Sync to Community Distribution
sync_target "$TARGET_COMMUNITY" "Community Edition"

echo -e "\n${GREEN}🚀 Cross-version synchronization successful!${NC}"
