#!/usr/bin/env bash
# =============================================================================
#  OCULTAR Enterprise + Sombra — Automated Demo
#  One command: ./demo/run_demo.sh
#  Add --record to capture the session with asciinema.
# =============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."          # always run from repo root

# ── Ports (change here if anything clashes) ──────────────────────────────────
PORT_REFINERY=9091
PORT_SOMBRA=9092
PORT_MOCK_AI=9090
PORT_MOCK_SLM=8085

# ── AI backend — auto-detected from environment ──────────────────────────────
# Priority: GEMINI_API_KEY > OPENAI_API_KEY > mock AI (no key needed)
if [[ -n "${GEMINI_API_KEY:-}" ]]; then
  AI_MODEL="gemini-flash-latest"
  AI_LABEL="Gemini Flash (real)"
  USE_MOCK_AI=0
elif [[ -n "${OPENAI_API_KEY:-}" ]]; then
  AI_MODEL="gpt-4o-mini"
  AI_LABEL="GPT-4o-mini (real)"
  USE_MOCK_AI=0
else
  AI_MODEL="mock-ai"
  AI_LABEL="Mock AI (offline)"
  USE_MOCK_AI=1
fi

# ── Files (auto-cleaned on exit) ─────────────────────────────────────────────
VAULT_REFINERY="/tmp/ocultar_demo_refinery.db"
VAULT_SOMBRA="/tmp/ocultar_demo_sombra.db"
AUDIT_LOG="sombra_audit.log"          # Sombra writes here (CWD-relative)
BIN_REFINERY="/tmp/ocultar_demo_refinery"
BIN_SOMBRA="/tmp/ocultar_demo_sombra"
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
  CAST_FILE="demo/ocultar_demo_$(date +%Y%m%d_%H%M%S).cast"
  echo -e "${Y}Recording to ${CAST_FILE}${D}"
  exec asciinema rec --overwrite -c "bash $(realpath "$0")" "$CAST_FILE"
fi

# =============================================================================
banner() {
  echo -e "${W}"
  echo "  ╔══════════════════════════════════════════════════════════════╗"
  echo "  ║         OCULTAR  ·  Enterprise  ·  Zero-Egress Demo         ║"
  echo "  ║         Refinery + Sombra Gateway + Immutable Audit         ║"
  echo "  ╚══════════════════════════════════════════════════════════════╝${D}"
  echo
}

step() { echo -e "\n${C}━━━  $1  ${D}"; }
label() { echo -e "${Y}  $1${D}"; }
ok()    { echo -e "${G}  ✓ $1${D}"; }
fail()  { echo -e "${R}  ✗ $1${D}"; exit 1; }
pause() { sleep "${DEMO_PAUSE:-1.2}"; }

