# OCULTAR Sombra Gateway — Setup & Usage Guide

> **Audience:** Developers and operators who need an intelligent, multi-model AI routing layer on top of OCULTAR — going beyond the raw proxy to support document connectors, multi-provider routing, and structured query workflows.

---

## Table of Contents

1. [What Is Sombra — and When to Use It](#1-what-is-sombra--and-when-to-use-it)
2. [Prerequisites](#2-prerequisites)
3. [Installation](#3-installation)
4. [Configuration (`configs/sombra.yaml`)](#4-configuration-configssombrayaml)
5. [Connectors: File & API Sources](#5-connectors-file--api-sources)
6. [Multi-Model Routing](#6-multi-model-routing)
7. [The `/query` Endpoint](#7-the-query-endpoint)
8. [How Sombra Integrates with OCULTAR](#8-how-sombra-integrates-with-ocultar)
9. [Running Sombra](#9-running-sombra)
10. [Troubleshooting](#10-troubleshooting)

---

## 1. What Is Sombra — and When to Use It

**OCULTAR** protects data in transit — it redacts PII from payloads before they reach an LLM, and rehydrates tokens in the response.

**Sombra** (`github.com/Edu963/sombra`) is the agentic layer **above** OCULTAR. It adds:

| Capability | OCULTAR Proxy | Sombra Gateway |
|---|---|---|
| Transparent HTTP PII redaction | ✅ | ✅ (built-in) |
| Re-hydration of LLM responses | ✅ | ✅ (built-in) |
| Multi-LLM routing (OpenAI, Gemini, Claude, local) | ❌ | ✅ |
| File connector (ingest documents before querying) | ❌ | ✅ |
| API connector (pull structured data from HTTP sources) | ❌ | ✅ |
| Pro Connectors (Slack, SharePoint via pkg/connector) | ❌ | ✅ |
| Single `/query` endpoint orchestrating everything | ❌ | ✅ |

**Use the raw OCULTAR proxy when:**
- You only need transparent PII scrubbing for a single upstream API
- You want zero-config sidecar deployment alongside an existing LLM client

**Use Sombra when:**
- You want to route queries across multiple AI providers based on content or cost
- You need to ingest files or data sources before querying the LLM
- You want a single orchestrated endpoint that handles: ingest → redact → route → respond → rehydrate

---

## 2. Prerequisites

| Requirement | Details |
|---|---|
| **Go** | 1.22+ |
| **OCULTAR** | The core refinery repo cloned and buildable (Sombra imports `github.com/Edu963/ocultar/pkg/refinery`) |
| **`OCU_MASTER_KEY`** | Same key used by OCULTAR — Sombra shares the vault |
| **`OCU_JWT_SECRET`** | HS256 secret for Bearer token validation. Generate: `openssl rand -hex 32`. If unset, any Bearer value is accepted as actor identity (dev only — insecure). |
| **LLM API key(s)** | At least one: `OPENAI_API_KEY`, `GEMINI_API_KEY`, or a local Ollama endpoint |

---

## 3. Installation

### Step 1 — Clone Sombra alongside OCULTAR

```bash
# Sombra must be a sibling directory of ocultar/
git clone https://github.com/Edu963/sombra.git ../sombra
```

Your directory layout should be:
```
~/dev/
  ocultar/     ← this repo
  sombra/        ← Sombra gateway
```

### Step 2 — Add Sombra to the Go workspace

From the `ocultar/` directory:
```bash
go work use ../sombra
```

Verify the workspace recognises both modules:
```bash
go work sync
go build ./...
```

### Step 3 — Create the Sombra config

```bash
cp ../sombra/configs/sombra.example.yaml configs/sombra.yaml
```

(If no example exists, create `configs/sombra.yaml` from the template in §4.)

---

## 4. Configuration (`configs/sombra.yaml`)

Sombra is configured entirely through a single YAML file. A fully-annotated template:

```yaml
sombra:
  # ── Server ─────────────────────────────────────────────────────────────────
  listen_port: "8081"           # Port Sombra listens on (distinct from OCULTAR's 8080)

  # ── Vault ─────────────────────────────────────────────────────────────────
  # Sombra shares OCULTAR's vault. Point to the same file.
  vault_path: "vault_data/vault.db"

  # ── AI Models ──────────────────────────────────────────────────────────────
  # List all LLM providers Sombra may route to. At least one is required.
  models:
    - name: gpt-4o
      provider: openai
      endpoint: https://api.openai.com/v1/chat/completions
      api_key_env: OPENAI_API_KEY       # name of the env var holding the key

    - name: gemini-pro
      provider: gemini
      endpoint: https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent
      api_key_env: GEMINI_API_KEY

    - name: claude-3-sonnet
      provider: anthropic
      endpoint: https://api.anthropic.com/v1/messages
      api_key_env: ANTHROPIC_API_KEY

    - name: local-ollama
      provider: ollama
      endpoint: http://localhost:11434/api/chat
      api_key_env: ""                   # Ollama needs no key

  # ── Connectors ─────────────────────────────────────────────────────────────
  # Data sources Sombra can ingest before sending to LLMs.
  connectors:
    - name: local-files
      type: file
      policy:
        redact: true                    # Apply OCULTAR redaction before querying

    - name: internal-api
      type: api
      endpoint: https://internal.corp/api/data
      policy:
        redact: true

    - name: slack-prod
      type: slack
      config:
        workspace_id: "T12345"
        token: "xoxb-..."

  # ── Identity Sync (Enterprise) ─────────────────────────────────────────────
  # Automatically poll external CRM/LDAP providers for protected names/VIPs.
  crm_endpoint: "https://crm.corp.internal/api/v1/identities"
  crm_api_key: "ocultar-sync-secret"
  sync_interval: "1h"
```

### Field Reference

#### `models[]`

| Field | Required | Description |
|---|---|---|
| `name` | ✅ | Logical name used in routing rules and `/query` requests |
| `provider` | ✅ | `openai`, `gemini`, `anthropic`, or `ollama` |
| `endpoint` | ✅ | Full URL of the model's completion API |
| `api_key_env` | ✅ | Name of the environment variable holding the API key. Leave empty (`""`) for unauthenticated local endpoints |

#### `connectors[]`

| Field | Required | Description |
|---|---|---|
| `name` | ✅ | Logical name referenced in `/query` requests |
| `type` | ✅ | `file` (local files) or `api` (HTTP data source) |
| `endpoint` | API type only | URL of the data source |
| `policy.redact` | ✅ | `true` = run all data through OCULTAR before sending to LLM |

---

## 5. Connectors: File & API Sources

Connectors allow Sombra to ingest data before querying an LLM — without you having to pre-process it.

### File Connector

Use to analyse local documents (CSV, JSON, plain text):

```json
POST /query
{
  "connector": "local-files",
  "file_path": "/data/patient_records.csv",
  "model": "gpt-4o",
  "prompt": "Summarise the key health risk factors in this dataset."
}
```

Sombra will:
1. Read the file
2. Pass it through the OCULTAR refinery (all PII → tokens)
3. Send the sanitised content + your prompt to `gpt-4o`
4. Rehydrate tokens in the response before returning it

### API Connector

Use to pull real-time data from an internal endpoint before querying:

```json
POST /query
{
  "connector": "internal-api",
  "model": "gemini-pro",
  "prompt": "Identify anomalies in today's transaction data."
}
```

Sombra fetches from `connector.endpoint`, redacts the response, and forwards it to the LLM.

---

## 6. Multi-Model Routing

Sombra's router lets a single `/query` call target any model defined in `sombra.yaml`. Routing is **explicit** — you name the model in the request:

```json
POST /query
{
  "model": "claude-3-sonnet",
  "prompt": "Review this contract clause: ..."
}
```

To route to a local model:
```json
POST /query
{
  "model": "local-ollama",
  "prompt": "Translate this to French: ..."
}
```

**All routes share the same OCULTAR redaction pipeline** — regardless of which LLM is targeted, PII is stripped before leaving your infrastructure.

---

## 7. The `/query` Endpoint

The single orchestration endpoint:

```
POST http://localhost:8081/query
Content-Type: application/json
Authorization: Bearer <user-id-or-token>
```

### Request Schema

```json
{
  "model":      "gpt-4o",             // required: model name from sombra.yaml
  "prompt":     "Your question here", // required: user prompt
  "connector":  "local-files",        // optional: data source connector name
  "file_path":  "/path/to/file.csv"  // required if connector=file
}
```

### Response Schema

```json
{
  "response":   "The AI's answer (PII rehydrated)",
  "metadata": {
    "model": "gpt-4o",
    "connector": "local-files",
    "pii_was_redacted": true,
    "ai_saw": "..."
  }
}
```

| Field | Description |
|---|---|
| `response` | The LLM's answer with original PII values restored from the vault |
| `metadata.model` | The model that handled the request |
| `metadata.pii_was_redacted` | `true` if at least one PII entity was detected and tokenised |
| `metadata.ai_saw` | **Level 4 Security:** Sombra redacts *both* the ingested file data and the user's `prompt`. This field shows exactly what string the AI models received (useful for debugging, can be removed in production). |

---

## 8. How Sombra Integrates with OCULTAR

Sombra imports and wraps the OCULTAR refinery directly:

```
Client → POST /query
  → Sombra reads connector data (if any)
  → Sombra calls refinery.ProcessInterface() — same pipeline as OCULTAR
      [Tier 0 Dictionary → Tier 1 Regex → Tier 1.1 Phone → Tier 1.2 Address → Tier 2 AI]
  → Sombra forwards sanitised data + prompt to the selected LLM
  → Sombra receives LLM response
  → Sombra calls refinery.DecryptToken() for each token in the response
  → Client receives rehydrated response
```

### Vault Sharing

Sombra and OCULTAR share the **same vault file** (`vault_path` in `sombra.yaml` must point to the same `vault.db` as OCULTAR). This ensures:
- Tokens generated by OCULTAR can be rehydrated by Sombra and vice versa
- A single consolidated audit log covers all processing

### Key Derivation

Both OCULTAR components use **HKDF-SHA256** for key derivation:

| Component | Key derivation |
|---|---|
| OCULTAR Proxy | `HKDF-SHA256(OCU_MASTER_KEY, OCU_SALT, "ocultar-aes-key")` |
| Sombra gateway | `HKDF-SHA256(OCU_MASTER_KEY, OCU_SALT, "ocultar-aes-key")` |

> ✅ Because both components use the same derivation function **and the same salt**, they produce the same AES key. Any token vaulted by the proxy can be re-hydrated by Sombra and vice versa, as long as they point to the same vault file and share `OCU_MASTER_KEY` + `OCU_SALT`.

---

## 9. Running Sombra

### Development (from source)

```bash
# From the sombra/ directory
export OCU_MASTER_KEY="your-key"
export OCU_SALT="your-salt"
export OPENAI_API_KEY="sk-..."

go run . --config ../ocultar/configs/sombra.yaml
```

Expected startup output:
```
[INFO] Sombra Gateway v1.x.x
[INFO] Vault: vault_data/vault.db
[INFO] Models loaded: gpt-4o, gemini-pro, local-ollama
[INFO] Connectors: local-files, internal-api
[INFO] Listening on :8081
```

### Quick Test

```bash
curl -s -X POST http://localhost:8081/query \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "prompt": "Hello, my email is test@example.com. What'\''s 2+2?"
  }' | jq .
```

Expected: the response contains `4`, and `redacted: true` with `pii_count: 1`. The email never reached OpenAI.

---

## 10. Troubleshooting

| Symptom | Likely Cause | Fix |
|---|---|---|
| `module not found: github.com/Edu963/sombra` | Sombra not in Go workspace | Run `go work use ../sombra` from `ocultar/` |
| `vault: no such file or directory` | `vault_path` in `sombra.yaml` is wrong | Check relative path; use an absolute path for reliability |
| `401 Unauthorized` from LLM | API key env var not set | `export OPENAI_API_KEY=sk-...` before running Sombra |
| `model "X" not found` | Model name in request doesn't match `sombra.yaml` | Check the `name:` field in your YAML; it is case-sensitive |
| Response contains raw tokens `[EMAIL_…]` | Rehydration vault path mismatch | Ensure Sombra's `vault_path` and OCULTAR's `OCU_VAULT_PATH` point to the same file |
| HKDF key mismatch warning in logs | Different `OCU_SALT` between Sombra and OCULTAR | Use the same `OCU_SALT` env var for both services |
