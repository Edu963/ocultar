---
name: ecosystem-state-tracker
description: A partitioned, dependency-aware metadata orchestrator for the Ocultar AI ecosystem. Prevents redundant skill executions by tracking deep hashes of both inputs and environmental dependencies.
---

# Ecosystem State Tracker (EST)

## Purpose
EST serves as the high-fidelity state store for the Ocultar multi-agent system. It maintains a deterministic record of skill executions to optimize performance, prevent redundant security scans, and ensure consistency across the agent lifecycle.

## Inputs
- `execution_context`:
    - `skill_id`: The unique identifier of the calling skill (e.g., `security-sanitizer`).
    - `agent_id`: Identifier of the agent requesting the state check.
    - `input_parameters`: Key-value map of all direct arguments passed to the skill.
    - `dependencies`: List of file paths or system configurations that influence the skill's outcome.
- `update_payload`: (Post-execution only)
    - `status`: SUCCESS | FAIL | WARNING.
    - `artifacts`: List of absolute paths to generated files.
    - `metadata`: Any additional context (e.g., lines of code scanned, vulnerabilities found).

## Outputs
- `redundancy_check`:
    - `is_redundant`: Boolean.
    - `cached_result`: The previously stored `skill_result` if `is_redundant` is TRUE.
    - `last_executed_at`: Timestamp of the cached execution.
- `state_receipt`: Confirmation of the registered state update.

---

## Instructions

### 1. Verification of Redundancy
Before executing any resource-intensive skill:
1.  **Compute Deep Hash**:
    - Generate a SHA-256 hash of all `input_parameters`.
    - Generate a Merkle-tree style recursive hash of all files listed in `dependencies`.
    - Combine these into a single `state_hash`.
2.  **Partition Lookup**:
    - Navigate to `.agents/state/{{skill_id}}/{{state_hash}}.json`.
3.  **Validate TTL**:
    - If the file exists, check the `expires_at` field. If not expired AND `status` was `SUCCESS`, return `is_redundant = TRUE`.

### 2. Registering State Updates
Upon skill completion:
1.  **Initialize Partition**: Create the directory `.agents/state/{{skill_id}}/` if it does not exist.
2.  **Atomic Write**:
    - Build the state JSON including `status`, `artifacts`, `state_hash`, and a calculated `expires_at` (based on the skill's specific TTL policy).
    - Write to a temporary file: `.agents/state/{{skill_id}}/{{state_hash}}.tmp`.
    - Atomically rename `.tmp` to `.json` to ensure integrity.
3.  **Audit Trail Integration**: If in `COMPLIANCE` mode, send a hash of this state record to the `compliance-integrity-suite`.

### 3. Cleanup & Pruning
- EST does not maintain a giant manifest. Individual files should be pruned by an external `maintenance-agent` or via a `garbage_collection` trigger if a partition exceeds a specific size (e.g., 50MB).

## Failure Handling
- **Missing Dependencies**: If a path in `dependencies` is unreachable, fail the redundancy check and force a skill re-run.
- **State Collision**: If two agents attempt to write the same hash simultaneously, use the atomic `rename` operation to let the filesystem resolve the race condition.

## Constraints
- **Absolute Paths Only**: All artifact and dependency paths must be absolute.
- **Fail-Safe**: If any error occurs during state lookup (e.g., permissions, corruption), EST MUST return `is_redundant = FALSE` to ensure the skill runs and system integrity is preserved.
