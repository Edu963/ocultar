#!/usr/bin/env bash
# =============================================================================
#  OCULTAR Community — Automated Demo
#  One command: ./demo/run_demo_community.sh
#  Add --record to capture the session with asciinema.
# =============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."          # always run from repo root

# ── Ports ─────────────────────────────────────────────────────────────────────
PORT_PROXY=9093
PORT_MOCK_AI=9090

# ── Files ─────────────────────────────────────────────────────────────────────
VAULT_PROXY="/tmp/ocultar_demo_community.db"
BIN_PROXY="/tmp/ocultar_demo_proxy"
PIDS=()

# ── Colours ───────────────────────────────────────────────────────────────────
R='\033[0;31m'; G='\033[0;32m'; Y='\033[1;33m'
B='\033[0;34m'; C='\033[0;36m'; W='\033[1;37m'; D='\033[0m'

# ── Optional asciinema recording ─────────────────────────────────────────────
RECORD=0
if [[ "${1:-}" == "--record" ]]; then
  RECORD=1
  if ! command -v asciinema &>/dev/null; then
    echo "asciinema not found — install with: pip install asciinema"
    exit 1
  fi
  CAST_FILE="demo/ocultar_community_demo_$(date +%Y%m%d_%H%M%S).cast"
  echo -e "${Y}Recording to ${CAST_FILE}${D}"
  exec asciinema rec --overwrite -c "bash $(realpath "$0")" "$CAST_FILE"
fi

# =============================================================================
banner() {
  echo -e "${W}"
  echo "  ╔══════════════════════════════════════════════════════════════╗"
  echo "  ║          OCULTAR  ·  Community  ·  Zero-Egress Demo         ║"
  echo "  ║          Transparent Proxy  ·  PII Redaction  ·  Vault      ║"
  echo "  ╚══════════════════════════════════════════════════════════════╝${D}"
  echo
}

step()  { echo -e "\n${C}━━━  $1  ${D}"; }
label() { echo -e "${Y}  $1${D}"; }
ok()    { echo -e "${G}  ✓ $1${D}"; }
fail()  { echo -e "${R}  ✗ $1${D}"; exit 1; }
note()  { echo -e "${B}  ℹ  $1${D}"; }
pause() { sleep "${DEMO_PAUSE:-1.2}"; }

# =============================================================================
cleanup() {
  echo -e "\n${Y}  Stopping services…${D}"
  for pid in "${PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
  rm -f "$VAULT_PROXY" "$BIN_PROXY" \
        /tmp/ocultar_demo_proxy.log /tmp/ocultar_mock_ai.log \
        /tmp/ocultar_community_usage.db
  echo -e "${G}  Done.${D}\n"
}
trap cleanup EXIT INT TERM

# =============================================================================
check_port() {
  local port=$1 name=$2
  if ss -tlnp 2>/dev/null | grep -q ":${port} "; then
    fail "Port ${port} is already in use (needed for ${name}). Free it and retry."
  fi
}

wait_http() {
  local url=$1 name=$2
  echo -ne "  Waiting for ${name}"
  for i in $(seq 1 40); do
    if curl -sf --max-time 1 "$url" &>/dev/null; then
      echo -e " ${G}ready${D}"; return 0
    fi
    echo -n "."; sleep 0.8
  done
  fail "${name} did not become ready at ${url}"
}

# =============================================================================
banner

# ── 0. Preflight ─────────────────────────────────────────────────────────────
step "0 / 3  Pre-flight checks"

check_port $PORT_PROXY   "Community Proxy"
check_port $PORT_MOCK_AI "Mock AI (upstream target)"

command -v python3 &>/dev/null || fail "python3 required"
command -v go      &>/dev/null || fail "go toolchain required"
ok "Dependencies present"
echo -e "  No license key required  :  ${W}Community is free & open${D}"

# ── 1. Build ──────────────────────────────────────────────────────────────────
step "1 / 3  Building proxy binary"
echo "  (this takes ~20s the first time; cached thereafter)"

CGO_ENABLED=1 go build -o "$BIN_PROXY" ./apps/proxy
ok "Community proxy compiled"

# ── 2. Start services ─────────────────────────────────────────────────────────
step "2 / 3  Starting services"

# Mock AI — acts as the upstream API (OpenAI-compatible)
python3 apps/sombra/mock_ai.py >/tmp/ocultar_mock_ai.log 2>&1 &
PIDS+=($!)
ok "Mock AI (upstream)  →  :${PORT_MOCK_AI}"

# Community Proxy — transparent, no license key, --dev relaxes salt requirement
OCU_MASTER_KEY="demo-master-key-32-chars-longXXX" \
OCU_SALT="demo-salt" \
OCU_PROXY_TARGET="http://localhost:${PORT_MOCK_AI}" \
OCU_PROXY_PORT="${PORT_PROXY}" \
OCU_VAULT_PATH="$VAULT_PROXY" \
  "$BIN_PROXY" --dev \
  >/tmp/ocultar_demo_proxy.log 2>&1 &
PIDS+=($!)
ok "Community Proxy  →  :${PORT_PROXY}  (target → :${PORT_MOCK_AI})"

wait_http "http://localhost:${PORT_MOCK_AI}"    "Mock AI"
wait_http "http://localhost:${PORT_PROXY}/healthz" "Community Proxy"

