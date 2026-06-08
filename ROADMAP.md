# OCULTAR | Project Roadmap

This document outlines the development trajectory of the OCULTAR engine.

---

## ✅ Phase 1: Core Sovereignty (RELEASED)
*Focus: Local inference, deterministic redaction, and encrypted storage.*

- [x] **Tier 1 — Deterministic Refinery** — High-speed regex and heuristic detection pipeline (EMAIL, SSN, PHONE, CC, IBAN, and 30+ entity types).
- [x] **Base64 / JWT Evasion Shield (Tier 0.1)** — Recursive decode-and-rescan loop defeats encoding obfuscation.
- [x] **Zero-Egress Proxy** — Transparent reverse proxy for OpenAI-compatible APIs.
- [x] **Sovereign Vault** — Encrypted local storage (DuckDB) with AES-256-GCM + HKDF-SHA256 key derivation.
- [x] **Tier 2 — Contextual AI** — Model-agnostic SLM interface; supports OpenAI Privacy Filter (default) and llama.cpp backends via `SLM_ADAPTER` env var.
- [x] **Dashboard** — React + Tailwind UI for live redaction testing and system monitoring.
- [x] **Sombra Gateway** — Multi-model AI router with domain allowlisting, data policy enforcement, and adapter registry (OpenAI, Claude, Gemini, Mistral, Local SLM).
- [x] **Ed25519 Immutable Audit Log** — SHA-256 hash-chained, Ed25519-signed audit trail.
- [x] **PostgreSQL HA Vault** — Multi-node identity storage for enterprise deployments.
- [x] **CRM/LDAP Ingestion** — Automated background polling for protected identities.

---

## ✅ Phase 2: Enterprise Hardening (COMPLETE)
*Focus: Scalability, distribution, and compliance readiness.*

- [x] **Test Suite Coverage** — Go and React test suites across all 8 modules: proxy, refinery, vault, slm-engine, sombra, automation_bridge, dashboard, internal/pii.
- [x] **Fail-Closed Integration Tests** — SLM timeout, vault failure, and empty `protected_entities.json` all verified to return HTTP 500 with zero PII leakage.
- [x] **SSRF / DNS Rebinding Protection** — RFC 1918, 169.254.169.254, and IPv6 loopback blocked in the proxy with full test coverage.
- [x] **Prometheus Metrics Endpoint** — `GET /metrics` with tier hit rates, latency quantiles, vault size, and queue depth.
- [x] **Vault Persistence Integration Test** — Close + reopen round-trip verified; token determinism and key isolation confirmed.
- [x] **GDPR Article 25 Compliance Pack** — DPA template, privacy-by-design mapping, and EU AI Act alignment.
- [x] **Claude MCP Extension** — Stdio MCP server for Claude Desktop — `refine_text` + `reveal_tokens` tools, fail-closed.
- [x] **Goose MCP Extension** — Stdio MCP server for Goose AI — `refine_text` tool, fail-closed.
- [x] **Mistral Le Chat MCP Connector** — MCP extension for Mistral's client runtime; same architecture as Claude/Goose extensions.

---

## 🏃 Phase 3: Ecosystem Expansion (IN PROGRESS)
*Focus: Developer adoption, connector ecosystem, and AI governance.*

- [x] **Persistent Entity Registry** — Session-spanning identity resolution that collapses all name variants (`"John"`, `"Doe"`, `"J. Doe"`) to a single canonical token (`[PERSON_1]`). Ships two new DuckDB/PostgreSQL tables, five new `Provider` interface methods, and three Sombra API endpoints (`POST /v1/entities`, `POST /v1/entities/seed`, `GET /v1/entities`).
- [ ] **Cursor / Windsurf MCP Connector** — Developer-facing extension for AI coding assistants.
- [ ] **Native CGO Decoding Loop** — Move SLM inference from Python sidecar to in-process CGO for ~40% latency reduction. Removes the Python runtime dependency in the community build.
- [ ] **Regulatory Intent Decoder** — Auto-generate Refinery detection rules from uploaded PDF regulations (CNIL decisions, GDPR guidance, sector-specific rules).
- [ ] **Shadow AI Discovery** — Passive scanning of egress traffic to identify unmanaged AI API usage within the network perimeter.
- [ ] **Remote Model Orchestration** — Migrate large model binaries to a private Hugging Face repository with automated download scripts; keeps the repository lightweight.
- [ ] **LangChain & Flowise Native SDK** — First-class Python/JS integrations for privacy-first AI pipelines.
- [ ] **Multi-Arch Docker Distribution** — `arm64` (Graviton/M-series) and `x86_64` via `docker buildx`.
- [ ] **GitHub Releases + GHCR Automated Pipeline** — Versioned releases on `v*` tags with signed binary assets and a one-command install path.

---

## 🎯 Phase 4: Production-Grade Hardening
*Focus: Regulated-market readiness, connector hardening, and operational evidence.*

### Cryptographic Compliance
- [ ] **FIPS 140-2 Cryptographic Module** — Swap `crypto/aes` for `GOEXPERIMENT=boringcrypto`. Required by US and EU financial regulators.

### Connector Hardening
- [x] **SharePoint Real Graph API** — Live OAuth2 client credentials flow, MS Graph `/drive/root/delta` incremental sync, and real file content extraction (DOCX, XLSX, PPTX, PDF, plain text).
- [ ] **Slack Events API (Push Mode)** — Upgrade from pull-mode to event-driven webhooks with exponential-backoff retry and dead-letter queue.

### Reliability & Observability
- [x] **Tier 2 Circuit Breaker** — Three-state circuit breaker (closed → open → half-open) with automatic Tier 1 fallback and `/api/health` exposure.
- [ ] **Refinery Regression Suite** — Labeled PII corpus tests covering all detection tiers (0–3).
- [ ] **Performance Benchmark Harness** — Reproducible P50/P95/P99 latency suite against the full pipeline.

### Integrations & Connectors
- [ ] **SIEM Log-Scrubbing Connectors (Splunk/Datadog)** — Intercept and redact PII before it reaches observability platforms.

---

## 🔭 Phase 5: Future Considerations
*These items have strategic merit but depend on scale and market signals from Phases 3–4.*

### Sovereign AI Blueprints (Additional Verticals)
- [ ] **Healthcare Blueprint (HIPAA + GDPR Art. 9)** — Pre-configured policies for HEALTH_ENTITY, SENSITIVE_EVENT, MEDICAL_RECORD, PATIENT_ID. Target: hospital systems and health-tech platforms.
- [ ] **Legal / Law Firm Blueprint** — Policies for CLIENT_ID, INTERNAL_PROJECT, CREDENTIAL. Target: law firms and legal-tech platforms deploying AI for contract review.
- [ ] **Insurance Blueprint** — Policies for financial identifiers, health data, and cross-border location PII. Target: EU insurers using AI for claims processing.

### Platform
- [ ] **Streaming / Event Pipeline Connector** — Kafka and Redpanda support for real-time SIEM scrubbing.
- [ ] **Sensitivity Classification Layer** — Tag each detected entity with a sensitivity band (`public` / `confidential` / `regulated` / `restricted`).
- [ ] **Public Benchmarking Portal** — Live recall, precision, latency, and fail-closed coverage dashboard.

### Distribution
- [ ] **Native Package Installers** — `apt`, `brew`, and `rpm` packages for Fedora, Ubuntu, Debian, and macOS.
- [ ] **Domain-Specific Fine-Tuned NER Models** — LATAM, Nordics, DACH, and fr-finance variants built on demand with regional design partner labeled data.

---

*Apache 2.0 — Self-hosting is free and always will be.*
