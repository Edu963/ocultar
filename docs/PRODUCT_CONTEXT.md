# OCULTAR | Product Context

This document provides a deep, technical, and conceptual overview of the OCULTAR ecosystem. It outlines the core principles of **Zero-Egress security**, **PII refinement**, and **regulatory compliance** that guide every design decision.

## 1. Product Mission

OCULTAR is an enterprise-grade **Global Data Refinery**. Its mission is to enable safe AI adoption and infrastructure-level compliance by intercepting, refining, and monitoring all data flows—from HTTP to Syslog—converting regulatory liabilities into audit-safe assets.

## 2. Core Components

| Component | Description |
|---|---|
| **Live Refinery** | The multi-tier engine (Regex, NLP, SLM) that detects and obfuscates PII in real-time. |
| **Sombra Gateway** | A high-performance proxy and agentic gateway with native connectors (Slack, SharePoint) enforcing "Fail-Closed" security. |
| **Identity Vault** | A local, encrypted database (DuckDB/PostgreSQL) that stores mappings for deterministic re-hydration. |
| **Enterprise Dashboard** | A centralized UI for monitoring the **Risk Matrix**, **ROI Analytics**, and **Audit Logs**. |
| **Governance & Audit** | **RBAC** and audit systems that control and log "re-hydration" (un-vaulting) of sensitive data. |
| **Dictionary Shield** | **Tier 0** high-speed protection against known internal terms or proprietary keywords. |

## 3. Fundamental Principles

- **Zero-Egress**: Sensitive data never leaves your infrastructure. Only non-sensitive tokens are sent to external LLM providers.
- **Fail-Closed**: If a security check fails or an error occurs, the system defaults to blocking the request to prevent accidental leakage.
- **Privacy-by-Design**: User anonymity is preserved through robust, deterministic pseudonymization.
- **Regulatory Alignment**: Every interaction is mapped to specific regulatory articles (e.g., **GDPR**, **HIPAA**, **SOC 2 Type II**, **PCI DSS v4.0**, **NIS2**, **EU AI Act**, **BSI C5**) for automated compliance auditing.
- **Specialized Vision (SR-SLM)**: Intelligence is focused on high-precision entity recognition (data "eyes") while software controls the logic.

## 4. Target Users

- **CISO & Compliance Officers**: Monitor the Risk Matrix and ensure data privacy standards are met.
- **Enterprise IT / DevOps**: Deploy and maintain the OCULTAR/Sombra infrastructure.
- **App Developers**: Seamlessly integrate applications with safe AI via the Sombra Gateway.
- **Pilot Managers**: Validate platform value using structured onboarding and ROI metrics.

## 5. Security Goals

- **No Secret Leakage**: Credentials and internal paths are never included in distributions.
- **Evasion-Proof**: The Refinery is hardened against Base64 or URL-encoding obfuscation attacks.
- **Immutable Audit**: Logs are tamper-proof and provide a clear regulatory trail.

---

> [!TIP]
> This context is used by the **Continuous AI Orchestrator** to verify that every code change aligns with the product's core security and privacy objectives.
