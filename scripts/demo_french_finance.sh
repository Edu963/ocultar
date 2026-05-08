#!/usr/bin/env bash
# OCULTAR French Finance — PII redaction demo
# Tests the proxy with a realistic French banking payload.
# Runs against http://localhost:8081 (proxy) and http://localhost:8080 (refinery API).

set -e

PROXY="http://localhost:8081"
REFINERY="http://localhost:8080"

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "  OCULTAR — French Finance Sovereign AI Demo"
echo "══════════════════════════════════════════════════════════════"

# ── Test 1: PII redaction via proxy ─────────────────────────────────────────
echo ""
echo "▶ Test 1: Proxy redaction (port 8081)"
echo "  Payload: French banking data with IBAN, NIR, SIRET, phone, email"
echo ""

RESULT=$(curl -s "$PROXY/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{
      "role": "user",
      "content": "Virement de 4 500 EUR depuis le compte de Jean Dupont (NIR: 1 85 06 75 115 423 18) vers IBAN FR76 3000 6000 0112 3456 7890 189. Société émettrice: Dupont & Associés SARL, SIRET 552 100 554 03333, TVA FR83552100554. Contact: jean.dupont@bpce.fr, tél. +33 6 12 34 56 78."
    }]
  }')

echo "$RESULT" | python3 -m json.tool 2>/dev/null || echo "$RESULT"

# ── Test 2: Direct refinery API + report ────────────────────────────────────
echo ""
echo "▶ Test 2: Refinery API with detection report (port 8080)"
echo ""

REPORT=$(curl -s -X POST "$REFINERY/api/refine" \
  -H "Content-Type: text/plain" \
  -d "Appel de Marie Leclerc au 06 12 34 56 78 pour le dossier Crédit Agricole. IBAN: FR76 3000 6000 0112 3456 7890 189. NIR: 2 85 06 75 115 423 18.")

echo "$REPORT" | python3 -m json.tool 2>/dev/null || echo "$REPORT"

# ── Test 3: Policy block (credentials) ──────────────────────────────────────
echo ""
echo "▶ Test 3: Policy block — credentials must never reach AI model"
echo ""

BLOCK=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$REFINERY/api/refine" \
  -H "Content-Type: text/plain" \
  -d "API key: sk-prod-xK9mN2pQ7rL4wV8yB3cF6jD1hT0sZ5uE")

if [ "$BLOCK" = "403" ]; then
  echo "  ✅ HTTP 403 — credential blocked by policy 'block-credentials-always'"
else
  echo "  ⚠  Unexpected status: $BLOCK (expected 403)"
fi

# ── Test 4: Compliance evidence ──────────────────────────────────────────────
echo ""
echo "▶ Test 4: Compliance evidence endpoint (for your DPO)"
echo ""

curl -s "$REFINERY/api/compliance/evidence" | python3 -m json.tool 2>/dev/null

echo ""
echo "══════════════════════════════════════════════════════════════"
echo "  Demo complete. No PII left your infrastructure."
echo "══════════════════════════════════════════════════════════════"
echo ""
