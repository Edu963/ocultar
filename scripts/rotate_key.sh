#!/bin/bash
# OCULTAR Master Key Rotation Utility
# This script generates a new OCU_MASTER_KEY and OCU_SALT and updates the .env file.

set -e

ENV_FILE=".env"
BACKUP_FILE=".env.bak.$(date +%Y%m%d%H%M%S)"

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: $ENV_FILE not found."
    exit 1
fi

echo "Creating backup of .env to $BACKUP_FILE..."
cp "$ENV_FILE" "$BACKUP_FILE"

NEW_KEY=$(openssl rand -hex 32)
NEW_SALT=$(openssl rand -hex 16)

echo "Generating new key and salt..."

# Update OCU_MASTER_KEY
if grep -q "OCU_MASTER_KEY=" "$ENV_FILE"; then
    sed -i "s/^OCU_MASTER_KEY=.*/OCU_MASTER_KEY=$NEW_KEY/" "$ENV_FILE"
else
    echo "OCU_MASTER_KEY=$NEW_KEY" >> "$ENV_FILE"
fi

# Update OCU_SALT
if grep -q "OCU_SALT=" "$ENV_FILE"; then
    sed -i "s/^OCU_SALT=.*/OCU_SALT=$NEW_SALT/" "$ENV_FILE"
else
    echo "OCU_SALT=$NEW_SALT" >> "$ENV_FILE"
fi

echo "Rotation complete."
echo "New Master Key: $NEW_KEY"
echo "New Salt: $NEW_SALT"
echo ""
echo "WARNING: Existing vault data in vault.db is now unrecoverable."
echo "Please restart all OCULTAR services to apply the changes."
