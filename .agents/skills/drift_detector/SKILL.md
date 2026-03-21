---
name: drift-detector
description: Expert monitor for Ocultar configuration integrity. Detects discrepancies between the active Sombra Proxy settings and the global regulatory policy.
---

# Drift Detector (v1.0)

## Purpose

The Drift Detector prevents "Security Drift" where new AI-integrated agents or endpoints are deployed without the necessary PII protection rules. It ensures the environment remains in a "Known-Good" compliant state.

## Prerequisites

- **Gateway Config**: Access to `apps/sombra/configs/sombra.yaml`.
- **Regulatory Policy**: Access to `security/regulatory_policy.json`.
- **Proxy Manifest**: Access to `docker-compose.proxy.yml` for upstream mappings.

## Inputs / Outputs

### Inputs:
- `current_config_path` (Path): Path to the active Sombra/Proxy configuration.
- `policy_source_path` (Path): Path to the `regulatory_policy.json` truth source.

### Outputs:
- `drift_analysis_report` (Artifact): Summary of mismatches.
- `unprotected_endpoints` (List): Routes detected without active PII filters.
- `alert_level` (Enum): `NONE`, `LOW` (Metadata drift), `CRITICAL` (Endpoint coverage gap).

---

## Instructions

### Step 1 – Schema Alignment
- Load the `regulatory_policy.json` to identify the "Required Protection Set" (All categories that MUST be audited/masked).
- Load the `sombra.yaml` and `docker-compose.proxy.yml` to identify the "Active Protection Set" (All endpoints and their assigned filters).

### Step 2 – Endpoint Discovery
- Extract all `OCU_PROXY_TARGET` and route mappings from the proxy configuration.
- Identify all `protected_entities` registered in Sombra.

### Step 3 – Policy Gap Analysis
- For every active endpoint, verify that the assigned protection tier covers the required regulations defined in the policy source.
- **Drift Check**:
    - Is an endpoint missing a `REDACT` or `VAULT` action for a category marked as `MANDATORY` in the policy?
    - Are the tokenization algorithms in use weaker than those required by the regulation (e.g., MD5 vs. SHA-256)?

### Step 4 – Metadata Integrity Check
- Verify that `OCU_SALT` and `OCU_MASTER_KEY` environment variables are present and consistent across the stack.
- **Gate**: If `OCU_SALT` is using the "built-in default salt", flag as **Critical Security Drift**.

### Step 5 – Alert Generation
- Consolidate findings into the `drift_analysis_report`.
- **Action Required**: If `alert_level == CRITICAL`, notify the `CISO & Compliance Officer` to halt deployment.

---

## Failure Handling

- **Missing Policy**: If `regulatory_policy.json` is missing, fail-closed and assume maximum drift.
- **Config Syntax Error**: If YAML cannot be parsed, halt and request a config audit.

---

## Examples

### Example 1: Clean State
- **Input**: Sombra YAML matches Policy JSON; all endpoints have GDPR-compliant filters.
- **Result**: `alert_level: NONE`.

### Example 2: Uncovered Endpoint
- **Input**: A new `/v1/chat/completions/internal` endpoint was added but not registered in the Sombra policy.
- **Result**: `alert_level: CRITICAL`. Unprotected route detected.
