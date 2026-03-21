---
name: roi-cost-efficiency-accountant
description: Analyzes token usage and calculates the financial impact of Ocultar’s security and cost-avoidance.
---

# ROI & Cost-Efficiency Accountant (v1.1)

## Purpose

The REA quantifies the monetary value of Ocultar. It measures token efficiency and potential fine-avoidance, allowing the CISO to justify security budgets through technical data.

## Inputs / Outputs

### Inputs
- `audit_logs`: Path to `app/audit.log` or SIEM stream.
- `pricing_model`: JSON mapping categories to risk-dollar values.
- `timeframe`: `1d`, `7d`, `30d`, `quarterly`.

### Outputs
- `roi_report` (Artifact): Detailed breakdown of savings and protected capital.
- `optimization_tips`: Specific changes to reduce token spend.

---

## Instructions

### 1. Fine-Avoidance Mapping
- Identify all `REDACT` or `VAULT` hits in `audit_logs`.
- **Arithmetic**: `Sum(hits * penalty_per_category)`.
- **Default Penalty**: Minor ($5k), Major ($100k - Tier 0).

### 2. Token Optimization Audit
- Measure the ratio of SLM (Local) vs LLM (Remote) calls.
- **Savings Logic**: Calculate `$ Total_Saved` by using local refinery vs. passing raw PII through remote provider filters (which often reject or over-redact).

### 3. Efficiency Reporting
- Generate a "Cost-per-Million-Tokens" (CPMT) metric.
- **Goal**: Highlight how the Enterprise Site License reduces CPMT compared to individual API keys.

## Failure Handling
- **`LOG_DATA_SPARSITY`**: If less than 1,000 requests are logged, provide a "Projected ROI" based on industry averages.

## Postconditions
- Data MUST be aggregated and sanitized (No PII in ROI reports).
- Results stored in `services/engine/analytics/roi/`.
