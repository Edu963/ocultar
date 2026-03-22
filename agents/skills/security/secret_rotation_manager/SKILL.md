---
name: secret-rotation-manager
description: Remediation expert for Ocultar infrastructure secrets. Automatically rotates cryptographically secure keys and salts to eliminate security drift.
---

# Secret Rotation Manager (v1.1)

## Purpose

SRM mitigates "Root-of-Trust Drift". It ensures that Ocultar's foundational secrets (Master Key, Salt) are rotated automatically according to policy, without manual intervention or plaintext exposure.

## Inputs / Outputs

### Inputs
- `target_files`: YAML or ENV files to update.
- `threshold_days`: Default 90.
- `trigger` (Enum): `DRIFT` | `SCHEDULE` | `BREACH`.

### Outputs
- `rotation_id`: Unique ID for the event.
- `checkpoint_hash`: SHA-256 of the new configuration.
- `verification_status`: `SUCCESS` | `FAILURE`.

## Preconditions
- Root/Sudo permission on the target directory.
- `openssl` installed.

---

## Instructions

### 1. Entropy Synthesis
- Generate `OCU_MASTER_KEY`: `openssl rand -hex 32`.
- Generate `OCU_SALT`: `openssl rand -hex 8`.

### 2. Safe Update Phase
- Create backup: `{{file}}.bak.{{timestamp}}`.
- Perform atomic string replacement (Sed).
- **Integrity Gate**: Verify the file remains a valid YAML/ENV object post-edit.

### 3. Deployment & Verification
- Signal the Sombra Gateway to reload.
- **Verification**: Query the `/health/security` endpoint. 
- Compare the reported `key_id_hash` from the live service against the calculated `checkpoint_hash`.
- **Verdict**: If they do not match, REVERT immediately.

## Failure Handling
- **`LOCK_CONFLICT`**: If the vault is currently being accessed, WAIT for clear state.
- **`RELOAD_FAILURE`**: If the Sombra Gateway fails to restart, restore the backup and alert PILOT_OPS.

## Postconditions
- Rotation event MUST be signed by `evidence-archiver`.
