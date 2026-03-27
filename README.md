# Ocultar Monorepo

Welcome to the unified Ocultar Refinery ecosystem. This monorepo contains all core services, applications, and enterprise extensions.

## Structure

- `/apps/` - User-facing applications (Proxy, Sombra, Web)
- `/services/` - Core backend services (Refinery, Vault)
- `/enterprise/` - Enterprise-only extensions
- `/experiments/` - Research and prototypes
- `/docs/` - Documentation (Architecture and API)
- `/business/` - Business strategy and roadmap
- `/tools/` - Development and build tools
- `/security/` - Regulatory policies and integrity manifests

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
    ./scripts/orchestrate.sh  # Run the 16-Step Ocultar Protocol validation
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

OCULTAR is maintained by a specialized ecosystem of 40+ AI Agent Skills. Every change to this repository is verified by the **16-Step Ocultar Protocol**, ensuring deterministic security and compliance.

- **Orchestration**: Managed by `/.agents/skills/`.
- **Policy**: Defined in `/security/regulatory_policy.json`.
- **Audit**: Verifiable logs signed with Ed25519.

## Development

- **Documentation**: See `/docs` for architecture and API references.
- **Testing**: Run `go test ./...` from the root.
