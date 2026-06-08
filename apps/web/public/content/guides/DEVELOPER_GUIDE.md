# OCULTAR | Developer Guide

> **Audience:** Go contributors, security researchers, and integrators embedding OCULTAR as a library.

---

## Table of Contents

1. [Module Structure](#1-module-structure)
2. [Development Environment](#2-development-environment)
3. [Running Tests](#3-running-tests)
4. [Package Overview](#4-package-overview)
5. [Extending the Refinery](#5-extending-the-refinery)
   - [Adding a Detection Rule (no recompile)](#51-adding-a-detection-rule-no-recompile)
   - [Adding a New Detection Tier (code)](#52-adding-a-new-detection-tier-code)
   - [Adding a Vault Backend](#53-adding-a-vault-backend)
6. [Embedding OCULTAR as a Go Library](#6-embedding-ocultar-as-a-go-library)
7. [Coding Conventions](#7-coding-conventions)
8. [Agentic Governance & Orchestration](#8-agentic-governance--orchestration)
9. [SLM AI Relay (Ollama Proxy)](#9-slm-ai-relay-ollama-proxy)
10. [CI / PR Checklist](#10-ci--pr-checklist)

---

## 1. Module Structure

OCULTAR uses a **Go Workspace** (`go.work`) to manage multiple modules in a single repository:

```
ocultar/                          ← root — shared library (github.com/Edu963/ocultar)
├── go.mod
├── go.work                         ← workspace definition
│
├── pkg/                            ← all importable packages
│   ├── config/       config.go     ← settings, regex/dict rules, fail-closed startup
│   ├── refinery/       refinery.go     ← core redaction pipeline (RefineString, ProcessInterface)
│   │                 phone_parser.go
│   │                 address_parser.go
│   ├── vault/        vault.go      ← Provider interface + factory (New)
│   │                 duckdb_provider.go
│   │                 postgres_provider.go
│   ├── proxy/        proxy.go      ← HTTP reverse proxy (Handler)
│   │                 vault.go      ← re-hydration helpers
│   ├── reporter/                   ← HTML risk-report generation (Enterprise)
│   └── license/                   ← license validation
│
├── cmd/                            ← CLI entrypoints
│   ├── ocultar/    main.go       ← shared CLI bootstrap (not directly runnable)
│   ├── proxy/        main.go       ← proxy-mode entrypoint
│   └── riskreport/   main.go       ← standalone report generator
│
├── dist/                           ← distribution entrypoints (buildable mains)
│   ├── community/    main.go       ← Entry point for Community Tier environment
│   └── enterprise/   main.go       ← Entry point for Enterprise Tier environment
│
├── ocultar-enterprise/           ← Enterprise-only sub-module
│   └── pkg/audit/                  ← SIEM audit logger
│
├── configs/
│   ├── config.yaml                 ← Enterprise runtime config (regexes, dicts, vault)
│   └── protected_entities.json     ← Tier 0 Dictionary Shield terms (required at startup)
│
├── scripts/                        ← Setup, smoke-test, and sync scripts
│   ├── orchestrate.sh              ← Main DEV Orchestrator (PII tests + Sync + Release)
│   ├── sync_cores.sh               ← Core Sync (DEV → Sombra Lab / Community)
│   ├── check_docs.sh               ← Documentation Link & Quality Checker
│   └── ocu-pre-commit.sh           ← Lead Shield Git Hook
└── documentation/                  ← All user-facing docs (you are here)
```

### `go.work` contents

```go
go 1.22

use (
    .
    ./dist/community
    ./dist/enterprise
)
```

> The Sombra gateway (`github.com/Edu963/sombra`) is a separate sibling repository, added to the workspace locally with `go work use ../sombra`.

---

## 2. Development Environment

### Prerequisites

| Tool | Version | Notes |
|---|---|---|
| **Go** | 1.22+ | `go version` to verify |
| **GCC / CGO** | Any modern | Required — DuckDB uses CGO. `gcc --version` to verify. |
| **Docker + Compose** | Latest | Only needed to run the proxy or full stack. |
| **Python** | 3.9+ | Optional — only for the audit/analysis scripts in the root. |

### Built-in Regex Rules

| Type | Pattern | Description |
|---|---|---|
| `EMAIL` | `(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}` | Standard email addresses |
| `URL` | `(?i)https?://[^\s/$.?#].[^\s]*|www\.[^\s/$.?#].[^\s]*` | HTTP/S URLs and bare `www.` domains |
| `SSN` | `\b\d{3}-\d{2}-\d{4}\b` | US Social Security Numbers |
| `CREDENTIAL` | `(?i)\bpassword\s*[:=]\s*[^\s,]+` | In-line passwords |
| `SECRET` | `(?i)\b(?:secret|key|token|api_key|auth_token|access_token|client_secret|private_key|refresh_token)\b\s*[:=]\s*[^\s,]+` | In-line secrets/keys |

### Clone and Build

```bash
# Clone
git clone https://github.com/Edu963/ocultar.git
cd ocultar

# (Optional) add Sombra if you need gateway development
git clone https://github.com/Edu963/sombra.git ../sombra
go work use ../sombra

# Verify the workspace
go build ./...
```

### First-Run Requirements

OCULTAR's startup is **fail-closed** — it will immediately abort if the Dictionary Shield file is missing:

```bash
# configs/protected_entities.json must exist and be non-empty before running
# The file is a simple JSON array of strings:
cat configs/protected_entities.json
# Example: ["internal-term-1", "internal-term-2"]
```

For development you can use the placeholder array already in the repository.

### Environment Setup

```bash
# Minimum required variable
export OCU_MASTER_KEY="dev-only-key-do-not-use-in-production"

# For proxy development
export OCU_PROXY_TARGET="http://localhost:11434"  # e.g. a local Ollama instance
export OCU_PROXY_PORT="8080"

# For Enterprise development (optional)
export OCU_LICENSE_KEY="your-enterprise-key"
```

---

## 3. Running Tests

```bash
# Run all unit tests (requires CGO + a writable tmp dir for DuckDB)
go test ./...

# Run a specific package
go test ./pkg/refinery/...
go test ./pkg/proxy/...
go test ./pkg/vault/...

# Run with race detector (recommended for proxy/refinery concurrency tests)
go test -race ./...

# Run fail-closed proxy tests specifically
go test -v ./pkg/proxy/ -run TestFailClosed

# Run with verbose output and coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Key Test Files

| File | What it covers |
|---|---|
| `pkg/refinery/refinery_test.go` | Core redaction correctness (email, phone, IBAN, address, base64, nested JSON) |
| `pkg/refinery/phone_parser_test.go` | International phone number parsing edge cases |
| `pkg/refinery/address_parser_test.go` | European/LATAM address heuristics |
| `pkg/proxy/proxy_test.go` | End-to-end proxy redaction with a mock upstream |
| `pkg/proxy/fail_closed_test.go` | Ensures refinery errors return 4xx/5xx and never forward un-redacted data |
| `pkg/vault/vault_test.go` | StoreToken idempotency, GetToken lookup, CountAll |

---

## 4. Package Overview

```
pkg/config   ──► loaded once at startup ──► pkg/refinery reads config.Global
                                        ──► pkg/vault reads cfg.VaultBackend
pkg/vault    ──► Provider interface ──► duckdbProvider (default)
                                   ──► postgresProvider (Enterprise)
pkg/refinery   ──► RefineString / ProcessInterface
             ──► depends on: pkg/config, pkg/vault, pkg/license
             ──► optional: AuditLogger, AIScanner (injected post-construction)
pkg/proxy    ──► http.Handler wrapping pkg/refinery
             ──► depends on: pkg/refinery, pkg/vault
```

**Dependency rules (enforced by import graph):**
- `pkg/config` has **zero** internal dependencies (it only uses stdlib).
- `pkg/vault` depends on `pkg/config` and `pkg/license` — never on `pkg/refinery`.
- `pkg/refinery` depends on `pkg/config`, `pkg/vault`, `pkg/license` — never on `pkg/proxy`.
- `pkg/proxy` sits at the top of the stack; it may import `pkg/refinery` and `pkg/vault`.

---

## 5. Extending the Refinery

### 5.1 Adding a Detection Rule (no recompile)

The fastest path — no Go code required. Edit `configs/config.yaml`:

```yaml
# Custom regex rule
regexes:
  - type: PASSPORT_NUMBER
    pattern: '\b[A-Z]{1,2}[0-9]{6,9}\b'

# Custom dictionary rule (Tier 0 — exact match)
dictionaries:
  - type: SECRET_PROJECT
    terms:
      - "Project Nightshade"
      - "Operation Dusk"
```

Restart the enterprise binary. Changes take effect at next startup (no recompilation).

> **Community note:** `configs/config.yaml` is ignored in Community mode. Custom rules require an Enterprise license.

---

### 5.2 Adding a New Detection Tier (code)

To add a new Tier inside `RefineString` in `pkg/refinery/refinery.go`:

1. **Write a parser function** following the existing pattern:
   ```go
   // ParseAndReplaceMyPII returns match index pairs (start, end) for each detected entity.
   func ParseAndReplaceMyPII(input string) [][]int { ... }
   ```

2. **Call `parseAndReplaceWithErr`** in `RefineString` at the appropriate tier position:
   ```go
   // TIER X: My Custom Shield
   refined, err = parseAndReplaceWithErr(refined, ParseAndReplaceMyPII, func(match string) (string, error) {
       return e.getOrSetSecureToken(match, "MY_TYPE", actor)
   })
   if err != nil {
       return "", err
   }
   ```

3. **Add tests** in `pkg/refinery/refinery_test.go` following the pattern in `TestRefineString`.

> **Important:** Never return partial output when `err != nil`. The caller always discards output on error.

---

### 5.3 Adding a Vault Backend

1. Create `pkg/vault/my_provider.go` implementing the `Provider` interface:
   ```go
   type myProvider struct { /* ... */ }

   func (p *myProvider) StoreToken(hash, token, encryptedPII string) (bool, error) { ... }
   func (p *myProvider) GetToken(hash string) (string, bool)                        { ... }
   func (p *myProvider) CountAll() int64                                             { ... }
   func (p *myProvider) Close() error                                                { ... }
   ```

2. Add a new `case` to the `New()` factory in `pkg/vault/vault.go`:
   ```go
   case "mybackend":
       return newMyProvider(cfg.MyDSN)
   ```

3. Add a `MyDSN string \`yaml:"my_dsn"\`` field to `config.Settings`.

4. Add test coverage in `pkg/vault/vault_test.go`.

---

## 6. Embedding OCULTAR as a Go Library

You can import the refinery directly into your own Go service:

```go
import (
    "github.com/Edu963/ocultar/pkg/config"
    "github.com/Edu963/ocultar/pkg/refinery"
    "github.com/Edu963/ocultar/pkg/vault"
    "crypto/sha256"
)

func main() {
    // 1. Load configuration (fail-closed: will fatal if protected_entities.json is missing)
    config.Load()

    // 2. Derive the 32-byte AES key
    rawKey := []byte("my-secret-master-key")
    hash := sha256.Sum256(rawKey)
    masterKey := hash[:]

    // 3. Open the vault
    v, err := vault.New(config.Global, "vault.db")
    if err != nil {
        panic(err)
    }
    defer v.Close()

    // 4. Construct the refinery
    eng := refinery.NewRefinery(v, masterKey)

    // 5. Refine a string
    refined, err := eng.RefineString("Call me at john@example.com", "system", nil)
    if err != nil {
        panic(err)
    }
    // refined → "Call me at [EMAIL_9c8f7a1b]"
    fmt.Println(refined)
}
```

> Add `github.com/Edu963/ocultar` to your `go.mod`:
> ```bash
> go get github.com/Edu963/ocultar@latest
> ```

---

## 7. Coding Conventions

| Area | Convention |
|---|---|
| **Error handling** | Always return errors; never panic in `pkg/` packages (only `main` and config loading may `log.Fatal`). |
| **Fail-closed** | Refinery errors must **block** processing — never forward partially-processed data. |
| **Thread safety** | All shared state (e.g. `Hits` map) must be protected by a mutex. Use `atomic.Int64` for counters. |
| **No side effects in tests** | Tests must not rely on disk state. Use `:memory:` for vault and `config.InitDefaults()`. |
| **Token format** | `[TYPE_XXXXXXXX]` where `XXXXXXXX` is the first 8 hex characters of the SHA-256 of the original PII. Never change this format — it breaks existing vaults. |
| **Logging** | Use stdlib `log`. Prefix proxy messages `[PROXY]`, refinery messages `[REFINERY]`, config messages `[config]`. |
| **Imports** | `stdlib` → `external` → `internal` (the standard Go import grouping). |

---

## 8. Agentic Governance & Orchestration

Ocultar development is supported by a 16-step **Continuous AI Orchestrator**. This system ensures that all code changes follow security and compliance best practices.

### Key Skills for Developers
- **`refinery-rule-generator`**: Use this when adding new PII detection rules. It automates the Go regex and config generation.
- **`sombra-gateway-policy-enforcer`**: An architectural linter. If you add new routes to Sombra, this skill will verify they are "Fail-Closed".
- **`change-impact-visualizer`**: Run this before opening a PR to generate a compliance impact summary for the reviewers.
- **`red-team-evasion-scanner`**: Proactively stress-tests your changes against obfuscation bypasses (Base64, URL-encoding).

### How to trigger
Skills are triggered automatically by the agent during the task lifecycle. Developers can also run the full suite manually via the orchestration pipeline:

```bash
./scripts/orchestrate.sh
```

### Security Gates
The orchestrator executes the following functional gates located in `tools/scripts/`:
- **`run_secret_scanner.sh`**: Scans for hardcoded API keys, tokens, and credentials using high-entropy regex pathfinding.
- **`run_arch_linter.sh`**: Enforces strict package boundaries (e.g., preventing `gateway` from directly importing `internal/pii`).
- **`run_zero_egress_validator.sh`**: Validates configuration manifests to ensure `ALLOW_ALL` policies are never committed.

---

## 9. Tier 2 SLM Engine Selection

OCULTAR Enterprise routes Tier 2 AI NER scans to the `slm-engine` sidecar over HTTP (`SLM_SIDECAR_URL`). The sidecar is engine-agnostic — its internal backend is selected at startup via `SLM_ENGINE`.

### Architecture

```
Refinery → RemoteScanner (HTTP) → slm-engine sidecar → inference backend
```

The refinery (`services/refinery/pkg/inference/remote.go`) and the enterprise extension (`enterprise/refinery-extensions/pkg/ai/scanner.go`) are fully decoupled from the inference backend. Only the sidecar changes between engines.

### Supported Engines

| `SLM_ENGINE` | Backend | CGO required | Model env var |
|---|---|---|---|
| `llama` (default) | llama.cpp native CGO | Yes | `SLM_MODEL_PATH` |
| `privacy-filter` | `openai/privacy-filter` via Python service | No | `PYTHON_SIDECAR_URL` / `PRIVACY_FILTER_MODEL_PATH` |

### Engine: llama (default)

```bash
export SLM_ENGINE=llama
export SLM_MODEL_PATH=models/qwen-1.5b-q4_k_m.gguf
```

Build requirements: `llama.cpp` headers and `libllama.a` in the library path. Optimized for Qwen 1.5B Q4_K_M GGUF (~1.2 GB VRAM); any GGUF-format model is compatible. A 5-second inference timeout is enforced via `llama_set_abort_callback`.

### Engine: privacy-filter

`openai/privacy-filter` is a bidirectional token classifier (Apache 2.0, ~1.5B params). Because it is Python-native (HuggingFace Transformers), run it as a separate Python service using the provided Dockerfile or script, and point the sidecar at it:

```bash
# To run locally via script:
pip install -r apps/slm-engine/python/requirements.txt
export PRIVACY_FILTER_MODEL_PATH=openai/privacy-filter # Or a local fine-tuned model path
python scripts/serve_privacy_filter.py   # listens on :8086

# Start the Go sidecar pointing at it
export SLM_ENGINE=privacy-filter
export PYTHON_SIDECAR_URL=http://localhost:8086
go run ./apps/slm-engine
```

No CGO required for the Go sidecar in this mode. The Python service must expose the same HTTP contract:
- `POST /scan  {"text":"..."}  →  {"PERSON":["John"],"EMAIL":["j@x.com"]}`
- `GET  /health               →  {"status":"ok"}`

---

## 10. CI / PR Checklist

Before opening a PR:

```bash
# 1. All tests pass with race detector
go test -race ./...

# 2. No compilation errors across the workspace
go build ./...

# 3. Vet clean
go vet ./...

# 4. Doc links valid
bash scripts/check_docs.sh

# 5. Full Orchestration Scan (PII Tests + Cross-Version Sync + Release Build)
# This is the mandatory "Source of Truth" sync before any release.
bash scripts/orchestrate.sh

> [!IMPORTANT]
> **Git Tracking:** The `orchestrate.sh` script generates binary artifacts in `dist/`. These are specifically **excluded from Git tracking** in `.gitignore`. Do not attempt to force-add them, as it will cause an infinite "Modified" loop during commits.

# 5. Smoke test passes (requires Docker)
docker compose -f docker-compose.proxy.yml up -d
bash scripts/smoke_test.sh
docker compose -f docker-compose.proxy.yml down
```

**PR requirements:**
- New detection tiers must include at least 3 test cases (true positive, false positive, boundary).
- New vault backends must implement all 4 `Provider` methods with test coverage.
- Changes to `RefineString`'s tier order must include a comment explaining the security rationale.
- Secrets must never appear in test fixtures — use placeholder-only test strings.
- **Agentic Audit**: Must pass the complete 16-step orchestrator sequence.
- **Impact Summary**: PR description should include the output from `change-impact-visualizer`.
