#!/bin/bash
# ── OCULTAR Automated Release Tool ──────────────────────────────────
# This script automates the compilation and packaging of OCULTAR
# components for client distribution.
# ───────────────────────────────────────────────────────────────────

set -e

# Configuration
VERSION=$(date +%Y.%m.%d)
DIST_DIR="./dist/release_$VERSION"
BUILD_TARGETS=("apps/proxy" "services/refinery" "apps/sombra" "apps/slm-engine")

echo "🚀 Starting OCULTAR Release v$VERSION..."

# 1. Workspace Sync
echo "📡 Syncing Go workspace..."
go work sync

# 2. Prepare Dist Directory
echo "📂 Preparing distribution directory: $DIST_DIR"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 3. Build Components
for TARGET in "${BUILD_TARGETS[@]}"; do
    NAME=$(basename "$TARGET")
    echo "🏗️ Building $NAME..."
    CGO_ENABLED=1 go build -o "$DIST_DIR/$NAME" "./$TARGET"
done

# 4. Include Config Templates
echo "📄 Including configuration templates..."
mkdir -p "$DIST_DIR/configs"
cp configs/*.yaml "$DIST_DIR/configs/" 2>/dev/null || true
cp .env.example "$DIST_DIR/"

# 5. Create Archive
echo "📦 Creating final archive..."
tar -czf "ocultar-release-v$VERSION.tar.gz" -C "$DIST_DIR" .

echo "✅ Release complete: ocultar-release-v$VERSION.tar.gz"
echo "⚠️  Reminder: Do NOT commit this .tar.gz to Git."
