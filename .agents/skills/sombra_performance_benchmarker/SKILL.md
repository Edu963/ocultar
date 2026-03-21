---
name: sombra-performance-benchmarker
description: Monitor latency overhead introduced by the Refinery and optimize the execution pipeline, ensuring AI requests feel fast while preserving security.
---

# Sombra Performance Benchmarker

## Purpose

This skill ensures that the security layers (Refinery, Dictionary Shield) do not become a bottleneck for real-time AI interactions. It provides deep visibility into the latency "tax" of security and identifies optimization targets.

## When To Use This Skill

Use this skill:
- During load testing or performance profiling of the Sombra Gateway.
- After adding complex SLM-based refinement rules.
- When users report "sluggish" AI responses.
- To validate that engine optimizations (e.g., regex caching) are effective.

## Instructions

1.  **Measure Pipeline Latency**: Instrument the request flow to track milliseconds spent in:
    - Tier 0: Dictionary Shield lookup.
    - Tier 1: Regex/Rule-based PII detection.
    - Tier 2: SLM-based refinement.
    - Gateway Overhead: Routing and TLS termination.
2.  **Identify Bottlenecks**: Pinpoint "Heavy Rules" (e.g., inefficient regex or slow SLM prompts) that contribute disproportionately to latency.
3.  **Validate Optimizations**: Compare "Before vs. After" metrics for engine patches.
4.  **Balance Security/Speed**: Suggest using "Fast-Track" rules for non-sensitive traffic while maintaining "Deep-Scan" for critical PII.
5.  **Generate Periodic Reports**: Produce performance snapshots for DevOps teams and clients to ensure transparency on AI responsiveness.

## Examples

### Regex Optimization
**Bottleneck**: A complex pattern for global banking codes is adding 50ms per request.
**Action**: Suggest replacing the regex with a pre-compiled Aho-Corasick dictionary if possible, or optimizing the pattern structure.

### SLM Latency Audit
**Bottleneck**: Tier 2 refinement is consistently taking >500ms.
**Action**: Recommend switching to a smaller, quantized SLM model or optimizing the refinement prompt for faster inference.
