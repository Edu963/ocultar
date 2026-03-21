---
name: ecosystem-state-tracker
description: A partitioned, dependency-aware metadata orchestrator for the Ocultar AI ecosystem. Prevents redundant skill executions by tracking deep hashes of both inputs and environmental dependencies.
---

# Ecosystem State Tracker (v2.1)

## Purpose

EST is the high-fidelity state store for the Ocultar multi-agent system. It maintains deterministic records of skill executions to optimize performance, prevent redundant security scans, and ensure consistency.

## Inputs / Outputs

### Inputs
- `execution_context`:
    - `skill_id`: Unique identifier (e.g., `secret-scanner`).
    - `input_hash`: Pre-calculated SHA-256 of all input parameters.
    - `dependency_manifest`: List of absolute paths to files influencing the outcome.
- `update_payload` (Post-execution):
    - `receipt`: Signed proof of execution result (`SUCCESS` | `FAIL`).
    - `artifacts`: List of generated absolute paths.

### Outputs
- `cache_status`: `HIT` | `MISS`.
- `receipt`: The stored receipt if `cache_status` == `HIT`.

## Preconditions
- Filesystem MUST support atomic `rename` operations.
- `skill_id` MUST be registered in `ttl_manifest.json`.

---

## Instructions

### 1. Verification (Redundancy Check)
1. **Compute State ID**: Combine `input_hash` and the Merkle-hash of the `dependency_manifest`.
2. **TTL Validation**:
   - Lookup `.agents/state/{{skill_id}}/{{state_id}}.json`.
   - Verify `expires_at` > `now()`.
   - If valid, return `cache_status = HIT`.

### 2. State Committal
1. **Atomic Write**:
   - Serialize `status`, `receipt`, and `artifacts`.
   - Calculate `expires_at` using the `ttl_manifest.json` policy.
   - Write to `.tmp` and rename to `.json`.
2. **Audit Push**:
   - If in `STRICT_AUDIT` mode, broadcast the state hash to the `evidence-archiver`.

### 3. Lifecycle Management
- **Pruning**: Execute `Prune()` on any partition exceeding 100MB.
- **Verification**: Periodically re-validate artifact hashes on disk against the stored manifest in EST.

## Failure Handling
- **Collision**: If a write collision occurs, the filesystem rename ensures only one state persists.
- **Corrupt Cache**: On parse error, immediately delete the JSON and return `MISS`.

## Constraints
- **Absolute Paths**: No relative pathing permitted.
- **Signatures**: Receipts MUST be signed by the executing agent’s Ed25519 key.
