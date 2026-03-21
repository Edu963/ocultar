---
name: audit-log-validator
description: Expert validator for AI audit log integrity and regulatory compliance. Enforces deterministic mapping between detection hits and global privacy standards (GDPR, HIPAA, ISO 27001).
---

# Audit Log Validator (v2.0)

## Purpose

The Audit Log Validator ensures that every security event recorded by the OCULTAR engine is accurately classified according to global regulatory frameworks. It transforms raw log data into a verified compliance manifest, identifying gaps in detection coverage or misclassified PI (Personal Information).

## Prerequisites

- **Log Format**: The audit log must be in structured JSON or JSONL format (`audit.log`).
- **Policy Context**: The `CISO & Compliance Officer` skill or a `regulatory_policy.json` must be available for mapping.
- **Fail-Closed Logic**: Access to the Sombra Gateway configuration is required to verify bypass prevention.

## Inputs / Outputs

### Inputs:
- `audit_log_path` (Path): Path to the current audit log.
- `target_regulations` (List): Array of regulations to validate against (e.g., `["GDPR", "HIPAA"]`).
- `policy_map` (JSON): A dictionary mapping hit categories (e.g., `PERSON_VIP`, `SSN`) to regulatory requirements.

### Outputs:
- `compliance_audit_report` (Artifact): A structured summary of verified hits.
- `exceptions_list` (List): All hits identified without a valid regulatory mapping.
- `tamper_check_status` (Boolean): Verification of log continuity and sequence.

---

## Instructions

### Step 1 – Log Ingestion & Sequence Check
- Load the audit log from `audit_log_path`.
- Verify timestamp continuity and monotonic sequence IDs to ensure no log entries were deleted or suppressed.
- **Gate**: If a gap in sequence IDs is detected, trigger a **Security Alert** for potential log tampering.

### Step 2 – Detection Hit Extraction
- Extract every detection event (`hit_category`, `hit_value_masked`, `action_taken`).
- Verify that every "Blocked" action includes `reason: "Policy Violation"` and `fail_closed: true` metadata.

### Step 3 – Regulatory Reconciliation
- Cross-reference each `hit_category` against the `policy_map`.
- **Validation Rules**:
  - `PERSON_VIP` -> Must map to `INTERNAL_PII_PROTECTION`.
  - `SSN` / `CREDIT_CARD` -> Must map to `GDPR_ART_6` and `PCI_DSS`.
  - `HEALTH_CONDITION` -> Must map to `HIPAA_TECHNICAL_SAFEGUARDS`.
- **Gate**: Any hit without a defined regulatory mapping must be moved to the `exceptions_list`.

### Step 4 – Fail-Closed Verification
- Compare "Blocked" events in the audit log against the Sombra Proxy's active ruleset.
- Confirm that the `Fail-Closed` logic successfully dropped the request before it reached the LLM.

### Step 5 – Compliance Report Generation
- Consolidate findings into a `compliance_audit_report`.
- Include a summary of:
  - Total Hits Protected
  - Blocked vs. Sanitized ratio
  - Regulatory Coverage % (Hits mapped vs. Total hits)

---

## Failure Handling

- **Malformed Log**: If the log cannot be parsed as JSON, halt and request a log rotation/repair.
- **Mapping Gap**: If >10% of hits are unmapped, flag a **Significant Compliance Risk** and escalate to the `Chief Revenue Officer` for roadmap update.

---

## Examples

### Example 1: Standard Audit
- **Input**: `audit.log` containing 500 hits; target `GDPR`.
- **Action**: Verify all 500 hits have GDPR-compliant mappings; check for "Fail-Closed" signatures on blocked requests.

### Example 2: New Hit Category Discovery
- **Input**: Log contains `US_PASSPORT` which is not in the current `policy_map`.
- **Action**: Log an exception, notify the `PII Detection Updater`, and complete the report for other categories.

### Example 3: Tamper Detection
- **Input**: Log sequence jumps from ID 1024 to 1050.
- **Action**: Halt validation; report potential unauthorized log deletion.
