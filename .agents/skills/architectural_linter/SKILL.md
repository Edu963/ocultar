---
name: architectural-linter
description: Deterministic validator for Ocultar's architectural integrity. Prevents terminology drift and enforces repository structure compliance.
---

# Architectural Linter

## Purpose
Ensures that all code and documentation conform to the canonical names and structural rules defined in `ecosystem.manifest.json`.

## Inputs / Outputs

### Inputs:
- `target_path` (Path): The file or directory to scan.
- `manifest_path` (Path): Path to `ecosystem.manifest.json`.

### Outputs:
- `lint_report` (Artifact): List of violations and suggested fixes.

## Instructions

### Step 1 – Terminology Scan
- Parse the `ecosystem.manifest.json` for `forbidden_terms`.
- Execute a case-insensitive `grep` on the `target_path` for these terms.
- **Fail Condition**: Any match for a forbidden term must be flagged.

### Step 2 – Structural Validation
- Verify that files are located in their canonical component paths.
- **Example**: A connector must reside in `services/engine/pkg/connector/`.

### Step 3 – Fix Generation
- For every violation, provide a `sed` or `mv` command to remediate.

## Failure Handling
- If linting fails, the `ai-development-protocol` **MUST** be blocked.
