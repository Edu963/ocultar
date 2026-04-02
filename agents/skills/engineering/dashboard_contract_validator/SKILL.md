---
name: dashboard-contract-validator
description: Ensures the JSON schema returned by Ocultar API endpoints matches the expected input of the React dashboard components.
---

# Dashboard Contract Validator

## Purpose
This skill prevents breaking changes in the API from reaching the dashboard. It validates the "Contract" between the Go backend and the Vite/React frontend, focusing on the `/api/refine`, `/api/metrics`, and `/api/roi` endpoints.

## Role
- **Category**: Integration Validator
- **Phase**: CI/CD / Pre-Release
- **Ecosystem Fit**: Acts as a bridge between `services/refinery` and `apps/dashboard`.

## Inputs / Outputs

### Inputs
| Name | Type | Description |
|---|---|---|
| `api_endpoint` | string | The endpoint to test (e.g., `/api/metrics`). |
| `expected_schema` | JSON | The JSON Schema that the endpoint must satisfy. |
| `license_tier` | enum | `community` \| `enterprise`. |

### Outputs
| Name | Type | Description |
|---|---|---|
| `contract_valid` | boolean | `true` if the API response matches the schema. |
| `breaking_changes` | array | List of missing or mismatched fields. |

## Preconditions
- The Refinery Refinery must be running in a test environment.
- The `Refinery-Rule-Generator` should have generated a stable set of test PII hits.

## Instructions

### 1. Validate Metrics Schema
- **Action**: Call `GET /api/metrics`.
- **Check**: Confirm the presence and type of:
    - [ ] `vault_entries` (integer)
    - [ ] `slm_status` (string: "Online"|"Offline")
    - [ ] `queue_depth` (integer)
    - [ ] `slm_host` (string: URL)

### 2. Validate ROI Schema (Enterprise Only)
- **Action**: Call `GET /api/roi` with an Enterprise license.
- **Check**: 
    - [ ] `status` (string: "online")
    - [ ] `savings` (number)
- **Validation**: If license is missing, verify the API returns `403 Forbidden`.

### 3. Validate Configuration Schema
- **Action**: Call `GET /api/config/regex`.
- **Check**: Verify the list of customer-defined PII types.
- **Action**: Call `POST /api/refine` with a sample JSON payload.
- **Check**:
    - [ ] `refined` (string)
    - [ ] `report` (object containing `pii_hits`, `total_pii_count`)

## Failure Handling
- **SCHEMA_VIOLATION**: If a field type changes (e.g., `savings` becomes a string), log a `BLOCKING_CONTRACT_FAILURE` and halt the release pipeline.
- **UNEXPECTED_FIELD**: If new fields are added without documentation, log a `STRICT_SCHEMA_WARNING`.
