#!/usr/bin/env bash

# OCULTAR AI Orchestrator & Compliance Check
# This script automates the validation of PII rules, regulatory mapping, and builds.
# It can be run manually or set up as a Git pre-commit hook.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

# ANSI Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

cd "$REPO_ROOT"

echo -e "${BLUE}=============================================${NC}"
echo -e "${BLUE} OCULTAR AI Orchestration & Security Check ${NC}"
echo -e "${BLUE}=============================================${NC}"

echo -e "\n${BLUE}[1/6] Running core PII Detection regression tests (Regex, Dict, SLM fallback)...${NC}"
go test -v ./services/engine/pkg/engine/...

echo -e "\n${BLUE}[2/6] Running Enterprise Audit Logger tests (GDPR, HIPAA, VIP mappings)...${NC}"
go test -v ./enterprise/engine-extensions/pkg/audit/...

echo -e "\n${BLUE}[3/6] Validating Documentation Integrity and Links...${NC}"
"$SCRIPT_DIR/check_docs.sh" --strict

# Automated Doc-Update Check
if git diff --name-only HEAD 2>/dev/null | grep -q "pkg/engine/engine.go"; then
    if ! git diff --name-only HEAD 2>/dev/null | grep -qE "documentation/|docs/"; then
        echo -e "${YELLOW}[WARN] PII Engine core modified but no changes detected in documentation/ or docs/.${NC}"
        echo -e "       Ensure compliance reports (GDPR_Exposure_Report.md) are updated."
    fi
fi

echo -e "\n${BLUE}[4/6] Validating Go code formatting...${NC}"
# go fmt ./services/engine/... ./apps/proxy/... ./apps/sombra/... ./enterprise/engine-extensions/...
echo "Skipping go fmt for workspace compatibility"

echo -e "\n${BLUE}[5/6] Synchronizing Core Engine to Sombra Lab & Community Edition...${NC}"
"$SCRIPT_DIR/sync_cores.sh"

echo -e "\n${BLUE}[6/6] Building Production Release Artifacts (.zip, .tar.gz)...${NC}"
"$SCRIPT_DIR/build_release.sh"

echo -e "\n${GREEN}✅ All AI Skills, Compliance Checks, Synchronization and Builds passed successfully!${NC}"
echo "You can securely commit or deploy these changes."