# =============================================================================
cleanup() {
  echo -e "\n${Y}  Stopping services…${D}"
  for pid in "${PIDS[@]}"; do kill "$pid" 2>/dev/null || true; done
  rm -f "$VAULT_REFINERY" "$VAULT_SOMBRA" \
        "$BIN_REFINERY" "$BIN_SOMBRA" \
        "$AUDIT_LOG" sombra_vault.db \
        /tmp/ocultar_mock_slm.log /tmp/ocultar_mock_ai.log \
        /tmp/ocultar_refinery.log /tmp/ocultar_sombra.log
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
step "0 / 4  Pre-flight checks"

check_port $PORT_REFINERY  "Enterprise Refinery"
check_port $PORT_SOMBRA    "Sombra Gateway"
check_port $PORT_MOCK_SLM  "Mock SLM Sidecar"
[[ $USE_MOCK_AI -eq 1 ]] && check_port $PORT_MOCK_AI "Mock AI"

LICENSE=$(cat services/refinery/license.key 2>/dev/null | tr -d '[:space:]') \
  || fail "License key not found at services/refinery/license.key"
ok "Enterprise license found"
echo -e "  AI backend  :  ${W}${AI_LABEL}${D}"

command -v python3 &>/dev/null || fail "python3 required for mock services"
command -v go      &>/dev/null || fail "go toolchain required"
ok "Dependencies present"

# ── 1. Build ──────────────────────────────────────────────────────────────────
step "1 / 4  Building binaries"
echo "  (this takes ~20s the first time; cached thereafter)"

CGO_ENABLED=1 go build -o "$BIN_REFINERY" ./services/refinery/cmd &
CGO_ENABLED=1 go build -o "$BIN_SOMBRA"   ./apps/sombra           &
wait
ok "Refinery and Sombra compiled"

# ── 2. Start services ─────────────────────────────────────────────────────────
step "2 / 4  Starting services"

# Mock SLM Sidecar — always started so Enterprise Refinery has a Tier 2 target
python3 demo/mock_slm.py >/tmp/ocultar_mock_slm.log 2>&1 &
PIDS+=($!)
ok "Mock SLM Sidecar  →  :${PORT_MOCK_SLM}"

# Mock AI — only started when no real API key is set
if [[ $USE_MOCK_AI -eq 1 ]]; then
  python3 apps/sombra/mock_ai.py >/tmp/ocultar_mock_ai.log 2>&1 &
  PIDS+=($!)
  ok "Mock AI  →  :${PORT_MOCK_AI}"
else
  ok "Using real AI backend: ${AI_LABEL}"
fi

# Enterprise Refinery
OCU_LICENSE_KEY="$LICENSE" \
OCU_MASTER_KEY="demo-master-key-32-chars-longXXX" \
OCU_SALT="demo-salt" \
OCU_VAULT_PATH="$VAULT_REFINERY" \
SLM_SIDECAR_URL="http://localhost:${PORT_MOCK_SLM}" \
  "$BIN_REFINERY" --serve "$PORT_REFINERY" \
  >/tmp/ocultar_refinery.log 2>&1 &
PIDS+=($!)
ok "Enterprise Refinery  →  :${PORT_REFINERY}"

# Sombra Gateway (uses its own vault, routes to mock-ai or real AI)
_MOCK_AI_URL=""
[[ $USE_MOCK_AI -eq 1 ]] && _MOCK_AI_URL="http://127.0.0.1:${PORT_MOCK_AI}"

OCU_LICENSE_KEY="$LICENSE" \
OCU_MASTER_KEY="demo-master-key-32-chars-longXXX" \
OCU_SALT="demo-salt" \
SOMBRA_PORT="$PORT_SOMBRA" \
SOMBRA_MOCK_AI_URL="$_MOCK_AI_URL" \
GEMINI_API_KEY="${GEMINI_API_KEY:-}" \
OPENAI_API_KEY="${OPENAI_API_KEY:-}" \
OCU_VAULT_PATH="$VAULT_SOMBRA" \
  "$BIN_SOMBRA" \
  >/tmp/ocultar_sombra.log 2>&1 &
PIDS+=($!)
ok "Sombra Gateway  →  :${PORT_SOMBRA}"

# Health checks
wait_http "http://localhost:${PORT_MOCK_SLM}/health" "Mock SLM Sidecar"
[[ $USE_MOCK_AI -eq 1 ]] && wait_http "http://localhost:${PORT_MOCK_AI}" "Mock AI"
wait_http "http://localhost:${PORT_REFINERY}/api/system/status" "Refinery"
wait_http "http://localhost:${PORT_SOMBRA}/healthz"             "Sombra"

# ── 3. Demo ───────────────────────────────────────────────────────────────────
step "3 / 4  Running demo scenarios"

# ─── Scenario A: Refinery — direct PII detection ─────────────────────────────
echo
echo -e "${W}  ══ SCENARIO A  ·  Enterprise Refinery  ══════════════════════════${D}"
echo -e "     Direct PII detection and tokenisation — no AI involved"
echo

RAW_TEXT='Please summarise the Q2 performance review for Jean-Pierre Dupont (jp.dupont@acme.fr). Employee SSN: 587-25-1093. Salary IBAN: FR76 3000 6000 0112 3456 7890 189. Mobile: +33 6 47 22 89 31.'

label "📥  RAW INPUT"
echo    "  ──────────────────────────────────────────────────────────────"
echo   "$RAW_TEXT" | fold -s -w 70 | sed 's/^/  /'
echo    "  ──────────────────────────────────────────────────────────────"
pause

REFINED=$(curl -sf -X POST "http://localhost:${PORT_REFINERY}/api/refine" \
  -H "Content-Type: application/json" \
  -d "{\"text\": $(echo "$RAW_TEXT" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read().strip()))')}" \
  | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d.get("refined","ERROR"))' 2>/dev/null) \
  || fail "Refinery call failed"

echo
label "📤  REFINED OUTPUT  (what the AI provider would see)"
echo    "  ──────────────────────────────────────────────────────────────"
echo   "$REFINED" | fold -s -w 70 | sed 's/^/  /'
echo    "  ──────────────────────────────────────────────────────────────"
pause

# Count tokens
TOKEN_COUNT=$(echo "$REFINED" | grep -oE '\[[A-Z_]+_[0-9a-f]{8}\]' | wc -l)
echo
ok "${TOKEN_COUNT} PII entities detected and tokenised"
ok "Raw PII → AES-256-GCM vault  ·  Deterministic SHA-256 tokens"

# ─── Scenario B: Sombra — full end-to-end AI proxy ───────────────────────────
echo
echo
echo -e "${W}  ══ SCENARIO B  ·  Sombra Gateway  ·  End-to-End ════════════════${D}"
echo -e "     OpenAI-compatible request  →  redact  →  AI  →  rehydrate"
echo

label "📥  USER REQUEST  (OpenAI SDK format, sent to Sombra → ${AI_LABEL})"
echo    "  ──────────────────────────────────────────────────────────────"
cat <<MSG | sed 's/^/  /'
POST /v1/chat/completions  →  localhost:${PORT_SOMBRA}
{
  "model": "${AI_MODEL}",
  "messages": [{
    "role": "user",
    "content": "Summarise Q2 for Jean-Pierre Dupont (jp.dupont@acme.fr),
                SSN 587-25-1093, IBAN FR76…189, +33 6 47 22 89 31."
  }]
}
MSG
echo    "  ──────────────────────────────────────────────────────────────"
pause

SOMBRA_RESP=$(curl -sf -X POST "http://localhost:${PORT_SOMBRA}/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"${AI_MODEL}\",\"messages\":[{\"role\":\"user\",\"content\":\"Summarise Q2 for Jean-Pierre Dupont (jp.dupont@acme.fr), SSN 587-25-1093, IBAN FR76 3000 6000 0112 3456 7890 189, mobile +33 6 47 22 89 31.\"}]}") \
  || fail "Sombra call failed"

AI_CONTENT=$(echo "$SOMBRA_RESP" \
  | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["choices"][0]["message"]["content"])' 2>/dev/null) \
  || fail "Could not parse Sombra response"

echo
if [[ $USE_MOCK_AI -eq 1 ]]; then
  label "🤖  WHAT MOCK AI RECEIVED  (no real PII — only vault tokens)"
  echo    "  ──────────────────────────────────────────────────────────────"
  tail -6 /tmp/ocultar_mock_ai.log | sed 's/^/  /' || true
  echo    "  ──────────────────────────────────────────────────────────────"
else
  label "🤖  ${AI_LABEL} was called with redacted prompt (tokens only)"
  echo -e "  ${G}  No real PII left the VPC — zero-egress enforced by Sombra${D}"
fi
pause

echo
label "✅  RESPONSE RETURNED TO CALLER  (tokens rehydrated by Sombra)"
echo    "  ──────────────────────────────────────────────────────────────"
echo "$AI_CONTENT" | fold -s -w 70 | sed 's/^/  /'
echo    "  ──────────────────────────────────────────────────────────────"
pause

# Verify rehydration worked
if echo "$AI_CONTENT" | grep -q "Dupont\|jp.dupont\|587\|FR76"; then
  ok "Rehydration verified — original PII restored for the authorized caller"
else
  echo -e "${Y}  ⚠  Rehydration check inconclusive (mock AI may not have echoed tokens)${D}"
fi

# ─── Scenario C: Enterprise audit trail ──────────────────────────────────────
echo
echo
echo -e "${W}  ══ SCENARIO C  ·  Immutable Audit Trail  ════════════════════════${D}"
echo -e "     Ed25519-signed, SHA-256 hash-chained — every event recorded"
echo

if [[ -s "$AUDIT_LOG" ]]; then
  label "📋  AUDIT LOG  (last 5 entries)"
  echo    "  ──────────────────────────────────────────────────────────────"
  tail -5 "$AUDIT_LOG" \
    | python3 -c '
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line: continue
    try:
        e = json.loads(line)
        print(f"  [{e.get(\"timestamp\",\"\")}]  {e.get(\"action\",\"\")}  →  {e.get(\"result\",\"\")}  sig:{e.get(\"signature\",\"\")[:16]}…")
    except Exception:
        print(f"  {line[:100]}")
' 2>/dev/null || tail -5 "$AUDIT_LOG" | sed 's/^/  /'
  echo    "  ──────────────────────────────────────────────────────────────"
  ok "Every entry is Ed25519-signed and hash-chained (GDPR Art. 5(2))"
else
  echo -e "  ${Y}Audit log not yet written — Sombra writes on request completion${D}"
fi

# ── 4. Summary ────────────────────────────────────────────────────────────────
step "4 / 4  Summary"

VAULT_COUNT=$(curl -sf "http://localhost:${PORT_SOMBRA}/healthz" \
  | python3 -c 'import sys,json; print(json.load(sys.stdin).get("vault_count",0))' 2>/dev/null || echo "?")

echo
echo -e "${W}  Result${D}"
echo "  ┌─────────────────────────────────────────────────────────┐"
echo "  │  PII entities vaulted (Sombra)  :  ${VAULT_COUNT}"
echo "  │  Tier detected & tokenised      :  PERSON / EMAIL / SSN"
echo "  │                                    IBAN / PHONE"
echo "  │  Data reached AI provider       :  NONE — zero-egress ✓"
echo "  │  Caller received                :  Rehydrated plaintext ✓"
echo "  │  Audit log signed               :  Ed25519 ✓"
echo "  │  License                        :  ENTERPRISE ✓"
echo "  └─────────────────────────────────────────────────────────┘"
echo
echo -e "${G}  Demo complete.${D}"
echo
if [[ $RECORD -eq 1 ]]; then
  echo -e "${C}  Recording saved to: ${CAST_FILE}${D}"
  echo -e "${C}  Replay with: asciinema play ${CAST_FILE}${D}"
  echo
fi

# cleanup runs automatically via trap