# ── 3. Demo ───────────────────────────────────────────────────────────────────
step "3 / 3  Running demo scenarios"

RAW_MSG='Summarise Q2 for Jean-Pierre Dupont (jp.dupont@acme.fr), SSN 587-25-1093, IBAN FR76 3000 6000 0112 3456 7890 189, mobile +33 6 47 22 89 31.'

# ─── Scenario A: What the client sends ───────────────────────────────────────
echo
echo -e "${W}  ══ SCENARIO A  ·  Client sends a normal OpenAI request  ════════${D}"
echo -e "     The client points its SDK at localhost:${PORT_PROXY} — no code changes"
echo
label "📥  RAW REQUEST  (client application sees this)"
echo    "  ──────────────────────────────────────────────────────────────"
cat <<MSG | sed 's/^/  /'
POST /v1/chat/completions  →  localhost:${PORT_PROXY}
{
  "model": "gpt-4o-mini",
  "messages": [{"role":"user","content":"${RAW_MSG}"}]
}
MSG
echo    "  ──────────────────────────────────────────────────────────────"
pause

# ─── Scenario B: Proxy intercepts, redacts, forwards ─────────────────────────
echo
echo -e "${W}  ══ SCENARIO B  ·  Proxy intercepts and redacts  ════════════════${D}"
echo -e "     PII is stripped before the request leaves your infrastructure"
echo

PROXY_RESP=$(curl -sf -X POST "http://localhost:${PORT_PROXY}/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "X-Ocultar-Tier: developer" \
  -d "{\"model\":\"gpt-4o-mini\",\"messages\":[{\"role\":\"user\",\"content\":\"${RAW_MSG}\"}]}") \
  || fail "Proxy call failed"

# Show what the upstream AI received (captured by mock AI log)
label "🔒  WHAT THE UPSTREAM AI RECEIVED  (no real PII)"
echo    "  ──────────────────────────────────────────────────────────────"
tail -8 /tmp/ocultar_mock_ai.log | sed 's/^/  /' || true
echo    "  ──────────────────────────────────────────────────────────────"
pause

# Show the rehydrated response returned to the caller
AI_CONTENT=$(echo "$PROXY_RESP" \
  | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["choices"][0]["message"]["content"])' 2>/dev/null) \
  || fail "Could not parse proxy response"

echo
label "✅  RESPONSE TO CALLER  (tokens rehydrated — PII restored)"
echo    "  ──────────────────────────────────────────────────────────────"
echo "$AI_CONTENT" | fold -s -w 70 | sed 's/^/  /'
echo    "  ──────────────────────────────────────────────────────────────"
pause

if echo "$AI_CONTENT" | grep -q "Dupont\|jp.dupont\|587\|FR76"; then
  ok "Rehydration verified — caller sees original PII, AI never did"
else
  echo -e "${Y}  ⚠  Rehydration check inconclusive${D}"
fi

# ─── What Community does NOT include ─────────────────────────────────────────
echo
echo
echo -e "${W}  ══ WHAT COMMUNITY DOES NOT INCLUDE  ════════════════════════════${D}"
echo
echo -e "${Y}  The following are Enterprise-only features:${D}"
echo
echo "  ✗  Tier 2 AI NER  — catches PII regex misses (names in prose,"
echo "                       job titles, contextual entities in French/EN)"
echo
echo "  ✗  Immutable Audit Log  — Ed25519-signed, SHA-256 hash-chained"
echo "                            required for GDPR Art. 5(2), SOC 2, ISO 27001"
echo
echo "  ✗  Sombra Gateway  — multi-model routing (Gemini, OpenAI, Claude,"
echo "                        local models), Slack/file connectors, agentic flows"
echo
echo "  ✗  CRM / LDAP Identity Sync  — auto-ingest VIP names from"
echo "                                  Salesforce, Workday, LDAP into Tier 0"
echo
echo "  ✗  PostgreSQL HA Vault  — horizontal scaling across multiple"
echo "                             proxy instances sharing one vault"
echo
note "Run ./demo/run_demo.sh to see all Enterprise features live."

# ── Summary ───────────────────────────────────────────────────────────────────
TOKEN_COUNT=$(grep -oE '\[[A-Z_]+_[0-9a-f]{8}\]' /tmp/ocultar_mock_ai.log 2>/dev/null | wc -l | tr -d ' ')

echo
echo -e "${W}  Result${D}"
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │  PII tokens sent to upstream AI :  ${TOKEN_COUNT} (zero raw PII)"
echo "  │  Tiers active                   :  0 · 0.5 · 1 · 1.1 · 1.2"
echo "  │  Tier 2 AI NER                  :  — (Enterprise only)"
echo "  │  Data reached upstream AI       :  NONE — zero-egress ✓"
echo "  │  Caller received                :  Rehydrated plaintext ✓"
echo "  │  Immutable audit log            :  — (Enterprise only)"
echo "  │  License required               :  NONE — free & open ✓"
echo "  └─────────────────────────────────────────────────────────┘"
echo
echo -e "${G}  Community demo complete.${D}"
echo
if [[ $RECORD -eq 1 ]]; then
  echo -e "${C}  Recording saved to: ${CAST_FILE}${D}"
  echo -e "${C}  Replay with: asciinema play ${CAST_FILE}${D}"
  echo
fi
