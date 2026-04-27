# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-04-27

### Added
- **Tier 1 — Deterministic Refinery**: High-speed regex and heuristic detection pipeline for EMAIL, SSN, PHONE, CC, IBAN, and 30+ entity types.
- **Base64 / JWT Evasion Shield (Tier 0.1)**: Recursive decode-and-rescan loop to defeat encoding obfuscation.
- **Zero-Egress Proxy**: Transparent reverse proxy for OpenAI-compatible APIs ensuring data never leaves the trust boundary.
- **Sovereign Vault**: Encrypted local storage (DuckDB) with AES-256-GCM + HKDF-SHA256 key derivation for secure PII tokenization.
- **Tier 2 — Contextual AI**: Model-agnostic SLM interface supporting OpenAI Privacy Filter (default) and llama.cpp backends.
- **Shield Manager UI (Dashboard)**: React-based dashboard for live redaction testing and system monitoring.
- **Sombra Gateway**: Multi-model AI router with domain allowlisting and data policy enforcement.
- **Ed25519 Immutable Audit Log**: SHA-256 hash-chained, Ed25519-signed audit trail for verifiable compliance.
- **Identity-Aware Auditing**: JWT header extraction for actor attribution in logs.
- **Doppler Integration**: Secure runtime secret injection and management.

### Changed
- Migrated from `.env` files to Doppler for production-grade secret management.
- Refactored refinery engine abstraction to support pluggable AI models.

### Fixed
- Improved Luhn-algorithm validation for credit card detection to reduce false positives.
- Hardened PII detection logic for nested JSON payloads.

---
[1.0.0]: https://github.com/Edu963/ocultar/releases/tag/v1.0.0
