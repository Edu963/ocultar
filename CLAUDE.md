# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Does

OCULTAR is a fully open-source, zero-egress PII detection and redaction proxy. It sits between a client and an upstream API (e.g., OpenAI, Gemini), intercepts requests, tokenizes all PII in-place using deterministic SHA-256 tokens (e.g., `[EMAIL_9c8f7a1b]`), and stores encrypted ciphertext in a local Vault. Responses can optionally rehydrate tokens back to plaintext for authorized callers. No raw PII ever reaches the upstream.

All components — including the Sombra agentic gateway — are open source. There is no enterprise/community split.

## Prerequisites

- **Go 1.25+** (workspace uses `go 1.25.8`; tested on `go1.26.3`)
- **CGO enabled** — DuckDB and libphonenumber require a C compiler (`gcc`/`clang`)
- **Node 18+** for the frontend apps

```bash
# Verify CGO is available
gcc --version
CGO_ENABLED=1 go build ./...
```

## Commands

```bash
# Build all Go modules (requires CGO)
make build

# Run all tests
make test

# Full workflow: sync workspace → provision model → build → test
# WARNING: `make all` provisions the SLM model — do not run in CI
make all

# Start the proxy (default port 8081)
go run ./apps/proxy

# Start the refinery HTTP server (port 8080)
go run ./services/refinery/cmd/main.go --serve 8080

# Start the SLM sidecar (local AI NER, port 8085)
go run ./apps/slm-engine/main.go

# Start the Sombra Gateway (port 8086)
go run ./apps/sombra

# Frontend (apps/dashboard or apps/web)
npm run dev
npm run build
```

Running a specific Go test:
```bash
cd services/refinery && CGO_ENABLED=1 go test ./... -run TestName
```

## Testing Conventions

All refinery detection logic must have table-driven tests with `input`/`wantRedacted` pairs. Use in-memory DuckDB (pass `""` as vault path) so tests are self-contained and don't leave `.db` files.

```go
func TestMyRule(t *testing.T) {
    v, _ := vault.New(config.Settings{VaultBackend: "duckdb"}, "")
    t.Cleanup(func() { v.Close() })
    config.InitDefaults()
    eng := refinery.NewRefinery(v, []byte("01234567890123456789012345678901"))

    cases := []struct {
        input string
        want  string // expected token type prefix, e.g. "[EMAIL_"
    }{
        {"send to alice@example.com", "[EMAIL_"},
    }
    for _, c := range cases {
        result, _ := eng.RefineString(c.input, "", nil)
        if !strings.Contains(result, c.want) {
            t.Errorf("input %q: want token %q in output, got %q", c.input, c.want, result)
        }
    }
}
```

**TestMain pattern** — any test package that calls `config.InitDefaults()` needs a `testmain_test.go` that `os.Chdir`s to the repo root so `configs/` relative paths resolve. See `services/refinery/pkg/proxy/testmain_test.go` as the reference.

**Entity Registry tests** — use `vault.Provider.RegisterEntity` to pre-seed canonical entities, then assert that all name variants collapse to a single `[TYPE_N]` token. See `services/refinery/pkg/refinery/refinery_entity_test.go`.

## Commit Style

Follow conventional commits with a scope matching the affected module:

```
feat(refinery): add FR_IBAN regex to Tier 1 rule engine
fix(vault): handle concurrent RegisterEntity race on DuckDB
docs: update ARCHITECTURE.md for v2.5
```

Common scopes: `proxy`, `refinery`, `vault`, `sombra`, `slm-engine`, `pii`, `gateway`, `docs`.

## Required Environment Variables

OCULTAR primarily uses **Doppler** for secret management. If running manually, copy `.env.example` to `.env`:

| Variable | Purpose |
|---|---|
| `OCU_MASTER_KEY` | 32-byte AES key for HKDF key derivation |
| `OCU_SALT` | Per-deployment salt |
| `OCU_PROXY_TARGET` | Upstream API base URL |
| `OCU_PROXY_PORT` | Proxy listen port (default `8081`) |
| `SLM_SIDECAR_URL` | SLM sidecar endpoint (default `http://localhost:8085`) |
| `SLM_ADAPTER` | Sidecar protocol: `privacy-filter` (default) or `openai-chat` (llama.cpp/Qwen) |
| `OCU_JWT_SECRET` | HS256 secret for Sombra JWT Bearer validation (generate: `openssl rand -hex 32`). If unset, Sombra is in insecure dev mode. |
| `OCU_AUDIT_PRIVATE_KEY` | Hex-encoded 32-byte Ed25519 seed for immutable audit log (generate: `openssl rand -hex 32`) |
| `OCU_AUDIT_LOG_PATH` | Audit log file path (default: `audit.log` alongside vault file) |
| `SLACK_SIGNING_SECRET` | Slack app signing secret for HMAC-SHA256 verification of incoming Slack events. Required when using the Slack connector; if unset, Sombra rejects all Slack webhook requests with HTTP 500. |

**Deprecated:** `TIER2_ENGINE` is a legacy alias for `SLM_ADAPTER` — the server logs `[DEPRECATED]` on startup if it is set. Use `SLM_ADAPTER` instead.

**Port note:** `apps/slm-engine` defaults to `:8085`; the refinery's `SLM_SIDECAR_URL` defaults to `http://localhost:8085`. Both services connect out of the box with no extra configuration.

**SLM sidecar (`apps/slm-engine`) — additional vars:**

| Variable | Purpose |
|---|---|
| `PYTHON_SIDECAR_URL` | URL of the Python privacy-filter process (default `http://localhost:8086`) |
| `PRIVACY_FILTER_MODEL_PATH` | HuggingFace model ID or local path (default `openai/privacy-filter`) |

