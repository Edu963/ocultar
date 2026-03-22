---
name: regulation-digest-ingestor
description: Expert AI Orchestrator for regulatory source monitoring. Fetches updates from official bodies and pre-processes them for the policy-schema-generator.
---

# Regulation Digest Ingestor (v1.1)

## Purpose

The RDI triggers the compliance pipeline by monitoring and collecting regulatory updates. It sanitizes noisy sources (web, PDF, email) into a format ready for semantic decoding.

## Inputs / Outputs

### Inputs
- `source_uri`: URL or absolute path of the source.
- `ingestion_mode` (Enum): `PUBLIC` | `INTERNAL` | `AUTO_SCAN`.

### Outputs
- `regulation_digest`: Markdown artifact containing sanitized text.
- `relevance_score` (Int): 0-100 based on keyword density (GDPR, PII, Encryption).
- `sources_verified` (Boolean): Cryptographic proof of source origin (if available).

## Preconditions
- Network access for `PUBLIC` mode.
- `read_url_content` or `view_file` permissions.

---

## Instructions

### 1. Source Acquisition
- Use `read_url_content` (Public) or `view_file` (Internal/PDF).
- **Halt**: If 404 or access denied, retry once with `RETRY_AFTER` header respect.

### 2. Sanitization Pipeline
- Strip HTML tags, headers/footers, and ad blocks.
- **Deterministic Format**: Structure as `## [SOURCE]`, `### [SECTION]`.
- Output MUST be a clean markdown string.

### 3. Relevance Calculation
Assign score based on weighted keywords:
- `PII`, `Personal Data`, `Healthcare`: +30 pts.
- `Encryption`, `SSN`, `Masking`: +40 pts.
- `Admin`, `Procedural`: +5 pts.

## Failure Handling
- **`UNSUPPORTED_FORMAT`**: Fail if binary data (other than PDF) is encountered.
- **`BOT_BLOCKED`**: If public sources block the crawler, report `ACCESS_DENIED`.

## Postconditions
- Artifact MUST be saved in `docs/compliance/digests/{{timestamp}}.md`.
- Metadata MUST be registered in `ecosystem-state-tracker`.
