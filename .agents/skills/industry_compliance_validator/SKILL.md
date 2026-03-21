---
name: industry-compliance-validator
description: Regulatory Auditor for OCULTAR. Maps active security rules to global compliance standards.
---

# Industry Compliance Validator (v1.1)

## Purpose

Bridges the gap between technical detection rules and legal requirements. This skill ensures that Ocultar deployments are "Audit-Ready" for specific verticals (Finance, Healthcare, Gov).

## Inputs / Outputs

### Inputs
- `industry` (Enum): `FINANCE` | `HEALTHCARE` | `GOVTECH` | `RETAIL`.
- `policy_snapshot`: JSON from `regulatory_policy.json`.
- `sombra_config`: YAML from `sombra.yaml`.

### Outputs
- `readiness_score` (Int): 0-100 percentage.
- `gap_manifest` (JSON): List of requirements missing technical safeguards.
- `compliance_audit`: Markdown artifact summarizing the verdict.

## Preconditions
- `regulatory_policy.json` MUST be the signed version.
- Access to the industry benchmark database.

---

## Instructions

### 1. Regulatory Mapping
Analyze active rules against Framework Benchmarks:
- **GDPR (Universal)**: Check for Art. 4(1) mappings.
- **HIPAA (Healthcare)**: Verify Sec. 164.514(b) for Patient Identifiers.
- **PCI-DSS (Finance)**: Ensure `FAIL_CLOSED` for credit card categories.

### 2. Shadow PII Detection
- Identify detected categories in the engine that lack a regulatory mapping.
- **Weight**: Each "Shadow Category" reduces `readiness_score` by 5 points.

### 3. Gap Analysis & Scoring
- `readiness_score` = 100 - (Gaps * 10) - (ShadowPII * 5).
- If `readiness_score` < 80, output `COMPLIANCE_WATCH` status.

## Failure Handling
- **`MISSING_BENCHMARK`**: If the requested industry benchmark is not found, default to `ISO-27001` generic audit.

## Postconditions
- Audit results MUST be pushed to the Enterprise Dashboard.
