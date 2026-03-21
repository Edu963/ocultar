---
name: industry-snapshot-generator
description: Programmatically provisions the Ocultar stack (Refinery, Shield, Dashboard) for specific verticals.
---

# Industry Snapshot Generator (v1.1)

## Purpose

The ISG automates the vertical-specific configuration of Ocultar. It ensures a "Day-0" high-fidelity posture for regulated industries by orchestrating rule injection and dashboard provisioning.

## Inputs / Outputs

### Inputs
- `industry_id`: `FINANCE` | `HEALTHCARE` | `LEGAL`.
- `target_env`: `DEMO` | `STAGING` | `PRODUCTION`.
- `mfa_token` (Optional): Required for `PRODUCTION`.

### Outputs
- `provisioning_id`: Unique trace for the event.
- `compliance_report`: Summary of rules applied.
- `verdict`: `SUCCESS` | `REVERTED`.

## Preconditions
- All target skills (`dictionary-shield-manager`, `refinery-architecture-manager`) must be reachable.

---

## Instructions

### 1. Transactional Provisioning
- **Step A**: Load industry snapshot JSON.
- **Step B**: Sync Dictionary Shield (VIPs, SEC-terms).
- **Step C**: Inject Regulatory Policies (Regex, SLM).
- **Validation**: Trigger `pii-regression-suite-runner`.

### 2. Atomic Rollback
- If Step C or Validation fails:
    - **Action**: Restore `regulatory_policy.json` and `protected_entities.json` from the pre-provisioning backup.
    - **Verdict**: Return `REVERTED`.

### 3. Production Hardening
- If `target_env` == `PRODUCTION`:
    - **MFA Gate**: Require `mfa_token` before applying policy changes.
    - **Audit**: Log the event as a "Production-State Change" in the compliance suite.

## Failure Handling
- **`DEPENDENCY_MISSING`**: If the Dictionary Shield Manager fails to sync, abort the entire transaction.

## Postconditions
- System state MUST be captured in a "System Health" snapshot post-provisioning.
