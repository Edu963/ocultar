---
name: ai-development-protocol
description: Production-grade orchestrator for the AI agent development lifecycle. Ensures repository consistency, security integrity, and release readiness through a deterministic pipeline of documentation, packaging, and sanitization gates.
---

# AI Development Protocol (v2.0)

## Purpose

The AI Development Protocol is the authoritative orchestrator for the project lifecycle. It transforms standard development activity into a production-ready state by enforcing strict validation gates and coordinating specialized skills. This protocol ensures that every "meaningful change" is documented, tested, sanitized, and optionally packaged without manual oversight or ambiguity.

## Prerequisites

Before executing this protocol, the following conditions **MUST** be met:
- **Git Status**: Working directory must be clean (no uncommitted changes).
- **Build Status**: The current codebase must pass all unit tests and linting.
- **Branch Policy**: Releases are only permitted from `main`, `stable`, or designated `release/*` branches.

## Inputs / Outputs

### Inputs:
- `project_root` (Path): The absolute path to the repository.
- `change_set` (Git Diff/Manifest): A summary of files modified in the current iteration.
- `release_intent` (Boolean): Whether the current change is intended for a versioned release or a milestone.

### Outputs:
- `validation_report` (Artifact): A report detailing the outcome of documentation, packaging, and security checks.
- `release_manifest` (JSON): (Optional) Metadata for the generated release artifact.

---

## Instructions

### Step 1 – Vision Oracle Alignment
**Dependency**: `Ocultar | Product Context`
- Before any change analysis, the AI agent **MUST** present the `proposal` to the `Product Context` skill.
- **Gate**: If the `alignment_report` returns a **REJECT** verdict, the protocol **MUST** halt immediately.
- **Validation**: Analyze the `alignment_report` for specific guardrail violations.

### Step 2 – Change Impact Analysis
**Dependency**: `change-impact-visualizer`
- Execute the `change-impact-visualizer` to map modified files to their functional domains.
- **Gate**: If the impact analysis identifies undocumented API changes, immediately flag a failure.

### Step 3 – Automated Documentation Sync
**Dependency**: `Ocultar | Documentation Updater`
- Identify all files modified in the `change_set`.
- Run the `Documentation Updater` to synchronize `TECH_DOCS.md`, `API_REFERENCE.md`, and any relevant `/documentation/*.md` guides.
- **Validation**: Confirm that the generated documentation reflects the current state of the code.

### Step 4 – Zero-Trust Security Shield
**Dependency**: `security-sanitizer`, `zero-egress-validator`, `architectural-linter`
- **4.1 Architectural Integrity**: Run the `architectural-linter` to check for terminology drift and structural errors.
- **4.2 Leak Detection**: Run the `zero-egress-validator` to ensure no unmasked PII is passed to external URIs.
- **4.3 Sanitization**: Scan the entire repository (or the staging area) for secrets, hardcoded credentials, and internal paths. Run the `security-sanitizer` to replace sensitive values with environment-safe placeholders.
- **Gate**: Failure in ANY security sub-step **MUST** block subsequent actions.

### Step 5 – Client Package Integrity
**Dependency**: `client-package-updater`
- If the `change_set` affects files designated for client distribution (.tar, .zip, .exe):
  - Run the `Client Package Updater` to verify build script alignment and file inclusions.
  - **Validation**: Ensure all new dependencies are accounted for in the packaging rules.

### Step 6 – Release Artifact Generation (Conditional)
**Dependency**: `release-artifact-builder`
- **Condition**: Only execute if `release_intent` is TRUE.
- Run the `Release Artifact Builder` to generate a versioned, reproducible archive.
- Ensure the version naming follows standard SemVer (e.g., `v1.2.3`).

### Step 7 – Post-Flight Verification
- Verify the following assertions:
  1. `alignment_report` verdict is APPROVE.
  2. `security_verdict` is PASS.
  3. `lint_report` is clean.
  4. `git_status` remains clean (unless documentation updates were committed).
  5. `release_artifact` (if generated) is checksum-validated and signature-ready.
  6. `TECH_DOCS.md` includes a reference to the latest changes.

## Failure Handling

- **Step Failure**: If any sub-skill fails or a validation gate is not met, the protocol **MUST** halt immediately and report the specific error.
- **Recovery**: Correct the underlying issue (e.g., fix a test, add a doc entry) and restart the protocol from Step 1.

---

## Examples

### Example 1: API Enhancement (Non-Release)
- **Input**: `change_set` includes internal API changes; `release_intent` = FALSE.
- **Action**: Run Step 1 (Impact), Step 2 (Docs), and Step 4 (Security). Skip Step 3 and 5.

### Example 2: Major Milestone (Release Candidate)
- **Input**: `change_set` includes critical feature completion; `release_intent` = TRUE.
- **Action**: Run the **Full Protocol (Steps 1-6)**. Generate a versioned tarball and update all compliance docs.

### Example 3: Bug Fix (Emergency Patch)
- **Input**: One-line fix in logic; `release_intent` = TRUE.
- **Action**: Run the **Full Protocol**. Ensure the patch is sanitized and the version is incremented appropriately.
