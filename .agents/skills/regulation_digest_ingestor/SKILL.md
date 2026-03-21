---
name: regulation-digest-ingestor
description: Expert AI Orchestrator for regulatory source monitoring. Fetches updates from official bodies and pre-processes them for the policy-schema-generator.
---

# Regulation Digest Ingestor (v1.0)

## Purpose

The Regulation Digest Ingestor is the "Front-End" of the compliance pipeline. It automates the collection of regulatory updates (e.g., from the EU Journal, HIPAA updates, or CISO internal memos), transforming them into a clean `regulation_digest` string for downstream analysis.

## Preconditions

- **Input**: A `source_url` (Public/Internal) or `raw_artifact` (PDF/DOCX/Email).
- **Access**: Internet access (for public sources) or repository access (for internal memos).
- **Tools**: `read_url_content`, `view_file`.

## Inputs / Outputs

### Inputs
- `source_url` (String, Optional): URL of the regulatory update.
- `target_artifact` (Path, Optional): Local file containing the update.

### Outputs
- `regulation_digest` (Artifact): A sanitized markdown file containing the raw regulatory text.
- `ingestion_metadata` (JSON): Source, timestamp, and relevance score.

---

## Instructions

### Step 1 – Source Acquisition
1.  **Public Sources**: Use `read_url_content` to fetch the markdown/HTML update.
2.  **Internal Sources**: Use `view_file` to read the CISO's memo or regulation PDF.
3.  **Halt on Fail**: If the source is unreachable or unreadable, retry once or fail with `SOURCE_UNAVAILABLE`.

### Step 2 – Content Sanitization
1.  Remove boilerplate (nav bars, legal footers, formatting noise).
2.  Preserve key legal citations (Articles, Sections, Requirements).
3.  Format as a single "Regulation Digest" markdown artifact.

### Step 3 – Relevance Scoring
1.  Determine if the update affects Ocultar's protected categories (SSN, PHI, etc.).
2.  Assign `relevance`:
    - `CRITICAL`: Directly impacts `regulatory_policy.json` mappings.
    - `LOW`: Informational or procedural.

---

## Failure Handling

- **`MALFORMED_SOURCE`**: If the source contains encrypted data or unreadable characters.
- **`NO_DATA_FETCHED`**: If the URL returns a 404 or empty content.

---

## Ecosystem Role
- **Role**: Ingestor / Crawler.
- **Target**: `regulatory-intent-decoder`.
- **Trigger**: Chron task or manual URL submission.
