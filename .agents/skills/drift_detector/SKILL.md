---
name: drift-detector
description: Expert monitor for Ocultar configuration integrity. Detects discrepancies between the active Sombra Proxy settings and the global regulatory policy.
---

# Drift Detector (v2.0)

## Purpose

The Drift Detector prevents "Security Drift" by ensuring that the active Sombra Gateway configuration and the Dictionary Shield are perfectly aligned with the Global Regulatory Policy. It serves as a declarative gate to prevent the exposure of PII through unprotected or misconfigured endpoints.

## Preconditions

- **Tools**: `yq` (v4+) and `jq` must be available in the environment for configuration parsing.
- **Access**: Read access to the repository root, including `.env` or equivalent environment manifests.

## Inputs / Outputs

### Inputs
- `config_root` (Path): Root directory for configuration files (default: `apps/sombra/configs`).
- `policy_path` (Path): Path to the `regulatory_policy.json` (default: `security/regulatory_policy.json`).
- `proxy_manifest` (Path): Path to the `docker-compose.proxy.yml` (default: `./docker-compose.proxy.yml`).

### Outputs
- `drift_analysis_report` (Artifact): A structured JSON report containing:
    - `status`: `CLEAN` | `DRIFT_DETECTED` | `CRITICAL_FAILURE`
    - `drift_types`: List of detected drift categories (Policy, Secret, Version, Entity).
    - `mismatches`: Array of specific field/value discrepancies.
- `alert_level` (Enum): `NONE` | `LOW` | `CRITICAL`.

---

## Instructions

### Step 1 – Validation & Discovery
1.  **Schema Check**: Validate that `regulatory_policy.json` and `sombra.yaml` are syntactically correct.
2.  **Version Alignment**: Extract the `version` from `regulatory_policy.json`. Cross-reference with the `OCU_VERSION` or the latest entry in `CHANGELOG.md`. 
    - **Drift**: If the policy version is older than the system version, flag as **Version Drift**.

### Step 2 – Entity Alignment (Dictionary Shield)
1.  Load `protected_entities.json` from the `config_root`.
2.  Extract the list of "Mandatory Protected Entities" (VIPs, Internal Projects).
3.  **Entity Drift Check**: Verify that every entity in the JSON list is represented in the Sombra Engine's active dictionary rules (via `REDACT` or `VAULT` logic in `sombra.yaml`).

### Step 3 – Policy Coverage Analysis
1.  **Extract Protection Requirements**: Parse `regulatory_policy.json` to identify all sensitive categories (e.g., `SSN`, `CREDIT_CARD`).
2.  **Verify Sombra Connectors**: For every connector in `sombra.yaml`:
    - Check if `strip_categories` includes ALL categories defined in the policy mappings.
    - **Policy Drift**: If a category exists in the policy but is NOT stripped or managed by a connector, flag as **CRITICAL**.
3.  **Algorithm Audit**: Ensure tokenization/masking algorithms meet the regulation standards (e.g., PCI_DSS requires specific hashing/encryption).

### Step 4 – Infrastructure Secret Audit
1.  Extract `OCU_SALT` and `OCU_MASTER_KEY` values (or their defaults) from `proxy_manifest`.
2.  **Secret Drift Check**:
    - **CRITICAL**: If `OCU_SALT` matches `ocultar-v112-kdf-salt-fixed-16`.
    - **CRITICAL**: If `OCU_MASTER_KEY` matches `default-dev-key-32-chars-long-!!!`.
    - **LOW**: If keys are present but haven't been rotated in > 90 days (check file modification metadata).

### Step 5 – Report Generation
1.  Consolidate findings into `drift_analysis_report.json`.
2.  Set `alert_level`:
    - `CRITICAL`: Any Secret Drift or Policy Coverage Gap.
    - `LOW`: Version Drift or Metadata Mismatch.
    - `NONE`: Full alignment.

---

## Failure Handling

- **Missing Dependency**: If a required config file is missing, fail-closed: `alert_level: CRITICAL`.
- **Parser Error**: If YAML/JSON parsing fails, exit with `CRITICAL_FAILURE` and provide the line number of the error.

## Validation Steps for Agent

To verify execution, the agent should:
```bash
# Verify Policy Schema
jq '.' security/regulatory_policy.json > /dev/null

# Verify Sombra Config
yq eval '.' apps/sombra/configs/sombra.yaml > /dev/null
```

---

## Ecosystem Role
- **Category**: Validator / Auditor.
- **Dependencies**: `security-advisory-scanner`, `repository-knowledge-map`.
- **Triggers**: `CISO & Compliance Officer` audit workflow.
