---
name: enterprise-dashboard-integrity-checker
description: Production-grade validator for the Ocultar Enterprise Dashboard. Ensures UI consistency and regulatory alignment.
---

# Enterprise Dashboard Integrity Checker (v1.1)

## Purpose

The EDIC ensures that the visual representation of compliance (the Dashboard) is technically accurate. It prevents "Green-Washing" by verifying that UI metrics are directly derived from the Refinery's state.

## Inputs / Outputs

### Inputs
- `dashboard_url`: Endpoint for verification.
- `active_license`: Verification token.
- `api_truth_endpoint`: Sombra metrics API.

### Outputs
- `integrity_status`: `VERIFIED` | `STALE_DATA` | `LICENSE_BYPASS`.
- `latency_spans`: P50/P99 UI response times.

---

## Instructions

### 1. Licensed Feature Lockdown
- Verify that "Enterprise-Only" tabs are gated by the `active_license`.
- **Violation**: If Tier 0 metrics are visible on a `COMMUNITY` tier, trigger an instant security incident.

### 2. The "Metrics Truth" Check
- Query the `api_truth_endpoint` for a specific count (e.g., `total_redact_events`).
- Compare against the number displayed in the Dashboard UI.
- **Drift**: If delta > 1%, mark as `STALE_DATA`.

### 3. Visual Regulatory Alignment
- Ensure that `GDPR` or `HIPAA` tiles only illuminate when their corresponding PII categories (from `regulatory_policy.json`) have been detected in the last 24h.

## Failure Handling
- **`UI_TIMING_DELAY`**: If the React app hasn't refreshed the state, wait 5s and re-verify before flagging drift.

## Postconditions
- All integrity results must be signed by `evidence-archiver`.
