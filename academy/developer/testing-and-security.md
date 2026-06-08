# 🧪 Testing & Security Gates

OCULTAR maintains high standards for code quality and security. Every change must pass a series of automated gates.

## 1. Running Unit Tests
We use standard Go testing tools. Since OCULTAR uses DuckDB and concurrency, there are a few extra flags to keep in mind.

```bash
# Run all tests
go test ./...

# Run with race detector (Highly Recommended)
# This is critical for catching concurrency bugs in the Sombra gateway.
go test -race ./...

# Run with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 2. Fail-Closed Verification
We have specific tests to ensure that if a component fails, the entire system blocks the request.

Check `pkg/proxy/fail_closed_test.go` for examples of how we simulate:
- SLM engine timeout.
- Vault connection failure.
- Empty boot-guard file.

---

## 3. The 16-Step AI Orchestrator
OCULTAR uses a 16-step Continuous AI Orchestrator (`scripts/orchestrate.sh`). This script runs:
1.  **PII Detection Tests**: Verifies no regressions in the refinery.
2.  **Cross-Version Sync**: Ensures Community and Enterprise tiers are in sync.
3.  **Secret Scanner**: Scans for hardcoded keys or high-entropy strings in code.
4.  **Architectural Linter**: Enforces package boundaries (no illegal imports).
5.  **Zero-Egress Validator**: Checks configuration manifests for dangerous settings.

```bash
./scripts/orchestrate.sh
```

---

## 4. PR Checklist
Before submitting a Pull Request, ensure:
- [ ] `go test -race ./...` passes.
- [ ] `go build ./...` succeeds across all workspace modules.
- [ ] No new `Linter` warnings.
- [ ] `scripts/orchestrate.sh` passes 100%.
- [ ] New detection rules have at least 3 test cases (Match, No Match, Boundary).

> [!IMPORTANT]
> **No Side Effects in Tests**: Tests must not rely on local disk state. Always use `:memory:` for DuckDB tests and `config.InitDefaults()` to reset global state.
