---
name: policy-schema-generator
description: Expert Instructions for the AI assistant. Automatically updates regulatory_policy.json when new regulations are added to the CISO & Compliance Officer knowledge base.
---

# Policy Schema Generator (v1.0)

## Purpose

The Policy Schema Generator ensures that the Global Regulatory Policy (`regulatory_policy.json`) stays in sync with the latest compliance requirements. It automates the transcription of human-readable regulatory digests (e.g., from the CISO) into the structured JSON format required by the Sombra Gateway and the Engine.

## Preconditions

- **Input**: A "Regulation Digest" or "CISO Instructions" via a knowledge resource or artifact.
- **Access**: Write access to `security/regulatory_policy.json`.
- **Validation**: Access to `jq` for JSON schema enforcement.

## Inputs / Outputs

### Inputs
- `regulation_digest` (Text): The raw text or document describing new compliance rules (e.g., "Add HIPAA Article 17 requirements for PHI").
- `policy_file` (Path): Path to the existing `regulatory_policy.json` (default: `security/regulatory_policy.json`).

### Outputs
- `updated_policy` (Artifact): The new `regulatory_policy.json` with incremented version.
- `change_summary` (String): A human-readable diff of the policy changes.

---

## Instructions

### Step 1 – Requirements Extraction
1.  Analyze the `regulation_digest` to identify the following:
    - **Category**: (e.g., `PHI`, `SSN`, `CREDIT_CARD`).
    - **Regulation**: (e.g., `HIPAA`, `GDPR`, `PCI_DSS`).
    - **Mandatory Action**: (e.g., `STRIP`, `VAULT`, `REDACT`).

### Step 2 – JSON Schema Mapping
1.  Load the current `regulatory_policy.json`.
2.  Map the extracted requirements to the `mappings` object.
    - If the category already exists, update its `description` or `regulation` metadata.
    - If it's new, append it as a new key in the `mappings` dictionary.

### Step 3 – Versioning & Metadata
1.  Read the `version` field from the root of the JSON.
2.  **Increment Logic**:
    - Major version if the `default_posture` changes.
    - Minor version if only `mappings` are added/updated.
3.  Ensure the `default_posture` remains compliant with the CISO's guidelines (default: `FAIL_CLOSED`).

### Step 4 – Validation & Atomic Write
1.  **Safety Gate**: Run `jq '.'` on the generated JSON to ensure it is valid.
2.  **Schema Check**: Verify that every mapping has a `regulation` and `description` field.
3.  Write the updated policy to the `policy_file`.

---

## Failure Handling

- **Ambiguous Digest**: If the input requirements are unclear, halt and request a "Clarification Artifact" from the `CISO & Compliance Officer`.
- **Schema Violation**: If the generated JSON lacks required root fields (`version`, `mappings`), discard changes and report the error.

---

## Ecosystem Role
- **Category**: Transformer / Orchestrator.
- **Dependencies**: `CISO & Compliance Officer` (Knowledge source), `drift-detector` (Validation).
- **Triggers**: Regulatory update event.
