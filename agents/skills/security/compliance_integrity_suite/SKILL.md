---
name: compliance-integrity-suite
description: Expert AI Auditor for Ocultar compliance. Combines static configuration drift detection with runtime audit log validation.
---

# Compliance Integrity Suite (v1.1)

## Purpose

CIS is the authoritative auditor. It reconciles intended policy, active architecture, and observed behavior to detect drift, tampering, or security evasions.

## Inputs / Outputs

### Inputs
- `policy_path`: Path to `regulatory_policy.json`.
- `config_root`: Sombra config directory.
- `audit_log_path`: System audit logs.

### Outputs
- `integrity_score` (Int): 0-100.
- `verdict` (Enum): `AUDIT_PASS` | `AUDIT_FAIL` | `TAMPER_DETECTED`.
- `findings`: Detailed JSON of discrepancies.

## Preconditions
- Logs MUST have been signed by the Sombra logging service.
- `compliance-certificate-signer` MUST be available to verify policy integrity.

---

## Instructions

### 1. Static Integrity (Configuration)
- Verify that the `regulatory_policy.json` contains a valid signature (via `compliance-certificate-signer`).
- Map each category in the policy to a corresponding block/mask rule in `sombra.yaml`.
- **Deduction**: -20 pts if a policy category has no technical enforcement.

### 2. Runtime Integrity (Audit Logs)
- Check sequence IDs in logs for continuity.
- **Cross-check**: Ensure that `hit_category` entries in the logs match the active policy.
- **Validation**: If a category is marked `STRIP` in config but appears unmasked in logs, return `AUDIT_FAIL`.

### 3. Cross-Domain Reconciliation
- Identify "Ghost Hits": Hits for categories that are NOT in the policy.
- Calculate the "Evasion Index": Ratio of allowed-to-blocked traffic for high-risk categories.

### 4. Verdict & Report
- Final Score: `100 - (StaticGaps * 15) - (RuntimeDrift * 20)`.
- If score < 90 OR `TAMPER_DETECTED`, exit with `AUDIT_FAIL`.

## Failure Handling
- **`TAMPER_ALERT`**: If log hashes don't match or gaps exist, trigger `LOCKDOWN`.
- **`UNREACHABLE_VAULT`**: If the Identity Vault cannot be queried for policy hashes, fail the audit.

## Postconditions
- Output Report MUST be available to the `continuous-ai-orchestrator`.
