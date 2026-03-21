---
name: change-impact-visualizer
description: Deterministic analysis of code changes to map architectural, security, and regulatory impacts. Step 12 of the Ocultar Protocol.
---

# Change Impact Visualizer (v1.1)

## Purpose

The CIV (Step 12) prevents "Security Drift" during refactoring. It analyzes the `git diff` to identify ripple effects across the Ocultar tiers and ensures that documentation remains a "Live" representation of the system.

## Inputs / Outputs

### Inputs
- `git_diff`: STAGED changes.
- `protocol_phase`: Current step in the 16-step protocol.

### Outputs
- `impact_matrix` (Artifact): Logical mapping of File -> Component -> Regulatory Impact.
- `drift_verdict`: `SYNCED` | `DOC_DRIFT_DETECTED`.

---

## Instructions

### 1. Component Attribution
- Map every changed file to its core component:
    - `pkg/engine` -> **Refinery Tier**.
    - `pkg/vault` -> **Identity Vault**.
    - `apps/sombra` -> **Gateway Layer**.

### 2. Regulatory Impact Check
- For changes in **Vault** or **Refinery**:
    - Identify if a PII category logic is modified.
    - **Logic**: Document if the "Zero-Egress" promise for that category is affected (Recall/Precision change).

### 3. Doc-Parity Validation
- Verify if `PII_DETECTION.md` or `ARCHITECTURE.md` requires an update based on the logic change.
- **Gate**: If a schema changed but `ARCHITECTURE.md` was not updated, return `DOC_DRIFT_DETECTED`.

## Failure Handling
- **`MONOLITHIC_DIFF`**: If >30 files changed, group impacts by directory and flag for "Package Isolation" review.

## Postconditions
- The matrix MUST be included in the PR description as the "Impact Manifest".
