---
name: compliance-docs-orchestrator
description: Synchronizes Ocultar technical, API, and compliance documentation. Ensures architectural changes are reflected in reports.
---

# Compliance Docs Orchestrator (v1.1)

## Purpose

CDO ensures the "Documentation-as-Code" principle. It synchronizes technical specifications with security narratives, maintaining auditor-ready documentation across the repository.

## Inputs / Outputs

### Inputs
- `change_set`: Modified code and config files.
- `impact_domain` (Enum): `CORE`, `API`, `PII`, `UI`.

### Outputs
- `doc_manifest` (JSON): Hashes of all updated markdown files.
- `sync_status` (Enum): `COMPLETE` | `PARTIAL` | `OUT_OF_SYNC`.

## Preconditions
- Access to `repository-knowledge-map` for path resolution.
- Read/Write permissions on `/documentation` and `/docs`.

---

## Instructions

### 1. Domain-Locked Sync
Trigger updates based on `impact_domain`:
- **CORE**: Update `TECH_DOCS.md` and `ARCHITECTURE.md`.
- **API**: Synchronize `API_REFERENCE.md` with new routes.
- **PII**: Update `PII_DETECTION.md` and horizontal regulatory mappings.

### 2. Consistency Cross-Check
- Verify that any new environment variable in code is documented in `SETUP_GUIDE.md`.
- Ensure all relative documentation links are valid using `repository-knowledge-map`.

### 3. Manifest Generation
- Generate `doc_manifest.json` containing: `{ "file": string, "sha256": string, "last_reviewed_by_ai": timestamp }`.

## Failure Handling
- **`DOCUMENTATION_DRIFT`**: If documentation is significantly behind code (measured by git history), flag `OUT_OF_SYNC`.
- **`LINK_ROT`**: If internal links are broken, return a `PARTIAL` status with a list of bad links.

## Postconditions
- `doc_manifest` MUST be signed by `evidence-archiver`.
