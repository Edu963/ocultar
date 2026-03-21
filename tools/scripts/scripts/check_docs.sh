#!/usr/bin/env bash
# scripts/check_docs.sh
# ─────────────────────────────────────────────────────────────────────────────
# OCULTAR Documentation Link & Quality Checker
#
# Validates all Markdown files under documentation/ and docs/ by:
#   1. Checking every internal relative link resolves to a real file
#   2. Ensuring no Mermaid code block is empty
#   3. Reporting a pass/fail summary
#
# Usage:
#   bash scripts/check_docs.sh          # from repo root
#   bash scripts/check_docs.sh --strict  # exit 1 on any warning (for CI)
#
# Dependencies: bash, grep, find — no external tools required
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STRICT=false
[[ "${1:-}" == "--strict" ]] && STRICT=true

ERRORS=0
WARNINGS=0

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
RESET='\033[0m'
BOLD='\033[1m'

error()   { echo -e "${RED}[ERROR]${RESET} $*"; ((ERRORS++)) || true; }
warn()    { echo -e "${YELLOW}[WARN]${RESET}  $*"; ((WARNINGS++)) || true; }
ok()      { echo -e "${GREEN}[OK]${RESET}    $*"; }
header()  { echo -e "\n${BOLD}$*${RESET}"; }

# ─── 1. Collect all Markdown files ───────────────────────────────────────────
header "Scanning Markdown files..."
DOC_FILES=()
while IFS= read -r -d '' f; do
    DOC_FILES+=("$f")
done < <(find "$REPO_ROOT/documentation" "$REPO_ROOT/docs" -name "*.md" -print0 2>/dev/null)

echo "  Found ${#DOC_FILES[@]} file(s)"

# ─── 2. Check internal relative links ────────────────────────────────────────
header "Checking internal links..."

for file in "${DOC_FILES[@]}"; do
    rel_file="${file#"$REPO_ROOT"/}"

    # Extract relative links: [text](./path) or [text](../path) or [text](path.md)
    # We skip http://, https://, mailto:, and anchors (#...)
    while IFS= read -r link; do
        # Resolve the link relative to the file's directory
        file_dir="$(dirname "$file")"
        # Strip any #anchor from the link
        link_path="${link%%#*}"
        target="$(realpath -m "$file_dir/$link_path" 2>/dev/null || true)"

        if [[ -z "$target" ]]; then
            warn "$rel_file → Could not resolve link: $link"
            continue
        fi

        if [[ ! -e "$target" ]]; then
            error "$rel_file → Broken link: $link"
            error "    Resolved to: ${target#"$REPO_ROOT"/}"
        else
            ok "$rel_file → $link"
        fi
    done < <(grep -oP '\[([^\]]+)\]\(\K[^)]+' "$file" \
              | grep -v '^https\?://' \
              | grep -v '^mailto:' \
              | grep -v '^#' \
              | grep -v '^\.' 2>/dev/null || true
    )

    # Also match links starting with ./ or ../
    while IFS= read -r link; do
        link_path="${link%%#*}"
        file_dir="$(dirname "$file")"
        target="$(realpath -m "$file_dir/$link_path" 2>/dev/null || true)"

        if [[ -z "$target" ]]; then
            warn "$rel_file → Could not resolve link: $link"
            continue
        fi

        if [[ ! -e "$target" ]]; then
            error "$rel_file → Broken link: $link"
            error "    Resolved to: ${target#"$REPO_ROOT"/}"
        else
            ok "$rel_file → $link"
        fi
    done < <(grep -oP '\[([^\]]+)\]\(\K[^)]+' "$file" \
              | grep '^[./]' \
              | grep -v '^https\?://' 2>/dev/null || true
    )
done

# ─── 3. Check Mermaid blocks are non-empty ────────────────────────────────────
header "Checking Mermaid diagram blocks..."

for file in "${DOC_FILES[@]}"; do
    rel_file="${file#"$REPO_ROOT"/}"
    # Find all ```mermaid blocks and ensure they contain at least one non-blank line
    awk '
      /^```mermaid/ { in_block=1; content=0; next }
      in_block && /^```/ {
        if (!content) print FILENAME": EMPTY mermaid block found"
        in_block=0; next
      }
      in_block && /[^[:space:]]/ { content=1 }
    ' "$file" | while IFS= read -r msg; do
        error "$msg"
    done
done

# ─── 4. Summary ───────────────────────────────────────────────────────────────
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [[ $ERRORS -gt 0 ]]; then
    echo -e "${RED}${BOLD}FAILED${RESET} — $ERRORS error(s), $WARNINGS warning(s)"
    exit 1
elif [[ $WARNINGS -gt 0 ]] && $STRICT; then
    echo -e "${YELLOW}${BOLD}FAILED (strict)${RESET} — 0 errors, $WARNINGS warning(s)"
    exit 1
else
    echo -e "${GREEN}${BOLD}PASSED${RESET} — 0 errors, $WARNINGS warning(s) across ${#DOC_FILES[@]} file(s)"
    exit 0
fi
