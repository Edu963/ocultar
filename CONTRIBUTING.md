# Contributing to Ocultar

## Running tests

Requires Go 1.25+ with CGO enabled (DuckDB and libphonenumber require a C compiler).

```bash
CGO_ENABLED=1 go test ./...
```

Run a specific test:

```bash
cd services/refinery && CGO_ENABLED=1 go test ./... -run TestMyRule
```

Tests use in-memory DuckDB (vault path `""`), so they leave no `.db` files behind.

## Running the sidecar locally

```bash
export OCU_MASTER_KEY=$(openssl rand -hex 32)
export OCU_SALT=$(openssl rand -hex 16)
export OCU_AUDITOR_TOKEN=$(openssl rand -hex 24)

go run ./services/refinery/cmd/main.go --serve 4141
```

The sidecar listens on `127.0.0.1:4141` only. Test it:

```bash
curl -s -X POST http://127.0.0.1:4141/api/refine \
  -H "Content-Type: application/json" \
  -d '"Hello Alice, contact me at alice@example.com"'
```

## PR checklist

- [ ] `CGO_ENABLED=1 go test ./...` passes with no failures
- [ ] No new `0.0.0.0` bindings — server listeners must use `127.0.0.1`
- [ ] No hardcoded secrets, API keys, or real PII in test fixtures
- [ ] Every new regex rule in `configs/config.yaml` has a corresponding table-driven test

## Code style

Standard Go formatting and vet:

```bash
go fmt ./...
go vet ./...
```

Follow conventional commits with a scope matching the affected module:

```
feat(refinery): add FR_IBAN regex to Tier 1 rule engine
fix(vault): handle concurrent RegisterEntity race on DuckDB
```

Common scopes: `proxy`, `refinery`, `vault`, `sombra`, `slm-engine`, `pii`, `docs`.

## Adding a new PII type

1. Add the regex to `configs/config.yaml` under `regexes`.
2. Add a table-driven test in `services/refinery/pkg/refinery/` with at least one positive
   and one negative example.
