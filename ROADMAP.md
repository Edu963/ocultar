# OCULTAR | Project Roadmap

This document outlines the development trajectory of the Unified OCULTAR Engine. 

---

## ✅ Phase 1: Core Sovereignty (RELEASED)
*Focus: Local inference, deterministic redaction, and encrypted storage.*

- [x] **Tier 1 — Deterministic Refinery**: High-speed regex and heuristic detection pipeline.
- [x] **Zero-Egress Proxy**: Transparent reverse proxy for OpenAI-compatible APIs.
- [x] **Sovereign Vault**: Encrypted local storage (DuckDB) for PII re-hydration.
- [x] **Tier 2 — Contextual AI**: Local SLM inference (Qwen/Phi via `llama.cpp`) for prose NER.
- [x] **Shield Manager UI**: Dashboard modals for managing regex and keyword dictionaries in real-time.
- [x] **CRM/LDAP Ingestion**: Automated background polling for protected identities (Enterprise).

## 🏃 Phase 2: Enterprise Hardening (IN PROGRESS)
*Focus: Scalability, high availability, and deeper auditability.*

- [/] **16-Step Ocultar Protocol**: Automating the full compliance validation DAG.
- [x] **PostgreSQL HA Vault**: Support for shared, multi-node identity storage.
- [ ] **Native CGO Decoding Loop**: Moving SLM inference from sidecar to in-process for 40% latency reduction.
- [x] **Identity-Aware Auditing**: JWT header extraction for precise actor attribution in logs.
- [ ] **Tamper-Proof Audit Trail**: Ed25519 signing for all audit log entries (Audit-Ready).

## 🚀 Phase 3: Ecosystem Expansion (BACKLOG)
*Focus: Broadening the reach of the refinery beyond HTTP proxying.*

- [ ] **Sombra Gateway v3**: Advanced multi-model routing with cost-aware load balancing.
- [ ] **Native Connectors**: Direct integration for Slack, SharePoint, and Microsoft Teams.
- [ ] **Shadow AI Discovery**: Passive scanning of network traffic to identify un-managed AI usage.
- [ ] **Regulatory Intent Decoder**: Auto-generating refinery rules directly from PDF regulation files.

---

*Last Updated: 2026-04-19*
