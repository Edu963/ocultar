---
name: industry-snapshot-generator
description: Programmatically provisions the Ocultar stack (Refinery, Shield, Dashboard) for specific verticals (Finance, Healthcare, GovTech, etc.).
---

# Industry Snapshot Generator

## Purpose
Automates the configuration of the Ocultar ecosystem for vertical-specific compliance requirements. This skill ensures that a prospect or new client starts with a high-fidelity, industry-standard security posture.

## Role
- **Category**: Provisioner / Orchestrator
- **Environment**: Staging / Demo / Initial Onboarding
- **Lifecycle**: Pre-Sales, Deployment

## Inputs & Preconditions

| Input | Type | Description |
| :--- | :--- | :--- |
| `INDUSTRY_ID` | String | `finance`, `healthcare`, `govtech`, `legal` |
| `TARGET_ENV` | Enum | Must be `DEMO` or `STAGING`. `PRODUCTION` requires MFA. |
| `MOCK_DATA` | Boolean | Whether to generate ROI/Risk data for the dashboard. |

### Preconditions
1. Ocultar Engine and Sombra Gateway are reachable.
2. `security/regulatory_policy.json` is initialized.
3. `pii-regression-suite-runner` is available for validation.

## Instructions

### 1. Load Declarative Snapshot
- Retrieve the JSON payload from `/configs/snapshots/[INDUSTRY_ID].json`.
- **Validation**: If the file does not exist, fail and report "Unsupported Industry".

### 2. Infuse Dictionary Shield
- Extract the `dictionary_entities` list from the snapshot.
- Delegate to `Ocultar | Dictionary Shield Manager`:
  - Command: `UpdateProtectedList(entities)`
  - Goal: Synchronize industry-specific terms.

### 3. Update Policy Posture
- Read PII mappings from the snapshot (regex and SLM classifiers).
- Update `security/regulatory_policy.json` with the industry-specific mappings.
- Delegate to `Ocultar | Refinery Architecture Manager`:
  - Command: `VerifyRuleIntegrity(rules)`
  - Trigger `sombra-gateway-policy-enforcer` to apply `strip_categories`.

### 4. Direct Compliance Mapping
- Delegate to `Ocultar | Industry Compliance Validator`:
  - Command: `ValidateIndustryPosture(regulatory_policy.json)`
  - Goal: Generate `COMPLIANCE_AUDIT.md`.

### 5. Provision Dashboard Analytics
- If `MOCK_DATA` is true:
  - Delegate to `Ocultar | Dashboard Scenario Generator`:
    - Command: `TriggerScenario(INDUSTRY_ID)`
    - Goal: Update `Mock-API` state and ROI metrics.

### 5. Verification & Health Check
- Run `pii-regression-suite-runner` against the newly applied rules.
- Verify Sombra Gateway connectivity via `sombra-performance-benchmarker`.

## Failure Handling
- **Invalid Snapshot**: Logs error and aborts before any file modification.
- **Validation Failure**: If regression tests fail, revert `regulatory_policy.json` to the previous version from Git/Backup.

## Postconditions
1. `regulatory_policy.json` contains industry-specific mappings.
2. Dictionary Shield is populated with vertical-specific terms.
3. Dashboard reflects realistic risk/ROI data for the vertical.
