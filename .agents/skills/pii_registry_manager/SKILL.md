---
description: Expert AI orchestrator for maintaining the Ocultar PII Central Registry.
---

# PII Registry Manager (Skill)

Expert orchestrator for maintaining the Ocultar PII Central Registry (`internal/pii/registry.go`). Validates new regex additions, ensures deterministic compliance, and enforces presence of checksum validators.

## Core Capabilities

- **Regex Quality Validation**: Ensures all patterns use word boundaries (`\b`) and handle separators correctly.
- **Normalization Enforcement**: Verifies that entities requiring validation are marked for normalization.
- **Validator Provisioning**: Detects missing validators for new entities and enforces their implementation in `validators.go`.
- **Conflict Detection**: Identifies overlapping or duplicate patterns in the registry.

## Workflow

### 1. Registry Audit
Scan `internal/pii/registry.go` for architectural violations.
- Check for `\b` boundaries.
- Verify `MinLength` is realistic.
- Ensure `Normalization` is true for validated entities.

### 2. Validator Sync
Cross-reference `registry.go` with `validators.go`.
- Every `EntityDef` with a `Validator` != `ValNone` must have a corresponding function in `validators.go`.
- Ensure `GetValidator` map is updated.

### 3. Pipeline Verification
Run regression tests to ensure no regressions in base detection.
- Execute `pii_regression_suite_runner`.

## Constraints

- NO hardcoded regex outside `registry.go`.
- ALL validators must be deterministic (no external APIs).
- Fail-Closed design: if a validator exists, it MUST pass for the entity to be detected.
