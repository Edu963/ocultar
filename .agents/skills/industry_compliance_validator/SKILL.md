---
name: industry-compliance-validator
description: Regulatory Auditor for OCULTAR. Maps active security rules to global compliance standards (GDPR, HIPAA, PCI-DSS).
---

# Industry Compliance Validator

## Purpose
Bridges the gap between technical regex/dictionary rules and legal regulatory requirements. This skill ensures that an OCULTAR deployment is not just "secure," but "compliant" for its specific industry.

## Role
- **Category**: Validator / Auditor
- **Ecosystem Positioning**: Governance & Compliance

## Instructions

### 1. Regulatory Mapping
- Analyze the active rules in `regulatory_policy.json` and `sombra.yaml`.
- Compare them against the required mappings for a specific framework:
  - **GDPR**: Are `EMAIL`, `PHONE`, and `SSN` mapped to Art. 4(1)?
  - **HIPAA**: Are `PATIENT_ID` and `MEDICAL_RECORD` mapped to Sec. 164.514(b)?
  - **PCI-DSS**: Is `CREDIT_CARD` set to `FAIL_CLOSED`?

### 2. Gap Analysis
- Identify "Shadow PII": Data types that are detected but NOT mapped to a regulatory article.
- Identify "Vulnerable Articles": Regulatory requirements that have no corresponding detection rule in the engine.

### 3. Reporting
- Generate a `COMPLIANCE_AUDIT.md` summary for the Enterprise Dashboard.
- Provide a "Readiness Score" (0-100%) for the target industry.

## Constraints
- Do not make legal advice; provide technical alignment to known regulatory schemas.
- Ensure all mappings are based on the latest version of `regulatory_policy.json`.
