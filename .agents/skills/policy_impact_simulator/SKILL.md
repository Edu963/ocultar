---
name: policy-impact-simulator
description: Quality Assurance (QA) Orchestrator for regulatory policy changes. Replays historical audit logs against a proposed "Draft Policy" to calculate the delta in blocked/sanitized requests.
---

# Policy Impact Simulator (v1.0)

## Purpose

The Policy Impact Simulator provides "Safety-First" validation. It predicts the results of a policy change *before* it hits the Sombra Gateway. By replaying the last N requests from the `audit.log` against the new `regulatory_policy.json`, it identifies potential breaking changes (e.g., blocking a mission-critical endpoint).

## Preconditions

- **Input**: A `draft_policy` (JSON) and access to `audit.log` (JSONL).
- **Tools**: `jq` and `cat`.

## Inputs / Outputs

### Inputs
- `draft_policy` (Path): Path to the proposed `regulatory_policy.json`.
- `sample_size` (Int): Number of historical events to replay (default: `1000`).
- `audit_log` (Path): Path to the system audit records (default: `logs/audit.log`).

### Outputs
- `simulation_report` (Artifact): A structured report containing:
    - `total_requests`: Number of events replayed.
    - `delta_blocked`: Number of *newly* blocked requests compared to the current policy.
    - `delta_sanitized`: Number of *newly* sanitized requests.
    - `risk_score`: `LOW` | `MEDIUM` | `HIGH` (based on delta percentage).

---

## Instructions

### Step 1 – Sample Extraction
1.  Read the last `sample_size` entries from `audit_log`.
2.  Extract the `hit_category` and `raw_value` (if available) from each entry.

### Step 2 – Draft Policy Replay
1.  For each extracted entry, evaluate the `hit_category` against the `mappings` in the `draft_policy`.
2.  **Simulation Rules**:
    - If the current policy allowed a hit, but the `draft_policy` would `BLOCK` it, log as a **New Block**.
    - If the `draft_policy` mandates a different `REDACT` or `STRIP` behavior, log as a **Metadata Change**.

### Step 3 – Risk Calculation
1.  Calculate the `impact_ratio`: `(New Blocks + New Sanitizations) / Total Requests`.
2.  Assign `risk_score`:
    - `HIGH`: > 5% of total requests are newly blocked.
    - `MEDIUM`: > 1% change in behavior.
    - `LOW`: < 1% impact.

### Step 4 – Visual Delta Generation
1.  Produce a human-readable comparison showing specific examples of "Before vs. After" for the most frequent hits.

---

## Failure Handling

- **`INSUFFICIENT_DATA`**: If `audit_log` has fewer than 100 entries.
- **`MALFORMED_DRAFT`**: If the `draft_policy` fails `jq` validation.

---

## Ecosystem Role
- **Role**: Validator / Simulation Agent.
- **Consumer**: `continuous-ai-orchestrator` (Before deployment).
- **Dependency**: `audit-log-validator`.
