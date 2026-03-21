---
name: ai-development-protocol
description: Production-grade orchestrator for the AI agent development lifecycle. Ensures repository consistency, security integrity, and release readiness through a deterministic pipeline.
---

# AI Development Protocol (v2.1)

## Purpose

The AI Development Protocol is the authoritative orchestrator for the project lifecycle. It transforms development activity into a production-ready state by enforcing strict validation gates and coordinating specialized skills for documentation, security, and packaging.

## Inputs / Outputs

### Inputs:
- `project_root` (Path): Absolute path to the repository.
- `change_set` (Git Diff): Summary of modifications.
- `release_intent` (Boolean): Defines if packaging/versioning is required.

### Outputs:
- `protocol_report` (Artifact): Summary of all gate outcomes.
- `release_manifest` (JSON): Metadata for the generated release (if applicable).
- `verdict` (Enum): `PROCEED` | `HALT`.

## Preconditions
- **Branch Eligibility**: Current branch must match `main`, `stable`, or `release/*` if `release_intent` is TRUE.
- **Git State**: Clean workspace required (`git status --porcelain` is empty).
- **Environment**: All required environment variables (`OCULTAR_KEY`, etc.) MUST be present.

---

## Instructions

### Step 1 — Product Strategy Alignment
**Dependency**: `Ocultar | Product Context`
- Submit the `change_set` and `intent` to the `Product Context` skill.
- **Gate**: If alignment is `REJECTED`, the protocol MUST halt.

### Step 2 — Impact Mapping
**Dependency**: `change-impact-visualizer`
- Map modified files to functional domains (Engine, UI, Policy, Docs).
- **Validation**: If changes affect `services/engine/` but no test updates are present, flag as `WARNING`.

### Step 3 — Integrated Governance Sync
**Dependency**: `compliance-docs-orchestrator`
- Trigger document synchronization for all identified domains.
- **Validation**: Ensure `TECH_DOCS.md` and `API_REFERENCE.md` deep hashes are updated in the manifest.

### Step 4 — Security & Integrity Shield
**Dependency**: `secret-scanner`, `zero-egress-validator`, `architectural-linter`
- **4.1 Linting**: Validate structural and terminology compliance.
- **4.2 Zero-Trust**: Verify no PII egress in new code paths.
- **4.3 Sanitization**: Scan and redact secrets from the staging area.

### Step 5 — Distribution Alignment (Conditional)
**Dependency**: `client-package-updater`, `release-artifact-builder`
- Execute only if `release_intent` is TRUE.
- **Action**: Increment version strings in `package.json`, `VERSION`, and `go.mod`.
- **Validation**: Verify that `client-package-updater` confirms build script parity.

### Step 6 — Final Verification & Receipt
- Aggregate all receipts from previous steps.
- **Postcondition**: `protocol_report` MUST include hashes of all updated files and generated artifacts.

## Failure Handling
- **Soft Fail**: Warnings (e.g., doc gaps) do not halt non-releases but MUST be reported.
- **Hard Fail**: Any security or alignment failure triggers immediate `HALT`.

## Examples

### Example: Security Patch
- **Input**: Fix for SSN masking; `release_intent` = TRUE.
- **Action**: Full protocol run. Ensure `zero-egress-validator` passes specifically for the new fix. 
appropriately.