## Architecture

### Go Workspace

The repo uses a Go workspace (`go.work`) linking 7 core modules:
- `apps/proxy` — reverse proxy entrypoint
- `apps/slm-engine` — local Small Language Model sidecar
- `apps/sombra` — agentic LLM gateway
- `services/refinery` — core PII detection and redaction engine
- `services/vault` — encrypted token storage (DuckDB or PostgreSQL)
- `internal/pii` — shared PII type registry and detection interfaces
- `pkg/gateway` — shared gateway client logic

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
| 1.5 | Contextual Shield | Interrogative name detection (e.g., "Where does [NAME] live?") and greeting/signature logic |
| 2 | AI NER | Sends text to SLM sidecar for deep named-entity recognition. Optimized for French Finance. |
| 3 | Structural Heuristics | Context-aware detection for structured document types |

### Vault and Tokenization

For each PII match:
1. `token_id = SHA-256(plaintext_PII)[:8 hex chars]` — deterministic, same input → same token
2. `ciphertext = AES-256-GCM(plaintext_PII, HKDF(masterKey, salt))`
3. Store `token_id → [TYPE_token_id] + ciphertext` in Vault
4. Replace PII in payload with `[TYPE_token_id]`

Vault backends: **DuckDB** (default, zero-config) or **PostgreSQL** (HA deployments). Configured in `configs/config.yaml`.

### Key Configuration Files

| File | Role |
|---|---|
| `configs/config.yaml` | Detection thresholds, vault backend, enabled tiers |
| `configs/protected_entities.json` | Named entities (VIPs, orgs) for Tier 0 dictionary matching |
| `configs/names.json` | Zero-Config local name dictionary |
| `configs/automation_commands.json` | Command definitions for the automation bridge |

### Entity Registry (Path 3)

The vault exposes a **Persistent Entity Registry** that maps name variants to a single canonical token across files, prompts, and sessions:

- `vault.Provider.RegisterEntity(entityType, canonicalName, variants)` → returns `[PERSON_1]`
- All fragments ("John", "Doe", "J. Doe") registered as variants produce the **same** token
- Tables: `canonical_entities` (id, type, name) + `entity_variants` (entity_id, variant)
- Rehydration: `[PERSON_1]` → `"John Doe"` via `DecryptToken` (entity tokens bypass AES; canonical name is returned directly)

Pre-seed entities via `RegisterEntity` before processing documents where identity fragmentation is expected (e.g., French finance invoices with first/last name on separate lines).

### Frontend

`apps/dashboard` and `apps/web` are independent Vite + React 19 + Tailwind CSS 4 apps. They are served separately from the Go backend.

| App | Dev port | API target |
|---|---|---|
| `apps/web` | `8080` (Vite) | Refinery HTTP on `8080` (same port — production build served by Go) |
| `apps/dashboard` | `3030` (Vite) | Refinery HTTP on `8090` (proxied via Vite `/api` prefix) |

```bash
# apps/web
cd apps/web && npm run dev        # → http://localhost:8080

# apps/dashboard
cd apps/dashboard && npm run dev  # → http://localhost:3030
# requires Refinery running on :8090
```

### What NOT to Do

- **Never commit `.env`** — use Doppler or `.env.example` only; `.env` is gitignored
- **Never run `make all` in CI** — it provisions the SLM model and has side effects
- **Never disable fail-closed** — if Vault or the Refinery returns an error, the request must be blocked, not passed through
- **Never add a new PII type without a test** — every regex in `configs/config.yaml` must have a corresponding table-driven test case in `services/refinery`
- **Never use `CGO_ENABLED=0`** — DuckDB requires CGO; the build will silently produce a binary that panics at runtime

---

## Development Personas (gstack Methodology)
When working on OCULTAR, adopt the following specialized roles as needed:

*   **CEO / Founder**: Focuses on "The Switzerland of Data" positioning. Use `/plan-ceo-review` for strategic shifts.
*   **Chief Security Officer (CSO)**: Mandates **Fail-Closed** logic. Use `/cso` for STRIDE audits on Go detection tiers.
*   **Staff Engineer**: Owns the Go Workspace (`go.work`) integrity. Use `/plan-eng-review` for architecture locks.
*   **QA Lead**: Use `/qa` to verify that PII tokenization (SHA-256) is deterministic and that responses are correctly rehydrated.
*   **Developer Experience (DX) Lead**: Optimizes the "5-minute deployment" path and the `Makefile` workflow. Use `/devex-review` for feedback.

## Sovereign Development Workflow
1.  **Office Hours**: Run `/office-hours` to reframe the task. Why are we building this?
2.  **Architecture Lock**: Use `/plan-eng-review` to map the flow across the 7 Go modules.
3.  **Security Audit**: Use `/cso` before coding to identify if this tier introduces false negatives or leaks.
4.  **Test-Driven Refinement**: Run `/review` to ensure every detection rule has corresponding tests.
5.  **Documentation Sync**: Use `/document-release` to ensure `ROADMAP.md` is updated.


## gstack (recommended)

This project uses [gstack](https://github.com/garrytan/gstack) for AI-assisted workflows.
Install it for the best experience:

```bash
git clone --depth 1 https://github.com/garrytan/gstack.git ~/.claude/skills/gstack
cd ~/.claude/skills/gstack && ./setup --team
```

Skills like /qa, /ship, /review, /investigate, and /browse become available after install.
Use /browse for all web browsing. Use ~/.claude/skills/gstack/... for gstack file paths.
