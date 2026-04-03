#!/usr/bin/env bash
set -e

echo "================================================="
echo " OCULTAR COMMUNITY (SMB Edition) - Setup"
echo "================================================="

# 1. Check Docker
if ! command -v docker &> /dev/null; then
    echo "[!] Docker is not installed. Please install Docker Desktop from https://www.docker.com/products/docker-desktop/"
    exit 1
fi

# 2. Check/Setup .env
if [ ! -f .env ]; then
    if [ ! -f .env.example ]; then
        echo "[!] .env.example not found in the current directory."
        exit 1
    fi
    echo "[i] Initialising secure environment (.env)..."
    cp .env.example .env
    
    # Generate cryptographic keys
    MASTER_KEY=$(openssl rand -hex 32)
    SALT=$(openssl rand -hex 16)
    
    # macOS and Linux compatible sed replacement
    if sed --version 2>/dev/null | grep -q GNU; then
        sed -i "s/replace-with-a-secure-32-byte-key/${MASTER_KEY}/g" .env
        sed -i "s/ocultar-v112-kdf-salt-fixed-16/${SALT}/g" .env
    else
        sed -i '' "s/replace-with-a-secure-32-byte-key/${MASTER_KEY}/g" .env
        sed -i '' "s/ocultar-v112-kdf-salt-fixed-16/${SALT}/g" .env
    fi
    echo "[✓] Secure keys generated and saved to .env"
else
    echo "[✓] Existing .env file found."
fi

# 3. Start services
echo "[i] Starting OCULTAR stack..."
docker compose up -d

# 4. Open dashboard
echo "[✓] OCULTAR Engine is running!"
echo "[i] Launching dashboard..."

if command -v xdg-open &> /dev/null; then
    xdg-open http://localhost:3000
elif command -v open &> /dev/null; then
    open http://localhost:3000
else
    echo "================================================="
    echo " -> Please open http://localhost:3000 in your browser."
    echo "================================================="
fi
