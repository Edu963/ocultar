---
name: sombra-performance-benchmarker
description: Monitor latency overhead introduced by the Refinery and optimize the execution pipeline, ensuring AI requests feel fast while preserving security.
---

# Sombra Performance Benchmarker (v1.1)

## Purpose

The SPB ensures that "Security Latency" stays within acceptable human-interactive boundaries. It automates the measurement of the security "tax" across all detection tiers (0, 1, 2) and flags performance regressions.

## Inputs / Outputs

### Inputs
- `latency_threshold_ms`: Max acceptable overhead (Default: `100ms` for Tier 0/1, `5000ms` for Tier 2).
- `sample_payload`: Path to a representative AI request.
- `tier_depth`: `0` (Dict), `1` (Regex), `2` (SLM).

### Outputs
- `bench_report` (Artifact): Per-tier latency breakdown.
- `performance_verdict`: `OPTIMAL` | `DEGRADED` | `SHALLOW_BYPASS_RECOMMENDED`.
- `optimization_patch`: Suggested rule modifications to reduce overhead.

## Preconditions
- Sombra Gateway must be running in `PROFILING_MODE`.

---

## Instructions

### 1. Instrumentation & Sampling
- Execute the `sample_payload` through the gateway with `X-Ocultar-Profiler: true`.
- Collect high-fidelity spans for:
    - **Tier 0 (Dictionary)**: Constant time $O(1)$ lookup.
    - **Tier 1 (Regex)**: Pattern matching time.
    - **Tier 2 (SLM)**: Inference time.
    - **Queue Delay**: Duration spent in the Hardened Concurrency Queue.

### 2. Threshold Violation Analysis
- If total latency > `latency_threshold_ms`:
    - **Action**: Identify the "Heavy Rule" with the highest span value.
    - **Refactor**: Suggest pre-compilation, dictionary-migration, or rule-pruning.

### 3. Verdict Generation
- `DEGRADED`: Latency is > threshold but within 2x limit.
- `SHALLOW_BYPASS_RECOMMENDED`: If security recall is high but latency is prohibitive, recommend moving rules from Tier 2 to Tier 1.

## Failure Handling
- **`COLD_START_BIAS`**: If the first request is slow due to cache Misses, perform 5 warm-up runs before measuring.

## Postconditions
- Artifact: Final performance snapshot must be saved to `services/refinery/bench/results.json`.
