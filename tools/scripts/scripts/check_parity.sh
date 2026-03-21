#!/usr/bin/env bash
# tools/scripts/scripts/check_parity.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Distribution Parity Checker
# ─────────────────────────────────────────────────────────────────────────────
# This script ensures that:
# 1. Enterprise bundle contains all Community files (parity).
# 2. Community bundle DOES NOT contain Enterprise-specific logic/binaries.
# 3. No sensitive internal/test files are present in either bundle.

set -euo pipefail

DIST_DIR="${1:-./dist}"
COMMUNITY_DIR="$DIST_DIR/community"
ENTERPRISE_DIR="$DIST_DIR/enterprise"

echo "Comparing Community and Enterprise bundles in $DIST_DIR..."

if [[ ! -d "$COMMUNITY_DIR" ]] || [[ ! -d "$ENTERPRISE_DIR" ]]; then
    echo "Error: Staging directories not found. Run build_release.sh first."
    exit 1
fi

# 1. Verify Parity (Enterprise should have everything Community has)
echo "[*] Verifying Enterprise parity..."
diff -r --brief "$COMMUNITY_DIR" "$ENTERPRISE_DIR" | grep "Only in $COMMUNITY_DIR" && {
    echo "[-] FAILURE: Enterprise bundle is missing files present in Community!"
    exit 1
} || echo "[+] Parity check passed."

# 2. Verify Leakage (Community should NOT have Enterprise-only components)
echo "[*] Searching for Enterprise leakage in Community..."
EXTERNAL_MARKER="ocultar-enterprise"
if [[ -f "$COMMUNITY_DIR/$EXTERNAL_MARKER" ]]; then
    echo "[-] FAILURE: Enterprise binary found in Community bundle!"
    exit 1
fi

# 3. Verify Sanitization (No test/dev data)
echo "[*] Verifying sanitization..."
LEAK_FILES=(".env" "testdata" ".git")
for file in "${LEAK_FILES[@]}"; do
    if find "$DIST_DIR" -name "$file" | grep -q .; then
        echo "[-] FAILURE: Sensitive file or directory '$file' found in distribution!"
        exit 1
    fi
done

echo "[+] SUCCESS: Distribution parity and sanitization verified."
exit 0
