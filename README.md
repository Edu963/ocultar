# 🛡️ OCULTAR Monorepo

> [!TIP]
> **Featured Article:** [OpenAI shipped a model. We built the system.](https://dev.to/oculter_dev/openai-shipped-a-model-we-built-the-system-33pl)

[![GitHub Release](https://img.shields.io/github/v/release/Edu963/ocultar?display_name=v1.0.0)](https://github.com/Edu963/ocultar/releases/latest)
[![Security Policy](https://img.shields.io/badge/security-policy-brightgreen)](SECURITY.md)
[![Changelog](https://img.shields.io/badge/changelog-v1.0.0-blue)](CHANGELOG.md)
[![PyPI](https://img.shields.io/pypi/v/ocultar-goose-mcp)](https://pypi.org/project/ocultar-goose-mcp/)

Welcome to the **Unified OCULTAR Engine**. This monorepo contains the core refinery, integrated applications, and enterprise security tiers.


## Structure

- `/apps/` - Applications (Proxy, SLM Engine, Dashboard, Automation Bridge, Web)
- `/services/` - Core backend logic (Refinery, Vault)
- `/enterprise/` - Enterprise security extensions & licensing logic
- `/internal/pii/` - Centralized PII detection engine & registry
- `/extensions/` - Third-party AI tool integrations (Goose MCP, etc.)
- `/docs/` - Technical and product documentation
- `/security/` - Regulatory policies and integrity manifests

### Key Hardening Features (v1.1.0+)
- **Fail-Closed Proxy**: Strict semaphore-based queueing ensures no data bypasses the refinery under load.
- **SSRF Protection**: Hardened IP/DNS validation with rebinding protection for all egress targets.
- **Secrets Enforcement**: Mandatory `OCU_MASTER_KEY` and `OCU_SALT` validation using HKDF-SHA256 derivation.
- **Enterprise Observability**: Prometheus metrics for real-time tracking of latency, queue depth, and PII hit rates.
- **Zero-Bypass AI**: Tier 2 AI deep scan (Local Sidecar) is strictly enforced for contextual PII detection.

## Key Features

- **Tier 0 — Dictionary Shield**: Integrated high-speed dictionary protection for VIPs and corporate secrets with automated CRM synchronization.
- **Multi-Tier Refinery**: 5-tier detection pipeline (Base64 Shield → Dictionary → Regex Registry → AI NER → Structural Heuristics).
- **Native SLM Inference**: Sidecar support for local, sovereign PII scanning using specialized LLMs without cloud egress.
- **Zero-Egress Architecture**: PII is tokenized *before* ever leaving the trust boundary, converting liabilities into audit-safe assets.

## Why OCULTAR Is Different

### Obfuscation-Resistant: Recursive Base64 Scanning
Most PII filters operate on plaintext. A sophisticated attacker — or a misbehaving upstream SDK — can embed sensitive data inside a Base64-encoded blob inside a JSON field, completely bypassing naive pattern matching. OCULTAR decodes and recursively scans every Base64 segment it encounters, running the full detection pipeline on the decoded content before re-encoding the output. The payload structure is preserved; the PII is not. This is implemented in `Tier 0.1` of the refinery pipeline (`RefineString → processInterfaceRecursive`) and applies to both inline strings and nested JSON payloads hidden inside Base64.

### Luhn-Validated Credit Card Detection
Regex alone catches digit sequences that *look* like card numbers, producing a high false-positive rate that poisons analytics and erodes trust in any filtering system. OCULTAR applies the **Luhn algorithm (mod-10 checksum)** to every credit card candidate before vaulting it — the same validation used by payment processors. A match that fails Luhn is silently passed through without redaction or vault storage. Finance and payments teams will recognize this as the difference between a real card filter and a noise generator.

### Deterministic Tokens Enable Privacy-Safe Cross-Document Analytics
OCULTAR tokens are not random UUIDs. Every token is derived deterministically from `SHA-256(original_PII)`, meaning the same input always produces the same token — across requests, across sessions, and across documents. This has a consequence that has not been documented anywhere until now: **you can run aggregations, joins, and frequency analysis on fully tokenized data without ever de-tokenizing it.** A database of `[EMAIL_9c8f7a1b]` values can be counted, grouped, and correlated exactly like the original emails — the analytical value is preserved, the privacy risk is not. Re-hydration to plaintext is only needed when a human must read the result.

## Integration Boundary

Ocultar's responsibility ends at `POST /refine`. It returns `cleanText` 
and a vault token map. It has no knowledge of role cards, judgment logs, 
or downstream AI decisions. Nous is the only authorized caller of this 
endpoint in the Nous/Ocultar stack. If Ocultar is unavailable, callers 
must fail loudly — never degrade gracefully by passing raw data through.

## Getting Started

1.  **Environment Setup**:
    ```bash
    cp .env.example .env
    # Define OCU_MASTER_KEY and OCU_SALT
    ```

2.  **Go Workspace**:
    This project uses Go workspaces. Ensure you have Go 1.22+ installed.
    ```bash
    go work sync
    ```

3.  **Build and Run**:
    Use the root `Makefile` to build all components:
    ```bash
    make build
    go run ./apps/proxy  # Start the Privacy Proxy
    ```

## AI-Driven Governance

OCULTAR is maintained by a specialized ecosystem of AI Agent Skills. Every change to this repository is verified by the **Ocultar Protocol**, ensuring deterministic security and compliance.

- **Orchestration**: Managed by `/agents/skills/`.
- **Policy**: Defined in `/services/refinery/pkg/config/data/regulatory_policy.json`.
- **Audit**: Verifiable logs signed with Ed25519.

## Development

- **Documentation**: See `/docs/reference` for architecture and PII detection details.
- **Testing**: Run `go test ./...` from the root to verify all modules.

## Discovery & Community

- **Topics**: `privacy`, `gdpr`, `pii`, `golang`, `ai-security`, `zero-trust`, `llm`, `data-privacy`
- **Article**: [Read our launch story on dev.to](https://dev.to/oculter_dev/openai-shipped-a-model-we-built-the-system-33pl)
- **License**: Apache 2.0 (Open-Core)

