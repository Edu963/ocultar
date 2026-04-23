# Ocultar Monorepo (Open-Core)

Welcome to the **Unified OCULTAR Engine**. This monorepo contains the core refinery, integrated applications, and enterprise security tiers.

## Structure

- `/apps/` - Applications (Proxy, SLM Engine, Dashboard, Automation Bridge, Web)
- `/services/` - Core backend logic (Refinery, Vault)
- `/enterprise/` - Enterprise security extensions & licensing logic
- `/internal/pii/` - Centralized PII detection engine & registry
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
