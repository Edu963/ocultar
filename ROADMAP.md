# OCULTAR | Project Roadmap

This document outlines the development trajectory of the Unified OCULTAR Engine.

---

## ✅ Phase 1: Core Sovereignty (RELEASED)
*Focus: Local inference, deterministic redaction, and encrypted storage.*

- [x] **Tier 1 — Deterministic Refinery**: High-speed regex and heuristic detection pipeline (EMAIL, SSN, PHONE, CC, IBAN, and 30+ entity types).
- [x] **Base64 / JWT Evasion Shield (Tier 0.1)**: Recursive decode-and-rescan loop defeats encoding obfuscation.
- [x] **Zero-Egress Proxy**: Transparent reverse proxy for OpenAI-compatible APIs.
- [x] **Sovereign Vault**: Encrypted local storage (DuckDB) with AES-256-GCM + HKDF-SHA256 key derivation.
- [x] **Tier 2 — Contextual AI**: Model-agnostic SLM interface; supports OpenAI Privacy Filter (default) and llama.cpp backends via `TIER2_ENGINE` env var.
- [x] **Shield Manager UI (Dashboard)**: React + Tailwind dashboard for live redaction testing and system monitoring.
- [x] **CRM/LDAP Ingestion**: Automated background polling for protected identities (Enterprise — `apps/sombra`).
- [x] **Sombra Gateway**: Multi-model AI router with domain allowlisting, data policy enforcement, and adapter registry.
- [x] **Ed25519 Immutable Audit Log**: SHA-256 hash-chained, Ed25519-signed audit trail in `services/refinery/pkg/audit`.
- [x] **Identity-Aware Auditing**: JWT header extraction for actor attribution in logs.
- [x] **PostgreSQL HA Vault**: Support for shared, multi-node identity storage (Enterprise).

---

## 🏃 Phase 2: Enterprise Hardening (IN PROGRESS)
*Focus: Scalability, distribution, and compliance readiness.*

- [/] **Test Suite Coverage**: Go and React test suites for all 8 modules (proxy, refinery, vault, slm-engine, sombra, automation_bridge, dashboard, internal/pii). Integration gap remains for fail-closed scenarios and SSRF protection.
- [ ] **Fail-Closed Integration Tests**: Verify that SLM timeout, vault failure, and empty protected_entities.json all result in 500 with no PII leakage (see TASK 0.4).
- [ ] **SSRF / DNS Rebinding Protection**: Block RFC 1918, 169.254.169.254, and IPv6 loopback in proxy. Write tests.
- [ ] **16-Step Ocultar Compliance Protocol**: Automating the full compliance validation DAG for enterprise onboarding.
- [ ] **Per-Tier API Rate Limiting**: Monthly call caps enforced by API key metadata; `X-Ocultar-Calls-Remaining` header.
- [ ] **Prometheus Metrics Endpoint**: `GET /metrics` with tier hit rates, latency quantiles, vault size, queue depth.
- [ ] **Native CGO Decoding Loop**: Move SLM inference from sidecar to in-process for ~40% latency reduction.
- [ ] **Vault Persistence Integration Test**: Docker Compose restart → re-hydration round-trip test.

---

## 🚀 Phase 3: Ecosystem Expansion (BACKLOG)
*Focus: Distribution, compliance certification, and AI governance.*

- [ ] **Goose MCP Extension**: Local MCP server wrapping `POST /refine` for Goose AI agent workflows.
- [ ] **Anthropic Claude MCP Connector**: Same architecture, listed in Anthropic connector directory.
- [ ] **AWS Marketplace Listing**: Container listing for Sombra + Refinery with Private Offer pricing.
- [ ] **GDPR Article 25 Compliance Pack**: DPA template, privacy-by-design mapping, EU AI Act alignment note.
- [ ] **Native Connectors**: Direct integrations for Slack, SharePoint, and Microsoft Teams message streams.
- [ ] **Regulatory Intent Decoder**: Auto-generate refinery rules from uploaded PDF regulations.
- [ ] **Shadow AI Discovery**: Passive scanning of network traffic to identify unmanaged AI usage.
- [ ] **Bpifrance Grant Application**: Innovation dossier for €100K Prêt d'Innovation (Grenoble/AuRA region).

---

*Last Updated: 2026-04-26*
