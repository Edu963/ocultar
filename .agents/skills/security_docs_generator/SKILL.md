---
name: security_docs_generator
description: Expert AI orchestrator for PII compliance and auditor-ready documentation.
---

# PII Security Docs Generator (v1.1)

## Purpose

The PSDG automates the generation of verifiable compliance documentation. It ensures that the current state of the Ocultar Engine (rules, filters, recall) is accurately represented for SOC2/HIPAA auditors.

## Inputs / Outputs

### Inputs
- `regression_metrics`: Accuracy data from `pii-regression-suite`.
- `policy_snapshot`: Active rules from `regulatory_policy.json`.
- `audit_summary`: Hit counts from `compliance-integrity-suite`.

### Outputs
- `security_assessment.md`: Top-level summary for the dashboard.
- `auditor_report.pdf`: Verifiable technical deep-dive.

## Preconditions
- All input artifacts MUST be cryptographically signed by `artifact-signer`.

---

## Instructions

### 1. Accuracy Verification
- Ingest `regression_metrics`.
- **Logic**: Document the current "Recall Rate" for Tier 0 entities (SSN, Keys). 
- **Disclaimer**: Explicitly list any "Heuristic Fallbacks" that are active.

### 2. Policy Mapping
- Cross-reference `policy_snapshot` against the `industry-compliance-validator` standards.
- Generate a "Regulatory Coverage Matrix" (Mapped to GDPR Art 32, HIPAA §164.306).

### 3. Drift Analysis
- Compare the current engine state against the previous `auditor_report`.
- Highlight "Hardening Events" (New regex rules, expanded dictionary).

## Failure Handling
- **`UNSIGNED_INPUT`**: If input metrics lack a valid signature, flag a "Doc-Integrity Risk" and request re-validation.

## Postconditions
- Final documents MUST be saved to `docs/compliance/` and registered in `ecosystem-state-tracker`.
