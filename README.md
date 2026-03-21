# Ocultar Monorepo

Welcome to the unified Ocultar Engine ecosystem. This monorepo contains all core services, applications, and enterprise extensions.

## Structure

- `/apps/` - User-facing applications (Proxy, Sombra, Web)
- `/services/` - Core backend services (Engine)
- `/enterprise/` - Enterprise-only extensions
- `/experiments/` - Research and prototypes
- `/docs/` - Documentation
- `/business/` - Business strategy and roadmap
- `/tools/` - Development and build tools

## Getting Started

1.  **Environment Setup**:
    ```bash
    cp .env.example .env
    # Fill in your master keys and API keys
    ```

2.  **Go Workspace**:
    This project uses Go workspaces. Ensure you have Go 1.25.7+ installed.
    ```bash
    go work sync
    ```

3.  **Build and Run**:
    Use the root `Makefile` (coming soon) or run individual apps:
    ```bash
    go run ./apps/proxy
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
