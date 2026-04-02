---
name: dictionary_shield_manager
description: Production-grade orchestrator for the "Tier 0" Dictionary Shield. Manages protected entity lists (VIPs, Projects, Internal Assets) with deterministic sync and validation.
---

# Dictionary Shield Manager (v1.1)

## Purpose

The DSM manages the "Platinum List" of protected entities. It ensures that critical strings are synchronized across all refinery nodes with 100% parity, using the **Operational Dashboard's Shield Manager** as the primary control interface.

## Inputs / Outputs

### Inputs
- `action` (Enum): `ADD` | `SYNC` | `AUDIT`.
- `entities` (List): New strings for protection.
- `enforce_parity` (Boolean): Default `TRUE`.

### Outputs
- `sync_verdict`: `SUCCESS` | `PARTIAL_SYNC` | `ROLLBACK_TRIGGERED`.
- `parity_hash`: SHA-256 of the unified `protected_entities.json`.

## Preconditions
- Access to all three config paths (Sombra, Gateway, Refinery).

---

## Instructions

### 1. Atomic Add & Sync
- **Staging**: Update `configs/config.yaml` (dictionaries block).
- **Propagation**: Hot-reload memory structures via `POST /api/config/dictionary`.
- **Validation**: Verify entity presence in the **Shield Manager** UI.
- **Integrity**: If sync fails, revert `configs/config.yaml` to last-known-good.

### 2. Fail-Safe Rollback
- If Re-sync fails:
    - **Action**: Restore all paths from `protected_entities.json.bak`.
    - **Verdict**: Return `ROLLBACK_TRIGGERED` and signal `continuous-ai-orchestrator`.

### 3. Entity Pruning (Auditor)
- Flag any entry shorter than 3 characters (Noise risk).
- Compare hits in `audit.log` vs Dictionary entries.
- **Action**: Suggest removal of zero-hit entries after 30 days to maintain performance.

## Failure Handling
- **`SYNC_CONFLICT`**: If a file is read-only or locked, abort the sync and retain the last-known-good (LKG).
- **`JSON_ERROR`**: If the provided entity list is malformed, reject the `ADD` action.

## Postconditions
- All refinery instances MUST be notified to reload the dictionary post-sync.
