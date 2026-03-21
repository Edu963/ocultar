---
name: pii-regression-suite-runner
description: Formalizes the verification of refinery rules to prevent regression in sensitive data detection. Executes automated test cases against the Ocultar Engine and verifies redaction/pseudonymization accuracy.
---

# PII Regression Suite Runner

## Purpose
This skill ensures that updates to PII detection logic, regex patterns, or refinery configurations do not break existing detection capabilities. It acts as a "Quality Gate" for privacy compliance before any changes are merged into the core engine or Sombra Gateway.

## Inputs
- `test_data_path`: Absolute path to the JSON/Text files containing ground-truth PII examples.
- `rule_diff`: The change-set of refinery rules to verify.
- `engine_version`: The binary version or build of the Ocultar Engine to test against.

## Outputs
- `regression_report`: A detailed summary of pass/fail counts, missed hits (FN), and false positives (FP).
- `verification_hash`: A SHA-256 hash of the rule-set + test results.

---

## Instructions

### 1. Load Ground Truth
- Locate the regression test suite in `services/engine/tests/ground_truth/`.
- Ensure test cases cover all supported PII categories (SSN, Email, IBAN, Credentials, etc.).

### 2. Configure Local Engine
- Spin up a local instance of the Ocultar Engine using the updated `refinery_rules.json`.
- Ensure the engine is running in `DEBUG_REFINERY` mode to capture detailed hit metadata.

### 3. Execute Suite
- Stream the test data through the engine's `/refine` endpoint.
- Compare the output tokens against the expected ground truth markers.
- **Rule**: 100% accuracy required for "Critical" categories (SSN, Passport, Secret Keys).

### 4. Analysis & Result
- Calculate detection metrics (Precision/Recall).
- Identify any regression where a previously caught token is now missed.
- **Gate**: Any regression in "Tier 0" or "Tier 1" data MUST result in a `FAIL`.

## Failure Handling
- **Missing Ground Truth**: If test data is missing, halt execution and alert the Privacy Officer.
- **Engine Crash**: If the engine fails under the test load, report a "Stability Regression".

## Examples

### Scenario: New SSN Regex
- **Input**: Updated regex for US SSN with dashes.
- **Process**: Run the suite against 10,000 anonymized samples.
- **Result**: Verify that the new regex catches all 10,000 samples without increasing false positives in serial numbers.
