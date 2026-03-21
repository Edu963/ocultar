---
name: compliance-integrity-suite
description: Expert AI Auditor for Ocultar compliance. Combines static configuration drift detection with runtime audit log validation to ensure end-to-end regulatory alignment.
---

# Compliance Integrity Suite (v1.0)

## Purpose

The Compliance Integrity Suite is the "Ultimate Truth" mechanism for the Ocultar ecosystem. It ensures that the system is not only *configured* correctly (Static Audit) but is also *performing* correctly (Runtime Audit). It identifies discrepancies between the intended policy, the active gateway settings, and the actual traffic observed in the logs.

## Preconditions

- **Inputs**: Access to `regulatory_policy.json`, `sombra.yaml`, and `audit.log`.
- **Tools**: `jq`, `yq`, `openssl`.

## Inputs / Outputs

### Inputs
- `policy_path` (Path): Path to the global policy.
- `config_root` (Path): Path to the Sombra configuration files.
- `audit_log_path` (Path): Path to the system audit logs.

### Outputs
- `integrity_score` (Int): 0-100 (where 100 is perfect alignment).
- `drift_findings` (JSON): Mismatches between policy and configuration.
- `runtime_violations` (JSON): Observations in the logs that violate the policy.
- `suite_report` (Artifact): A comprehensive audit report for the CISO.

---

## Instructions

### Phase 1 – Static Configuration Audit (Legacy Drift Detector)
1.  **Version Validation**: Ensure `regulatory_policy.json` version matches the system manifest.
2.  **Mapping Coverage**: Verify that every category in the policy is explicitly handled in `sombra.yaml` (STRIP, REDACT, etc.).
3.  **Secret Integrity**: Check for default `OCU_SALT` and `OCU_MASTER_KEY` values in the infrastructure.

### Phase 2 – Runtime Performance Audit (Legacy Audit Log Validator)
1.  **Log Continuity**: Verify sequence IDs in `audit.log` to detect tampering.
2.  **Detection Efficacy**: Cross-reference `hit_category` from logs against the policy mappings.
3.  **Action Verification**: Confirm that "Blocked" events in the log were actually blocked by the gateway logic.

### Phase 3 – Cross-Domain Reconciliation (New)
1.  **Shadow Category Detection**: Identify hits in the `audit.log` that have NO mapping in the `regulatory_policy.json`.
2.  **Evasion Check**: Look for patterns where a category is blocked in config but allowed in logs (indicates a bypass or misconfiguration).
3.  **Statistical Baseline**: Calculate the "Sanitize-to-Block" ratio and flag sudden spikes as potential security incidents.

### Phase 4 – Integrity Scoring
1.  Score the deployment:
    - `-20 points` per Critical Drift (Secret/Policy gap).
    - `-10 points` per Log Tampering evidence.
    - `-5 points` per Unmapped Hit Category.
2.  Generate the `suite_report`.

---

## Failure Handling

- **`MALFORMED_INPUTS`**: If any of the required files are unreadable.
- **`CRITICAL_TAMPERING`**: If log sequence gaps are detected, escalate to `PILOT_OPS` and lock the release.

---

## Ecosystem Role
- **Role**: Auditor / Validator (Level 2).
- **Triggers**: Commit, PR, or Weekly Cron.
- **Dependencies**: `repository-knowledge-map`, `compliance-certificate-signer`.
