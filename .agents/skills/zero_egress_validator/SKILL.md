---
name: zero-egress-validator
description: Specialized security linter to detect unmasked PII leakage and unauthorized external network calls.
---

# Zero-Egress Validator

## Purpose
Enforces the "Platinum Rule": No unmasked PII ever leaves the client's VPC.

## Inputs / Outputs

### Inputs:
- `change_set` (Git Diff): The code diff to evaluate.

### Outputs:
- `security_verdict` (Enum): [PASS | FAIL]
- `leak_report` (Artifact): Detailed breakdown of potential egress points.

## Instructions

### Step 1 – Network Call Detection
- Search the `change_set` for external communication patterns:
    - JavaScript: `fetch(`, `axios(`, `https.request(`
    - Go: `http.Post(`, `http.Get(`, `url.Parse(`
    - Python: `requests.`, `urllib.`

### Step 2 – Data Payload Audit
- For every network call identified in Step 1, verify if the data passed is:
    - Redacted/Anonymized
    - Internal-only (e.g., local health checks)
- **Gate**: If unmasked raw data from `services/engine` or `services/vault` is passed to an external URI, the verdict is **FAIL**.

### Step 3 – Verdict Generation
- Issue a `security_verdict`.
- **Action**: If FAIL, provide specific remediation (e.g., "Use Sombra's redaction hook before calling this API").

## Failure Handling
- A FAIL verdict **MUST** halt any deployment or release process.
