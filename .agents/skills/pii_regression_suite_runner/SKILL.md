---
name: pii-regression-suite-runner
description: Production-grade orchestrator for the PII Regression Suite.
---

# PII Regression Suite Runner (v1.1)

## Purpose

PRSR acts as the ultimate quality gate (Step 11). It ensures that no change to detection rules or engine logic degrades the "Recall" of sensitive data. It enforces a strict "No Regressions" policy for Tier 0 categories.

## Inputs / Outputs

### Inputs
- `test_suite_id`: Dataset ID in `ground_truth/`.
- `rule_set_hash`: SHA-256 of the refinery rules under test.
- `enforcement`: `FAIL_ON_ANY` | `FAIL_ON_TIER_0`.

### Outputs
- `verdict`: `PASS` | `FAIL`.
- `metrics` (JSON): Pass rate, False Negatives (FN), Latency P95.
- `state_id`: SHA-256 key for `ecosystem-state-tracker`.

## Preconditions
- Ephemeral Clean-Room environment (No Internet).
- Mount read-only ground-truth volume.

---

## Instructions

### 1. Integrity Prep
- Verify `refinery-architecture-manager` output hash.
- **Constraint**: Never log raw PII. Use token offsets for reporting.

### 2. Differential Execution
- Run the engine over the ground-truth set.
- Compare hits.
- **Regressions**:
    - **Tier 0 (SSN, Keys)**: 1 missed hit = `FAIL`.
    - **Tier 1 (Phone, Email)**: >1% degradation = `FAIL`.

### 3. State Registration
- Commit results to `ecosystem-state-tracker`.
- Use `state_id` derived from `(engine_version + rule_set_hash)`.

## Failure Handling
- **`ENGINE_CRASH`**: Immediate `FAIL`.
- **`RESOURCE_EXHAUSTED`**: If P95 latency > 200ms, flag as `PERFORMANCE_REGRESSION`.

## Postconditions
- Verdict MUST be signed by `artifact-signer` for release eligibility.
