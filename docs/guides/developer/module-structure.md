# 🏗️ Monorepo & Module Structure

OCULTAR is organized as a **Go Workspace** to ensure clean separation of concerns while allowing shared development across multiple modules.

## The Import Graph (Dependency Rules)

To maintain security and stability, we enforce strict dependency rules:

1.  **`pkg/config`**: Zero internal dependencies. Every other package can import it.
2.  **`pkg/vault`**: Depends on `pkg/config`. It should **never** import `pkg/refinery`.
3.  **`pkg/refinery`**: The core logic. Depends on `pkg/config` and `pkg/vault`. It should **never** import `pkg/proxy`.
4.  **`apps/sombra`**: The top-level application. It can import `pkg/refinery`, `pkg/vault`, and `pkg/config`.

---

## Folder Map

| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `/apps/sombra` | API Gateway | `main.go`, `pkg/handler/gateway.go` |
| `/apps/slm-engine` | AI Sidecar | `main.go`, `python/serve_privacy_filter.py` |
| `/services/refinery`| Detection Engine | `pkg/refinery/refinery.go` |
| `/services/vault` | Secure Storage | `vault.go`, `duckdb_provider.go` |
| `/internal/pii` | Detection Registry| `registry.go`, `engine.go`, `validators.go` |
| `/pkg/audit` | Audit Logging | `immutable_logger.go` |
| `/pkg/policy` | Policy Engine | Runtime policy evaluation and access control rules |

---

## Go Workspace (`go.work`)

The workspace allows you to work on multiple modules simultaneously without `replace` directives.

```go
go 1.25.8

use (
    ./apps/proxy
    ./apps/sombra
    ./internal/pii
    ./pkg/gateway
    ./services/refinery
    ./services/vault
)
```

### Pro Tip: Local Sombra Development
If you are developing Sombra and OCULTAR in parallel as sibling directories:
```bash
go work use ../sombra
```

---

## Building the Project

We use a `Makefile` to simplify common tasks.

- `make build`: Compiles all binaries.
- `make test`: Runs all unit tests.
- `make docker`: Builds the Docker images.

> [!WARNING]
> **CGO is required**. OCULTAR uses DuckDB for the vault, which requires a C compiler (GCC) to build the Go bindings.
