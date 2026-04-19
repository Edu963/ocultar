# Ocultar Monorepo (Open-Core)

Welcome to the **Unified OCULTAR Engine**. This monorepo contains the core refinery, all 6 integrated applications, and enterprise security tiers.

## Structure

- `/apps/` - Core applications (Sombra, SLM Engine, Dashboard, Proxy, Automation Bridge, Web)
- `/services/` - Core backend logic (Refinery, Vault)
- `/enterprise/` - Enterprise security extensions & licensing logic
- `/experiments/` - Research and prototypes
- `/docs/` - Documentation
- `/business/` - Business strategy and roadmap
- `/tools/` - Development and build tools
- `/security/` - Regulatory policies and integrity manifests

### Key Hardening Features (v1.1.0+)
- **Fail-Closed Proxy**: Replaced performance-shield load shedding with a strict semaphore-based queue.
- **SSRF Protection**: Strict IP/DNS validation with rebinding protection for all egress targets.
- **Secrets Enforcement**: Mandatory `OCU_MASTER_KEY` and `OCU_SALT` validation for production deployments.
- **Enterprise Observability**: Integrated Prometheus metrics for real-time tracking of latency, queue depth, and PII hit rates.
- **Zero-Bypass AI**: Tier 2 AI deep scan is now mandatory and strictly enforced (no silent bypasses).

## Key Features

- **Tier 0 — Dictionary Shield**: Integrated high-speed dictionary protection for VIPs, internal project codes, and corporate secrets. Includes automated CRM/LDAP synchronization.
- **Native SLM Inference (CGO)**: Embedded `llama.cpp` decoding loop for local, sovereign PII scanning without cloud egress.
- **Identity-Aware Auditing**: Transparent `Authorization` header extraction for precise actor attribution in all refinery logs.
- **Zero-Egress Architecture**: PII is masked *before* ever leaving the trust boundary or hitting an LLM.

## Getting Started

1.  **Environment Setup**:
    ```bash
    cp .env.example .env
    # Fill in your master keys and API keys
    ./scripts/orchestrate.sh  # Run the Ocultar Protocol validation (In-Progress)
    ```

2.  **Go Workspace**:
    This project uses Go workspaces. Ensure you have Go 1.25.7+ installed.
    ```bash
    go work sync
    ```

3.  **Build and Run**:
    Use the root `Makefile` to build all components:
    ```bash
    make build
    go run ./apps/sombra  # See apps/sombra/README.md for configuration
    ```

## AI-Driven Governance

OCULTAR is maintained by a specialized ecosystem of 40+ AI Agent Skills. Every change to this repository is verified by the **Ocultar Protocol** (v1.0), ensuring deterministic security and compliance. See [ROADMAP.md](./ROADMAP.md) for the implementation status of the full 16-step sequence.

- **Orchestration**: Managed by `/.agents/skills/`.
- **Policy**: Defined in `/security/regulatory_policy.json`.
- **Audit**: Verifiable logs signed with Ed25519.

## Development

- **Documentation**: See `/docs` for architecture and API references.
- **Roadmap**: See [ROADMAP.md](./ROADMAP.md) for upcoming features and releases.
- **Testing**: Run `go test ./...` from the root.
