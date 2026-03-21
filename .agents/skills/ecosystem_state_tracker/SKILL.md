---
name: ecosystem-state-tracker
description: A centralized metadata store to prevent redundant skill executions across different agents. Maintains a high-fidelity audit trail of all skill outcomes and system-wide state transitions.
---

# Ecosystem State Tracker

## Purpose
This skill acts as the "Memory" of the Ocultar AI agent ecosystem. It records the status, inputs, and outputs of every skill executed. This prevents expensive or redundant tasks (like re-scanning a file that hasn't changed) and provides a single source of truth for the `continuous-ai-orchestrator`.

## Inputs
- `execution_context`: Metadata about the current task (AgentID, SkillID, Timestamp).
- `skill_result`: SUCCESS/FAIL/WARNING and any generated artifacts.
- `state_query`: (Optional) A query to retrieve the last known state of a specific component or skill.

## Outputs
- `state_report`: The updated or retrieved state manifest.
- `redundancy_check`: Boolean (TRUE if the task can be skipped based on previous successful execution).

---

## Instructions

### 1. Register Execution
- On every skill call, record the **SHA-256 hash** of the inputs (e.g., `git diff` hash).
- Store the starting timestamp and the requesting agent's identity.

### 2. Check for Redundancy
- Before a skill starts, compare the current input hash against the `state_manifest.json` located in `.agents/state/`.
- **Condition**: If the hash matches a previous `SUCCESS` result within the last 15 minutes, flag `redundancy_check = TRUE`.

### 3. Update State Manifest
- Once a skill completes, append the result and any output artifact paths to the manifest.
- Ensure the manifest is cryptographically signed if in `COMPLIANCE_MODE`.

### 4. Garbage Collection
- Prune log entries older than 24 hours to maintain performance within the `state_manifest.json`.

## Failure Handling
- **Manifest Lock**: If `state_manifest.json` is locked by another process, wait 500ms and retry.
- **Corrupt Manifest**: If the JSON is invalid, reset the manifest and log a "State Loss" warning.

## Examples

### Scenario: Redundant Sanitization
- **Input**: `security-sanitizer` called twice on the same `dist/` bundle.
- **Process**: `Ecosystem-State-Tracker` identifies the matching hash.
- **Result**: `redundancy_check = TRUE`, allowing the orchestrator to skip the second scan.
