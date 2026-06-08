# OCULTAR Testing Guide

This guide helps you test OCULTAR in all modes: CLI/Dashboard, and the HTTP Proxy. No coding experience required.

---

## Part 1: Dashboard Test

The refinery runs the full detection pipeline (Tiers 0–2, including local AI NER when the SLM sidecar is active).

### Step 1 — Start the refinery

```bash
# If running from the distribution zip:
bash scripts/setup.sh

# If running from source:
Navigate to: **http://localhost:3000**

### Step 3 — Paste test data

In the **Raw Input (Liability)** box, paste:

```text
From: sarah.connor@cyberdyne.systems
To: john.smith@resistance.org

Hello John,

Please contact me at +1 (555) 019-8372 or my FR line: 06 12 34 56 78.
My current IP is 192.168.1.45.
Please wire funds to: IBAN DE89370400440532013000.
Our office is at 14 Rue de Rivoli, 75001 Paris.

Regards,
Sarah Connor
```

### Step 4 — Run and verify

Click **Execute Redaction**. The **Clean Asset** panel should show all of the above replaced with tokens like `[EMAIL_a1b2c3]`, `[PHONE_9e8f7a]`, `[IBAN_...]`, `[ADDRESS_...]`, `[PERSON_...]`.

---

## Part 2: Compliance Audit Dashboard

Enable SLM deep-scan (contextual NER) and the SIEM audit log by setting `SLM_SIDECAR_URL` and `OCU_AUDIT_PRIVATE_KEY`.

### Step 1 — Start the refinery

```bash
# From source:
go run ./services/refinery/cmd/main.go --serve 9091
```

Or use a different port to run alongside an existing instance.

### Step 2 — Open the Compliance Audit view

Navigate to: **http://localhost:9091/index.html**

### Step 3 — Paste risk data

```json
[
  {
    "patient_name": "Maria Garcia Rodriguez",
    "contact_email": "m.garcia@gmail.com",
    "home_address": "123 Calle Principal, Madrid",
    "medical_diagnosis": "Mild hypertension, new prescription required.",
    "passport_id": "ES123456789"
  },
  {
    "employee_name": "James Rutherford",
    "work_phone": "+44 20 7946 0958",
    "corporate_card": "4532 0156 8923 1144",
    "notes": "James visited the London office on Tuesday."
  }
]
```

### Step 4 — Execute and verify

Click **Execute Policy Audit**. You should see:
- **Extraction Breakdown**: entity counts by type (`HEALTH`, `CREDIT_CARD`, `PERSON`, `EMAIL`, etc.)
- **Global Regulatory Risk Matrix**: green checkmarks per framework (GDPR, HIPAA, AI Act…)
- **"Payload Successfully Anonymized"** banner when all PII is caught

### Step 5 — Check the audit log

```bash
tail -f audit.log
```

You should see structured JSON entries like:
```json
{"timestamp":"2026-03-01T18:00:00Z","actor":"127.0.0.1","action":"vaulted","token":"[EMAIL_3a9f2b01]"}
```

---

## Part 3: Proxy Mode Test

The proxy sits transparently in front of any upstream API (e.g. OpenAI).

### Step 1 — Start the proxy cluster

```bash
docker compose -f docker-compose.proxy.yml up -d
```

### Step 2 — Run the automated smoke test

```bash
./scripts/smoke_test.sh
```

Expected output:
```
[+] Proxy is healthy!
[*] Running smoke test with leaky payload...
[+] SUCCESS: PII successfully intercepted and redacted!
```

### Step 3 — Manual test with curl

```bash
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Email me at leaky@example.com"}]
  }'
```

Inspect the upstream target's received payload — `leaky@example.com` should appear as `[EMAIL_XXXXXXXX]`.

---

## Part 4: CLI Dry-Run (Quick Sanity Check)

```bash
# Scan a file without writing to the vault:
./ocultar --dry-run < raw_emails.json

# Scan and print a JSON PII report:
./ocultar --report < raw_emails.json
```

The `--report` flag outputs a JSON summary of all PII types detected, useful for CI/CD gates.
