---
name: pii-regression-suite-runner
description: Production-grade orchestrator for the PII Regression Suite. Validates Ocultar Engine detection accuracy against ground-truth datasets to prevent privacy regressions.
---

# PII Regression Suite Runner

## Purpose
This skill acts as a mandatory Quality Gate (Step 11 of the Ocultar 16-Step Protocol) to ensure that updates to Refinery rules, NLP models, or regex patterns do not degrade detection capabilities. It enforces a "Fail-Closed" policy for regressions in critical (Tier 0/1) data categories.

## Preconditions
1.  **Environment Isolation**: Execution MUST occur in a "Zero-Egress" ephemeral container.
2.  **License Validation**: A valid Ocultar Enterprise license must be active (Verified via `tier-compliance-checker`).
3.  **Artifact Parity**: The `engine_version` and `refinery_rules.json` must be compatible (Verified via `distribution-integrity-validator`).

## Inputs
- `test_suite_id`: [REQUIRED] ID of the ground-truth dataset in `services/engine/tests/ground_truth/`.
- `rule_set_hash`: [REQUIRED] SHA-256 hash of the refinery rules to be tested.
- `engine_binary_path`: [REQUIRED] Absolute path to the Ocultar Engine binary.
- `concurrency_limit`: [OPTIONAL] Max parallel requests to the engine (Default: 4).

## Outputs
- `regression_report`: Detailed JSON containing:
    - `pass_rate`: Percentage of correctly redacted tokens.
    - `fn_list`: List of missed tokens (False Negatives) with their categories.
    - `fp_list`: List of over-redacted tokens (False Positives).
    - `latency_p95`: 95th percentile latency per request.
- `verification_signature`: Ed25519 signature of the report, linked to the `rule_set_hash`.
- `state_token`: UUID registered in the `ecosystem-state-tracker`.

---

## Instructions

### 1. Initialize Clean-Room Environment
- Provision an ephemeral workspace with NO external network access.
- Mount the ground-truth dataset as a READ-ONLY volume.
- **Security Rule**: Never log the actual content of ground-truth PII; only log the token positions and category IDs.

### 2. Verify Engine Integrity
- Execute `ocultar-engine --version` and verify against requirements.
- Run a `Health-Check` on the engine's `/heartbeat` endpoint.
- Ensure `DEBUG_REFINERY=1` is set to enable hit-metadata extraction.

### 3. Execute Batch Processing
- Partition the ground-truth data into chunks based on `concurrency_limit`.
- Iterate through each chunk, sending requests to the `/refine` endpoint.
- **Fail-Fast**: Stop execution if the engine latency exceeds 500ms for more than 5% of requests.

### 4. Differential Analysis (Ground-Truth Mapping)
- Compare actual engine redactions against expected ground-truth markers.
- Classify mismatches into:
    - **Tier A Regression**: Critical missed hit (SSN, Passport, Secret Key) -> **IMMEDIATE FAIL**.
    - **Tier B Regression**: High-sensitivity missed hit (Email, Address) -> **FAIL**.
    - **Functional Drift**: False Positive in non-PII context -> **WARNING**.

### 5. Finalize Verification & State Sync
- Generate the `regression_report`.
- Call `artifact-signer` to sign the report hash.
- Register the results in the `ecosystem-state-tracker` with the current `rule_set_hash` as the key.

## Failure Handling
- **Engine Instability**: If the engine crashes (Signal 9/11), report as "Critical Stability Regression".
- **Schema Mismatch**: If the ground-truth format is invalid, halt and trigger `policy-schema-generator` for update.
- **Security Alert**: If any network egress is detected during the suite run, immediate SHUTDOWN and alert CISO.

## Quality Metadata
- Role: Validator / Gatekeeper
- Lifecycle: CI-CD / Pre-Release
- Maturity: Production
- Zero-Egress: MANDATORY
