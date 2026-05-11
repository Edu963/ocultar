# 🛡️ OCULTAR Monorepo

Zero-egress PII refinery for AI pipelines.
Runs in your infrastructure. Your data never leaves.

> [!IMPORTANT]
> **Featured Article:** [OpenAI shipped a model. We built the system.](https://dev.to/oculter_dev/openai-shipped-a-model-we-built-the-system-33pl)
> 📖 OpenAI shipped a model. We built the system. [Read on dev.to](https://dev.to/oculter_dev/openai-shipped-a-model-we-built-the-system-33pl)

[![GitHub Release](https://img.shields.io/github/v/release/Edu963/ocultar?display_name=v1.0.0)](https://github.com/Edu963/ocultar/releases/latest)
[![Security Policy](https://img.shields.io/badge/security-policy-brightgreen)](SECURITY.md)
[![Changelog](https://img.shields.io/badge/changelog-v1.0.0-blue)](CHANGELOG.md)
[![PyPI goose](https://img.shields.io/pypi/v/ocultar-goose-mcp?label=goose-mcp)](https://pypi.org/project/ocultar-goose-mcp/)
[![PyPI claude](https://img.shields.io/pypi/v/ocultar-claude-mcp?label=claude-mcp)](https://pypi.org/project/ocultar-claude-mcp/)
[![PyPI mistral](https://img.shields.io/pypi/v/ocultar-mistral-mcp?label=mistral-mcp)](https://pypi.org/project/ocultar-mistral-mcp/)

### Quick Security Stats
| Stat | Value |
|------|-------|
| SSRF bypass vectors found + fixed | 2 |
| Fail-closed scenarios tested | 6 |
| Vault persistence | Named Docker volume |
| Tier 2 engine | OpenAI Privacy Filter (Apache 2.0) |
| Key management | Doppler |

Welcome to the **Unified OCULTAR Engine**. This monorepo contains the core refinery, integrated applications, and enterprise security tiers.

## Structure

- `/apps/` - Applications (Proxy, Sombra Gateway, SLM Engine, Dashboard, Automation Bridge, Web)
- `/services/` - Core backend logic (Refinery, Vault, Mock API)
- `/enterprise/` - Enterprise security extensions & licensing logic
- `/internal/pii/` - Centralized PII detection engine & registry
- `/extensions/` - Third-party AI tool integrations (Goose MCP, etc.)
- `/docs/` - Technical and product documentation
- `/security/` - Regulatory policies and integrity manifests

## Security Model

OCULTAR is built on a **Zero-Trust for Data** architecture. It is designed for senior security engineers who require verifiable guarantees before connecting internal data to external AI providers.

- **Zero-Egress**: A hard architectural guarantee. All PII detection and tokenization happen within your trust boundary. No network calls are made to third-party detection providers.
- **Fail-Closed**: 6 critical failure modes are rigorously tested (SLM timeout, vault failure, empty boot-guard, queue saturation, refinery internal error, re-hydration failure). In all cases, OCULTAR blocks the request rather than degrading to plaintext exposure.
- **SSRF Protection**: Hardened IP/DNS validation blocking RFC 1918 and `169.254.169.254` (IMDS) ranges with active DNS rebinding safety. 2 bypass vectors (including IPv6 loopback and non-standard decimal encoding) were identified and patched during adversarial testing.
- **Secure Vault**: AES-256-GCM encryption with keys derived via HKDF-SHA256. The vault is persisted via a named Docker volume to survive redeployments while keeping the master key in memory.
- **Ed25519 Audit Logs**: Tamper-proof, hash-chained audit trails signed with Ed25519. Every vault event (matching or vaulting) is logged for SIEM ingestion and compliance verification.

## Multi-Tier Refinery Pipeline

Tokenization is handled via a defense-in-depth pipeline that runs before any payload reaches an upstream AI provider.

| Tier | Shield | Technical Description |
|------|--------|-----------------------|
| 0.1 | Base64 Evasion | Decodes, scans, and re-encodes PII hidden inside Base64/JWT blobs. |
| 0 | Dictionary | High-speed protection for VIPs, internal projects, and sensitive org names. |
| 0.5 | Pattern + Entropy | Shannon scoring for high-entropy strings, catching keys and tokens. |
| 1 | Rule Engine | EMAIL, SSN, IBAN (MOD97), CC (Luhn mod-10), 50+ national ID types. |
| 1.1 | Phone Shield | libphonenumber validation to reduce false positives on digit sequences. |
| 1.2 | Address Shield | Heuristic street address parser supporting EN/FR/ES/DE. |
| 1.5 | Greeting/Signature | Detects names in salutations ("Regards, Jean") and intro sentences. |
| **2** | **AI NER** | **OpenAI Privacy Filter — 1.5B param, local inference. Optimized for French Finance.** |
| 3 | Structural Heuristics | Proximity expansion: `[TOKEN] ET Dupont` → re-tokenized as single entity. |

## Why OCULTAR Is Different

### Obfuscation-Resistant: Recursive Base64 Scanning
Most PII filters operate on plaintext. A sophisticated attacker can embed sensitive data inside a Base64-encoded blob inside a JSON field, bypassing naive pattern matching. OCULTAR decodes and recursively scans every Base64 segment, running the full pipeline on the decoded content.

### Luhn-Validated Credit Card Detection
OCULTAR applies the **Luhn algorithm (mod-10 checksum)** to every credit card candidate before vaulting it. A match that fails Luhn is passed through without redaction or vault storage, eliminating the noise typical of regex-only filters.

### Deterministic Tokens for Privacy-Safe Analytics
Tokens are derived from `SHA-256(original_PII)`. The same input always produces the same token. This allows you to run aggregations, joins, and frequency analysis on fully tokenized data without de-tokenizing it — preserving analytical value while eliminating privacy risk.

## Extensions

### Mistral Le Chat Extension
Zero-egress PII protection for Mistral Le Chat workflows.
```bash
pip install ocultar-mistral-mcp
```

### Claude Desktop Extension
Zero-egress PII protection for Claude AI workflows.
```bash
pip install ocultar-claude-mcp
```

### Goose AI Workflow Integration
Zero-egress PII protection for Goose AI workflows.
```bash
pip install ocultar-goose-mcp
```

### Ki! Desktop Distribution 🫏
The "Zero-Friction" desktop gateway for Ocultar. Built with Tauri and Rust, it bundles the Refinery engine into a native desktop app for individual researchers and enterprise workstations.
[Explore the Ki! Repo](https://github.com/Edu963/kii)

Read the launch story: [OpenAI shipped a model. We built the system.](https://dev.to/oculter_dev/openai-shipped-a-model-we-built-the-system-33pl)


## Integration Boundary

Ocultar's responsibility ends at `POST /refine`. It returns `cleanText` and a vault token map. It has no knowledge of downstream AI decisions. Callers must fail loudly if Ocultar is unavailable — never degrade gracefully by passing raw data.

## Quick Start

Prerequisites: [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2.

```bash
git clone https://github.com/Edu963/ocultar && cd ocultar
docker compose up --build
```

That's it. The proxy is live at **http://localhost:8081**. First build takes ~5 minutes (compiling Go + CGO for DuckDB). Every subsequent start is instant.

Send a test request through the proxy:

```bash
curl -s http://localhost:8081/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"My email is alice@example.com and SSN is 123-45-6789"}]}'
```

PII is tokenized before reaching the upstream — check the echo response to confirm `[EMAIL_...]` and `[SSN_...]` tokens in place of the originals.

**Add AI-powered Tier 2 NER** (downloads ~500 MB model on first run):

```bash
OCU_PILOT_MODE=0 docker compose --profile ai up --build
```

**Use real API keys / custom secrets:**

```bash
cp .env.example .env   # edit with your keys
docker compose up --build
```

**Prometheus metrics**: http://localhost:9090/metrics

## Development

- **Documentation**: See `/docs/reference` for architecture details.
- **Testing**: Run `go test ./...` to verify all modules (requires CGO: `cd services/refinery && CGO_ENABLED=1 go test ./...`).

## Discovery & Community

- **Topics**: `privacy`, `gdpr`, `pii`, `golang`, `ai-security`, `zero-trust`, `llm`, `data-privacy`
- **License**: Apache 2.0 (Open-Core)
