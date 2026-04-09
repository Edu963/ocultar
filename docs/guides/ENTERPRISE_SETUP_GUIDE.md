# OCULTAR Enterprise Tier — Complete Setup & Usage Guide

> **Audience:** Technical leads and operators deploying OCULTAR Enterprise for the first time, or graduating from Community to Enterprise.

---

## Table of Contents

1. [Community vs Enterprise: What's Different](#1-community-vs-enterprise-whats-different)
2. [Prerequisites](#2-prerequisites)
3. [Deployment Mode A: Standalone Binary + Dashboard](#3-deployment-mode-a-standalone-binary--dashboard)
4. [Deployment Mode B: Docker Compose Proxy Stack](#4-deployment-mode-b-docker-compose-proxy-stack)
5. [Enterprise Configuration (`configs/config.yaml`)](#5-enterprise-configuration-configsconfigyaml)
6. [Dictionary Shield (`configs/protected_entities.json`)](#6-dictionary-shield-configsprotected_entitiesjson)
7. [PostgreSQL HA Vault (Optional)](#7-postgresql-ha-vault-optional)
8. [SIEM Audit Log](#8-siem-audit-log)
9. [Enterprise Compliance Dashboard](#9-enterprise-compliance-dashboard)
10. [Key Rotation](#10-key-rotation)
11. [Shutting Down](#11-shutting-down)
12. [Troubleshooting](#12-troubleshooting)
13. [Activating Your Enterprise License](#13-activating-your-enterprise-license)

---

| Feature | Community (Pilot) | Enterprise (Enforcement) |
|---|---|---|

---

## 2. Prerequisites

| Requirement | Details |
|---|---|
| **OS** | Linux or macOS (Windows via WSL 2 or Docker Desktop) |
| **Docker** | Docker Refinery + Compose plugin, or Docker Desktop |
| **RAM** | 8 GB minimum (for Tier 2 local AI model) |
| **Disk** | ~2 GB free for the GGUF model on first run |
| **Go** | 1.22+ only if building from source |
| **`OCU_LICENSE_KEY`** | Obtain from the License Generator (`scripts/keygen.go`) |

---

## 3. Deployment Mode A: Standalone Binary + Dashboard

Use this mode to run the Enterprise refinery locally with the browser dashboard — ideal for demos, pilots, and batch processing.

### Step 1 — Prepare the environment

```bash
# Unpack the enterprise archive
tar -xzf ocultar-enterprise.tar.gz
cd ocultar-enterprise

# Create your secrets file
cp .env.example .env
```

Edit `.env`:
```bash
# ── Required ─────────────────────────────────────────────────────────────────
OCU_MASTER_KEY=<output of: openssl rand -hex 32>
OCU_SALT=<output of: openssl rand -hex 16>
OCU_LICENSE_KEY=<your enterprise license key>
```

> ⚠️ **Important:** `OCU_MASTER_KEY` and `OCU_SALT` derive your vault encryption key. **Changing either value after first run invalidates all vault entries.** Back them up securely before going to production.

### Step 2 — Start the Enterprise binary

```bash
source .env

# Start the dashboard on port 3030
./ocultar-enterprise --serve 3030
```

Or for batch file processing from the CLI:
```bash
# Refine a file and print the result
./ocultar-enterprise < my_data.json

# Dry-run scan (no vault writes, outputs JSON report)
./ocultar-enterprise --dry-run < my_data.json

# Report mode (refines + appends PII report to stderr)
./ocultar-enterprise --report < my_data.json
```

### Step 3 — Open the Enterprise Dashboard

Navigate to **http://localhost:3030/index_v2.html**

The Enterprise dashboard shows:
- **Extraction Breakdown** — entity counts by type
- **Global Regulatory Risk Matrix** — GDPR/HIPAA/AI Act/NIS2 compliance status per dataset
- **"Payload Successfully Anonymized"** banner when all PII is caught
- **Live Vault Metrics** — entry count, vault reuse rate, SLM health

---

## 4. Deployment Mode B: Docker Compose Proxy Stack

Use this mode to sit OCULTAR transparently in front of any upstream API (OpenAI, Azure OpenAI, local Ollama, etc.) — ideal for production integration.

### Step 1 — Prepare the environment

```bash
cp .env.example .env
```

Edit `.env` with all required values:

```bash
# ── Required ─────────────────────────────────────────────────────────────────
OCU_MASTER_KEY=<output of: openssl rand -hex 32>
OCU_SALT=<output of: openssl rand -hex 16>
OCU_PROXY_TARGET=https://api.openai.com   # your upstream LLM API
OCU_LICENSE_KEY=<your enterprise license key>

# ── Optional ─────────────────────────────────────────────────────────────────
OCU_PROXY_PORT=8081                        # host port the proxy listens on
```

### Step 2 — Launch the cluster

```bash
docker compose -f docker-compose.proxy.yml up -d
```

**What happens on first run:**
1. `init-slm` (alpine) downloads the `Qwen 1.5B Q4_K_M` GGUF model (~900 MB) from HuggingFace into the `slm_data` volume.
2. `ocultar-proxy` runs pre-flight validation (master key entropy, vault directory, SLM model path accessibility), then starts listening.
3. The refinery uses **Native CGO bindings** to perform inference directly within the proxy process.

Watch progress:
```bash
docker compose -f docker-compose.proxy.yml logs -f
```

Wait for:
```
ocultar-proxy  | [+] All pre-flight checks passed! Starting Proxy.
ocultar-proxy  | [INFO] SLM: Native llama.cpp initialized (model: /app/models/qwen-1.5b-q4_k_m.gguf)
ocultar-proxy  | [INFO] OCULTAR proxy listening on :8080
```

### Step 3 — Verify with the smoke test

```bash
bash scripts/smoke_test.sh
```

Expected:
```
[+] Proxy is healthy!
[*] Running smoke test with leaky payload...
[+] SUCCESS: PII successfully intercepted and redacted!
```

### Step 4 — Point your application at the proxy

Change your application's LLM base URL from:
```
https://api.openai.com
```
to:
```
http://localhost:8081   (or whatever OCU_PROXY_PORT you set)
```

No other change needed. The proxy preserves all headers, paths, and query strings.

**Optional — per-request upstream override:**
```bash
curl -X POST http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Ocultar-Target: https://other-llm.example.com" \
  -d '{"messages":[{"role":"user","content":"My name is Alice Martin."}]}'
```

---

## 5. Enterprise Configuration (`configs/config.yaml`)

Enterprise users can extend PII detection without recompiling. Edit `configs/config.yaml` and restart:

```yaml
# configs/config.yaml
# ─────────────────────────────────────────────────────────────────────────────
# This file is ONLY evaluated when a valid OCU_LICENSE_KEY is present.
# In Community mode, this file is silently ignored.
# ─────────────────────────────────────────────────────────────────────────────

# Minimum SLM confidence score for Tier 2 NER detections (0.0–1.0).
presidio_confidence: 0.75

# Custom regex rules (Tier 1) ─────────────────────────────────────────────────
# type: token label in UPPER_SNAKE_CASE
# pattern: Go regexp/syntax compatible regex (no look-aheads/look-behinds)
regexes:
  - type: PATIENT_ID
    pattern: '(?i)PTN-\d{6}'

  - type: INTERNAL_EMPLOYEE_ID
    pattern: '\b(ID|EMP)[-_][0-9]{4,6}\b'

  - type: TRANSACTION_CODE
    pattern: 'TXN-[A-Z0-9]{8}'

# Custom dictionary rules (Tier 0) — exact match, case-insensitive ────────────
dictionaries:
  - type: INTERNAL_PROJECT
    terms:
      - "Project Apollo"
      - "Operation Starlight"
```

After editing, apply by restarting:
```bash
# Proxy stack:
docker compose -f docker-compose.proxy.yml restart ocultar-proxy

# Standalone binary — just re-run it; config is loaded at startup.
```

**Testing a new rule:**
```bash
echo "Patient PTN-001234 was admitted." | ./ocultar-enterprise
# Expected: Patient [PATIENT_ID_3f1a9c2b] was admitted.
```

---

## 6. Dictionary Shield (`configs/protected_entities.json`)

`configs/protected_entities.json` is the **Tier 0 Dictionary Shield** — a mandatory fail-closed dependency. The refinery **will not start** if this file is missing or empty.

```json
["Alice Martin", "Project Phoenix", "Ouroboros Protocol"]
```

These terms are matched case-insensitively before any regex or AI scan, guaranteeing 100% recall for known sensitive entities.

> **Required even if empty-looking:** The file must contain at least one entry. An empty JSON array `[]` causes a fatal startup error by design.

When using Docker Compose, this file is baked into the Docker image at build time. If you need to update it, rebuild the image:
```bash
docker compose -f docker-compose.proxy.yml build ocultar-proxy
docker compose -f docker-compose.proxy.yml up -d ocultar-proxy
```

---

## 7. PostgreSQL HA Vault (Optional)

By default, OCULTAR uses an embedded DuckDB file (`vault.db`). For production multi-node deployments, switch to PostgreSQL:

### Step 1 — Provision PostgreSQL

Use any PostgreSQL 14+ instance: AWS RDS, Google CloudSQL, Azure Database, or a Docker container.

### Step 2 — Configure the vault backend

In `configs/config.yaml`:
```yaml
vault_backend: postgres
postgres_dsn: "host=db.corp.internal port=5432 user=ocultar password=<secret> dbname=ocultar_vault sslmode=require"
```

The `vault` table schema is created automatically on first startup:
```sql
-- Auto-created by OCULTAR on startup:
CREATE TABLE IF NOT EXISTS vault (
    hash          TEXT PRIMARY KEY,
    token         TEXT NOT NULL,
    encrypted_pii TEXT NOT NULL
);
```

### Step 3 — Deploy multiple proxy instances

With a shared PostgreSQL vault, you can run as many `ocultar-proxy` containers as needed behind a load balancer. Each instance can handle up to 15 concurrent requests (semaphore-limited to match the PostgreSQL connection pool).

---

## 8. SIEM Audit Log

When `OCU_LICENSE_KEY` activates the Enterprise tier, every vault event is written as a structured JSON line to `audit.log`.

**Format:**
```json
{"timestamp":"2026-03-06T14:00:00Z","actor":"192.168.1.1","action":"vaulted","token":"[EMAIL_9c8f7a1b]"}
{"timestamp":"2026-03-06T14:00:01Z","actor":"192.168.1.1","action":"matched","token":"[EMAIL_9c8f7a1b]"}
```

| Field | Description |
|---|---|
| `timestamp` | ISO-8601 UTC |
| `actor` | Client IP or `X-Forwarded-For` value |
| `action` | `"vaulted"` = new PII stored; `"matched"` = existing token returned from cache |
| `token` | The token string (never the original PII) |

**Location in Docker:**
```bash
docker exec ocultar-proxy tail -f /app/audit.log
```

**SIEM integration:**
- **Splunk**: Use the `monitor` stanza to ingest the log file
- **Elastic**: Filebeat with `json.message_key: token` for structured parsing
- **Datadog**: Fluent Bit sidecar with `Parser json`

Satisfies **GDPR Article 32(1)(d)** — logging of all processing events without exposing raw PII.

---

## 9. Enterprise Compliance Dashboard

The compliance dashboard is available in **Deployment Mode A** (standalone binary) at:

**http://localhost:3030/index_v2.html**

### Running an Audit

1. Open the dashboard
2. Paste or upload your data in the **Raw Input** panel
3. Click **Execute Policy Audit**

The dashboard renders:
- **Extraction Breakdown**: PII entity counts by type (EMAIL, PERSON, HEALTH, CREDIT_CARD, etc.)
- **Regulatory Risk Matrix**: Per-framework compliance status (GDPR, HIPAA, AI Act, BSI C5, NIS2, ISO 27001)
- **Redacted Output**: The clean, token-substituted version of your data
- **Vault Metrics**: Live count of unique PII entries, reuse rate, Deep Scan (SLM) health status

### CLI Audit for CI/CD Gates

```bash
# Generate a JSON compliance report (non-zero exit if PII detected):
./ocultar-enterprise --dry-run < dataset.json
echo "Exit code: $?"

# Or pipe to jq to extract the PII count:
./ocultar-enterprise --report < dataset.json 2>&1 | grep '"total_pii_count"'
```

---

## 10. Key Rotation

> ⚠️ Key rotation is a **destructive operation** if your vault already contains entries.

**If you must rotate `OCU_MASTER_KEY` or `OCU_SALT`:**

1. Export all vault tokens before rotation (re-hydrate while the old key is still active).
2. Stop all OCULTAR services.
3. Clear the vault: `rm vault.db` (DuckDB) or `TRUNCATE vault;` (PostgreSQL).
4. Update `OCU_MASTER_KEY` / `OCU_SALT` in `.env`.
5. Restart services — a fresh vault will be created.

---

## 11. Shutting Down

**Proxy stack:**
```bash
docker compose -f docker-compose.proxy.yml down        # stops containers, keeps volumes
docker compose -f docker-compose.proxy.yml down -v     # stops containers + deletes vault & model
```

**Standalone binary:** `Ctrl+C` — the vault file (`vault.db`) is preserved.

---

## 12. Troubleshooting

| Symptom | Likely Cause | Fix |
|---|---|---|
| `[FATAL] Failed reading protected_entities.json!` | File missing from `configs/` | Create the file with at least one entry (see §6) |
| `[FATAL] protected_entities.json … contains zero entries` | File is `[]` | Add at least one string to the JSON array |
| `[!] FATAL: OCU_MASTER_KEY must be set` | `.env` not loaded or key missing | Confirm `source .env` before running binary, or check Docker env vars |
| SLM container not healthy after 5 minutes | Model download incomplete or `slm_data` volume corrupt | `docker compose -f docker-compose.proxy.yml down -v && docker compose -f docker-compose.proxy.yml up -d` |
| `Active license tier … does not permit postgres` | `OCU_LICENSE_KEY` missing or expired | Add/renew your license key in `.env` |
| Proxy returns `429 Too Many Requests` | More than 15 concurrent requests | Scale horizontally with PostgreSQL vault (see §7) |
| Audit log is empty | Community binary or missing license key | Confirm `OCU_LICENSE_KEY` is set and `Tier: enterprise` is activated (check startup logs) |

---

## 13. Activating Your Enterprise License

Your enterprise license is a **signed token** delivered to you by the OCULTAR team. It works entirely offline — no license server, no internet call required at runtime.

### What it looks like

The token is a single string in this format:
```
<signature>.<encoded_payload>
```

Example (truncated):
```
zjIA64nAUHpY...BA==.eyJDdXN0b21lck5hbWUi...MH0=
```

### Where to put it

Add one line to your `.env` file:
```bash
OCU_LICENSE_KEY=<your_token_here>
```

Alternatively, you can place the token in a file named `license.key` in the same directory as the binary:
```bash
echo "<your_token_here>" > license.key
```

The refinery checks `OCU_LICENSE_KEY` first, then falls back to reading `license.key`.

### Confirming activation

Restart OCULTAR and look for this line in the startup log:
```
[INFO] License verified. Tier: ENTERPRISE, Customer: Your Company Name
```

If you see this instead:
```
[INFO] No license key found. Reverting to Community Mode.
[WARN] Invalid or expired license key. Reverting to Community Mode.
```
…the token is either missing, malformed, or expired — see below.

### Without a license key

OCULTAR will **not crash** — it silently falls back to Community Mode. This means:
- Custom rules in `configs/config.yaml` are ignored
- Local AI NER scan (Tier 2) is disabled
- The audit log (`audit.log`) will not be written
- PostgreSQL HA vault is not available

### Renewing an expiring license

When your license approaches expiry, the OCULTAR team will send you a new token. To apply it:
1. Replace the `OCU_LICENSE_KEY` value in your `.env` (or update `license.key`)
2. Restart OCULTAR — no rebuild required
3. Confirm the new expiry in the startup log

> ⚠️ **Do not share your license token.** It is issued to your organisation and tied to your customer name. Each deployment of OCULTAR in your organisation may use the same token.
