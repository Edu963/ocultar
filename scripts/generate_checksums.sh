#!/bin/bash
# Ocultar Asset Integrity Generator
# Generates SHA256 hashes for dashboard distribution files.

TARGET_DIR="dist/enterprise/dashboard"
MANIFEST_FILE="security/dashboard_integrity.json"

if [ ! -d "$TARGET_DIR" ]; then
    echo "[ERROR] Distribution directory $TARGET_DIR not found. Build the frontend first."
    exit 1
fi

echo "[INFO] Scanning $TARGET_DIR for integrity..."

# Create a JSON manifest
echo "{" > $MANIFEST_FILE
echo "  \"generated_at\": \"$(date -u +"%Y-%m-%dT%H:%M:%SZ")\"," >> $MANIFEST_FILE
echo "  \"checksums\": {" >> $MANIFEST_FILE

# Find all files, calculate SHA256, and format as JSON
find "$TARGET_DIR" -type f | sort | while read -r file; do
    hash=$(sha256sum "$file" | awk '{print $1}')
    rel_path=${file#$TARGET_DIR/}
    echo "    \"$rel_path\": \"$hash\"," >> $MANIFEST_FILE
done

# Remove last comma and close JSON
sed -i '$ s/,$//' $MANIFEST_FILE
echo "  }" >> $MANIFEST_FILE
echo "}" >> $MANIFEST_FILE

echo "[SUCCESS] Dashboard integrity manifest created at $MANIFEST_FILE"
