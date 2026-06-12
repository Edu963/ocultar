# 🧪 Testing & Security Gates

OCULTAR maintains high standards for code quality and security. Every change must pass a series of automated gates.

## 1. Running Unit Tests
We use standard Go testing tools. Since OCULTAR uses DuckDB and concurrency, there are a few extra flags to keep in mind.

```bash
# Run all tests (CGO required for DuckDB)
CGO_ENABLED=1 go test ./...

# Run with race detector (Highly Recommended)
# This is critical for catching concurrency bugs in the Sombra gateway.
CGO_ENABLED=1 go test -race ./...

# Run with coverage report
CGO_ENABLED=1 go test -coverprofile=coverage.out ./...
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

## 3. CI Security Gates
The following gates run automatically on every PR via GitHub Actions:

1.  **Race-detected tests**: `CGO_ENABLED=1 go test -race ./...`
2.  **Static analysis**: `golangci-lint run` (errcheck, govet, staticcheck, gosec)
3.  **Vulnerability scan**: `govulncheck ./...`
4.  **Secret scan**: `gitleaks detect --source . --config .gitleaks.toml`
5.  **Architectural linter**: `tools/scripts/scripts/run_arch_linter.sh` (no illegal cross-package imports)

Run them locally before pushing:
```bash
CGO_ENABLED=1 go test -race ./...
go vet ./...
golangci-lint run
```

---

## 4. PR Checklist
Before submitting a Pull Request, ensure:
- [ ] `CGO_ENABLED=1 go test -race ./...` passes.
- [ ] `CGO_ENABLED=1 go build ./...` succeeds across all workspace modules.
- [ ] No new linter warnings (`golangci-lint run`).
- [ ] New detection rules have at least 3 test cases (Match, No Match, Boundary).

> [!IMPORTANT]
> **No Side Effects in Tests**: Tests must not rely on local disk state. Always use `:memory:` for DuckDB tests and `config.InitDefaults()` to reset global state.
