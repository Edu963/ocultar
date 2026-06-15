# 🛠️ Developer Learning Path

Welcome to the OCULTAR Developer Path. This track is designed for software engineers, security researchers, and integrators who want to understand the inner workings of OCULTAR and contribute to its development.

## 🎯 Learning Objectives
By the end of this path, you will be able to:
- Navigate the OCULTAR monorepo and understand its component boundaries.
- Build and run OCULTAR in a development environment.
- Add new PII detection rules and validation logic.
- Extend the vaulting system with new backends.
- Understand and enforce the "Fail-Closed" security model in code.

---

## 📚 Curriculum

### 1. [Monorepo & Module Structure](./module-structure.md)
Understand how we use Go Workspaces to manage multiple modules and where each component lives.

### 2. [Extending the Refinery](./extending-refinery.md)
Learn how to add new PII types, from simple regex rules to complex checksum-validated entities.

### 3. [Testing & Security Gates](./testing-and-security.md)
How to run the test suite, use the race detector, and pass the CI security gates.

### 4. [SLM Engine Deep-Dive](../architecture/refinery-pipeline.md#tier-2-ai-ner-small-language-model)
Understanding the local inference engine and how to swap backends (llama.cpp vs. Python sidecar).

---

## 🛠️ Environment Setup

### Option A — Docker (quickest, no Go toolchain needed)
Prerequisite: Docker + Docker Compose v2.

```bash
git clone https://github.com/ocultar-dev/ocultar.git
cd ocultar
docker compose up --build
```

Refinery is live at `http://localhost:4141`. No secrets setup required — dev defaults are pre-filled in `.env.example` and picked up automatically.

For full AI Tier 2 NER:
```bash
OCU_PILOT_MODE=0 docker compose --profile ai up --build
```

### Option B — Native Go (for active contributors)
Required tools:
- **Go 1.24+**
- **GCC / CGO** (required for DuckDB)
- **Docker + Compose** (for integration tests)

```bash
git clone https://github.com/ocultar-dev/ocultar.git
cd ocultar
go work sync
make build
```

Secrets: copy `.env.example` to `.env` and fill in values, then `make build && ./bin/proxy`.

---

## 📝 Coding Conventions
- **Fail-Closed**: Always block on error. Never return partially cleaned data.
- **Thread Safety**: All shared state must be protected by mutexes or atomics.
- **Deterministic**: Tokens must be derived via `SHA-256(PII)`.
- **No Panics**: Never panic in `pkg/` packages. Return errors to the caller.
