---
name: Ocultar | Product Context
description: Expert Instructions (prompt-based persona) for the AI assistant. Core repository of project metadata, component responsibilities, and architectural boundaries. Ensure that all AI assistant responses align with the Ocultar architectural vision.

> [!NOTE]
> This skill consists of **Expert Instructions** for the AI assistant. It is a prompt-based persona, not an autonomous background service.
---

# Overview
Ocultar is an enterprise-grade "Global Data Refinery" providing local-first, zero-egress PII detection and redaction.

# Core Components
- **Proxy**: Transparent HTTP layer for real-time traffic scrubbing.
- **Engine**: The detection logic (Regex, Dictionaries, SLM).
- **Vault**: AES-256 encrypted storage for original PII (DuckDB/Postgres).
- **Sombra**: Agentic gateway for secure LLM orchestration.

# Architectural Rules
- **Fail-Closed**: If a scan fails or times out (>5s), the request must be blocked or fully redacted.
- **Zero-Egress**: Data never leaves the client's VPC in an unmasked state.
- **Pure Refinery**: The engine remains agnostic of its data sources.

# Responsibilities
- Maintain the source of truth for the project's technical state.
- Guide new contributors through the system architecture.
- Ensure consistency in branding and terminology ("Refinery," "Shield," "Vault").
