# Contributing to Ocultar

Thank you for your interest in improving Ocultar. This document covers everything you need
to go from zero to a merged pull request.

---

## Table of Contents

1. [Before you start](#before-you-start)
2. [Development setup](#development-setup)
3. [Running the sidecar locally](#running-the-sidecar-locally)
4. [Running tests](#running-tests)
5. [Branch naming](#branch-naming)
6. [Commit messages](#commit-messages)
7. [Pull request process](#pull-request-process)
8. [Code style](#code-style)
9. [Adding a new PII entity type](#adding-a-new-pii-entity-type)
10. [Adding a new regulatory framework](#adding-a-new-regulatory-framework)
11. [Reporting bugs](#reporting-bugs)
12. [Reporting security vulnerabilities](#reporting-security-vulnerabilities)
13. [Developer Certificate of Origin](#developer-certificate-of-origin)

---

## Before you start

- Check [open issues](https://github.com/ocultar-dev/ocultar/issues) to avoid duplicate work.
- For large changes, open an issue first to discuss the approach before writing code.
- Security vulnerabilities must **not** be reported as public issues — see [Reporting security vulnerabilities](#reporting-security-vulnerabilities).

---

## Development setup

**Prerequisites:**

| Tool | Minimum version | Notes |
|---|---|---|
| Go | 1.22 | CGO must be enabled — DuckDB and libphonenumber require a C compiler |
| GCC / Clang | any recent | Linux: `sudo apt install build-essential`; macOS: Xcode CLI tools |
| Docker | 20+ | Optional — required only for running the full stack via `docker compose` |
| Node.js | 18+ | Required only for `apps/dashboard` development |

**Clone and verify:**

```bash
git clone https://github.com/ocultar-dev/ocultar.git
cd ocultar
CGO_ENABLED=1 go build ./...
CGO_ENABLED=1 go test ./...
```

Both commands should complete with no errors before you start making changes.

---

## Running the sidecar locally

Generate fresh keys — never reuse keys across environments:

```bash
export OCU_MASTER_KEY=$(openssl rand -hex 32)
export OCU_SALT=$(openssl rand -hex 16)
export OCU_AUDITOR_TOKEN=$(openssl rand -hex 24)
```

Start the sidecar:

```bash
go run ./services/refinery/cmd/ --serve 4141
```

The sidecar binds to `127.0.0.1:4141` only. Test it:

```bash
# Health check
curl -s http://127.0.0.1:4141/api/health | jq .

# Refine (mask PII)
curl -s -X POST http://127.0.0.1:4141/api/refine \
  -H "Content-Type: application/json" \
  -d '"Hello Alice, contact me at alice@example.com and call +1-555-867-5309"'
```

---

## Running tests

```bash
# All tests across the monorepo
CGO_ENABLED=1 go test ./...

# Specific package
cd services/refinery && CGO_ENABLED=1 go test ./... -run TestMyRule

# With race detector (recommended before submitting a PR)
CGO_ENABLED=1 go test -race ./...

# Verbose output
CGO_ENABLED=1 go test -v ./...
```

Tests use in-memory DuckDB (vault path `""`), so they leave no `.db` files behind.

---

## Branch naming

```
feat/short-description        # new feature
fix/short-description         # bug fix
docs/short-description        # documentation only
test/short-description        # tests only
chore/short-description       # tooling, CI, deps
```

Examples:
```
feat/de-personalausweis-detection
fix/phone-year-guard-false-positive
docs/deployment-guide
```

---

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) with a scope matching
the affected module:

```
feat(refinery): add FR_IBAN regex to Tier 1 rule engine
fix(vault): handle concurrent RegisterEntity race on DuckDB
test(pii): add negative fixtures for DE_PERSONALAUSWEIS
docs(readme): fix Go version badge
chore(ci): pin setup-go to commit SHA
```

**Common scopes:** `refinery`, `vault`, `sombra`, `slm-engine`, `proxy`, `pii`, `dashboard`, `ci`, `docs`.

Keep the subject line under 72 characters. Use the body for the *why*, not the *what*.

---

## Pull request process

1. **Fork** the repository and create your branch from `main`.
2. **Write tests** for any new behaviour. PRs that reduce test coverage will not be merged.
3. **Run the full test suite** locally before pushing: `CGO_ENABLED=1 go test -race ./...`
4. **Run code style checks**: `go fmt ./... && go vet ./...`
5. **Sign your commits** with a DCO sign-off — see [Developer Certificate of Origin](#developer-certificate-of-origin).
6. **Open the PR** against `main` with a clear description of what changed and why.
7. **Fill out the PR checklist** (the template will appear automatically).
8. A maintainer will review within 5 business days.

**PR checklist (also in the PR template):**

- [ ] `CGO_ENABLED=1 go test -race ./...` passes with no failures
- [ ] `go fmt ./...` and `go vet ./...` produce no output
- [ ] No new `0.0.0.0` bindings — server listeners must use `127.0.0.1`
- [ ] No hardcoded secrets, API keys, or real PII in test fixtures
  (use `555-01-XXXX` SSNs, RFC 5737 IPs `192.0.2.x`, clearly synthetic names)
- [ ] Every new PII entity type in `internal/pii/registry.go` has at least one positive and one negative test fixture
- [ ] Any change to the privacy pipeline has a corresponding entry in the CHANGELOG

---

## Code style

Standard Go tooling only — no custom linter config required:

```bash
go fmt ./...
go vet ./...
```

**Key conventions:**

- Server listeners bind to `127.0.0.1` explicitly — never `0.0.0.0` or `localhost`.
- Errors are returned, not logged-and-swallowed. Log at the call site if needed.
- New packages get a `_test.go` file before or alongside the implementation.
- Avoid global mutable state. Pass dependencies explicitly.

---

## Adding a new PII entity type

The canonical source of truth is `internal/pii/registry.go`. Every entity type is an `EntityDef` struct.

**Step 1 — Add the EntityDef to `registry.go`:**

```go
{
    Name:    "FR_NIR",                          // canonical token prefix: [FR_NIR_xxxx]
    Pattern: regexp.MustCompile(`\b[12]\d{14}\b`),
    Score:   2,                                 // higher score wins tie-breaks
    Validator: validateLuhn,                    // optional; nil if no checksum
},
```

**Step 2 — Add table-driven tests in `internal/pii/engine_test.go`:**

```go
{name: "FR_NIR positive",  input: "NIR: 1 85 02 75 056 048 46", wantType: "FR_NIR"},
{name: "FR_NIR negative",  input: "code 123456789",              wantType: ""},
```

**Step 3 — Update `data/pii_coverage.json`** (in the `kii` repo) if the new type maps to a
regulatory framework in the coverage registry.

**Step 4 — Add a fixture** to the relevant file in `tests/regulatory_coverage/fixtures/`
(or create a new fixture file for a new jurisdiction).

---

## Adding a new regulatory framework

New frameworks (e.g. a new EU member-state law or a US state privacy act) require changes
in three places:

1. **`internal/pii/registry.go`** — ensure all entity types required by the framework
   have entries. Add comments citing the specific article or safe-harbour identifier.

2. **`tests/regulatory_coverage/coverage_test.go`** — add a `FrameworkMinimum` entry
   specifying the minimum number of required types that must pass for the framework to be
   considered covered:

   ```go
   {Framework: "MY_LAW", MinRequired: 5},
   ```

3. **`tests/regulatory_coverage/fixtures/`** — add a `.json` fixture file with at least
   one realistic (but clearly synthetic) PII sample per required entity type.

Run `CGO_ENABLED=1 go test ./tests/regulatory_coverage/...` to confirm all framework
minimums pass before submitting the PR.

---

## Reporting bugs

Open a [GitHub issue](https://github.com/ocultar-dev/ocultar/issues/new?template=bug_report.md)
with:

- Ocultar version (`/api/health` → `version` field)
- Operating system and architecture
- Minimal reproduction case (a `curl` command is ideal)
- Expected vs actual behaviour
- Any relevant logs (redact any real PII before pasting)

---

## Reporting security vulnerabilities

**Do not open a public issue for security vulnerabilities.**

Email [security@getki.ai](mailto:security@getki.ai) with:

- A description of the vulnerability
- Steps to reproduce
- Potential impact assessment

We acknowledge within 48 hours and aim to ship a fix within 14 days for critical issues.
You will be credited in the release notes unless you prefer to remain anonymous.

See [SECURITY.md](SECURITY.md) for the full disclosure policy.

---

## Developer Certificate of Origin

By contributing to Ocultar you agree to the
[Developer Certificate of Origin v1.1](https://developercertificate.org/).

Add a sign-off to every commit:

```bash
git commit -s -m "feat(refinery): add FR_NIR detection"
```

This appends `Signed-off-by: Your Name <your@email.com>` to the commit message.
If you forget, you can add it to the last commit with `git commit --amend -s`.
