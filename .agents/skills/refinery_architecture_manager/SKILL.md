---
name: refinery-architecture-manager
description: Master architect for the Ocultar Engine's detection layer. Manages the end-to-end lifecycle of PII detection rules and architectural integrity.
---

# Refinery Architecture Manager

## Purpose
The Ocultar Engine's detection layer requires a unified, proactive architectural oversight. This skill merges rule generation (proactive) and detection updates (reactive) into a single, high-fidelity lifecycle manager. It ensures that "Refinement" is the first and most robust line of defense.

## Role
- **Category**: Architect / Security Maintainer
- **Ecosystem Positioning**: Core Engine / Sombra Gateway
- **Primary Goal**: 100% Recall for sensitive entities with minimal performance overhead.

## Responsibilities

### 1. Rule Lifecycle Management
- **Identification**: Analyze `audit.log` and "Refinement" failures to discover new PII types or bypass techniques.
- **Generation**: Propose optimized regex patterns and high-entropy dictionary strings for the Ocultar Engine.
- **Integration**: Update `protected_entities.json` and `PII_DETECTION.md`.
- **Validation**: Trigger the `pii-regression-suite-runner` to verify that new rules do not break existing detection.

### 2. Architectural Integrity (Refinery-First)
- **Gatekeeper**: Ensure all data flows through the Ocultar Engine before LLM egress.
- **Policy Audit**: Regularly review the `strip_categories` policy in `sombra.yaml` to prioritize "Removal" over "Pseudonymization" for Tier 0 data (SSN, Passwords).
- **Scanner Optimization**: Maintain SHA-256 keying standards and ensure the AI-based NER fallback is only used when deterministic rules fail.

### 3. Edge Case Mitigation
- Generate specific regex patterns instead of broad "catch-all" rules.
- Balance security (recall) against system latency.

## Inputs & Dependencies
- **Inputs**: `audit.log`, user-reported leaks, new regulatory requirements.
- **Dependencies**: `pii-regression-suite-runner`, `dictionary-shield-manager`.

## Failure Handling
- If a new rule causes a detection regression (>5%), revert the `regulatory_policy.json` and refine the rule logic.
