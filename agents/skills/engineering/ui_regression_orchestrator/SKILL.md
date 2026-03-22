---
name: ui-regression-orchestrator
description: Runs headless browser tests to verify the "Zero-Egress" visual boundaries and UI integrity of the Ocultar Enterprise Dashboard.
---

# UI Regression Orchestrator

## Purpose
This skill automates the visual and functional verification of the Ocultar dashboard. It ensures that the "Zero-Egress Zone" and "Risk Matrix" are not just functionally correct but visually accurate for the end-user.

## Role
- **Category**: Regression Tester
- **Phase**: Pre-Release / QA
- **Ecosystem Fit**: High-level orchestrator for frontend stability.

## Inputs / Outputs

### Inputs
| Name | Type | Description |
|---|---|---|
| `dashboard_url` | string | URL of the dashboard instance to test. |
| `browser_type` | enum | `chromium` \| `firefox` \| `webkit`. |
| `test_suite` | path | Path to the visual regression test scripts. |

### Outputs
| Name | Type | Description |
|---|---|---|
| `visual_diffs` | array | List of screenshots showing regressions. |
| `regression_status` | string | `PASS` \| `FAIL` \| `WARN`. |

## Preconditions
- The `Enterprise-Dashboard-Integrity-Checker` must have passed API validation.
- The `Manage-Ocultar-License` skill must have provisioned a test license.

## Instructions

### 1. Zero-Egress Boundary Verification
- **Action**: Launch the browser and navigate to the dashboard.
- **Check**: Verify that the "Zero-Egress Zone" (Particles/Canvas) is rendering without errors and that the dashed boundary is visible.

### 2. Risk Matrix Interactivity
- **Action**: Click on a regulatory tile (e.g., GDPR).
- **Check**: Confirm that the audit logs are correctly filtered to show only GDPR-related PII hits.

### 3. ROI Calculator Alignment
- **Action**: Adjust the sliders in the ROI module.
- **Check**: Verify that the derived savings match the recalculated values from the `/api/roi` endpoint (via console/network check).

## Failure Handling
- **UI_RENDER_FAILURE**: If core components (Risk Matrix, Gauges) fail to render, log a `CRITICAL_UI_REGRESSION`.
- **DATA_DRIFT**: If the UI shows a "Pass" but the backend returns a "Fail" for the same dataset, log a `COMPLIANCE_VISUAL_MISMATCH`.
