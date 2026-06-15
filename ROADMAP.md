# Roadmap

This document describes what we are working on and what is coming next. It is updated as priorities change.

For bugs and feature requests, open an issue. For security vulnerabilities, see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Released — v1.14.0 (June 2026)

Initial public release.

- Tier 1 deterministic refinery: 63 PII/PHI entity types across 12 categories
- Tier 2 contextual AI (SLM): 114/117 entity types with a local NER sidecar
- Base64 and JWT evasion shield (Tier 0.1)
- Sovereign vault: AES-256-GCM + HKDF-SHA256, DuckDB backend
- Ed25519-signed immutable audit log
- Fail-closed design: vault failure blocks requests, never forwards raw text
- Regulatory framework coverage: HIPAA, GDPR, CCPA, PCI-DSS, SOX, FERPA, BIPA, NYDFS
- Persistent entity registry with pre-seeding support
- Zero-egress proxy (Sombra gateway)
- Tier 2 circuit breaker: automatic Tier 1 fallback when SLM sidecar is unavailable
- Claude MCP extension (`extensions/claude/`)
- Goose MCP extension (`extensions/goose/`)
- Docker image published to GHCR (`ghcr.io/ocultar-dev/ocultar`)
- Cosign-signed Docker images (Sigstore keyless)
- SBOM published with each release (CycloneDX JSON)

---

## Near-term (v1.15)

### PII coverage

- [ ] Raise Tier 1 (deterministic) coverage from 63 → 80+ entity types
- [ ] Fix 5 partially-detected types: PASSPORT, VEHICLE_ID, BIOMETRIC_ID, GENETIC_ID, DEVICE_ID
- [ ] Add test fixtures for FERPA, BIPA, and NYDFS regulatory frameworks

### Testing

- [ ] Expand adversarial test suite: Base64-encoded PII, JSON-nested PII, mixed-language (FR+EN), Unicode obfuscation
- [ ] Integration test: Sombra zero-egress — verify no raw PII reaches the upstream AI
- [ ] Circuit breaker test: SLM down → Tier 1 only, no hang, no error

### Developer experience

- [ ] Cursor / Windsurf MCP connector
- [ ] Structured error codes in API responses (machine-readable failure reasons)
- [ ] `golangci-lint` pass across all modules

---

## Medium-term (v1.16 – v2.0)

### Observability

- [ ] Prometheus metrics endpoint (`/metrics`): request count, refine latency, vault size, tier hit rates
- [ ] OpenTelemetry trace export (optional, off by default)

### Entity coverage

- [ ] 117/117 Tier 1 coverage (close the remaining SLM-only gap)
- [ ] Custom entity type API: define new PII types at runtime via the entity registry

### Resilience

- [ ] Request-level timeout configuration (`OCU_REFINE_TIMEOUT`)
- [ ] Vault compaction — prune tokens older than a configurable TTL

### Deployment

- [ ] Helm chart for Kubernetes
- [ ] ARM64 Docker image (Apple Silicon / AWS Graviton)
- [ ] `apt`, `brew`, and `rpm` native packages

---

## Long-term (v2.x+)

- [ ] Policy engine: per-connector rules defining which PII types to mask and which to pass through
- [ ] Streaming refinery: mask text in real-time as it streams from an AI model response
- [ ] FIPS 140-2 cryptographic module (`GOEXPERIMENT=boringcrypto`)
- [ ] SOC 2 Type II audit
- [ ] GDPR Article 25 DPA template reviewed by EU privacy counsel
- [ ] Third-party penetration test — publish summary (redacted)
- [ ] ANSSI CSPN evaluation (French regulated market)

---

## Not planned

- **Cloud-hosted version** — Ocultar is designed to be sovereign and local-only.
- **Built-in AI model** — the SLM sidecar is intentionally user-supplied to avoid bundling large model weights.
- **Automatic upstream AI credential management** — Ocultar is a sidecar; it does not manage AI provider API keys or routing.

---

*Apache 2.0 — Self-hosting is free and always will be.*
