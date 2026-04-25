#!/usr/bin/env bash
# start.sh — Safe launcher for the SOMBRA gateway.
# Kills any stale process holding the vault lock or port 8084, then starts the gateway.

set -e
cd "$(dirname "$0")"

echo "[*] Clearing any stale Sombra processes..."
pkill -f "go-build.*sombra" 2>/dev/null || true
pkill -f "go-build.*main" 2>/dev/null || true
pkill -f "go run main.go" 2>/dev/null || true
pkill -f "exe/main" 2>/dev/null || true

# Kill any process holding port 8086
PORT_PID=$(lsof -ti :8086 2>/dev/null || true)
if [ -n "$PORT_PID" ]; then
  echo "[*] Killing process on port 8086 (PID: $PORT_PID)..."
  kill -9 $PORT_PID 2>/dev/null || true
fi

sleep 0.5

# Load enterprise .env (src/.env has GEMINI_API_KEY, OCU_LICENSE_KEY, etc.)
if [ -f src/.env ]; then
  echo "[*] Loading enterprise credentials from src/.env..."
  set -a
  source src/.env
  set +a
elif [ -f .env ]; then
  echo "[*] Loading .env..."
  set -a
  source .env
  set +a
fi

# Enable Enterprise Mode to activate Tier 2 AI
export OCU_FORCE_ENTERPRISE=true

echo "[*] Starting SOMBRA (Enterprise Mode)..."
exec go run main.go "$@"
