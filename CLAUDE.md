# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Does

OCULTAR is a zero-egress PII detection and redaction proxy. It sits between a client and an upstream API (e.g., OpenAI, Gemini), intercepts requests, tokenizes all PII in-place using deterministic SHA-256 tokens (e.g., `[EMAIL_9c8f7a1b]`), and stores encrypted ciphertext in a local Vault. Responses can optionally rehydrate tokens back to plaintext for authorized callers. No raw PII ever reaches the upstream.

## Commands

```bash
# Build all Go modules (requires CGO)
make build

# Run all tests
make test

# Full workflow: sync workspace → provision model → build → test
make all

# Start the proxy (default port 8081)
go run ./apps/proxy

# Start the refinery HTTP server (port 8080)
go run ./services/refinery/cmd/main.go --serve 8080

# Start the SLM sidecar (local AI NER, port 8085)
go run ./apps/slm-engine/main.go

# Frontend (apps/dashboard or apps/web)
npm run dev
npm run build
```

Running a specific Go test:
```bash
cd services/refinery && CGO_ENABLED=1 go test ./... -run TestName
```

## Required Environment Variables

Copy `.env.example` to `.env` before running:

| Variable | Purpose |
|---|---|
| `OCU_MASTER_KEY` | 32-byte AES key for HKDF key derivation |
| `OCU_SALT` | Per-deployment salt |
| `OCU_PROXY_TARGET` | Upstream API base URL |
| `OCU_PROXY_PORT` | Proxy listen port (default `8081`) |
| `SLM_SIDECAR_URL` | SLM sidecar endpoint (default `http://localhost:8085`) |
| `OCU_LICENSE_KEY` | Base64-encoded Ed25519 license (enterprise only) |

## Architecture

### Go Workspace

The repo uses a Go workspace (`go.work`) linking 8 independent modules:
- `apps/proxy` — reverse proxy entrypoint
- `apps/slm-engine` — local Small Language Model sidecar
- `apps/sombra` — agentic LLM gateway (enterprise)
- `apps/automation_bridge` — REST automation command runner
- `services/refinery` — core PII detection and redaction engine
- `services/vault` — encrypted token storage (DuckDB or PostgreSQL)
- `internal/pii` — shared PII type registry and detection interfaces
- `enterprise/refinery-extensions` — licensed enterprise extensions

### Detection Pipeline (5 Tiers)

Requests flow through the Refinery pipeline in order:

| Tier | Name | What it does |
|---|---|---|
| 0.1 | Base64 Evasion Shield | Recursively decodes Base64/JWT/URL-encoded payloads and rescans decoded content |
| 0 | Dictionary Shield | Matches names from `configs/protected_entities.json` |
| 0.5 | Pattern + Entropy Shield | Regex patterns and Shannon entropy scoring |
| 1 | Rule Engine | Regex rules from `configs/config.yaml` (EMAIL, SSN, PHONE, CC, etc.) |
| 1.1 | Phone Shield | libphonenumber validation with Luhn-style checksum reduction |
| 1.2 | Address Shield | Heuristic address parser |
| 1.5 | Greeting/Signature Shield | Detects PII embedded in email salutations and signatures |
| 2 | AI NER (Enterprise) | Sends text to SLM sidecar for deep named-entity recognition |
| 3 | Structural Heuristics | Context-aware detection for structured document types |

### Vault and Tokenization

For each PII match:
1. `token_id = SHA-256(plaintext_PII)[:8 hex chars]` — deterministic, same input → same token
2. `ciphertext = AES-256-GCM(plaintext_PII, HKDF(masterKey, salt))`
3. Store `token_id → [TYPE_token_id] + ciphertext` in Vault
4. Replace PII in payload with `[TYPE_token_id]`

Vault backends: **DuckDB** (default, zero-config) or **PostgreSQL** (enterprise HA). Configured in `configs/config.yaml`.

### Key Configuration Files

| File | Role |
|---|---|
| `configs/config.yaml` | Detection thresholds, vault backend, enabled tiers |
| `configs/protected_entities.json` | Named entities (VIPs, orgs) for Tier 0 dictionary matching |
| `configs/automation_commands.json` | Command definitions for the automation bridge |

### Tiers and Licensing

- **Community**: Proxy + Refinery (Tiers 0–1.5) + DuckDB vault + CLI
- **Enterprise**: Adds Tier 2 AI NER, SIEM audit logger, CRM/SharePoint connectors, PostgreSQL backend, and `enterprise/refinery-extensions`

### Frontend

`apps/dashboard` and `apps/web` are independent Vite + React 19 + Tailwind CSS 4 apps. They are served separately from the Go backend and communicate via the Refinery HTTP API on port 8080.
