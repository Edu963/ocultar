---
name: tier_compliance_checker
description: Cross-references the active OCULTAR license with the Sombra gateway policy. Ensures Pro features are only enabled when backed by a valid Enterprise license.
---

# Tier Compliance Checker (v1.1)

## Purpose

The TCC prevents "Feature Drift". It ensures that the Sombra Gateway configuration is technically aligned with the active license capabilities, preventing unmanaged usage of Pro Connectors (Slack, SharePoint).

## Inputs / Outputs

### Inputs
- `license_token`: The active signed license.
- `sombra_config_path`: Path to `sombra.yaml`.

### Outputs
- `compliance_verdict` (Enum): `COMPLIANT` | `VIOLATION` | `EXPIRY_RISK`.
- `violations` (List): Description of unlicensed features currently enabled.
- `active_tier`: `COMMUNITY` | `ENTERPRISE`.

## Preconditions
- `license_validation_cli` MUST be run first to ensure token authenticity.

---

## Instructions

### 1. Extract Capabilities
- Decode the `Capabilities` bitmask from the `license_token`.
- **Mapping**:
    - Bit 0: Slack Ingestion.
    - Bit 1: SharePoint/Teams Ingestion.
    - Bit 2: Custom SLM Refinement (Tier 2).

### 2. Config Audit
- Parse `sombra.yaml`.
- Search for active connectors and their types.
- **Reconciliation**:
    - If `connector.type == "slack"` and Bit 0 is `0`, record violation.
    - If `connector.type == "sharepoint"` and Bit 1 is `0`, record violation.

### 3. Verdict Generation
- If `len(violations) > 0`, return `VIOLATION`.
- **Remediation**: Append required `sed` commands to the output to disable the violating connectors.

## Failure Handling
- **`MISSING_CONFIG`**: If `sombra.yaml` is not found, assume `VIOLATION` (Fail-Closed).
- **`TIER_MISMATCH`**: If `tier == "community"` but enterprise connectors are present, return `CRITICAL_VIOLATION`.

## Postconditions
- Results MUST be recorded in the `ecosystem-state-tracker`.
