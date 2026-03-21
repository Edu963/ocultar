---
name: roi-cost-efficiency-accountant
description: Analyzes token usage across the Sombra Gateway and calculates the financial impact of Ocultar’s AI security, including potential fines avoided.
---

# ROI & Cost-Efficiency Accountant

## Purpose

This skill provides the data-driven justification for AI security investments. It transforms technical logs into financial metrics (ROI, Cost-Avoidance) that C-suite executives can use to measure the value of the Ocultar platform.

## When To Use This Skill

Use this skill:
- Monthly or quarterly to generate compliance and budget reports.
- When evaluating the performance of different LLM/SLM providers.
- During infrastructure reviews to identify cost-saving opportunities in the refinement pipeline.
- When justifying the "Security Tax" of AI protection to the CFO or CISO.

## Instructions

1.  **Extract Usage Metrics**: Aggregate token counts for both external LLM requests and internal SLM refinement operations.
2.  **Calculate Fine-Avoidance**: Map every blocked PII leak or Dictionary Shield hit to a potential regulatory penalty (e.g., $4k per breach for minor GDPR violations).
3.  **Perform Cost Benchmarking**: Compare the cost of using a high-end LLM for simple tasks versus routing them through cheaper, specialized SLMs (e.g., using a local regex/SLM vs. a remote GPT-4 call for PII detection).
4.  **Identify Resource Leaks**: Detect inefficient prompts or redundant refinement loops that are inflating token costs.
5.  **Generate Executive Summary**: Produce a high-level report showing:
    - **Total Tokens Refined**.
    - **Security ROI** ($ Fines Avoided / $ Ocultar Cost).
    - **Optimization Savings** ($ Saved by switching to cheaper SLMs).
6.  **Maintain Privacy**: Ensure that all cost reports use aggregated data and never reveal the content of the refined PII.

## Examples

### Quarterly ROI Report
**Action**: Aggregate 12 million refined tokens and 500 blocked leaks. Calculate $2M in potential fine-avoidance based on vertical-specific penalties.

### Model Swap Recommendation
**Change**: Discovered that 80% of PII detection is being handled by a remote LLM.
**Action**: Recommend switching to the local "Refinery Rule" engine to reduce token latency and cost by 95%.
