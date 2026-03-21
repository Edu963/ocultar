---
name: enterprise-dashboard-integrity-checker
description: Production-grade validator for the Ocultar Enterprise Dashboard. Ensures UI consistency, regulatory alignment, and license-tier enforcement.
---

# Enterprise Dashboard Integrity Checker

## Purpose
This skill ensures the **Ocultar Enterprise Dashboard** (React/Vite) remains a high-fidelity representation of the underlying **Refinery Engine** and **Audit Vault**. It prevents "Visual Compliance" drift where the UI misrepresents regulatory risk or ROI metrics.

## Role
- **Category**: Validator / Integrity Checker
- **Phase**: Verification / Post-Deployment
- **Ecosystem Fit**: Connects the `services/engine` API state to the `apps/web` visual presentation.

## Inputs / Outputs

### Inputs
| Name | Type | Description |
|---|---|---|
| `dashboard_url` | string | Base URL of the running dashboard (e.g., `http://localhost:9091`). |
| `license_tier` | string | Expected license state (`community` or `enterprise`). |
| `regulatory_policy` | path | Path to `security/regulatory_policy.json`. |
| `vault_count` | integer | (Optional) Known entry count in the active vault. |

### Outputs
| Name | Type | Description |
|---|---|---|
| `integrity_report` | JSON | Results of connectivity, schema, and visual alignment checks. |
| `p50_latency` | string | Measured frontend response time for core metrics. |
| `drift_detected` | boolean | Set to `true` if UI metrics mismatch backend truth. |

## Preconditions
- The Ocultar Engine must be running with the `--serve` flag.
- A valid `OCU_LICENSE_KEY` must be active for Enterprise-tier validation.
- The `dist/enterprise/dashboard` assets must be properly built and served.

## Instructions

### 1. Validate License-Tier Enforcement
- **Action**: Access `/api/status`.
- **Validation**: If `tier` is `enterprise`, confirm the following UI elements are visible and unlocked:
    - [ ] **Risk Matrix** (Regulatory breakdown)
    - [ ] **SIEM Log Stream** (Audit tail)
    - [ ] **PostgreSQL Vault Metrics**
- **Action**: If `tier` is `community`, verify that Enterprise features are either hidden or present a "Upgrade Required" barrier on the `/` route.

### 2. Verify Regulatory Alignment (Risk Matrix)
- **Action**: Cross-reference the `PII_TYPE` mapping in the UI with `services/engine/pkg/engine/engine.go:getComplianceMapping`.
- **Validation**:
    - [ ] `EMAIL` or `PHONE` hits must map to **GDPR** tiles.
    - [ ] `PATIENT_ID` or `MEDICAL_RECORD` hits must map to **HIPAA** tiles.
    - [ ] Ensure the "Risk Score" correctly reflects the **K-Anonymity** logic defined in `pkg/audit/risk.go`.

### 3. Match ROI Metrics (The "Truth" Check)
- **Action**: Compare the "Annual Savings" displayed in the internal Enterprise Dashboard against the `/api/roi` response.
- **Validation**: 
    - [ ] Internal Dashboard Savings must NOT be hardcoded.
    - [ ] `savings = vault_entries * multiplicative_constant`.
    - [ ] Confirm that the public `roi_calc.html` remains static (Manual Entry) to maintain Zero-Egress optics.

### 4. Navigation & Link Integrity
- **Action**: Crawl the sidebar and breadcrumbs to ensure zero broken links.
- **Validation**:
    - [ ] `DOCS` links to `https://ocultar.dev/docs`.
    - [ ] `ROOT` correctly renders `dist/enterprise/dashboard/index.html`.

## Failure Handling
- **CONNECTION_FAILURE**: If `dashboard_url` is unreachable, retry 3 times then log `FATAL_SYSTEM_UNAVAILABLE`.
- **SCHEMA_MISMATCH**: If `/api/metrics` returns fields not recognized by the current React build, log `VERSION_DRIFT_DETECTED`.
- **LICENSE_BYPASS**: If Enterprise metrics are visible in Community mode, immediately trigger a `SECURITY_VIOLATION` alert.

## Examples

### Validating GDPR Tile Consistency
**Action**: Inject a test payload containing `alice.martin@example.com` via `/api/refine` and confirm the `GDPR` risk tile in the dashboard increments its hit count while the `Identity Vault` gauge reflects a new entry.
```
